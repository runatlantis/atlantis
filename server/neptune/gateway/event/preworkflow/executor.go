package preworkflow

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"os"
	"os/exec"
)

type HookExecutor struct {
}

func (e *HookExecutor) Execute(hook *valid.PreWorkflowHook, repo models.Repo, path string) error {
	cmd := exec.Command("sh", "-c", hook.RunCommand) // #nosec
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

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "running hook")
	}
	return nil
}
