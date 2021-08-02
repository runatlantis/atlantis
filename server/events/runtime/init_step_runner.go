package runtime

import (
	"os"
	"path/filepath"
	"strings"

	version "github.com/hashicorp/go-version"
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

	terraformInitVerb := []string{"init"}
	terraformInitArgs := []string{"-input=false"}

	// If we're running < 0.9 we have to use `terraform get` instead of `init`.
	if MustConstraint("< 0.9.0").Check(tfVersion) {
		ctx.Log.Info("running terraform version %s so will use `get` instead of `init`", tfVersion)
		terraformInitVerb = []string{"get"}
		terraformInitArgs = []string{}
	}

	terraformInitArgs = append(terraformInitArgs, "-no-color")

	lockfilePath := filepath.Join(path, ".terraform.lock.hcl")
	if MustConstraint("< 0.14.0").Check(tfVersion) || fileDoesNotExists(lockfilePath) {
		terraformInitArgs = append(terraformInitArgs, "-upgrade")
	}

	// work if any of the core args have been overridden
	finalArgs := []string{}
	usedExtraArgs := []string{}
	for _, arg := range terraformInitArgs {
		override := ""
		prefix := arg
		argSplit := strings.Split(arg, "=")
		if len(argSplit) == 2 {
			prefix = argSplit[0]
		}
		for _, extraArg := range extraArgs {
			if strings.HasPrefix(extraArg, prefix) {
				override = extraArg
			}
		}
		if override != "" {
			finalArgs = append(finalArgs, override)
			usedExtraArgs = append(usedExtraArgs, override)
		} else {
			finalArgs = append(finalArgs, arg)
		}
	}
	// add any extra args that are not overrides
	for _, extraArg := range extraArgs {
		if !stringInSlice(usedExtraArgs, extraArg) {
			finalArgs = append(finalArgs, extraArg)
		}
	}

	terraformInitCmd := append(terraformInitVerb, finalArgs...)

	out, err := i.TerraformExecutor.RunCommandWithVersion(ctx, path, terraformInitCmd, envs, tfVersion, ctx.Workspace)
	// Only include the init output if there was an error. Otherwise it's
	// unnecessary and lengthens the comment.
	if err != nil {
		return out, err
	}
	return "", nil
}

func fileDoesNotExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func stringInSlice(stringSlice []string, target string) bool {
	for _, value := range stringSlice {
		if value == target {
			return true
		}
	}
	return false
}
