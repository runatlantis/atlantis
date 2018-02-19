package events

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/run"
	"github.com/runatlantis/atlantis/server/events/terraform"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_pre_executor.go ProjectPreExecutor

// ProjectPreExecutor executes before the plan and apply executors. It handles
// the setup tasks that are common to both plan and apply.
type ProjectPreExecutor interface {
	// Execute executes the pre plan/apply tasks.
	Execute(ctx *CommandContext, repoDir string, project models.Project) PreExecuteResult
}

// DefaultProjectPreExecutor implements ProjectPreExecutor.
type DefaultProjectPreExecutor struct {
	Locker       locking.Locker
	ConfigReader ProjectConfigReader
	Terraform    terraform.Client
	Run          run.Runner
}

// PreExecuteResult is the result of running the pre execute.
type PreExecuteResult struct {
	ProjectResult    ProjectResult
	ProjectConfig    ProjectConfig
	TerraformVersion *version.Version
	LockResponse     locking.TryLockResponse
}

// Execute executes the pre plan/apply tasks.
func (p *DefaultProjectPreExecutor) Execute(ctx *CommandContext, repoDir string, project models.Project) PreExecuteResult {
	workspace := ctx.Command.Workspace
	lockAttempt, err := p.Locker.TryLock(project, workspace, ctx.Pull, ctx.User)
	if err != nil {
		return PreExecuteResult{ProjectResult: ProjectResult{Error: errors.Wrap(err, "acquiring lock")}}
	}
	if !lockAttempt.LockAcquired && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return PreExecuteResult{ProjectResult: ProjectResult{Failure: fmt.Sprintf(
			"This project is currently locked by #%d. The locking plan must be applied or discarded before future plans can execute.",
			lockAttempt.CurrLock.Pull.Num)}}
	}
	ctx.Log.Info("acquired lock with id %q", lockAttempt.LockKey)

	// Check if config file is found, if not we continue the run.
	var config ProjectConfig
	absolutePath := filepath.Join(repoDir, project.Path)
	if p.ConfigReader.Exists(absolutePath) {
		config, err = p.ConfigReader.Read(absolutePath)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: err}}
		}
		ctx.Log.Info("parsed atlantis config file in %q", absolutePath)
	}

	// Check if terraform version is >= 0.9.0.
	terraformVersion := p.Terraform.Version()
	if config.TerraformVersion != nil {
		terraformVersion = config.TerraformVersion
	}
	constraints, _ := version.NewConstraint(">= 0.9.0")
	if constraints.Check(terraformVersion) {
		ctx.Log.Info("determined that we are running terraform with version >= 0.9.0. Running version %s", terraformVersion)
		if len(config.PreInit) > 0 {
			_, err := p.Run.Execute(ctx.Log, config.PreInit, absolutePath, workspace, terraformVersion, "pre_init")
			if err != nil {
				return PreExecuteResult{ProjectResult: ProjectResult{Error: errors.Wrapf(err, "running %s commands", "pre_init")}}
			}
		}
		_, err := p.Terraform.Init(ctx.Log, absolutePath, workspace, config.GetExtraArguments("init"), terraformVersion)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: err}}
		}
	} else {
		ctx.Log.Info("determined that we are running terraform with version < 0.9.0. Running version %s", terraformVersion)
		if len(config.PreGet) > 0 {
			_, err := p.Run.Execute(ctx.Log, config.PreGet, absolutePath, workspace, terraformVersion, "pre_get")
			if err != nil {
				return PreExecuteResult{ProjectResult: ProjectResult{Error: errors.Wrapf(err, "running %s commands", "pre_get")}}
			}
		}
		terraformGetCmd := append([]string{"get", "-no-color"}, config.GetExtraArguments("get")...)
		_, err := p.Terraform.RunCommandWithVersion(ctx.Log, absolutePath, terraformGetCmd, terraformVersion, workspace)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: err}}
		}
	}

	stage := fmt.Sprintf("pre_%s", strings.ToLower(ctx.Command.Name.String()))
	var commands []string
	if ctx.Command.Name == Plan {
		commands = config.PrePlan
	} else {
		commands = config.PreApply
	}
	if len(commands) > 0 {
		_, err := p.Run.Execute(ctx.Log, commands, absolutePath, workspace, terraformVersion, stage)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: errors.Wrapf(err, "running %s commands", stage)}}
		}
	}
	return PreExecuteResult{ProjectConfig: config, TerraformVersion: terraformVersion, LockResponse: lockAttempt}
}
