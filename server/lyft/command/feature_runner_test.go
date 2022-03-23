package command_test

import (
	"testing"

	. "github.com/petergtz/pegomock"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/command/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	lyftCommand "github.com/runatlantis/atlantis/server/lyft/command"
	featureMocks "github.com/runatlantis/atlantis/server/lyft/feature/mocks"
	featureMatchers "github.com/runatlantis/atlantis/server/lyft/feature/mocks/matchers"
)

var dbUpdater *events.DBUpdater
var pullUpdater *events.PullUpdater
var autoMerger *events.AutoMerger
var policyCheckCommandRunner *events.PolicyCheckCommandRunner
var planCommandRunner *events.PlanCommandRunner
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner

func TestFeatureAllocatorRunner(t *testing.T) {
	featureAllocator := featureMocks.NewMockAllocator()
	testLogger := logging.NewNoopLogger(t)
	// platformModeRunner := plan.NewRunner(vcsClient)
	cases := []struct {
		description         string
		platformModeEnabled bool
		allocated           bool
		err                 error
	}{
		{
			description:         "feature allocated and platform mode enabled",
			platformModeEnabled: true,
			allocated:           true,
		},
		{
			description:         "feature allocated and platform mode disabled",
			platformModeEnabled: false,
			allocated:           true,
		},
		{
			description:         "feature not allocated and platform mode enabled",
			platformModeEnabled: true,
			allocated:           false,
		},
		{
			description:         "feature not allocated and platform mode disabled",
			platformModeEnabled: false,
			allocated:           false,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)

			allocatedRunner := mocks.NewMockRunner()
			unallocatedRunner := mocks.NewMockRunner()
			featuredRunner := lyftCommand.NewPlatformModeFeatureRunner(
				featureAllocator,
				c.platformModeEnabled,
				testLogger,
				allocatedRunner,
				unallocatedRunner,
			)

			When(featureAllocator.ShouldAllocate(
				featureMatchers.AnyFeatureName(),
				AnyString(),
			)).ThenReturn(c.allocated, c.err)

			ctx := &command.Context{
				HeadRepo: models.Repo{
					FullName: "test-repo",
				},
			}
			cmd := &command.Comment{}
			featuredRunner.Run(ctx, cmd)

			if c.platformModeEnabled && c.allocated {
				allocatedRunner.VerifyWasCalledOnce().Run(ctx, cmd)
				unallocatedRunner.VerifyWasCalled(Never()).Run(ctx, cmd)
			} else {
				allocatedRunner.VerifyWasCalled(Never()).Run(ctx, cmd)
				unallocatedRunner.VerifyWasCalledOnce().Run(ctx, cmd)
			}
		})

	}
}
