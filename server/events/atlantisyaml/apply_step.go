package atlantisyaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runatlantis/atlantis/server/events/vcs"
)

// ApplyStep runs `terraform apply`.
type ApplyStep struct {
	ExtraArgs         []string
	VCSClient         vcs.ClientProxy
	TerraformExecutor TerraformExec
	Meta              StepMeta
}

func (a *ApplyStep) Run() (string, error) {
	planPath := filepath.Join(a.Meta.AbsolutePath, a.Meta.Workspace+".tfplan")
	stat, err := os.Stat(planPath)
	if err != nil || stat.IsDir() {
		return "", fmt.Errorf("no plan found at path %q and workspace %q–did you run plan?", a.Meta.DirRelativeToRepoRoot, a.Meta.Workspace)
	}

	tfApplyCmd := append(append(append([]string{"apply", "-no-color"}, a.ExtraArgs...), a.Meta.ExtraCommentArgs...), planPath)
	return a.TerraformExecutor.RunCommandWithVersion(a.Meta.Log, a.Meta.AbsolutePath, tfApplyCmd, a.Meta.TerraformVersion, a.Meta.Workspace)
}
