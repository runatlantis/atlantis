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
	// todo: move this to a common library
	planFileName := fmt.Sprintf("%s.tfplan", ctx.Workspace)
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.Name != nil {
		planFileName = fmt.Sprintf("%s-%s", *ctx.ProjectConfig.Name, planFileName)
	}
	planFile := filepath.Join(path, planFileName)
	stat, err := os.Stat(planFile)
	if err != nil || stat.IsDir() {
		return "", fmt.Errorf("no plan found at path %q and workspace %q–did you run plan?", ctx.RepoRelPath, ctx.Workspace)
	}

	tfApplyCmd := append(append(append([]string{"apply", "-no-color"}, extraArgs...), ctx.CommentArgs...), planFile)
	var tfVersion *version.Version
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.TerraformVersion != nil {
		tfVersion = ctx.ProjectConfig.TerraformVersion
	}
	return a.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, tfApplyCmd, tfVersion, ctx.Workspace)
}
