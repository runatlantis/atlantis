package activities

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/execute"
)

type executeCommandActivities struct{}

type ExecuteCommandResponse struct {
	Output string
}

type ExecuteCommandRequest struct {
	Step    execute.Step
	Path    string
	EnvVars map[string]string
}

func (t *executeCommandActivities) ExecuteCommand(ctx context.Context, request ExecuteCommandRequest) (ExecuteCommandResponse, error) {
	cmd := exec.Command("sh", "-c", request.Step.RunCommand) // #nosec
	cmd.Dir = request.Path

	finalEnvVars := os.Environ()
	for key, val := range request.EnvVars {
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
