package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	DefaultTFVersion *version.Version
}

func (r *RunStepRunner) Run(ctx models.ProjectCommandContext, command string, path string) (string, error) {
	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path
	tfVersion := r.DefaultTFVersion.String()
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion.String()
	}
	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"WORKSPACE":                  ctx.Workspace,
		"ATLANTIS_TERRAFORM_VERSION": tfVersion,
		"DIR":                        path,
		"PLANFILE":                   filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName)),
		"BASE_REPO_NAME":             ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":            ctx.BaseRepo.Owner,
		"HEAD_REPO_NAME":             ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":            ctx.HeadRepo.Owner,
		"HEAD_BRANCH_NAME":           ctx.Pull.HeadBranch,
		"BASE_BRANCH_NAME":           ctx.Pull.BaseBranch,
		"PULL_NUM":                   fmt.Sprintf("%d", ctx.Pull.Num),
		"PULL_AUTHOR":                ctx.Pull.Author,
		"USER_NAME":                  ctx.User.Username,
		"PROJECT_NAME":               ctx.ProjectName,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, out)
		ctx.Log.Debug("error: %s", err)
		return string(out), err
	}
	ctx.Log.Info("successfully ran %q in %q", command, path)
	return string(out), nil
}
