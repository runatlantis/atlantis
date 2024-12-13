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

//go:generate pegomock generate --package mocks -o mocks/mock_pre_workflows_hook_runner.go PreWorkflowHookRunner
type PreWorkflowHookRunner interface {
	Run(ctx models.WorkflowHookCommandContext, command string, shell string, shellArgs string, path string) (string, string, error)
}

type DefaultPreWorkflowHookRunner struct {
	OutputHandler jobs.ProjectCommandOutputHandler
}

func (wh DefaultPreWorkflowHookRunner) Run(ctx models.WorkflowHookCommandContext, command string, shell string, shellArgs string, path string) (string, string, error) {
	outputFilePath := filepath.Join(path, "OUTPUT_STATUS_FILE")

	shellArgsSlice := append(strings.Split(shellArgs, " "), command)
	cmd := exec.Command(shell, shellArgsSlice...) // #nosec
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
		"PULL_URL":           ctx.Pull.URL,
		"USER_NAME":          ctx.User.Username,
		"OUTPUT_STATUS_FILE": outputFilePath,
		"COMMAND_NAME":       ctx.CommandName,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}

	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	outString := strings.ReplaceAll(string(out), "\n", "\r\n")
	wh.OutputHandler.SendWorkflowHook(ctx, outString, false)
	wh.OutputHandler.SendWorkflowHook(ctx, "\n", true)

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, shell+" "+shellArgs+" "+command, path, out)
		ctx.Log.Debug("error: %s", err)
		return string(out), "", err
	}

	// Read the value from the "outputFilePath" file
	// to be returned as a custom description.
	var customStatusOut []byte
	if _, err := os.Stat(outputFilePath); err == nil {
		var customStatusErr error
		customStatusOut, customStatusErr = os.ReadFile(outputFilePath)
		if customStatusErr != nil {
			err = fmt.Errorf("%s: running %q in %q: \n%s", err, shell+" "+shellArgs+" "+command, path, out)
			ctx.Log.Debug("error: %s", err)
			return string(out), "", err
		}
	}

	ctx.Log.Info("Successfully ran '%s' in '%s'", shell+" "+shellArgs+" "+command, path)
	return string(out), strings.Trim(string(customStatusOut), "\n"), nil
}
