package runtime

import (
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
)

// EnvStepRunner set environment variables.
type EnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the env step command.
// value is the value for the environment variable. If set this is returned as
// the value. Otherwise command is run and its output is the value returned.
func (r *EnvStepRunner) Run(ctx command.ProjectContext, command string, value string, path string, envs map[string]string) (string, error) {
	if value != "" {
		return value, nil
	}
	// Pass `false` for streamOutput because this isn't interesting to the user reading the build logs
	// in the web UI.
	res, err := r.RunStepRunner.Run(ctx, command, path, envs, false, valid.PostProcessRunOutputShow)
	// Trim newline from res to support running `echo env_value` which has
	// a newline. We don't recommend users run echo -n env_value to remove the
	// newline because -n doesn't work in the sh shell which is what we use
	// to run commands.
	return strings.TrimSuffix(res, "\n"), err
}
