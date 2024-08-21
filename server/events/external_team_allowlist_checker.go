package events

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

type ExternalTeamAllowlistChecker struct {
	Command   string
	ExtraArgs []string
}

func (checker *ExternalTeamAllowlistChecker) HasRules() bool {
	return true
}

func (checker *ExternalTeamAllowlistChecker) IsCommandAllowedForTeam(ctx models.TeamAllowlistCheckerContext, team string, command string) bool {
	// Build command string
	// Format is "$external_cmd $external_args $command $repo $team"
	cmdArr := append([]string{checker.Command}, checker.ExtraArgs...)
	teamStr := strings.Join([]string{ctx.BaseRepo.Owner, team}, "/")
	cmd := strings.Join(append(cmdArr, command, ctx.BaseRepo.FullName, teamStr), " ")

	out, err := checker.run(ctx, "sh", "-c", cmd)
	if err != nil {
		return false
	}

	return checker.checkOutputResults(out)
}

func (checker *ExternalTeamAllowlistChecker) IsCommandAllowedForAnyTeam(ctx models.TeamAllowlistCheckerContext, teams []string, command string) bool {
	// Build command string
	// Format is "$external_cmd $external_args $command $repo $teams"
	cmdArr := append([]string{checker.Command}, checker.ExtraArgs...)
	orgTeams := make([]string, len(teams))
	for i, team := range teams {
		orgTeams[i] = fmt.Sprintf("%s/%s", ctx.BaseRepo.Owner, team)
	}

	teamStr := strings.Join(orgTeams, " ")
	cmd := strings.Join(append(cmdArr, command, ctx.BaseRepo.FullName, teamStr), " ")

	out, err := checker.run(ctx, "sh", "-c", cmd)
	if err != nil {
		return false
	}

	return checker.checkOutputResults(out)
}

func (checker *ExternalTeamAllowlistChecker) run(ctx models.TeamAllowlistCheckerContext, shell, shellArgs, command string) (string, error) {
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

func (checker *ExternalTeamAllowlistChecker) checkOutputResults(output string) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lastLine := lines[len(lines)-1]
	return strings.EqualFold(lastLine, "pass")
}
