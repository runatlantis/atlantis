package runtime

import (
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

// EnvStepRunner set environment variables.
type EnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the env step command.
// value is the value for the environment variable. If set this is returned as
// the value. Otherwise command is run and its output is the value returned.
func (r *EnvStepRunner) Run(ctx models.ProjectCommandContext, command string, value string, path string, envs map[string]string) (string, error) {
	if value != "" {
		return value, nil
	}
	res, err := r.RunStepRunner.Run(ctx, command, path, envs)
	// Trim newline from res to support running `echo env_value` which has
	// a newline. We don't recommend users run echo -n env_value to remove the
	// newline because -n doesn't work in the sh shell which is what we use
	// to run commands.
	return strings.TrimSuffix(res, "\n"), err
}
