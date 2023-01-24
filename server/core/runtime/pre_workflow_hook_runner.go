package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_pre_workflows_hook_runner.go PreWorkflowHookRunner
type PreWorkflowHookRunner interface {
	Run(ctx models.WorkflowHookCommandContext, command string, path string) (string, string, error)
}

type DefaultPreWorkflowHookRunner struct {
	OutputHandler jobs.ProjectCommandOutputHandler
}

func (wh DefaultPreWorkflowHookRunner) Run(ctx models.WorkflowHookCommandContext, command string, path string) (string, string, error) {
	outputFilePath := filepath.Join(path, "OUTPUT_STATUS_FILE")

	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"BASE_BRANCH_NAME":   ctx.Pull.BaseBranch,
		"BASE_REPO_NAME":     ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":    ctx.BaseRepo.Owner,
		"COMMENT_ARGS":       strings.Join(ctx.EscapedCommentArgs, ","),
		"DIR":                path,
		"HEAD_BRANCH_NAME":   ctx.Pull.HeadBranch,
		"HEAD_COMMIT":        ctx.Pull.HeadCommit,
		"HEAD_REPO_NAME":     ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":    ctx.HeadRepo.Owner,
		"PULL_AUTHOR":        ctx.Pull.Author,
		"PULL_NUM":           fmt.Sprintf("%d", ctx.Pull.Num),
		"USER_NAME":          ctx.User.Username,
		"OUTPUT_STATUS_FILE": outputFilePath,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}

	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, out)
		ctx.Log.Debug("error: %s", err)
		return "", "", err
	}

	// Read the value from the "outputFilePath" file
	// to be returned as a custom description.
	var customStatusOut []byte
	if _, err := os.Stat(outputFilePath); err == nil {
		var customStatusErr error
		customStatusOut, customStatusErr = os.ReadFile(outputFilePath)
		if customStatusErr != nil {
			err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, out)
			ctx.Log.Debug("error: %s", err)
			return "", "", err
		}
	}

	wh.OutputHandler.SendWorkflowHook(ctx, fmt.Sprintf("%s\n", string(out)), true)

	ctx.Log.Info("successfully ran %q in %q", command, path)
	return string(out), strings.Trim(string(customStatusOut), "\n"), nil
}
