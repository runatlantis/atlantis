package runtime

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// MinimumVersionStepRunnerDelegate ensures that a given step runner can't run unless the command version being used
// is greater than a provided minimum
type MinimumVersionStepRunnerDelegate struct {
	minimumVersion   *version.Version
	defaultTfVersion *version.Version
	delegate         Runner
}

func NewMinimumVersionStepRunnerDelegate(minimumVersionStr string, defaultVersion *version.Version, delegate Runner) (Runner, error) {
	minimumVersion, err := version.NewVersion(minimumVersionStr)

	if err != nil {
		return &MinimumVersionStepRunnerDelegate{}, errors.Wrap(err, "initializing minimum version")
	}

	return &MinimumVersionStepRunnerDelegate{
		minimumVersion:   minimumVersion,
		defaultTfVersion: defaultVersion,
		delegate:         delegate,
	}, nil
}

func (r *MinimumVersionStepRunnerDelegate) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := r.defaultTfVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	if tfVersion.LessThan(r.minimumVersion) {
		return fmt.Sprintf("Version: %s is unsupported for this step. Minimum version is: %s", tfVersion.String(), r.minimumVersion.String()), nil
	}

	return r.delegate.Run(ctx, extraArgs, path, envs)
}
