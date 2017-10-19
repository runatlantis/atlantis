package events

import (
	"path/filepath"
	"fmt"
	"github.com/hootsuite/atlantis/server/events/locking"
	"github.com/pkg/errors"
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/events/terraform"
	"github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/server/events/run"
	"strings"
)

type ProjectPreExecute struct {
	Locker          locking.Locker
	ConfigReader *ConfigReader
	Terraform    *terraform.Client
	Run          *run.Run
}

type PreExecuteResult struct {
	ProjectResult ProjectResult
	ProjectConfig ProjectConfig
	TerraformVersion *version.Version
	LockResponse locking.TryLockResponse
}

func (p *ProjectPreExecute) Execute(ctx *CommandContext, repoDir string, project models.Project) PreExecuteResult {
	tfEnv := ctx.Command.Environment
	lockAttempt, err := p.Locker.TryLock(project, tfEnv, ctx.Pull, ctx.User)
	if err != nil {
		return PreExecuteResult{ProjectResult: ProjectResult{Error: errors.Wrap(err, "acquiring lock")}}
	}
	if lockAttempt.LockAcquired == false && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return PreExecuteResult{ProjectResult: ProjectResult{Failure: fmt.Sprintf(
			"This project is currently locked by #%d. The locking plan must be applied or discarded before future plans can execute.",
			lockAttempt.CurrLock.Pull.Num)}}
	}
	ctx.Log.Info("acquired lock with id %q", lockAttempt.LockKey)

	// check if config file is found, if not we continue the run
	var config ProjectConfig
	absolutePath := filepath.Join(repoDir, project.Path)
	if p.ConfigReader.Exists(absolutePath) {
		config, err = p.ConfigReader.Read(absolutePath)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: err}}
		}
		ctx.Log.Info("parsed atlantis config file in %q", absolutePath)
	}

	// check if terraform version is >= 0.9.0
	terraformVersion := p.Terraform.Version()
	if config.TerraformVersion != nil {
		terraformVersion = config.TerraformVersion
	}
	constraints, _ := version.NewConstraint(">= 0.9.0")
	if constraints.Check(terraformVersion) {
		ctx.Log.Info("determined that we are running terraform with version >= 0.9.0. Running version %s", terraformVersion)
		_, err := p.Terraform.RunInitAndEnv(ctx.Log, absolutePath, tfEnv, config.GetExtraArguments("init"), terraformVersion)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: err}}
		}
	} else {
		ctx.Log.Info("determined that we are running terraform with version < 0.9.0. Running version %s", terraformVersion)
		terraformGetCmd := append([]string{"get", "-no-color"}, config.GetExtraArguments("get")...)
		_, err := p.Terraform.RunCommandWithVersion(ctx.Log, absolutePath, terraformGetCmd, terraformVersion, tfEnv)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: err}}
		}
	}

	stage := fmt.Sprintf("pre_%s", strings.ToLower(ctx.Command.Name.String()))
	var commands []string
	if ctx.Command.Name == Plan {
		commands = config.PrePlan.Commands
	} else {
		commands = config.PreApply.Commands
	}
	if len(commands) > 0 {
		_, err := p.Run.Execute(ctx.Log, commands, absolutePath, tfEnv, terraformVersion, stage)
		if err != nil {
			return PreExecuteResult{ProjectResult: ProjectResult{Error: errors.Wrapf(err, "running pre %s commands", stage)}}
		}
	}
	return PreExecuteResult{ProjectConfig: config, TerraformVersion: terraformVersion, LockResponse: lockAttempt}
}
