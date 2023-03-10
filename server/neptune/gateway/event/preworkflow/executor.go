package preworkflow

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	subprocess_exec "github.com/runatlantis/atlantis/server/neptune/exec"
	"os"
)

type HookExecutor struct {
	Logger logging.Logger
}

func (e *HookExecutor) Execute(ctx context.Context, hook *valid.PreWorkflowHook, repo models.Repo, path string) error {
	cmd := subprocess_exec.Command(e.Logger, "sh", "-c", hook.RunCommand)
	cmd.Dir = path
	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"BASE_BRANCH_NAME": repo.DefaultBranch,
		"BASE_REPO_NAME":   repo.FullName,
		"BASE_REPO_OWNER":  repo.Owner,
		"DIR":              path,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.RunWithNewProcessGroup(ctx)
	if err != nil {
		return errors.Wrap(err, "running command in separate process group")
	}
	return nil
}
