package runtime

import (
	"context"
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/runtime/common"
)

// InitStep runs `terraform init`.
type InitStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (i *InitStepRunner) Run(ctx context.Context, prjCtx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	lockFileName := ".terraform.lock.hcl"
	terraformLockfilePath := filepath.Join(path, lockFileName)
	terraformLockFileTracked, err := common.IsFileTracked(path, lockFileName)
	if err != nil {
		prjCtx.Log.Warnf("Error checking if %s is tracked in %s", lockFileName, path)

	}
	// If .terraform.lock.hcl is not tracked in git and it exists prior to init
	// delete it as it probably has been created by a previous run of
	// terraform init
	if common.FileExists(terraformLockfilePath) && !terraformLockFileTracked {
		delErr := os.Remove(terraformLockfilePath)
		if delErr != nil {
			prjCtx.Log.Infof("Error Deleting `%s`", lockFileName)
		}
	}

	tfVersion := i.DefaultTFVersion
	if prjCtx.TerraformVersion != nil {
		tfVersion = prjCtx.TerraformVersion
	}

	terraformInitVerb := []string{"init"}
	terraformInitArgs := []string{"-input=false"}

	// If we're running < 0.9 we have to use `terraform get` instead of `init`.
	if MustConstraint("< 0.9.0").Check(tfVersion) {
		prjCtx.Log.Infof("running terraform version %s so will use `get` instead of `init`", tfVersion)
		terraformInitVerb = []string{"get"}
		terraformInitArgs = []string{}
	}

	if MustConstraint("< 0.14.0").Check(tfVersion) || !common.FileExists(terraformLockfilePath) {
		terraformInitArgs = append(terraformInitArgs, "-upgrade")
	}

	finalArgs := common.DeDuplicateExtraArgs(terraformInitArgs, extraArgs)

	terraformInitCmd := append(terraformInitVerb, finalArgs...)

	out, err := i.TerraformExecutor.RunCommandWithVersion(ctx, prjCtx, path, terraformInitCmd, envs, tfVersion, prjCtx.Workspace)
	// Only include the init output if there was an error. Otherwise it's
	// unnecessary and lengthens the comment.
	if err != nil {
		return out, err
	}
	return "", nil
}
