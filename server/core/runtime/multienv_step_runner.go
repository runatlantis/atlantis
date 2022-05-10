package runtime

import (
	"fmt"
	"strings"
)

// EnvStepRunner set environment variables.
type MultiEnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the multienv step command.
// The command must return a json string containing the array of name-value pairs that are being added as extra environment variables
func (r *MultiEnvStepRunner) Run(ctx command.ProjectContext, command string, path string, envs map[string]string) (string, error) {
	res, err := r.RunStepRunner.Run(ctx, command, path, envs)
	if err == nil {
		envVars := strings.Split(res, ",")
		if len(envVars) > 0 {
			var sb strings.Builder
			sb.WriteString("Dynamic environment variables added:\n")
			for _, item := range envVars {
				nameValue := strings.Split(item, "=")
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
