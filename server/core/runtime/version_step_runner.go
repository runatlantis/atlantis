package runtime

import (
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// VersionStepRunner runs a version command given a ctx
type VersionStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

// Run ensures a given version for the executable, builds the args from the project context and then runs executable returning the result
func (v *VersionStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := v.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	versionCmd := []string{"version"}
	return v.TerraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), versionCmd, envs, tfVersion, ctx.Workspace)
}
