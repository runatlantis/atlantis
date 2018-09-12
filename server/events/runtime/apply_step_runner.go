package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// ApplyStepRunner runs `terraform apply`.
type ApplyStepRunner struct {
	TerraformExecutor TerraformExec
}

func (a *ApplyStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string) (string, error) {
	planPath := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectConfig))
	stat, err := os.Stat(planPath)
	if err != nil || stat.IsDir() {
		return "", fmt.Errorf("no plan found at path %q and workspace %q–did you run plan?", ctx.RepoRelDir, ctx.Workspace)
	}

	tfApplyCmd := append(append(append([]string{"apply", "-no-color", "-input=false"}, extraArgs...), ctx.CommentArgs...), planPath)
	var tfVersion *version.Version
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.TerraformVersion != nil {
		tfVersion = ctx.ProjectConfig.TerraformVersion
	}
	out, tfErr := a.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, tfApplyCmd, tfVersion, ctx.Workspace)

	// If the apply was successful, delete the plan.
	if tfErr == nil {
		ctx.Log.Info("apply successful, deleting planfile")
		if err := os.Remove(planPath); err != nil {
			ctx.Log.Warn("failed to delete planfile after successful apply: %s", err)
		}
	}
	return out, tfErr
}
