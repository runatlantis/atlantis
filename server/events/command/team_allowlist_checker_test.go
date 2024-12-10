package command_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"

	. "github.com/runatlantis/atlantis/testing"
)

func TestNewTeamAllowListChecker(t *testing.T) {
	allowlist := `bob:plan, dave:apply`
	_, err := command.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
}

func TestNewTeamAllowListCheckerEmpty(t *testing.T) {
	allowlist := ``
	checker, err := command.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
	Equals(t, false, checker.HasRules())
}

func TestIsCommandAllowedForTeam(t *testing.T) {
	allowlist := `bob:plan, dave:apply, connie:plan, connie:apply`
	checker, err := command.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
	Equals(t, true, checker.IsCommandAllowedForTeam(models.TeamAllowlistCheckerContext{}, "connie", "plan"))
	Equals(t, true, checker.IsCommandAllowedForTeam(models.TeamAllowlistCheckerContext{}, "connie", "apply"))
	Equals(t, true, checker.IsCommandAllowedForTeam(models.TeamAllowlistCheckerContext{}, "dave", "apply"))
	Equals(t, true, checker.IsCommandAllowedForTeam(models.TeamAllowlistCheckerContext{}, "bob", "plan"))
	Equals(t, false, checker.IsCommandAllowedForTeam(models.TeamAllowlistCheckerContext{}, "bob", "apply"))
}

func TestIsCommandAllowedForAnyTeam(t *testing.T) {
	allowlist := `alpha:plan,beta:release,*:unlock,nobody:*`
	teams := []string{`alpha`, `beta`}
	checker, err := command.NewTeamAllowlistChecker(allowlist)
	Ok(t, err)
	Equals(t, true, checker.IsCommandAllowedForAnyTeam(models.TeamAllowlistCheckerContext{}, teams, `plan`))
	Equals(t, true, checker.IsCommandAllowedForAnyTeam(models.TeamAllowlistCheckerContext{}, teams, `release`))
	Equals(t, true, checker.IsCommandAllowedForAnyTeam(models.TeamAllowlistCheckerContext{}, teams, `unlock`))
	Equals(t, false, checker.IsCommandAllowedForAnyTeam(models.TeamAllowlistCheckerContext{}, teams, `noop`))
}
