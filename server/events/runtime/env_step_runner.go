package runtime

import (
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// EnvStepRunner set environment variables.
type EnvStepRunner struct {
	DefaultTFVersion *version.Version
}

func (r *EnvStepRunner) Run(ctx models.ProjectCommandContext, name string, command string, value string, path string, envs map[string]string) (string, string, error) {
	if value != "" {
		return name, value, nil
	}

	runStepRunner := RunStepRunner{DefaultTFVersion: r.DefaultTFVersion}
	res, err := runStepRunner.Run(ctx, command, path, envs)

	return name, res, err
}
