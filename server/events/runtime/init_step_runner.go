package runtime

import (
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
	terraformInitCmd := append([]string{"init", "-no-color", "input=false"}, extraArgs...)

	// If we're running < 0.9 we have to use `terraform get` instead of `init`.
	if MustConstraint("< 0.9.0").Check(tfVersion) {
		ctx.Log.Info("running terraform version %s so will use `get` instead of `init`", tfVersion)
		terraformInitCmd = append([]string{"get", "-no-color"}, extraArgs...)
	}

	out, err := i.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, terraformInitCmd, tfVersion, ctx.Workspace)
	// Only include the init output if there was an error. Otherwise it's
	// unnecessary and lengthens the comment.
	if err != nil {
		return out, err
	}
	return "", nil
}
