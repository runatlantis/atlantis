package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	// TerraformBinDir is the directory where Atlantis downloads Terraform binaries.
	TerraformBinDir         string
	ProjectCmdOutputHandler jobs.ProjectCommandOutputHandler
}

func (r *RunStepRunner) Run(ctx command.ProjectContext, command string, path string, envs map[string]string, streamOutput bool) (string, error) {
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
		"POLICYCHECKFILE":            filepath.Join(path, ctx.GetPolicyCheckResultFileName()),
		"PROJECT_NAME":               ctx.ProjectName,
		"PULL_AUTHOR":                ctx.Pull.Author,
		"PULL_NUM":                   fmt.Sprintf("%d", ctx.Pull.Num),
		"PULL_URL":                   ctx.Pull.URL,
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

	runner := models.NewShellCommandRunner(command, finalEnvVars, path, streamOutput, r.ProjectCmdOutputHandler)
	output, err := runner.Run(ctx)

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, output)
		ctx.Log.Debug("error: %s", err)
		return "", err
	}
	return output, nil
}
