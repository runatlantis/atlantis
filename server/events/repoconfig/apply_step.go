package repoconfig

import (
	"fmt"
	"os"
	"path/filepath"
)

// ApplyStep runs `terraform apply`.
type ApplyStep struct {
	ExtraArgs []string
	Meta      StepMeta
}

func (a *ApplyStep) Run() (string, error) {
	planPath := filepath.Join(a.Meta.AbsolutePath, a.Meta.Workspace+".tfplan")
	stat, err := os.Stat(planPath)
	if err != nil || stat.IsDir() {
		return "", fmt.Errorf("no plan found at path %q and workspace %q–did you run plan?", a.Meta.DirRelativeToRepoRoot, a.Meta.Workspace)
	}

	tfApplyCmd := append(append(append([]string{"apply", "-no-color"}, a.ExtraArgs...), a.Meta.ExtraCommentArgs...), planPath)
	return a.Meta.TerraformExecutor.RunCommandWithVersion(a.Meta.Log, a.Meta.AbsolutePath, tfApplyCmd, a.Meta.TerraformVersion, a.Meta.Workspace)
}
