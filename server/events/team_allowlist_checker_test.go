package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewTeamAllowListChecker(t *testing.T) {
	allowlist := `bob:plan, dave:apply`
	_, err := events.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
}

func TestNewTeamAllowListCheckerEmpty(t *testing.T) {
	allowlist := ``
	checker, err := events.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
	Equals(t, false, checker.HasRules())
}

func TestIsCommandAllowedForTeam(t *testing.T) {
	allowlist := `bob:plan, dave:apply, connie:plan, connie:apply`
	checker, err := events.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
	Equals(t, true, checker.IsCommandAllowedForTeam("connie", "plan"))
	Equals(t, true, checker.IsCommandAllowedForTeam("connie", "apply"))
	Equals(t, true, checker.IsCommandAllowedForTeam("dave", "apply"))
	Equals(t, true, checker.IsCommandAllowedForTeam("bob", "plan"))
	Equals(t, false, checker.IsCommandAllowedForTeam("bob", "apply"))
}

func TestIsCommandAllowedForAnyTeam(t *testing.T) {
	allowlist := `alpha:plan,beta:release,*:unlock,nobody:*`
	teams := []string{`alpha`, `beta`}
	checker, err := events.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
	Equals(t, true, checker.IsCommandAllowedForAnyTeam(teams, `plan`))
	Equals(t, true, checker.IsCommandAllowedForAnyTeam(teams, `release`))
	Equals(t, true, checker.IsCommandAllowedForAnyTeam(teams, `unlock`))
	Equals(t, false, checker.IsCommandAllowedForAnyTeam(teams, `noop`))
}
