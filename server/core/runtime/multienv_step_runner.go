package runtime

import (
	"encoding/json"

	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

type KeyValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EnvStepRunner set environment variables.
type MultiEnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the env step command.
// value is the value for the environment variable. If set this is returned as
// the value. Otherwise command is run and its output is the value returned.
func (r *MultiEnvStepRunner) Run(ctx models.ProjectCommandContext, command string, path string, envs map[string]string) (string, error) {
	res, err := r.RunStepRunner.Run(ctx, command, path, envs)
	if err == nil {
		var customVars []KeyValue
		err = json.Unmarshal([]byte(res), &customVars)
		if err == nil {
			if len(customVars) > 0 {
				var sb strings.Builder
				sb.WriteString("Dynamic environment variables added:\n")
				for _, item := range customVars {
					envs[item.Name] = item.Value
					sb.WriteString("name: ")
					sb.WriteString(item.Name)
					sb.WriteString("  ")
					sb.WriteString("value: ")
					sb.WriteString(item.Value)
					sb.WriteString("\n")
				}
				return sb.String(), nil
			} else {
				return "No environment variable added", nil
			}
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
