package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	DefaultTFVersion *version.Version
}

func (r *RunStepRunner) Run(ctx models.ProjectCommandContext, command []string, path string) (string, error) {
	if len(command) < 1 {
		return "", errors.New("no commands for run step")
	}

	cmd := exec.Command("sh", "-c", strings.Join(command, " ")) // #nosec
	cmd.Dir = path
	tfVersion := r.DefaultTFVersion.String()
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.TerraformVersion != nil {
		tfVersion = ctx.ProjectConfig.TerraformVersion.String()
	}
	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"WORKSPACE":                  ctx.Workspace,
		"ATLANTIS_TERRAFORM_VERSION": tfVersion,
		"DIR":                        path,
		"PLANFILE":                   filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectConfig)),
		"BASE_REPO_NAME":             ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":            ctx.BaseRepo.Owner,
		"HEAD_REPO_NAME":             ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":            ctx.HeadRepo.Owner,
		"HEAD_BRANCH_NAME":           ctx.Pull.HeadBranch,
		"BASE_BRANCH_NAME":           ctx.Pull.BaseBranch,
		"PULL_NUM":                   fmt.Sprintf("%d", ctx.Pull.Num),
		"PULL_AUTHOR":                ctx.Pull.Author,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	commandStr := strings.Join(command, " ")
	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, commandStr, path, out)
		ctx.Log.Debug("error: %s", err)
		return string(out), err
	}
	ctx.Log.Info("successfully ran %q in %q", commandStr, path)
	return string(out), nil
}
