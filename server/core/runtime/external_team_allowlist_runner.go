package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate --package mocks -o mocks/mock_external_team_allowlist_runner.go ExternalTeamAllowlistRunner
type ExternalTeamAllowlistRunner interface {
	Run(ctx models.TeamAllowlistCheckerContext, shell, shellArgs, command string) (string, error)
}

type DefaultExternalTeamAllowlistRunner struct{}

func (r DefaultExternalTeamAllowlistRunner) Run(ctx models.TeamAllowlistCheckerContext, shell, shellArgs, command string) (string, error) {
	shellArgsSlice := append(strings.Split(shellArgs, " "), command)
	cmd := exec.CommandContext(context.TODO(), shell, shellArgsSlice...) // #nosec

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"BASE_BRANCH_NAME": ctx.Pull.BaseBranch,
		"BASE_REPO_NAME":   ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":  ctx.BaseRepo.Owner,
		"COMMENT_ARGS":     strings.Join(ctx.EscapedCommentArgs, ","),
		"HEAD_BRANCH_NAME": ctx.Pull.HeadBranch,
		"HEAD_COMMIT":      ctx.Pull.HeadCommit,
		"HEAD_REPO_NAME":   ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":  ctx.HeadRepo.Owner,
		"PULL_AUTHOR":      ctx.Pull.Author,
		"PULL_NUM":         fmt.Sprintf("%d", ctx.Pull.Num),
		"PULL_URL":         ctx.Pull.URL,
		"USER_NAME":        ctx.User.Username,
		"COMMAND_NAME":     ctx.CommandName,
		"PROJECT_NAME":     ctx.ProjectName,
		"REPO_ROOT":        ctx.RepoDir,
		"REPO_REL_PATH":    ctx.RepoRelDir,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s: running %q: \n%s", err, shell+" "+shellArgs+" "+command, out)
		ctx.Log.Debug("error: %s", err)
		return string(out), err
	}

	return strings.TrimSpace(string(out)), nil
}
