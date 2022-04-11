package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pre_workflows_hook_runner.go PreWorkflowHookRunner
type PreWorkflowHookRunner interface {
	Run(ctx context.Context, preCtx models.PreWorkflowHookCommandContext, command string, path string) (string, error)
}

type DefaultPreWorkflowHookRunner struct{}

func (wh DefaultPreWorkflowHookRunner) Run(ctx context.Context, preCtx models.PreWorkflowHookCommandContext, command string, path string) (string, error) {
	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"BASE_BRANCH_NAME": preCtx.Pull.BaseBranch,
		"BASE_REPO_NAME":   preCtx.BaseRepo.Name,
		"BASE_REPO_OWNER":  preCtx.BaseRepo.Owner,
		"DIR":              path,
		"HEAD_BRANCH_NAME": preCtx.Pull.HeadBranch,
		"HEAD_COMMIT":      preCtx.Pull.HeadCommit,
		"HEAD_REPO_NAME":   preCtx.HeadRepo.Name,
		"HEAD_REPO_OWNER":  preCtx.HeadRepo.Owner,
		"PULL_AUTHOR":      preCtx.Pull.Author,
		"PULL_NUM":         fmt.Sprintf("%d", preCtx.Pull.Num),
		"USER_NAME":        preCtx.User.Username,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars

	// pre-workflow hooks operate different than our terraform steps
	// it's up to the underlying implementation to log errors/output accordingly.
	// The only required step is to share Stdout and Stderr with the underlying
	// process, so that our logging sidecar can forward the logs to kibana
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}
	return "", nil
}
