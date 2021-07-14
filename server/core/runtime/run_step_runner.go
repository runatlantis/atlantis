package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	// TerraformBinDir is the directory where Atlantis downloads Terraform binaries.
	TerraformBinDir string
}

func (r *RunStepRunner) Run(ctx models.ProjectCommandContext, command string, path string, envs map[string]string) (string, error) {
	tfVersion := r.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	err := r.TerraformExecutor.EnsureVersion(ctx.Log, tfVersion)
	if err != nil {
		err = fmt.Errorf("%s: Downloading terraform Version %s", err, tfVersion.String())
		ctx.Log.Debug("error: %s", err)
		return "", err
	}

	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"ATLANTIS_TERRAFORM_VERSION": tfVersion.String(),
		"BASE_BRANCH_NAME":           ctx.Pull.BaseBranch,
		"BASE_REPO_NAME":             ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":            ctx.BaseRepo.Owner,
		"COMMENT_ARGS":               strings.Join(ctx.EscapedCommentArgs, ","),
		"DIR":                        path,
		"HEAD_BRANCH_NAME":           ctx.Pull.HeadBranch,
		"HEAD_COMMIT":                ctx.Pull.HeadCommit,
		"HEAD_REPO_NAME":             ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":            ctx.HeadRepo.Owner,
		"PATH":                       fmt.Sprintf("%s:%s", os.Getenv("PATH"), r.TerraformBinDir),
		"PLANFILE":                   filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName)),
		"SHOWFILE":                   filepath.Join(path, ctx.GetShowResultFileName()),
		"PROJECT_NAME":               ctx.ProjectName,
		"PULL_AUTHOR":                ctx.Pull.Author,
		"PULL_NUM":                   fmt.Sprintf("%d", ctx.Pull.Num),
		"REPO_REL_DIR":               ctx.RepoRelDir,
		"USER_NAME":                  ctx.User.Username,
		"WORKSPACE":                  ctx.Workspace,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range envs {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, out)
		ctx.Log.Debug("error: %s", err)
		return "", err
	}
	ctx.Log.Info("successfully ran %q in %q", command, path)
	return string(out), nil
}
