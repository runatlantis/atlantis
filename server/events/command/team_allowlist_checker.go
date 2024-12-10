package command

import (
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

// Wildcard matches all teams and all commands
const wildcard = "*"

// mapOfStrings is an alias for map[string]string
type mapOfStrings map[string]string

type TeamAllowlistChecker interface {
	// HasRules returns true if the checker has rules defined
	HasRules() bool

	// IsCommandAllowedForTeam determines if the specified team can perform the specified action
	IsCommandAllowedForTeam(ctx models.TeamAllowlistCheckerContext, team, command string) bool

	// IsCommandAllowedForAnyTeam determines if any of the specified teams can perform the specified action
	IsCommandAllowedForAnyTeam(ctx models.TeamAllowlistCheckerContext, teams []string, command string) bool
}

// DefaultTeamAllowlistChecker implements checking the teams and the operations that the members
// of a particular team are allowed to perform
type DefaultTeamAllowlistChecker struct {
	rules []mapOfStrings
}

// NewTeamAllowlistChecker constructs a new checker
func NewTeamAllowlistChecker(allowlist string) (*DefaultTeamAllowlistChecker, error) {
	var rules []mapOfStrings
	pairs := strings.Split(allowlist, ",")
	if pairs[0] != "" {
		for _, pair := range pairs {
			values := strings.Split(pair, ":")
			team := strings.TrimSpace(values[0])
			command := strings.TrimSpace(values[1])
			m := mapOfStrings{team: command}
			rules = append(rules, m)
		}
	}
	return &DefaultTeamAllowlistChecker{
		rules: rules,
	}, nil
}

func (checker *DefaultTeamAllowlistChecker) HasRules() bool {
	return len(checker.rules) > 0
}

// IsCommandAllowedForTeam returns true if the team is allowed to execute the command
// and false otherwise.
func (checker *DefaultTeamAllowlistChecker) IsCommandAllowedForTeam(_ models.TeamAllowlistCheckerContext, team string, command string) bool {
	for _, rule := range checker.rules {
		for key, value := range rule {
			if (key == wildcard || strings.EqualFold(key, team)) && (value == wildcard || strings.EqualFold(value, command)) {
				return true
			}
		}
	}
	return false
}

// IsCommandAllowedForAnyTeam returns true if any of the teams is allowed to execute the command
// and false otherwise.
func (checker *DefaultTeamAllowlistChecker) IsCommandAllowedForAnyTeam(ctx models.TeamAllowlistCheckerContext, teams []string, command string) bool {
	if len(teams) == 0 {
		for _, rule := range checker.rules {
			for key, value := range rule {
				if (key == wildcard) && (value == wildcard || strings.EqualFold(value, command)) {
					return true
				}
			}
		}
	} else {
		for _, t := range teams {
			if checker.IsCommandAllowedForTeam(ctx, t, command) {
				return true
			}
		}
	}
	return false
}
