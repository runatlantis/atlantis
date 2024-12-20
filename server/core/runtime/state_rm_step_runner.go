package runtime

import (
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/utils"
)

type stateRmStepRunner struct {
	terraformExecutor     TerraformExec
	defaultTFDistribution terraform.Distribution
	defaultTFVersion      *version.Version
}

func NewStateRmStepRunner(terraformExecutor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTfVersion *version.Version) Runner {
	runner := &stateRmStepRunner{
		terraformExecutor:     terraformExecutor,
		defaultTFDistribution: defaultTfDistribution,
		defaultTFVersion:      defaultTfVersion,
	}
	return NewWorkspaceStepRunnerDelegate(terraformExecutor, defaultTfDistribution, defaultTfVersion, runner)
}

func (p *stateRmStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfDistribution := p.defaultTFDistribution
	tfVersion := p.defaultTFVersion
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution)
	}
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	stateRmCmd := []string{"state", "rm"}
	stateRmCmd = append(stateRmCmd, extraArgs...)
	stateRmCmd = append(stateRmCmd, ctx.EscapedCommentArgs...)
	out, err := p.terraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), stateRmCmd, envs, tfDistribution, tfVersion, ctx.Workspace)

	// If the state rm was successful and a plan file exists, delete the plan.
	planPath := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	if err == nil {
		if _, planPathErr := os.Stat(planPath); !os.IsNotExist(planPathErr) {
			ctx.Log.Info("state rm successful, deleting planfile")
			if removeErr := utils.RemoveIgnoreNonExistent(planPath); removeErr != nil {
				ctx.Log.Warn("failed to delete planfile after successful state rm: %s", removeErr)
			}
		}
	}
	return out, err
}
