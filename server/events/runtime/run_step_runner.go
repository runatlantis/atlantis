package runtime

import (
	"fmt"
	"os"
	"os/exec"
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
	customEnvVars := []string{
		fmt.Sprintf("WORKSPACE=%s", ctx.Workspace),
		fmt.Sprintf("ATLANTIS_TERRAFORM_VERSION=%s", tfVersion),
		fmt.Sprintf("DIR=%s", path),
	}
	finalEnvVars := append(baseEnvVars, customEnvVars...)
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
