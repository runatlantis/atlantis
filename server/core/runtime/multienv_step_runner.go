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

type MultiEnvCallerResult struct {
	Success      bool
	ErrorMessage string `json:"errorMessage"`
	Result       []KeyValue
}

// EnvStepRunner set environment variables.
type MultiEnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the multienv step command.
// The command must return a json string containing the array of name-value pairs that are being added as extra environment variables
func (r *MultiEnvStepRunner) Run(ctx models.ProjectCommandContext, command string, path string, envs map[string]string) (string, error) {
	res, err := r.RunStepRunner.Run(ctx, command, path, envs)
	if err == nil {
		var callerResult MultiEnvCallerResult
		err = json.Unmarshal([]byte(res), &callerResult)
		if err == nil {
			if callerResult.Success {
				if len(callerResult.Result) > 0 {
					var sb strings.Builder
					sb.WriteString("Dynamic environment variables added:\n")
					for _, item := range callerResult.Result {
						envs[item.Name] = item.Value
						sb.WriteString(item.Name)
						sb.WriteString("\n")
					}
					return sb.String(), nil
				}
				return "No dynamic environment variable added", nil
			}
			return callerResult.ErrorMessage, nil
		}
		return "Parsing the json result of the multienv step failed, json content: " + res, err
	}
	return "", err
}
