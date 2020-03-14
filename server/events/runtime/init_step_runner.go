package runtime

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// InitStep runs `terraform init`.
type InitStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (i *InitStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := i.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}
	terraformInitCmd := append([]string{"init", "-input=false", "-no-color", "-upgrade"}, extraArgs...)

	// If we're running < 0.9 we have to use `terraform get` instead of `init`.
	if MustConstraint("< 0.9.0").Check(tfVersion) {
		ctx.Log.Info("running terraform version %s so will use `get` instead of `init`", tfVersion)
		terraformInitCmd = append([]string{"get", "-no-color", "-upgrade"}, extraArgs...)
	}

	ctx.Log.Info("Trying to generate environment file for workspace %s", ctx.Workspace)
	if err := generateEnvironmentFile(path, ctx.Workspace); err != nil {
		return "", nil
	}

	out, err := i.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, terraformInitCmd, envs, tfVersion, ctx.Workspace)
	// Only include the init output if there was an error. Otherwise it's
	// unnecessary and lengthens the comment.
	if err != nil {
		return out, err
	}
	return "", nil
}

// generateEnvironmentFile generates a .terraformrc file containing config for tfeToken
// and hostname tfeHostname.
// It will create the file in home/.terraformrc.
func generateEnvironmentFile(path string, workspace string) error {
	const tfFolderName = ".terraform"
	const tfConfigName = "environment"

	tfFolder := filepath.Join(path, tfFolderName)
	tfConfig := filepath.Join(tfFolder, tfConfigName)

	os.MkdirAll(tfFolder, os.ModePerm)

	config := fmt.Sprintf(workspace)

	if err := ioutil.WriteFile(tfConfig, []byte(config), 0600); err != nil {
		return errors.Wrapf(err, "writing generated %s file with TFE token to %s", tfConfigName, tfConfig)
	}
	return nil
}
