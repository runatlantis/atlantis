package gateway_test

import (
	"errors"
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
	"testing"
)

var autoplanValidator gateway.AutoplanValidator
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
var projectCommandBuilder *mocks.MockProjectCommandBuilder
var drainer *events.Drainer
var commitStatusUpdater *mocks.MockCommitStatusUpdater
var workingDir *mocks.MockWorkingDir

func setupAutoplan(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	commitStatusUpdater = mocks.NewMockCommitStatusUpdater()
	workingDir = mocks.NewMockWorkingDir()
	vcsClient := vcsmocks.NewMockClient()
	drainer = &events.Drainer{}
	pullUpdater := &events.PullUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     &events.MarkdownRenderer{},
	}
	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyPtrToEventsCommandContext())).ThenReturn(nil)
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{})
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	autoplanValidator = gateway.AutoplanValidator{
		Logger:                        logger,
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
	}
	return vcsClient
}

func TestIsValid_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	_ = setupAutoplan(t)
	drainer.ShutdownBlocking()
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
}

func TestIsValid_DeleteCloneError(t *testing.T) {
	t.Log("delete clone error")
	_ = setupAutoplan(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
}

func TestIsValid_ProjectBuilderError(t *testing.T) {
	t.Log("project builder error")
	vcsClient := setupAutoplan(t)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{}, errors.New("err"))
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "**Plan Error**\n```\nerr\n```\n", "plan")
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
}

func TestIsValid_TerraformChanges(t *testing.T) {
	t.Log("verify returns true if terraform changes exist")
	_ = setupAutoplan(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(nil)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)

	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == true, "should have terraform changes")
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsCommitStatus(),
		matchers.AnyModelsCommandName(),
		AnyInt(),
		AnyInt())
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
}

func TestPullRequestHasTerraformChanges_NoTerraformChanges(t *testing.T) {
	t.Log("verify returns false if terraform changes don't exist")
	vcsClient := setupAutoplan(t)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(nil)
	Assert(t, containsTerraformChanges == false, "should have no terraform changes")
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	commitStatusUpdater.VerifyWasCalled(Times(3)).UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsCommitStatus(),
		matchers.AnyModelsCommandName(),
		AnyInt(),
		AnyInt())
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
}
