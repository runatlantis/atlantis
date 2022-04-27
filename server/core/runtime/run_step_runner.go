package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	// TerraformBinDir is the directory where Atlantis downloads Terraform binaries.
	TerraformBinDir string
}

func (r *RunStepRunner) Run(ctx context.Context, prjCtx command.ProjectContext, command string, path string, envs map[string]string) (string, error) {
	tfVersion := r.DefaultTFVersion
	if prjCtx.TerraformVersion != nil {
		tfVersion = prjCtx.TerraformVersion
	}

	err := r.TerraformExecutor.EnsureVersion(prjCtx.Log, tfVersion)
	if err != nil {
		err = fmt.Errorf("%s: Downloading terraform Version %s", err, tfVersion.String())
		prjCtx.Log.Errorf("error: %s", err)
		return "", err
	}

	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"ATLANTIS_TERRAFORM_VERSION": tfVersion.String(),
		"BASE_BRANCH_NAME":           prjCtx.Pull.BaseBranch,
		"BASE_REPO_NAME":             prjCtx.BaseRepo.Name,
		"BASE_REPO_OWNER":            prjCtx.BaseRepo.Owner,
		"COMMENT_ARGS":               strings.Join(prjCtx.EscapedCommentArgs, ","),
		"DIR":                        path,
		"HEAD_BRANCH_NAME":           prjCtx.Pull.HeadBranch,
		"HEAD_COMMIT":                prjCtx.Pull.HeadCommit,
		"HEAD_REPO_NAME":             prjCtx.HeadRepo.Name,
		"HEAD_REPO_OWNER":            prjCtx.HeadRepo.Owner,
		"PATH":                       fmt.Sprintf("%s:%s", os.Getenv("PATH"), r.TerraformBinDir),
		"PLANFILE":                   filepath.Join(path, GetPlanFilename(prjCtx.Workspace, prjCtx.ProjectName)),
		"SHOWFILE":                   filepath.Join(path, prjCtx.GetShowResultFileName()),
		"PROJECT_NAME":               prjCtx.ProjectName,
		"PULL_AUTHOR":                prjCtx.Pull.Author,
		"PULL_NUM":                   fmt.Sprintf("%d", prjCtx.Pull.Num),
		"REPO_REL_DIR":               prjCtx.RepoRelDir,
		"USER_NAME":                  prjCtx.User.Username,
		"WORKSPACE":                  prjCtx.Workspace,
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
		prjCtx.Log.Errorf("error: %s", err)
		return "", err
	}
	prjCtx.Log.Infof("successfully ran %q in %q", command, path)
	return string(out), nil
}
