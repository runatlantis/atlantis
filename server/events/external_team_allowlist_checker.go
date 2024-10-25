package events

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/core/runtime"

	"github.com/runatlantis/atlantis/server/events/models"
)

type ExternalTeamAllowlistChecker struct {
	Command                     string
	ExtraArgs                   []string
	ExternalTeamAllowlistRunner runtime.ExternalTeamAllowlistRunner
}

func (checker *ExternalTeamAllowlistChecker) HasRules() bool {
	return true
}

func (checker *ExternalTeamAllowlistChecker) IsCommandAllowedForTeam(ctx models.TeamAllowlistCheckerContext, team string, command string) bool {
	cmd := checker.buildCommandString(ctx, []string{team}, command)
	out, err := checker.ExternalTeamAllowlistRunner.Run(ctx, "sh", "-c", cmd)
	if err != nil {
		ctx.Log.Err("Command '%s' error '%s'", cmd, err)
		return false
	}

	outputResults := checker.checkOutputResults(out)
	if !outputResults {
		ctx.Log.Info("command '%s' returns '%s'", cmd, out)
	}

	return outputResults
}

func (checker *ExternalTeamAllowlistChecker) IsCommandAllowedForAnyTeam(ctx models.TeamAllowlistCheckerContext, teams []string, command string) bool {
	cmd := checker.buildCommandString(ctx, teams, command)
	out, err := checker.ExternalTeamAllowlistRunner.Run(ctx, "sh", "-c", cmd)
	if err != nil {
		ctx.Log.Err("Command '%s' error '%s'", cmd, err)
		return false
	}

	outputResults := checker.checkOutputResults(out)
	if !outputResults {
		ctx.Log.Info("command '%s' returns '%s'", cmd, out)
	}

	return outputResults
}

func (checker *ExternalTeamAllowlistChecker) buildCommandString(ctx models.TeamAllowlistCheckerContext, teams []string, command string) string {
	// Build command string
	// Format is "$external_cmd $external_args $command $repo $teams"
	cmdArr := append([]string{checker.Command}, checker.ExtraArgs...)
	orgTeams := make([]string, len(teams))
	for i, team := range teams {
		orgTeams[i] = fmt.Sprintf("%s/%s", ctx.BaseRepo.Owner, team)
	}

	teamStr := strings.Join(orgTeams, " ")
	return strings.Join(append(cmdArr, command, ctx.BaseRepo.FullName, teamStr), " ")
}

func (checker *ExternalTeamAllowlistChecker) checkOutputResults(output string) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lastLine := lines[len(lines)-1]
	return strings.EqualFold(lastLine, "pass")
}
