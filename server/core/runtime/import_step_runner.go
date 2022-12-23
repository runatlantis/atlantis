package runtime

import (
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
)

type ImportStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (p *ImportStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := p.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	importCmd := []string{"import"}
	importCmd = append(importCmd, extraArgs...)
	importCmd = append(importCmd, ctx.EscapedCommentArgs...)
	out, err := p.TerraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), importCmd, envs, tfVersion, ctx.Workspace)

	// If the import was successful and a plan file exists, delete the plan.
	planPath := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	if err == nil {
		if _, planPathErr := os.Stat(planPath); !os.IsNotExist(planPathErr) {
			ctx.Log.Info("import successful, deleting planfile")
			if removeErr := os.Remove(planPath); removeErr != nil {
				ctx.Log.Warn("failed to delete planfile after successful import: %s", removeErr)
			}
		}
	}
	return out, err
}
