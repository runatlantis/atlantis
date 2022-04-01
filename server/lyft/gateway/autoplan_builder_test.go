package gateway_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/gateway"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

var autoplanValidator gateway.AutoplanValidator
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
var projectCommandBuilder *mocks.MockProjectCommandBuilder
var drainer *events.Drainer
var commitStatusUpdater *mocks.MockCommitStatusUpdater
var workingDir *mocks.MockWorkingDir
var workingDirLocker *mocks.MockWorkingDirLocker

func setupAutoplan(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	commitStatusUpdater = mocks.NewMockCommitStatusUpdater()
	workingDir = mocks.NewMockWorkingDir()
	workingDirLocker = mocks.NewMockWorkingDirLocker()
	vcsClient := vcsmocks.NewMockClient()
	drainer = &events.Drainer{}
	pullUpdater := &events.PullUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     &events.MarkdownRenderer{},
	}
	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(nil)
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{})
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	autoplanValidator = gateway.AutoplanValidator{
		Scope:                         scope,
		VCSClient:                     vcsClient,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		Drainer:                       drainer,
		GlobalCfg:                     globalCfg,
		AllowForkPRsFlag:              "allow-fork-prs-flag",
		PullUpdater:                   pullUpdater,
		PrjCmdBuilder:                 projectCommandBuilder,
		CommitStatusUpdater:           commitStatusUpdater,
		WorkingDir:                    workingDir,
		WorkingDirLocker:              workingDirLocker,
	}
	return vcsClient
}

func TestIsValid_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	_ = setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	drainer.ShutdownBlocking()
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalled(Never()).TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_DeleteCloneError(t *testing.T) {
	t.Log("delete clone error")
	_ = setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_LockError(t *testing.T) {
	t.Log("lock error")
	_ = setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(nil, errors.New("can't lock"))
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_ProjectBuilderError(t *testing.T) {
	t.Log("project builder error")
	vcsClient := setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{}, errors.New("err"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "**Plan Error**\n```\nerr\n```\n", "plan")
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_ProjectBuilderLockError(t *testing.T) {
	t.Log("project builder lock error")
	_ = setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{}, errors.New("err"))
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(nil, errors.New("can't lock"))
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_TerraformChanges(t *testing.T) {
	t.Log("verify returns true if terraform changes exist")
	_ = setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(nil)
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)

	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == true, "should have terraform changes")
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsCommitStatus(),
		matchers.AnyModelsCommandName(),
		AnyInt(),
		AnyInt())
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestPullRequestHasTerraformChanges_NoTerraformChanges(t *testing.T) {
	t.Log("verify returns false if terraform changes don't exist")
	vcsClient := setupAutoplan(t)
	log := logging.NewNoopLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(nil)
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should have no terraform changes")
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	commitStatusUpdater.VerifyWasCalled(Times(3)).UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsCommitStatus(),
		matchers.AnyModelsCommandName(),
		AnyInt(),
		AnyInt())
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}
