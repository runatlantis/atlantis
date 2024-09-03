package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/petergtz/pegomock/v4"
	runtime_mocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

var extTeamAllowlistChecker events.ExternalTeamAllowlistChecker
var extTeamAllowlistCheckerRunner *runtime_mocks.MockExternalTeamAllowlistRunner

func externalTeamAllowlistCheckerSetup(t *testing.T) {
	RegisterMockTestingT(t)
	extTeamAllowlistCheckerRunner = runtime_mocks.NewMockExternalTeamAllowlistRunner()

	extTeamAllowlistChecker = events.ExternalTeamAllowlistChecker{
		ExternalTeamAllowlistRunner: extTeamAllowlistCheckerRunner,
	}
}

func TestIsCommandAllowedForTeam(t *testing.T) {
	ctx := models.TeamAllowlistCheckerContext{
		Log: logging.NewNoopLogger(t),
	}

	t.Run("allowed", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[string](),
			Any[string]())).ThenReturn("pass\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForTeam(ctx, "foo", "plan")
		Equals(t, true, res)
	})

	t.Run("denied", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[string](),
			Any[string]())).ThenReturn("nothing found\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForTeam(ctx, "foo", "plan")
		Equals(t, false, res)
	})
}

func TestIsCommandAllowedForAnyTeam(t *testing.T) {
	ctx := models.TeamAllowlistCheckerContext{
		Log: logging.NewNoopLogger(t),
	}

	t.Run("allowed", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[string](),
			Any[string]())).ThenReturn("pass\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, []string{"foo"}, "plan")
		Equals(t, true, res)
	})

	t.Run("denied", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[string](),
			Any[string]())).ThenReturn("nothing found\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, []string{"foo"}, "plan")
		Equals(t, false, res)
	})
}
