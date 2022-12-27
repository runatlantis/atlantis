package runtime

import (
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
)

type StateRmStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (p *StateRmStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := p.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	stateRmCmd := []string{"state", "rm"}
	stateRmCmd = append(stateRmCmd, extraArgs...)
	stateRmCmd = append(stateRmCmd, ctx.EscapedCommentArgs...)
	out, err := p.TerraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), stateRmCmd, envs, tfVersion, ctx.Workspace)

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
