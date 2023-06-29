package runtime

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
)

// EnvStepRunner set environment variables.
type MultiEnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the multienv step command.
// The command must return a json string containing the array of name-value pairs that are being added as extra environment variables
func (r *MultiEnvStepRunner) Run(ctx command.ProjectContext, command string, path string, envs map[string]string) (string, error) {
	res, err := r.RunStepRunner.Run(ctx, command, path, envs, false, valid.PostProcessRunOutputShow)
	if err == nil {
		if len(res) > 0 {
			var sb strings.Builder
			sb.WriteString("Dynamic environment variables added:\n")

			envVars := strings.Split(res, ",")
			for _, item := range envVars {
				// Only split after the first = found in case the environment variable value has
				// = in it (as might be the case with access tokens)
				nameValue := strings.SplitN(strings.TrimRight(item, "\n"), "=", 2)
				if len(nameValue) == 2 {
					envs[nameValue[0]] = nameValue[1]
					sb.WriteString(nameValue[0])
					sb.WriteString("\n")
				} else {
					return "", fmt.Errorf("Invalid environment variable definition: %s", item)
				}
			}
			return sb.String(), nil
		}
		return "No dynamic environment variable added", nil
	}
	return "", err
}
