package runtime

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
)

// VersionStepRunner runs a version command given a prjCtx
type VersionStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

// Run ensures a given version for the executable, builds the args from the project context and then runs executable returning the result
func (v *VersionStepRunner) Run(ctx context.Context, prjCtx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := v.DefaultTFVersion
	if prjCtx.TerraformVersion != nil {
		tfVersion = prjCtx.TerraformVersion
	}

	versionCmd := []string{"version"}
	return v.TerraformExecutor.RunCommandWithVersion(ctx, prjCtx, filepath.Clean(path), versionCmd, envs, tfVersion, prjCtx.Workspace)
}
