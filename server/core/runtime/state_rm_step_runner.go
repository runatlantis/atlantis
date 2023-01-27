package runtime

import (
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
)

type stateRmStepRunner struct {
	terraformExecutor TerraformExec
	defaultTFVersion  *version.Version
}

func NewStateRmStepRunner(terraformExecutor TerraformExec, defaultTfVersion *version.Version) Runner {
	runner := &stateRmStepRunner{
		terraformExecutor: terraformExecutor,
		defaultTFVersion:  defaultTfVersion,
	}
	return NewWorkspaceStepRunnerDelegate(terraformExecutor, defaultTfVersion, runner)
}

func (p *stateRmStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := p.defaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	stateRmCmd := []string{"state", "rm"}
	stateRmCmd = append(stateRmCmd, extraArgs...)
	stateRmCmd = append(stateRmCmd, ctx.EscapedCommentArgs...)
	out, err := p.terraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), stateRmCmd, envs, tfVersion, ctx.Workspace)

	// If the state rm was successful and a plan file exists, delete the plan.
	planPath := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	if err == nil {
		if _, planPathErr := os.Stat(planPath); !os.IsNotExist(planPathErr) {
			ctx.Log.Info("state rm successful, deleting planfile")
			if removeErr := os.Remove(planPath); removeErr != nil {
				ctx.Log.Warn("failed to delete planfile after successful state rm: %s", removeErr)
			}
		}
	}
	return out, err
}
