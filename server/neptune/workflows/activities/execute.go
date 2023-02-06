package activities

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
)

type executeCommandActivities struct{}

type ExecuteCommandResponse struct {
	Output string
}

type ExecuteCommandRequest struct {
	Step           execute.Step
	Path           string
	EnvVars        map[string]string
	DynamicEnvVars []EnvVar
}

func (t *executeCommandActivities) ExecuteCommand(ctx context.Context, request ExecuteCommandRequest) (ExecuteCommandResponse, error) {
	cmd := exec.Command("sh", "-c", request.Step.RunCommand) // #nosec
	cmd.Dir = request.Path

	finalEnvVars := os.Environ()

	requestEnvVars, err := getEnvs(request.EnvVars, request.DynamicEnvVars)
	if err != nil {
		return ExecuteCommandResponse{}, err
	}
	for key, val := range requestEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}

	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ExecuteCommandResponse{}, errors.Wrapf(err, "running %q in %q: \n%s", request.Step.RunCommand, request.Path, out)
	}

	return ExecuteCommandResponse{
		Output: string(out),
	}, nil
}

// StringCommand is a command that outputs a string
type StringCommand struct {
	Command        string
	Dir            string
	AdditionalEnvs map[string]string
}

func (l *StringCommand) Load() (string, error) {
	cmd := exec.Command("sh", "-c", l.Command) // #nosec
	cmd.Dir = l.Dir

	finalEnvVars := os.Environ()
	for key, val := range l.AdditionalEnvs {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}

	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", errors.Wrap(err, "executing cmd")
	}

	// Trim newline from `out` to support running `echo env_value` which has
	// a newline. We don't recommend users run echo -n env_value to remove the
	// newline because -n doesn't work in the sh shell which is what we use
	// to run commands.
	return strings.TrimSuffix(string(out), "\n"), err
}

// EnvVar is a container used for activity requests
// and we use basic json serialization so we need to be explicit
// about our fields and can't do any clever interfacing here.
type EnvVar struct {
	Name    string
	Value   string
	Command StringCommand
}

func (v EnvVar) GetValue() (string, error) {
	if v.Value != "" {
		return v.Value, nil
	}

	return v.Command.Load()
}
