package status_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/status"
	"github.com/runatlantis/atlantis/server/logging"
)

// TestInterfaceCompatibility tests that events.DefaultCommitStatusUpdater
// automatically satisfies status.CommitStatusUpdater interface
func TestInterfaceCompatibility(t *testing.T) {
	// Create a real CommitStatusUpdater from events package
	realUpdater := &events.DefaultCommitStatusUpdater{
		StatusName: "test-atlantis",
	}

	// This should compile without any explicit type conversion
	// because DefaultCommitStatusUpdater automatically satisfies the interface
	var statusUpdater status.CommitStatusUpdater = realUpdater

	// Verify the interface is satisfied (the assignment above is the real test)
	_ = statusUpdater
}

// TestStatusManagerCreation tests that we can create a StatusManager with real dependencies
func TestStatusManagerCreation(t *testing.T) {
	// Real dependencies
	realUpdater := &events.DefaultCommitStatusUpdater{
		StatusName: "test-atlantis",
	}

	policy := status.NewSilencePolicy(false, false, false, false)
	logger := logging.NewNoopLogger(t)

	// Should create successfully - testing the Go implicit interface satisfaction
	manager := status.NewStatusManager(realUpdater, policy, logger)

	if manager == nil {
		t.Fatal("failed to create StatusManager")
	}
}

// TestSilencePolicyDecisions tests that the policy makes correct decisions
func TestSilencePolicyDecisions(t *testing.T) {
	tests := []struct {
		name                       string
		silenceNoProjects          bool
		silenceVCSStatusNoPlans    bool
		silenceVCSStatusNoProjects bool
		silenceForkPRErrors        bool
		expectedOperation          status.StatusOperation
	}{
		{
			name:                       "no silence flags - should set pending",
			silenceNoProjects:          false,
			silenceVCSStatusNoPlans:    false,
			silenceVCSStatusNoProjects: false,
			silenceForkPRErrors:        false,
			expectedOperation:          status.OperationSet,
		},
		{
			name:                       "silence VCS status no projects - should silence",
			silenceNoProjects:          false,
			silenceVCSStatusNoPlans:    false,
			silenceVCSStatusNoProjects: true,
			silenceForkPRErrors:        false,
			expectedOperation:          status.OperationSilence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := status.NewSilencePolicy(
				tt.silenceNoProjects,
				tt.silenceVCSStatusNoPlans,
				tt.silenceVCSStatusNoProjects,
				tt.silenceForkPRErrors,
			)

			// Create minimal context for testing
			ctx := &command.Context{
				Log: logging.NewNoopLogger(t),
			}

			decision := policy.DecideOnStart(ctx, command.Plan)

			if decision.Operation != tt.expectedOperation {
				t.Errorf("expected operation %v, got %v", tt.expectedOperation, decision.Operation)
			}
		})
	}
}
