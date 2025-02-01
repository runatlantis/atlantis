package runtime

import (
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
)

// VersionStepRunner runs a version command given a ctx
type VersionStepRunner struct {
	TerraformExecutor     TerraformExec
	DefaultTFDistribution terraform.Distribution
	DefaultTFVersion      *version.Version
}

// Run ensures a given version for the executable, builds the args from the project context and then runs executable returning the result
func (v *VersionStepRunner) Run(ctx command.ProjectContext, _ []string, path string, envs map[string]string) (string, error) {
	tfDistribution := v.DefaultTFDistribution
	tfVersion := v.DefaultTFVersion
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution)
	}
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	versionCmd := []string{"version"}
	return v.TerraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), versionCmd, envs, tfDistribution, tfVersion, ctx.Workspace)
}
