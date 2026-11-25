// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/core/runtime"

	"github.com/runatlantis/atlantis/server/events/models"
)

// externalTeamAllowlistArgv0 is the value of $0 inside the "sh -c" script.
const externalTeamAllowlistArgv0 = "atlantis-team-authz"

type ExternalTeamAllowlistChecker struct {
	Command                     string
	ExtraArgs                   []string
	ExternalTeamAllowlistRunner runtime.ExternalTeamAllowlistRunner
}

func (checker *ExternalTeamAllowlistChecker) HasRules() bool {
	return true
}

func (checker *ExternalTeamAllowlistChecker) IsCommandAllowedForTeam(ctx models.TeamAllowlistCheckerContext, team string, command string) bool {
	out, err := checker.ExternalTeamAllowlistRunner.Run(ctx, "sh", checker.buildShellArgs(ctx, []string{team}, command))
	if err != nil {
		return false
	}

	return checker.checkOutputResults(out)
}

func (checker *ExternalTeamAllowlistChecker) IsCommandAllowedForAnyTeam(ctx models.TeamAllowlistCheckerContext, teams []string, command string) bool {
	out, err := checker.ExternalTeamAllowlistRunner.Run(ctx, "sh", checker.buildShellArgs(ctx, teams, command))
	if err != nil {
		return false
	}

	return checker.checkOutputResults(out)
}

func (checker *ExternalTeamAllowlistChecker) AllTeams() []string {
	return []string{}
}

// buildShellArgs returns the arguments for "sh" that run the external command
// as "$external_cmd $external_args $command $repo $teams".
//
// Only the operator-configured command and args are interpreted by the shell,
// so commands that embed their own arguments keep working. The command name,
// repository full name and team names come from the VCS and are passed as
// positional parameters expanded with "$@", so the shell never parses them as
// source. Repository and project names on some VCS hosts can contain shell
// metacharacters.
func (checker *ExternalTeamAllowlistChecker) buildShellArgs(ctx models.TeamAllowlistCheckerContext, teams []string, command string) []string {
	script := strings.Join(append([]string{checker.Command}, checker.ExtraArgs...), " ") + ` "$@"`
	args := []string{"-c", script, externalTeamAllowlistArgv0, command, ctx.BaseRepo.FullName}
	for _, team := range teams {
		args = append(args, fmt.Sprintf("%s/%s", ctx.BaseRepo.Owner, team))
	}
	return args
}

func (checker *ExternalTeamAllowlistChecker) checkOutputResults(output string) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lastLine := lines[len(lines)-1]
	return strings.EqualFold(lastLine, "pass")
}
