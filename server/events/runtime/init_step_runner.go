package runtime

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// InitStep runs `terraform init`.
type InitStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (i *InitStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string) (string, error) {
	tfVersion := i.DefaultTFVersion
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.TerraformVersion != nil {
		tfVersion = ctx.ProjectConfig.TerraformVersion
	}
	terraformInitCmd := append([]string{"init", "-input=false", "-no-color"}, extraArgs...)

	// If we're running < 0.9 we have to use `terraform get` instead of `init`.
	if MustConstraint("< 0.9.0").Check(tfVersion) {
		ctx.Log.Info("running terraform version %s so will use `get` instead of `init`", tfVersion)
		terraformInitCmd = append([]string{"get", "-no-color"}, extraArgs...)
	}

	// Remove any error file from any previous init
	initErrorFile := filepath.Join(path, GetProjectFilenamePrefix(ctx.Workspace, ctx.ProjectConfig)+".tfinit-error")
	_ = os.Remove(initErrorFile) // safe to ignore return result

	out, err := i.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, terraformInitCmd, tfVersion, ctx.Workspace)
	// Only include the init output if there was an error. Otherwise it's
	// unnecessary and lengthens the comment.
	if err != nil {
		// If there was an error, write the result out to the '.tfinit-error' file in
		// the workspace. This may be used later to either retrieve the reason for the
		// failure, or to prevent automerging.
		writeErr := ioutil.WriteFile(initErrorFile, []byte(out), 0644)
		if writeErr != nil {
			panic(writeErr)
		}
		ctx.Log.Info("Failed init output has been written to %s", initErrorFile)

		return out, err
	}
	return "", nil
}
