package events_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/matchers"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
	"github.com/runatlantis/atlantis/server/events/runtime"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	logmocks "github.com/runatlantis/atlantis/server/logging/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

var wh events.DefaultPreWorkflowHooksCommandRunner
var whWorkingDir *mocks.MockWorkingDir
var whWorkingDirLocker *mocks.MockWorkingDirLocker
var whDrainer *events.Drainer

func preWorkflowHooksSetup(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logmocks.NewMockSimpleLogging()
	whWorkingDir = mocks.NewMockWorkingDir()
	whWorkingDirLocker = mocks.NewMockWorkingDirLocker()
	whDrainer = &events.Drainer{}

	wh = events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             vcsClient,
		Logger:                logger,
		WorkingDirLocker:      whWorkingDirLocker,
		WorkingDir:            whWorkingDir,
		Drainer:               whDrainer,
		PreWorkflowHookRunner: &runtime.PreWorkflowHookRunner{},
	}
	return vcsClient
}

func TestPreWorkflowHooksCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	vcsClient := preWorkflowHooksSetup(t)
	logger := wh.Logger.NewLogger("log", false, logging.LogLevel(1))

	When(whWorkingDir.Clone(
		logger,
		fixtures.GithubRepo,
		fixtures.Pull,
		events.DefaultWorkspace,
	)).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")

	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	_, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), fmt.Sprintf("comment should be about a goroutine panic but was %q", comment))
}

// Test that if one plan fails and we are using automerge, that
// we delete the plans.
func TestRunPreHooks_Clone(t *testing.T) {
	preWorkflowHooksSetup(t)
	logger := wh.Logger.NewLogger("log", false, logging.LogLevel(1))

	When(whWorkingDir.Clone(logger, fixtures.GithubRepo, fixtures.Pull, events.DefaultWorkspace)).
		ThenReturn("path/to/repo", false, nil)
	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
}

func TestRunPreHooks_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := preWorkflowHooksSetup(t)
	whDrainer.ShutdownBlocking()
	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Atlantis server is shutting down, please try again later.", "pre_workflow_hooks")
}

func TestRunPreHooks_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occured")
	preWorkflowHooksSetup(t)
	When(whWorkingDir.Clone(logger, fixtures.GithubRepo, fixtures.Pull, events.DefaultWorkspace)).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")
	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	whWorkingDir.VerifyWasCalledOnce().Clone(logger, fixtures.GithubRepo, fixtures.Pull, events.DefaultWorkspace)
	Equals(t, 0, whDrainer.GetStatus().InProgressOps)
}
