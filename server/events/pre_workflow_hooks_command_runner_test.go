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
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	logmocks "github.com/runatlantis/atlantis/server/logging/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

var wh events.DefaultPreWorkflowHooksCommandRunner
var whWorkingDir *mocks.MockWorkingDir
var whWorkingDirLocker *mocks.MockWorkingDirLocker
var whDrainer *events.Drainer
var whGlobalCfg valid.GlobalCfg
var whLogger *logmocks.MockSimpleLogging
var whPullLogger *logging.SimpleLogger

func preWorkflowHooksSetup(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	whLogger = logmocks.NewMockSimpleLogging()
	whWorkingDir = mocks.NewMockWorkingDir()
	whWorkingDirLocker = mocks.NewMockWorkingDirLocker()
	whDrainer = &events.Drainer{}
	pullLogger = logging.NewSimpleLogger("runatlantis/atlantis#1", true, logging.Info)
	When(whLogger.GetLevel()).ThenReturn(logging.Info)
	When(whLogger.NewLogger("runatlantis/atlantis#1", true, logging.Info)).
		ThenReturn(pullLogger)

	wh = events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             vcsClient,
		Logger:                whLogger,
		WorkingDirLocker:      whWorkingDirLocker,
		WorkingDir:            whWorkingDir,
		Drainer:               whDrainer,
		PreWorkflowHookRunner: &runtime.PreWorkflowHookRunner{},
		GlobalCfg: valid.GlobalCfg{
			Repos: []valid.Repo{
				valid.Repo{
					ID: "github.com/runatlantis/atlantis",
					PreWorkflowHooks: []*valid.PreWorkflowHook{
						&valid.PreWorkflowHook{
							StepName:   "run",
							RunCommand: `echo "exit with error"; exit 1`,
						},
					},
				},
			},
		},
	}
	return vcsClient
}

func TestPreWorkflowHooksCommandLogErrors(t *testing.T) {
	preWorkflowHooksSetup(t)

	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, strings.Contains(pullLogger.History.String(), "[EROR] Pre workflow hook run error results: exit status 1:"), "Failing pre workflow should log error")
}

func TestPreWorkflowHooksDoesntRun(t *testing.T) {
	preWorkflowHooksSetup(t)
	wh.GlobalCfg = valid.GlobalCfg{
		Repos: []valid.Repo{
			valid.Repo{
				ID:               "github.com/runatlantis/atlantis",
				PreWorkflowHooks: []*valid.PreWorkflowHook{},
			},
		},
	}

	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	whWorkingDirLocker.VerifyWasCalled(Times(0)).TryLock(AnyString(), AnyInt(), AnyString())
	whWorkingDir.VerifyWasCalled(Times(0)).Clone(matchers.AnyPtrToLoggingSimpleLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())
}

func TestPreWorkflowHooksCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	vcsClient := preWorkflowHooksSetup(t)

	When(whWorkingDir.Clone(
		pullLogger,
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

	wh.RunPreHooks(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	whWorkingDir.VerifyWasCalledOnce().Clone(pullLogger, fixtures.GithubRepo, fixtures.Pull, events.DefaultWorkspace)
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
	whWorkingDir.VerifyWasCalledOnce().Clone(pullLogger, fixtures.GithubRepo, fixtures.Pull, events.DefaultWorkspace)
	Equals(t, 0, whDrainer.GetStatus().InProgressOps)
}
