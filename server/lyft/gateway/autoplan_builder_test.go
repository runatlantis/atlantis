package gateway_test

import (
	"context"
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
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/lyft/gateway"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/vcs/markdown"
	. "github.com/runatlantis/atlantis/testing"
)

var autoplanValidator gateway.AutoplanValidator
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
var projectCommandBuilder *mocks.MockProjectCommandBuilder
var drainer *events.Drainer
var vcsStatusUpdater *mocks.MockVCSStatusUpdater
var workingDir *mocks.MockWorkingDir
var workingDirLocker *mocks.MockWorkingDirLocker
var featureAllocator *testFeatureAllocator

type testFeatureAllocator struct {
	Enabled bool
	Err     error
}

func (t *testFeatureAllocator) ShouldAllocate(featureID feature.Name, featureCtx feature.FeatureContext) (bool, error) {
	return t.Enabled, t.Err
}

func setupAutoplan(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	vcsStatusUpdater = mocks.NewMockVCSStatusUpdater()
	workingDir = mocks.NewMockWorkingDir()
	workingDirLocker = mocks.NewMockWorkingDirLocker()
	vcsClient := vcsmocks.NewMockClient()
	drainer = &events.Drainer{}
	featureAllocator = &testFeatureAllocator{}
	pullUpdater := &events.PullOutputUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     &markdown.Renderer{},
	}
	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(nil)
	globalCfg := valid.NewGlobalCfg("somedir")
	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	autoplanValidator = gateway.AutoplanValidator{
		Scope:                         scope,
		VCSClient:                     vcsClient,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		Drainer:                       drainer,
		GlobalCfg:                     globalCfg,
		OutputUpdater:                 pullUpdater,
		PrjCmdBuilder:                 projectCommandBuilder,
		VCSStatusUpdater:              vcsStatusUpdater,
		WorkingDir:                    workingDir,
		WorkingDirLocker:              workingDirLocker,
		Allocator:                     featureAllocator,
	}
	return vcsClient
}

func TestIsValid_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	_ = setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
	drainer.ShutdownBlocking()
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalled(Never()).TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_DeleteCloneError(t *testing.T) {
	t.Log("delete clone error")
	_ = setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_LockError(t *testing.T) {
	t.Log("lock error")
	_ = setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(nil, errors.New("can't lock"))
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_ProjectBuilderError(t *testing.T) {
	t.Log("project builder error")
	vcsClient := setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{}, errors.New("err"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "**Plan Error**\n```\nerr\n```\n", "plan")
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_ProjectBuilderLockError(t *testing.T) {
	t.Log("project builder lock error")
	_ = setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{}, errors.New("err"))
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(errors.New("failed to delete clone"))
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(nil, errors.New("can't lock"))
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should be false when an error occurs")
	workingDir.VerifyWasCalled(Never()).Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_TerraformChanges(t *testing.T) {
	t.Log("verify returns true if terraform changes exist")
	_ = setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
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

	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == true, "should have terraform changes")
	vcsStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsVcsStatus(),
		matchers.AnyModelsCommandName(),
		AnyInt(),
		AnyInt(),
		AnyString())
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}

func TestIsValid_TerraformChanges_PlatformMode(t *testing.T) {
	cases := []struct {
		Description           string
		WorkflowMode          valid.WorkflowModeType
		IsPlatformModeEnabled bool
	}{
		{
			Description:           "default workflow mode and platform mode is not enabled",
			WorkflowMode:          valid.DefaultWorkflowMode,
			IsPlatformModeEnabled: false,
		},
		{
			Description:           "platform workflow mode and platform mode is not enabled",
			WorkflowMode:          valid.PlatformWorkflowMode,
			IsPlatformModeEnabled: false,
		},
		{
			Description:           "default workflow mode and platform mode is enabled",
			WorkflowMode:          valid.DefaultWorkflowMode,
			IsPlatformModeEnabled: true,
		},
		{
			Description:           "platform workflow mode and platform mode is enabled",
			WorkflowMode:          valid.PlatformWorkflowMode,
			IsPlatformModeEnabled: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			_ = setupAutoplan(t)
			featureAllocator.Enabled = c.IsPlatformModeEnabled
			log := logging.NewNoopCtxLogger(t)
			When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(nil)
			When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
			When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
				ThenReturn([]command.ProjectContext{
					{
						CommandName:      command.Plan,
						WorkflowModeType: c.WorkflowMode,
					},
					{
						CommandName:      command.Plan,
						WorkflowModeType: c.WorkflowMode,
					},
				}, nil)

			containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
			Assert(t, containsTerraformChanges == true, "should have terraform changes")
			vcsStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
				matchers.AnyContextContext(),
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				matchers.AnyModelsVcsStatus(),
				matchers.AnyModelsCommandName(),
				AnyInt(),
				AnyInt(),
				AnyString())

			// Should only happen if both the conditions are met
			if c.WorkflowMode == valid.PlatformWorkflowMode && c.IsPlatformModeEnabled {
				vcsStatusUpdater.VerifyWasCalled(Once()).UpdateCombined(
					matchers.AnyContextContext(),
					matchers.AnyModelsRepo(),
					matchers.AnyModelsPullRequest(),
					matchers.AnyModelsVcsStatus(),
					matchers.AnyModelsCommandName(),
					AnyString(),
					AnyString(),
				)
			}

			workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
			workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
		})
	}
}

func TestIsValid_PreworkflowHookError(t *testing.T) {
	t.Log("verify returns false if preworkflow hook fails")
	_ = setupAutoplan(t)
	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(errors.New("error"))
	log := logging.NewNoopCtxLogger(t)
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

	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should not have terraform changes")
	vcsStatusUpdater.VerifyWasCalled(Times(1)).UpdateCombined(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsVcsStatus(),
		matchers.AnyModelsCommandName(),
		AnyString(),
		AnyString(),
	)
}

func TestPullRequestHasTerraformChanges_NoTerraformChanges(t *testing.T) {
	t.Log("verify returns false if terraform changes don't exist")
	vcsClient := setupAutoplan(t)
	log := logging.NewNoopCtxLogger(t)
	When(workingDir.Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(nil)
	When(workingDirLocker.TryLock(AnyString(), AnyInt(), AnyString())).ThenReturn(func() {}, nil)
	containsTerraformChanges := autoplanValidator.InstrumentedIsValid(context.TODO(), log, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	Assert(t, containsTerraformChanges == false, "should have no terraform changes")
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	vcsStatusUpdater.VerifyWasCalled(Times(3)).UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsVcsStatus(),
		matchers.AnyModelsCommandName(),
		AnyInt(),
		AnyInt(),
		AnyString())
	workingDir.VerifyWasCalledOnce().Delete(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())
	workingDirLocker.VerifyWasCalledOnce().TryLock(AnyString(), AnyInt(), AnyString())
}
