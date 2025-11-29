package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

func TestResetCommandRunner_Run(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tests := []struct {
		name           string
		cleanUpErr     error
		expComment     string
		expAutoplan    bool
		expCleanUpCall bool
	}{
		{
			name:           "success - clears locks and triggers replan",
			cleanUpErr:     nil,
			expComment:     "PR state has been reset. All locks cleared and replanning triggered.",
			expAutoplan:    true,
			expCleanUpCall: true,
		},
		{
			name:           "failure - cleanup fails",
			cleanUpErr:     errors.New("cleanup failed"),
			expComment:     "Failed to reset PR state",
			expAutoplan:    false,
			expCleanUpCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			pullCleaner := mocks.NewMockPullCleaner()
			vcsClient := vcsmocks.NewMockClient()
			commandRunner := mocks.NewMockCommandRunner()

			// Create the reset command runner
			resetRunner := events.NewResetCommandRunner(pullCleaner, vcsClient)
			resetRunner.SetCommandRunner(commandRunner)

			// Setup test data
			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
			modelPull := models.PullRequest{
				BaseRepo: testdata.GithubRepo,
				State:    models.OpenPullState,
				Num:      testdata.Pull.Num,
			}
			ctx := &command.Context{
				User:     testdata.User,
				Log:      logger,
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}
			cmd := &events.CommentCommand{Name: command.Reset}

			// Setup mock behavior
			When(pullCleaner.CleanUpPull(
				Any[logging.SimpleLogging](),
				Eq(testdata.GithubRepo),
				Eq(modelPull),
			)).ThenReturn(tt.cleanUpErr)

			// Run the reset command
			resetRunner.Run(ctx, cmd)

			// Verify CleanUpPull was called
			if tt.expCleanUpCall {
				pullCleaner.VerifyWasCalledOnce().CleanUpPull(
					Any[logging.SimpleLogging](),
					Eq(testdata.GithubRepo),
					Eq(modelPull),
				)
			}

			// Verify autoplan was triggered (only on success)
			if tt.expAutoplan {
				commandRunner.VerifyWasCalledOnce().RunAutoplanCommand(
					Eq(testdata.GithubRepo),
					Eq(testdata.GithubRepo),
					Eq(modelPull),
					Eq(testdata.User),
				)
			} else {
				commandRunner.VerifyWasCalled(Never()).RunAutoplanCommand(
					Any[models.Repo](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.User](),
				)
			}

			// Verify comment was created
			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](),
				Eq(testdata.GithubRepo),
				Eq(modelPull.Num),
				Eq(tt.expComment),
				Eq("reset"),
			)
		})
	}
}

func TestResetCommandRunner_SetCommandRunner(t *testing.T) {
	RegisterMockTestingT(t)

	pullCleaner := mocks.NewMockPullCleaner()
	vcsClient := vcsmocks.NewMockClient()
	commandRunner := mocks.NewMockCommandRunner()

	resetRunner := events.NewResetCommandRunner(pullCleaner, vcsClient)

	// Initially commandRunner should be nil, so SetCommandRunner should set it
	resetRunner.SetCommandRunner(commandRunner)

	// The commandRunner is now set - we can verify this by running the reset
	// command and checking that RunAutoplanCommand is called
	logger := logging.NewNoopLogger(t)
	scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    scopeNull,
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Reset}

	When(pullCleaner.CleanUpPull(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
	)).ThenReturn(nil)

	resetRunner.Run(ctx, cmd)

	// If SetCommandRunner worked, RunAutoplanCommand should have been called
	commandRunner.VerifyWasCalledOnce().RunAutoplanCommand(
		Any[models.Repo](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[models.User](),
	)
}

func TestResetCommandRunner_CommentCreationFails(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	pullCleaner := mocks.NewMockPullCleaner()
	vcsClient := vcsmocks.NewMockClient()
	commandRunner := mocks.NewMockCommandRunner()

	resetRunner := events.NewResetCommandRunner(pullCleaner, vcsClient)
	resetRunner.SetCommandRunner(commandRunner)

	scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    scopeNull,
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Reset}

	// Setup: cleanup succeeds but comment creation fails
	When(pullCleaner.CleanUpPull(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
	)).ThenReturn(nil)

	When(vcsClient.CreateComment(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[int](),
		Any[string](),
		Any[string](),
	)).ThenReturn(errors.New("comment creation failed"))

	// Run should not panic even if comment creation fails
	Assert(t, func() bool {
		resetRunner.Run(ctx, cmd)
		return true
	}(), "Run should not panic when comment creation fails")

	// Verify cleanup and autoplan were still called
	pullCleaner.VerifyWasCalledOnce().CleanUpPull(
		Any[logging.SimpleLogging](),
		Eq(testdata.GithubRepo),
		Eq(modelPull),
	)
	commandRunner.VerifyWasCalledOnce().RunAutoplanCommand(
		Any[models.Repo](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[models.User](),
	)
}

func TestResetCommandRunner_CleanupFailsCommentFails(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	pullCleaner := mocks.NewMockPullCleaner()
	vcsClient := vcsmocks.NewMockClient()
	commandRunner := mocks.NewMockCommandRunner()

	resetRunner := events.NewResetCommandRunner(pullCleaner, vcsClient)
	resetRunner.SetCommandRunner(commandRunner)

	scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    scopeNull,
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Reset}

	// Setup: both cleanup and comment creation fail
	When(pullCleaner.CleanUpPull(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
	)).ThenReturn(errors.New("cleanup failed"))

	When(vcsClient.CreateComment(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[int](),
		Any[string](),
		Any[string](),
	)).ThenReturn(errors.New("comment creation failed"))

	// Run should not panic even when both operations fail
	Assert(t, func() bool {
		resetRunner.Run(ctx, cmd)
		return true
	}(), "Run should not panic when both cleanup and comment creation fail")

	// Verify cleanup was called
	pullCleaner.VerifyWasCalledOnce().CleanUpPull(
		Any[logging.SimpleLogging](),
		Eq(testdata.GithubRepo),
		Eq(modelPull),
	)

	// Verify autoplan was NOT called since cleanup failed
	commandRunner.VerifyWasCalled(Never()).RunAutoplanCommand(
		Any[models.Repo](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[models.User](),
	)

	// Verify error comment was attempted
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](),
		Eq(testdata.GithubRepo),
		Eq(modelPull.Num),
		Eq("Failed to reset PR state"),
		Eq("reset"),
	)
}
