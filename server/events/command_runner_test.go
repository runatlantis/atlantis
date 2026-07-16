// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"

	"github.com/google/go-github/v88/github"
	. "github.com/petergtz/pegomock/v4"
	lockingmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

var projectCommandBuilder *mocks.MockProjectCommandBuilder
var projectCommandRunner *mocks.MockProjectCommandRunner
var eventParsing *mocks.MockEventParsing
var azuredevopsGetter *mocks.MockAzureDevopsPullGetter
var githubGetter *mocks.MockGithubPullGetter
var gitlabGetter *mocks.MockGitlabMergeRequestGetter
var ch events.DefaultCommandRunner
var workingDir events.WorkingDir
var pendingPlanFinder *mocks.MockPendingPlanFinder
var drainer *events.Drainer
var deleteLockCommand *mocks.MockDeleteLockCommand
var commitUpdater *mocks.MockCommitStatusUpdater
var pullReqStatusFetcher *vcsmocks.MockPullReqStatusFetcher

// TODO: refactor these into their own unit tests.
// these were all split out from default command runner in an effort to improve
// readability however the tests were kept as is.
var dbUpdater *events.DBUpdater
var pullUpdater *events.PullUpdater
var autoMerger *events.AutoMerger
var policyCheckCommandRunner *events.PolicyCheckCommandRunner
var approvePoliciesCommandRunner *events.ApprovePoliciesCommandRunner
var planCommandRunner *events.PlanCommandRunner
var applyLockChecker *lockingmocks.MockApplyLockChecker
var lockingLocker *lockingmocks.MockLocker
var applyCommandRunner *events.ApplyCommandRunner
var unlockCommandRunner *events.UnlockCommandRunner
var importCommandRunner *events.ImportCommandRunner
var stateCommandRunner *events.StateCommandRunner
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
var postWorkflowHooksCommandRunner events.PostWorkflowHooksCommandRunner
var cancellationTracker *mocks.MockCancellationTracker

type TestConfig struct {
	parallelPoolSize           int
	SilenceNoProjects          bool
	silenceVCSStatusNoPlans    bool
	silenceVCSStatusNoProjects bool
	StatusName                 string
	discardApprovalOnPlan      bool
	database                   db.Database
	DisableUnlockLabel         string
	DisableAutomergeLabel      string
	PendingApplyStatus         bool
	applyLockCheckerReturn     locking.ApplyCommandLock
	applyLockCheckerErr        error
	workingDirLocker           events.WorkingDirLocker
	planLocker                 locking.Locker
	livePullHeadFetcher        events.LivePullHeadFetcher
}

type configuredPreWorkflowHooksCommandRunner struct {
	hasHooks bool
	err      error
	calls    int
}

func (r *configuredPreWorkflowHooksCommandRunner) RunPreHooks(_ *command.Context, _ *events.CommentCommand) error {
	r.calls++
	return r.err
}

func (r *configuredPreWorkflowHooksCommandRunner) HasPreWorkflowHooks(_ *command.Context) bool {
	return r.hasHooks
}

func setup(t *testing.T, options ...func(testConfig *TestConfig)) *vcsmocks.MockClient {
	RegisterMockTestingT(t)

	// create an empty DB
	tmp := t.TempDir()
	defaultBoltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		defaultBoltDB.Close()
	})
	Ok(t, err)

	testConfig := &TestConfig{
		parallelPoolSize:      1,
		SilenceNoProjects:     false,
		StatusName:            "atlantis-test",
		discardApprovalOnPlan: false,
		database:              defaultBoltDB,
		DisableUnlockLabel:    "do-not-unlock",
	}

	for _, op := range options {
		op(testConfig)
	}

	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	eventParsing = mocks.NewMockEventParsing()
	vcsClient := vcsmocks.NewMockClient()
	githubGetter = mocks.NewMockGithubPullGetter()
	gitlabGetter = mocks.NewMockGitlabMergeRequestGetter()
	azuredevopsGetter = mocks.NewMockAzureDevopsPullGetter()
	logger := logging.NewNoopLogger(t)
	projectCommandRunner = mocks.NewMockProjectCommandRunner()
	workingDir = mocks.NewMockWorkingDir()
	pendingPlanFinder = mocks.NewMockPendingPlanFinder()
	commitUpdater = mocks.NewMockCommitStatusUpdater()
	pullReqStatusFetcher = vcsmocks.NewMockPullReqStatusFetcher()
	cancellationTracker = mocks.NewMockCancellationTracker()

	drainer = &events.Drainer{}
	deleteLockCommand = mocks.NewMockDeleteLockCommand()
	lockCtrl := gomock.NewController(t)
	applyLockChecker = lockingmocks.NewMockApplyLockChecker(lockCtrl)
	lockingLocker = lockingmocks.NewMockLocker(lockCtrl)
	// Allow incidental calls to CheckApplyLock (called internally during apply operations).
	// Tests that need specific return values should set applyLockCheckerReturn/applyLockCheckerErr in TestConfig.
	applyLockChecker.EXPECT().CheckApplyLock().Return(testConfig.applyLockCheckerReturn, testConfig.applyLockCheckerErr).AnyTimes()
	// Plan cleanup only deletes selected on_plan locks through owner-scoped cleanup.
	lockingLocker.EXPECT().UnlockIfOwnedByPull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	lockingLocker.EXPECT().GetLock(gomock.Any()).Return(nil, nil).AnyTimes()
	planLocker := testConfig.planLocker
	if planLocker == nil {
		planLocker = lockingLocker
	}

	dbUpdater = &events.DBUpdater{
		Database: testConfig.database,
	}

	pullUpdater = &events.PullUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     events.NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}

	autoMerger = &events.AutoMerger{
		VCSClient:       vcsClient,
		GlobalAutomerge: false,
	}

	policyCheckCommandRunner = events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		commitUpdater,
		projectCommandRunner,
		testConfig.parallelPoolSize,
		testConfig.silenceVCSStatusNoProjects,
		false,
	)

	workingDirLocker := testConfig.workingDirLocker
	if workingDirLocker == nil {
		workingDirLocker = events.NewDefaultWorkingDirLocker()
	}
	planCommandRunner = events.NewPlanCommandRunner(
		testConfig.silenceVCSStatusNoPlans,
		testConfig.silenceVCSStatusNoProjects,
		vcsClient,
		pendingPlanFinder,
		workingDir,
		workingDirLocker,
		commitUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		cancellationTracker,
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		autoMerger,
		testConfig.parallelPoolSize,
		testConfig.SilenceNoProjects,
		testConfig.database,
		planLocker,
		testConfig.discardApprovalOnPlan,
		pullReqStatusFetcher,
		testConfig.PendingApplyStatus,
	)

	applyCommandRunner = events.NewApplyCommandRunner(
		vcsClient,
		false,
		applyLockChecker,
		commitUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		cancellationTracker,
		autoMerger,
		pullUpdater,
		dbUpdater,
		testConfig.database,
		testConfig.parallelPoolSize,
		testConfig.SilenceNoProjects,
		testConfig.silenceVCSStatusNoProjects,
		workingDirLocker,
		pullReqStatusFetcher,
		testConfig.livePullHeadFetcher,
		testConfig.DisableAutomergeLabel,
	)

	approvePoliciesCommandRunner = events.NewApprovePoliciesCommandRunner(
		commitUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		pullUpdater,
		dbUpdater,
		testConfig.SilenceNoProjects,
		testConfig.silenceVCSStatusNoProjects,
		vcsClient,
	)

	unlockCommandRunner = events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
		testConfig.SilenceNoProjects,
		testConfig.DisableUnlockLabel,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		testConfig.parallelPoolSize,
		testConfig.SilenceNoProjects,
	)

	importCommandRunner = events.NewImportCommandRunner(
		pullUpdater,
		dbUpdater,
		pullReqStatusFetcher,
		projectCommandBuilder,
		projectCommandRunner,
		testConfig.SilenceNoProjects,
	)

	stateCommandRunner = events.NewStateCommandRunner(
		pullUpdater,
		dbUpdater,
		projectCommandBuilder,
		projectCommandRunner,
	)

	commentCommandRunnerByCmd := map[command.Name]events.CommentCommandRunner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
		command.Import:          importCommandRunner,
		command.State:           stateCommandRunner,
	}

	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	postWorkflowHooksCommandRunner = mocks.NewMockPostWorkflowHooksCommandRunner()

	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{})
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	ch = events.DefaultCommandRunner{
		VCSClient:                      vcsClient,
		CommentCommandRunnerByCmd:      commentCommandRunnerByCmd,
		EventParser:                    eventParsing,
		FailOnPreWorkflowHookError:     false,
		GithubPullGetter:               githubGetter,
		GitlabMergeRequestGetter:       gitlabGetter,
		AzureDevopsPullGetter:          azuredevopsGetter,
		Logger:                         logger,
		StatsScope:                     scope,
		GlobalCfg:                      globalCfg,
		AllowForkPRs:                   false,
		AllowForkPRsFlag:               "allow-fork-prs-flag",
		Drainer:                        drainer,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		PullStatusFetcher:              testConfig.database,
		CommitStatusUpdater:            commitUpdater,
	}

	return vcsClient
}

func TestRunCommentCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	vcsClient := setup(t)
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenPanic(
		"panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, 1, &events.CommentCommand{Name: command.Plan})
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), fmt.Sprintf("comment should be about a goroutine panic but was %q", comment))
}

func TestRunCommentCommand_GithubPullErr(t *testing.T) {
	t.Log("if getting the github pull request fails an error should be logged")
	vcsClient := setup(t)
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("`Error: making pull request API call to GitHub: err`"), Eq(""))
}

func TestRunCommentCommand_GitlabMergeRequestErr(t *testing.T) {
	t.Log("if getting the gitlab merge request fails an error should be logged")
	vcsClient := setup(t)
	When(gitlabGetter.GetMergeRequest(Any[logging.SimpleLogging](), Eq(testdata.GitlabRepo.FullName), Eq(testdata.Pull.Num))).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(testdata.GitlabRepo, &testdata.GitlabRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GitlabRepo), Eq(testdata.Pull.Num), Eq("`Error: making merge request API call to GitLab: err`"), Eq(""))
}

func TestRunCommentCommand_GithubPullParseErr(t *testing.T) {
	t.Log("if parsing the returned github pull request fails an error should be logged")
	vcsClient := setup(t)
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(testdata.Pull, testdata.GithubRepo, testdata.GitlabRepo, errors.New("err"))

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("`Error: extracting required fields from comment data: err`"), Eq(""))
}

func TestRunCommentCommand_CommentPreWorkflowHookFailure(t *testing.T) {
	t.Log("if there is a pre-workflowhook failure with failure enabled it is commented back on the pull request")
	vcsClient := setup(t)
	ch.FailOnPreWorkflowHookError = true

	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, 1, &events.CommentCommand{Name: command.Plan})
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()

	Assert(t, strings.Contains(comment, "Error: Pre-workflow hook failed"), fmt.Sprintf("comment should be about a pre-workflow hook failure but was %q", comment))
}

func TestRunCommentCommand_TeamAllowListChecker(t *testing.T) {
	t.Run("nil checker", func(t *testing.T) {
		vcsClient := setup(t)
		// by default these are false so don't need to reset
		ch.TeamAllowlistChecker = nil
		var pull github.PullRequest
		modelPull := models.PullRequest{
			BaseRepo: testdata.GithubRepo,
			State:    models.OpenPullState,
		}
		When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
		When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

		ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
		vcsClient.VerifyWasCalled(Never()).GetTeamNamesForUser(ch.Logger, testdata.GithubRepo, testdata.User)
		vcsClient.VerifyWasCalledOnce().CreateComment(
			Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq("Ran Plan for 0 projects:"), Eq("plan"))
	})

	t.Run("no rules", func(t *testing.T) {
		vcsClient := setup(t)
		// by default these are false so don't need to reset
		ch.TeamAllowlistChecker = &command.DefaultTeamAllowlistChecker{}
		var pull github.PullRequest
		modelPull := models.PullRequest{
			BaseRepo: testdata.GithubRepo,
			State:    models.OpenPullState,
		}
		When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
		When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

		ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
		vcsClient.VerifyWasCalled(Never()).GetTeamNamesForUser(ch.Logger, testdata.GithubRepo, testdata.User)
		vcsClient.VerifyWasCalledOnce().CreateComment(
			Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq("Ran Plan for 0 projects:"), Eq("plan"))
	})
}

func TestRunCommentCommand_ForkPRDisabled(t *testing.T) {
	t.Log("if a command is run on a forked pull request and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	// by default these are false so don't need to reset
	ch.AllowForkPRs = false
	ch.SilenceForkPRErrors = false
	var pull github.PullRequest
	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
	}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)

	headRepo := testdata.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, headRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	commentMessage := fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s  or, to disable this message, set --%s", ch.AllowForkPRsFlag, ch.SilenceForkPRErrorsFlag)
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(commentMessage), Eq(""))
}

func TestRunCommentCommand_ForkPRDisabled_SilenceEnabled(t *testing.T) {
	t.Log("if a command is run on a forked pull request and forks are disabled and we are silencing errors do not comment with error")
	vcsClient := setup(t)
	ch.AllowForkPRs = false // by default it's false so don't need to reset
	ch.SilenceForkPRErrors = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)

	headRepo := testdata.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, headRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunCommentCommandPlan_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if a plan command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	planCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.Plan),
		Eq(models.ProjectCounts{}),
	)
}

func TestRunCommentCommandPlan_NoProjectsWritesCurrentEmptyPullStatus(t *testing.T) {
	_ = setup(t)
	planCommandRunner.SilenceNoProjects = true

	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "old-project",
			Workspace:  events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)

	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunCommentCommandPlan_NoProjectsMissingPullDirWritesEmptyPullStatus(t *testing.T) {
	_ = setup(t)
	planCommandRunner.SilenceNoProjects = true

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn("", os.ErrNotExist)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunCommentCommandPlan_NoProjectsRetainsOldPlanFilesAndClearsPullStatus(t *testing.T) {
	_ = setup(t)
	planCommandRunner.SilenceNoProjects = true

	tmp := t.TempDir()
	oldPlanDir := filepath.Join(tmp, "old-project")
	Ok(t, os.MkdirAll(oldPlanDir, 0700))
	oldPlanPath := filepath.Join(oldPlanDir, runtime.GetPlanFilename(events.DefaultWorkspace, ""))
	Ok(t, os.WriteFile(oldPlanPath, []byte("old plan"), 0600))
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "old-project",
			Workspace:  events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)

	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	contents, err := os.ReadFile(oldPlanPath)
	Ok(t, err)
	Equals(t, "old plan", string(contents))
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunCommentCommandPlan_NoProjectsPreservesActiveGenerationPlan(t *testing.T) {
	vcsClient := setup(t)
	planCommandRunner.SilenceNoProjects = true

	modelPull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "abc123",
		BaseBranch: "main",
	}
	project := models.ProjectStatus{
		Workspace:   events.DefaultWorkspace,
		RepoRelDir:  "active-project",
		ProjectName: "active-project",
	}
	_, err := dbUpdater.Database.BeginPlanGeneration(modelPull, []models.ProjectStatus{project}, "generation-1")
	Ok(t, err)

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, runtime.GetPlanFilename(project.Workspace, project.ProjectName))
	marker := []byte("active generation marker")
	Ok(t, os.WriteFile(planPath, marker, 0600))

	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Plan}
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(nil, nil)

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected no-project replacement to reject the active generation")
	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	contents, readErr := os.ReadFile(planPath)
	Ok(t, readErr)
	Equals(t, marker, contents)
	status, statusErr := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, statusErr)
	Assert(t, status != nil, "expected active generation status")
	active := projectStatus(t, status, project.Workspace, project.RepoRelDir, project.ProjectName)
	Equals(t, models.ErroredPlanStatus, active.Status)
	Equals(t, "generation-1", active.PlanGeneration)
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(modelPull.BaseRepo), Eq(modelPull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "active plan generation"), "got: %s", comment)
}

func TestRunCommentCommandPlan_NoProjectsPreservesGenerationStartedAfterReplacement(t *testing.T) {
	underlying := newTestBoltDB(t)
	modelPull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "abc123",
		BaseBranch: "main",
	}
	project := models.ProjectStatus{
		Workspace:   events.DefaultWorkspace,
		RepoRelDir:  "active-project",
		ProjectName: "active-project",
	}
	planRoot := t.TempDir()
	planDir := filepath.Join(planRoot, project.RepoRelDir)
	Ok(t, os.MkdirAll(planDir, 0700))
	planPath := filepath.Join(planDir, runtime.GetPlanFilename(project.Workspace, project.ProjectName))
	marker := []byte("generation started after replacement")
	database := &beginGenerationAfterReplaceDB{
		Database: underlying,
		project:  project,
		planPath: planPath,
		marker:   marker,
	}
	_ = setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	planCommandRunner.SilenceNoProjects = true

	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Plan}
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(nil, nil)

	planCommandRunner.Run(ctx, cmd)

	Assert(t, database.started, "expected generation to begin after empty status replacement")
	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	contents, err := os.ReadFile(planPath)
	Ok(t, err)
	Equals(t, marker, contents)
	status, err := underlying.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, status != nil, "expected active generation status")
	active := projectStatus(t, status, project.Workspace, project.RepoRelDir, project.ProjectName)
	Equals(t, models.ErroredPlanStatus, active.Status)
	Assert(t, active.PlanGeneration != "", "expected active generation marker")
}

func TestRunCommentCommandPlan_NoProjectsNonSilencedRetainsOldPlanFiles(t *testing.T) {
	_ = setup(t)

	tmp := t.TempDir()
	oldPlanDir := filepath.Join(tmp, "old-project")
	Ok(t, os.MkdirAll(oldPlanDir, 0700))
	oldPlanPath := filepath.Join(oldPlanDir, runtime.GetPlanFilename(events.DefaultWorkspace, ""))
	Ok(t, os.WriteFile(oldPlanPath, []byte("old plan"), 0600))
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "old-project",
			Workspace:  events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)

	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	contents, err := os.ReadFile(oldPlanPath)
	Ok(t, err)
	Equals(t, "old plan", string(contents))
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunCommentCommandPlan_NoProjectsDoesNotUnlockActiveLocks(t *testing.T) {
	vcsClient := setup(t)
	installPlanCommandRunnerLocker(vcsClient, failUnlockByPullLocker{t: t}, true)

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
}

func TestRunCommentCommandPlan_NoProjectsNonSilencedDoesNotUnlockByPull(t *testing.T) {
	vcsClient := setup(t)
	installPlanCommandRunnerLocker(vcsClient, failUnlockByPullLocker{t: t}, false)

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
}

func TestRunCommentCommandPlan_NoProjectsEmptyPullStatusWriteFailureIsUserVisible(t *testing.T) {
	vcsClient := setup(t)
	planCommandRunner.SilenceNoProjects = true

	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	cmd := &events.CommentCommand{Name: command.Plan}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	closer := dbUpdater.Database.(interface{ Close() error })
	Ok(t, closer.Close())

	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected empty PullStatus write failure to mark command errored")
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Plan))
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Plan Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "writing empty plan status"), "got: %s", comment)
}

func TestRunCommentCommandPlan_NoProjectsTarget_SilenceEnabled(t *testing.T) {
	// TODO
	t.Log("if a plan command is run against a project and SilenceNoProjects is enabled, we are silencing all comments if the project is not in the repo config")
	vcsClient := setup(t)
	planCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan, ProjectName: "meow"})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.Plan),
		Eq(models.ProjectCounts{}),
	)
}

func TestRunCommentCommandApply_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an apply command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	applyCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123", BaseBranch: "main"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, nil)
	Ok(t, err)
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.Apply),
		Eq(models.ProjectCounts{}),
	)
}

func TestRunApply_NoProjectsAfterEmptyPullStatusNoOpsSafely(t *testing.T) {
	vcsClient := setup(t)
	applyCommandRunner.SilenceNoProjects = true

	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, nil)
	Ok(t, err)

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(args []Param) ReturnValues {
		ctx := args[0].(*command.Context)
		Assert(t, ctx.PullStatus != nil, "expected command context to include empty PullStatus")
		Equals(t, 0, len(ctx.PullStatus.Projects))
		return ReturnValues{[]command.ProjectContext{}, nil}
	})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.Apply),
		Eq(models.ProjectCounts{}),
	)
}

func TestRunApply_NoProjectsAfterEmptyPullStatusDoesNotSucceedWhilePlanInFlight(t *testing.T) {
	vcsClient := setup(t)
	applyCommandRunner.SilenceNoProjects = true

	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, nil)
	Ok(t, err)

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn(nil, fmt.Errorf("a plan is currently running for this pull request; wait for it to finish before applying"))

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Apply))
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("apply")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Apply Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "plan is currently running"), "got: %s", comment)
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.SuccessCommitStatus), Eq(command.Apply), Eq(models.ProjectCounts{}))
}

func TestRunApply_DoesNotReturnZeroProjectsWhilePlanInFlight(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.SilenceNoProjects = true
		tc.workingDirLocker = locker
	})
	unlock, err := locker.TryLockPull(testdata.GithubRepo.FullName, testdata.Pull.Num, command.Plan)
	Ok(t, err)
	defer unlock()

	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	projectCommandBuilder.VerifyWasCalled(Never()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Apply))
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("apply")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Apply Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "currently locked by"), "got: %s", comment)
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.SuccessCommitStatus), Eq(command.Apply), Eq(models.ProjectCounts{}))
}

func TestPlanCommandRunner_HoldsPlanInFlightLockAcrossCleanupPlanAndDBWrite(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	setup(t, func(tc *TestConfig) {
		tc.workingDirLocker = locker
	})
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	cmd := &events.CommentCommand{Name: command.Plan}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  ".",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Eq(cmd))).Then(func(args []Param) ReturnValues {
		Assert(t, locker.HasCommandLock(testdata.GithubRepo.FullName, testdata.Pull.Num, command.Plan), "expected plan lock during command build")
		return ReturnValues{[]command.ProjectContext{projectCtx}, nil}
	})
	When(projectCommandRunner.Plan(projectCtx)).Then(func(args []Param) ReturnValues {
		Assert(t, locker.HasCommandLock(testdata.GithubRepo.FullName, testdata.Pull.Num, command.Plan), "expected plan lock during project plan")
		return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
	})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

	Assert(t, !locker.HasCommandLock(testdata.GithubRepo.FullName, testdata.Pull.Num, command.Plan), "expected plan lock to be released")
}

func TestPlanCommandRunner_HoldsPlanLockDuringStalePlanCleanup(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	setup(t, func(tc *TestConfig) {
		tc.workingDirLocker = locker
	})
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	cmd := &events.CommentCommand{Name: command.Plan}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  ".",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	tmp := t.TempDir()

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Eq(cmd))).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).Then(func([]Param) ReturnValues {
		Assert(t, locker.HasCommandLock(testdata.GithubRepo.FullName, testdata.Pull.Num, command.Plan), "expected plan lock during stale plan cleanup")
		return ReturnValues{[]events.PendingPlan{}, nil}
	})
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
}

func TestPlanCommandRunner_HoldsPlanLockDuringPullStatusWrite(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	realDB := newTestBoltDB(t)
	writeObserved := false
	database := assertPlanLockDB{
		Database:     realDB,
		t:            t,
		locker:       locker,
		repoFullName: testdata.GithubRepo.FullName,
		pullNum:      testdata.Pull.Num,
		called:       &writeObserved,
	}
	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.workingDirLocker = locker
	})
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	cmd := &events.CommentCommand{Name: command.Plan}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  ".",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	tmp := t.TempDir()

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Eq(cmd))).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

	Assert(t, writeObserved, "expected pull status write to be observed")
}

func TestPlanCommandRunner_AutoplanHoldsPlanLockDuringStalePlanCleanup(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	setup(t, func(tc *TestConfig) {
		tc.workingDirLocker = locker
	})
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  ".",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	tmp := t.TempDir()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).Then(func([]Param) ReturnValues {
		Assert(t, locker.HasCommandLock(testdata.GithubRepo.FullName, testdata.Pull.Num, command.Plan), "expected autoplan lock during stale plan cleanup")
		return ReturnValues{[]events.PendingPlan{}, nil}
	})
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
}

func TestPlanCommandRunner_AutoplanHoldsPlanLockDuringPullStatusWrite(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	realDB := newTestBoltDB(t)
	writeObserved := false
	database := assertPlanLockDB{
		Database:     realDB,
		t:            t,
		locker:       locker,
		repoFullName: testdata.GithubRepo.FullName,
		pullNum:      testdata.Pull.Num,
		called:       &writeObserved,
	}
	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.workingDirLocker = locker
	})
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  ".",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	tmp := t.TempDir()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	Assert(t, writeObserved, "expected autoplan pull status write to be observed")
}

func TestPlanCommandRunner_AutoplanPersistsPullStatusBeforePublishingSuccess(t *testing.T) {
	database := newTestBoltDB(t)
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	modelPull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "abc123",
		BaseBranch: "main",
	}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "gitea-autoplan",
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	tmp := t.TempDir()
	statusVisibleWhenCommented := false

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(vcsClient.CreateComment(
		Any[logging.SimpleLogging](),
		Eq(testdata.GithubRepo),
		Eq(modelPull.Num),
		Any[string](),
		Eq("plan"),
	)).Then(func([]Param) ReturnValues {
		pullStatus, err := database.GetPullStatus(modelPull)
		Ok(t, err)
		if pullStatus != nil {
			for _, project := range pullStatus.Projects {
				if project.Workspace == projectCtx.Workspace &&
					project.RepoRelDir == projectCtx.RepoRelDir &&
					project.ProjectName == projectCtx.ProjectName &&
					project.Status == models.PlannedPlanStatus {
					statusVisibleWhenCommented = true
					break
				}
			}
		}
		return ReturnValues{nil}
	})

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	Assert(t, statusVisibleWhenCommented, "expected persisted PullStatus before successful autoplan comment was published")
}

func TestPlanCommandRunner_AutoplanCompletionPrecedesSuccessVisibilityAndImmediateApply(t *testing.T) {
	underlying := newTestBoltDB(t)
	database := &blockingPlanCompletionDB{
		Database: underlying,
		started:  make(chan struct{}),
		release:  make(chan struct{}),
	}
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
	}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "immediate-autoplan-apply",
		BaseRepo:    pull.BaseRepo,
		Pull:        pull,
		Steps:       valid.DefaultPlanStage.Steps,
	}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.AutoTrigger,
	}
	tmp := t.TempDir()
	var successCommentVisible atomic.Bool
	var successStatusVisible atomic.Bool
	When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(vcsClient.CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("plan"),
	)).Then(func([]Param) ReturnValues {
		successCommentVisible.Store(true)
		return ReturnValues{nil}
	})
	When(commitUpdater.UpdateCombinedCount(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(models.SuccessCommitStatus), Eq(command.Plan), Any[models.ProjectCounts](),
	)).Then(func([]Param) ReturnValues {
		successStatusVisible.Store(true)
		return ReturnValues{nil}
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		planCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Plan})
	}()
	<-database.started

	Assert(t, !successCommentVisible.Load(), "successful plan comment became visible before PullStatus completion")
	Assert(t, !successStatusVisible.Load(), "successful plan status became visible before PullStatus completion")
	inProgress, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, inProgress != nil, "expected durable in-progress PullStatus")
	Equals(t, models.ErroredPlanStatus, inProgress.Projects[0].Status)
	Assert(t, inProgress.Projects[0].PlanGeneration != "", "expected active generation marker")

	close(database.release)
	<-done
	Assert(t, successCommentVisible.Load(), "expected successful plan comment after PullStatus completion")
	Assert(t, successStatusVisible.Load(), "expected successful plan status after PullStatus completion")
	persisted, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, persisted != nil, "expected persisted PullStatus")
	Equals(t, models.PlannedPlanStatus, persisted.Projects[0].Status)

	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockApply := mocks.NewMockStepRunner()
	repoDir := t.TempDir()
	applyCtx := projectCtx
	applyCtx.CommandName = command.Apply
	applyCtx.Steps = valid.DefaultApplyStage.Steps
	applyCtx.Log = logging.NewNoopLogger(t)
	projectDir := filepath.Join(repoDir, applyCtx.RepoRelDir)
	Ok(t, os.MkdirAll(projectDir, 0755))
	planPath := filepath.Join(projectDir, runtime.GetPlanFilename(applyCtx.Workspace, applyCtx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("persisted autoplan generation"), 0600))
	When(mockWorkingDir.GetWorkingDir(applyCtx.Pull.BaseRepo, applyCtx.Pull, applyCtx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(applyCtx.Pull.BaseRepo, applyCtx.Pull, applyCtx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(applyCtx.Pull), Any[models.User](), Eq(applyCtx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)
	When(mockApply.Run(Any[command.ProjectContext](), Any[[]string](), Eq(projectDir), Any[map[string]string]())).ThenReturn("applied immediate autoplan", nil)
	applyRunner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: database},
	}

	applyResult := applyRunner.Apply(applyCtx)

	Ok(t, applyResult.Error)
	Equals(t, "applied immediate autoplan", applyResult.ApplySuccess)
}

func TestApprovePoliciesCommandRunner_CannotReplaceActivePlanGeneration(t *testing.T) {
	underlying := newTestBoltDB(t)
	database := &blockingPlanCompletionDB{
		Database: underlying,
		started:  make(chan struct{}),
		release:  make(chan struct{}),
	}
	released := false
	defer func() {
		if !released {
			close(database.release)
		}
	}()
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
	}
	planCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "concurrent-policy-approval",
		BaseRepo:    pull.BaseRepo,
		Pull:        pull,
		Steps:       valid.DefaultPlanStage.Steps,
	}
	_, err := underlying.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(planCtx.RepoRelDir, planCtx.Workspace, planCtx.ProjectName),
	})
	Ok(t, err)
	approveCtx := planCtx
	approveCtx.CommandName = command.ApprovePolicies
	approveCtx.ProjectPolicyStatus = []models.PolicySetStatus{{
		PolicySetName: "production",
		Hashes:        []string{"policy-hash"},
	}}
	planCommandCtx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	planCmd := &events.CommentCommand{Name: command.Plan, ProjectName: planCtx.ProjectName}
	When(projectCommandBuilder.BuildPlanCommands(planCommandCtx, planCmd)).ThenReturn([]command.ProjectContext{planCtx}, nil)
	When(projectCommandRunner.Plan(planCtx)).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "final generation plan"},
	})

	planDone := make(chan struct{})
	go func() {
		defer close(planDone)
		planCommandRunner.Run(planCommandCtx, planCmd)
	}()
	<-database.started

	activeBefore, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, activeBefore != nil, "expected active generation")
	activeProjectBefore := projectStatus(t, activeBefore, planCtx.Workspace, planCtx.RepoRelDir, planCtx.ProjectName)
	Assert(t, activeProjectBefore.PlanGeneration != "", "expected nonempty generation")
	activeGeneration := activeProjectBefore.PlanGeneration

	approveCommandCtx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	approveCmd := &events.CommentCommand{Name: command.ApprovePolicies, ProjectName: approveCtx.ProjectName}
	When(projectCommandBuilder.BuildApprovePoliciesCommands(approveCommandCtx, approveCmd)).ThenReturn([]command.ProjectContext{approveCtx}, nil)
	When(projectCommandRunner.ApprovePolicies(approveCtx)).ThenReturn(command.ProjectCommandOutput{
		PolicyCheckResults: &models.PolicyCheckResults{PolicySetResults: []models.PolicySetResult{{
			PolicySetName: "production",
			Passed:        true,
			Hashes:        []string{"policy-hash"},
		}}},
	})

	approvePoliciesCommandRunner.Run(approveCommandCtx, approveCmd)

	Assert(t, approveCommandCtx.CommandHasErrors, "expected concurrent approval persistence to fail the command")
	activeAfter, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, activeAfter != nil, "expected active generation after approval attempt")
	activeProjectAfter := projectStatus(t, activeAfter, planCtx.Workspace, planCtx.RepoRelDir, planCtx.ProjectName)
	Equals(t, activeGeneration, activeProjectAfter.PlanGeneration)
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq(models.SuccessCommitStatus),
		Eq(command.PolicyCheck),
		Any[models.ProjectCounts](),
	)
	_, _, _, approveComment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("approve_policies")).GetCapturedArguments()
	Assert(t, strings.Contains(approveComment, "active plan generation"), "got: %s", approveComment)

	close(database.release)
	released = true
	<-planDone
	Assert(t, !planCommandCtx.CommandHasErrors, "expected matching plan generation completion to succeed")
	completed, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, completed != nil, "expected completed plan generation")
	completedProject := projectStatus(t, completed, planCtx.Workspace, planCtx.RepoRelDir, planCtx.ProjectName)
	Equals(t, models.PlannedPlanStatus, completedProject.Status)
	Equals(t, "", completedProject.PlanGeneration)

	finalPlanMarker := "matching generation final plan"
	repoDir := t.TempDir()
	projectDir := filepath.Join(repoDir, planCtx.RepoRelDir)
	Ok(t, os.MkdirAll(projectDir, 0755))
	planPath := filepath.Join(projectDir, runtime.GetPlanFilename(planCtx.Workspace, planCtx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte(finalPlanMarker), 0600))
	applyCtx := planCtx
	applyCtx.CommandName = command.Apply
	applyCtx.Steps = valid.DefaultApplyStage.Steps
	applyCtx.Log = logging.NewNoopLogger(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	When(mockWorkingDir.GetWorkingDir(applyCtx.Pull.BaseRepo, applyCtx.Pull, applyCtx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(applyCtx.Pull.BaseRepo, applyCtx.Pull, applyCtx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(
		Any[logging.SimpleLogging](),
		Eq(applyCtx.Pull),
		Any[models.User](),
		Eq(applyCtx.Workspace),
		Any[models.Project](),
		AnyBool(),
	)).ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)
	applyStep := &planMarkerApplyStepRunner{t: t, planPath: planPath, marker: finalPlanMarker}
	applyRunner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           applyStep,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: database},
	}

	applyResult := applyRunner.Apply(applyCtx)

	Ok(t, applyResult.Error)
	Equals(t, "applied matching generation", applyResult.ApplySuccess)
	Equals(t, 1, applyStep.calls)
}

func TestPolicyCheckCommandRunner_ActivePlanGenerationWriteFailsCommand(t *testing.T) {
	database := newTestBoltDB(t)
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
	}
	projectCtx := command.ProjectContext{
		CommandName: command.PolicyCheck,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "concurrent-policy-check",
		BaseRepo:    pull.BaseRepo,
		Pull:        pull,
	}
	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(projectCtx.RepoRelDir, projectCtx.Workspace, projectCtx.ProjectName),
	})
	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace:   projectCtx.Workspace,
		RepoRelDir:  projectCtx.RepoRelDir,
		ProjectName: projectCtx.ProjectName,
	}}, "generation-1")
	Ok(t, err)
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	When(projectCommandRunner.PolicyCheck(projectCtx)).ThenReturn(command.ProjectCommandOutput{
		PolicyCheckResults: &models.PolicyCheckResults{PolicySetResults: []models.PolicySetResult{{
			PolicySetName: "production",
			Passed:        true,
		}}},
	})

	policyCheckCommandRunner.Run(ctx, []command.ProjectContext{projectCtx})

	Assert(t, ctx.CommandHasErrors, "expected policy-check persistence rejection to fail command")
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "expected active generation")
	project := projectStatus(t, status, projectCtx.Workspace, projectCtx.RepoRelDir, projectCtx.ProjectName)
	Equals(t, models.ErroredPlanStatus, project.Status)
	Equals(t, "generation-1", project.PlanGeneration)
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(models.FailedCommitStatus), Eq(command.PolicyCheck))
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq(models.SuccessCommitStatus),
		Eq(command.PolicyCheck),
		Any[models.ProjectCounts](),
	)
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("policy_check")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "active plan generation"), "got: %s", comment)
}

func TestApplyCommandRunner_ActivePlanGenerationWriteFailsCommandAndProjectStatus(t *testing.T) {
	database := newTestBoltDB(t)
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
	}
	applyCtx := command.ProjectContext{
		CommandName: command.Apply,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "concurrent-apply",
		BaseRepo:    pull.BaseRepo,
		Pull:        pull,
		Log:         logging.NewNoopLogger(t),
	}
	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(applyCtx.RepoRelDir, applyCtx.Workspace, applyCtx.ProjectName),
	})
	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace:   applyCtx.Workspace,
		RepoRelDir:  applyCtx.RepoRelDir,
		ProjectName: applyCtx.ProjectName,
	}}, "generation-1")
	Ok(t, err)
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Apply, ProjectName: applyCtx.ProjectName}
	projectOutput := command.ProjectCommandOutput{ApplySuccess: "stale apply result"}
	When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{applyCtx}, nil)
	When(projectCommandRunner.Apply(applyCtx)).ThenReturn(projectOutput)
	jobURLSetter := mocks.NewMockJobURLSetter()
	jobMessageSender := mocks.NewMockJobMessageSender()
	wrappedProjectRunner := &events.ProjectOutputWrapper{
		ProjectCommandRunner: projectCommandRunner,
		JobURLSetter:         jobURLSetter,
		JobMessageSender:     jobMessageSender,
	}
	runner := events.NewApplyCommandRunner(
		vcsClient,
		false,
		applyLockChecker,
		commitUpdater,
		projectCommandBuilder,
		wrappedProjectRunner,
		cancellationTracker,
		autoMerger,
		pullUpdater,
		dbUpdater,
		database,
		1,
		false,
		false,
		events.NewDefaultWorkingDirLocker(),
		pullReqStatusFetcher,
		nil,
		"",
	)

	runner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected apply persistence rejection to fail command")
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "expected active generation")
	project := projectStatus(t, status, applyCtx.Workspace, applyCtx.RepoRelDir, applyCtx.ProjectName)
	Equals(t, models.ErroredPlanStatus, project.Status)
	Equals(t, "generation-1", project.PlanGeneration)
	jobURLSetter.VerifyWasCalledOnce().SetJobURLWithStatus(applyCtx, command.Apply, models.PendingCommitStatus, nil)
	jobURLSetter.VerifyWasCalledOnce().SetJobURLWithStatus(
		Eq(applyCtx), Eq(command.Apply), Eq(models.FailedCommitStatus), Any[*command.ProjectCommandOutput]())
	jobURLSetter.VerifyWasCalled(Never()).SetJobURLWithStatus(
		Eq(applyCtx), Eq(command.Apply), Eq(models.SuccessCommitStatus), Any[*command.ProjectCommandOutput]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(models.FailedCommitStatus), Eq(command.Apply))
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq(models.SuccessCommitStatus),
		Eq(command.Apply),
		Any[models.ProjectCounts](),
	)
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("apply")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "active plan generation"), "got: %s", comment)
}

func TestPlanCommandRunner_DeletePlansFailureAfterGenerationStartFailsPlanAndApply(t *testing.T) {
	for _, trigger := range []command.Trigger{command.AutoTrigger, command.CommentTrigger} {
		t.Run(fmt.Sprintf("trigger-%d", trigger), func(t *testing.T) {
			underlying := newTestBoltDB(t)
			vcsClient := setup(t, func(tc *TestConfig) {
				tc.database = underlying
			})
			pull := models.PullRequest{
				BaseRepo:   testdata.GithubRepo,
				State:      models.OpenPullState,
				Num:        testdata.Pull.Num,
				HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				BaseBranch: "main",
			}
			projectCtx := command.ProjectContext{
				CommandName: command.Plan,
				RepoRelDir:  "infrastructure",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "cleanup-failure",
				BaseRepo:    pull.BaseRepo,
				Pull:        pull,
				Steps:       valid.DefaultPlanStage.Steps,
			}
			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
				Pull:     pull,
				HeadRepo: pull.BaseRepo,
				Trigger:  trigger,
			}
			cmd := &events.CommentCommand{Name: command.Plan}
			if trigger == command.AutoTrigger {
				When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
			} else {
				When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
			}
			pullDir := t.TempDir()
			When(workingDir.GetPullDir(pull.BaseRepo, pull)).ThenReturn(pullDir, nil)
			When(pendingPlanFinder.Find(pullDir)).ThenReturn(nil, errors.New("injected pending plan lookup failure"))

			planCommandRunner.Run(ctx, cmd)

			Assert(t, ctx.CommandHasErrors, "expected cleanup failure to fail %s plan", trigger)
			status, err := underlying.GetPullStatus(pull)
			Ok(t, err)
			Assert(t, status != nil, "expected durable incomplete generation")
			project := projectStatus(t, status, projectCtx.Workspace, projectCtx.RepoRelDir, projectCtx.ProjectName)
			Assert(t, project.PlanGeneration != "", "expected cleanup failure to retain generation")
			projectCommandRunner.VerifyWasCalled(Never()).Plan(projectCtx)
			commitUpdater.VerifyWasCalledOnce().UpdateCombined(
				Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(models.FailedCommitStatus), Eq(command.Plan))
			commitUpdater.VerifyWasCalledOnce().UpdateCombined(
				Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(models.FailedCommitStatus), Eq(command.Apply))
			_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
			Assert(t, strings.Contains(comment, "injected pending plan lookup failure"), "got: %s", comment)
		})
	}
}

func TestPlanCommandRunner_AbortOnExecutionOrderFailureCompletesPlanGeneration(t *testing.T) {
	database := newTestBoltDB(t)
	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.parallelPoolSize = 2
	})
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
	}
	first := command.ProjectContext{
		CommandName:               command.Plan,
		RepoRelDir:                "first",
		Workspace:                 events.DefaultWorkspace,
		ProjectName:               "first",
		BaseRepo:                  pull.BaseRepo,
		Pull:                      pull,
		Steps:                     valid.DefaultPlanStage.Steps,
		ExecutionOrderGroup:       0,
		ParallelPlanEnabled:       true,
		AbortOnExecutionOrderFail: true,
	}
	second := first
	second.RepoRelDir = "second"
	second.ProjectName = "second"
	second.ExecutionOrderGroup = 1
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Plan}
	pullDir := t.TempDir()
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{first, second}, nil)
	When(workingDir.GetPullDir(pull.BaseRepo, pull)).ThenReturn(pullDir, nil)
	When(pendingPlanFinder.Find(pullDir)).ThenReturn(nil, nil)
	When(projectCommandRunner.Plan(first)).ThenReturn(command.ProjectCommandOutput{Error: errors.New("first execution group failed")})

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected failed execution group to fail plan")
	projectCommandRunner.VerifyWasCalledOnce().Plan(first)
	projectCommandRunner.VerifyWasCalled(Never()).Plan(second)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "expected completed PullStatus")
	firstStatus := projectStatus(t, status, first.Workspace, first.RepoRelDir, first.ProjectName)
	secondStatus := projectStatus(t, status, second.Workspace, second.RepoRelDir, second.ProjectName)
	Equals(t, models.ErroredPlanStatus, firstStatus.Status)
	Equals(t, models.ErroredPlanStatus, secondStatus.Status)
	Equals(t, "", firstStatus.PlanGeneration)
	Equals(t, "", secondStatus.PlanGeneration)
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "first execution group failed"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "earlier execution-order group failed"), "got: %s", comment)
	Assert(t, !strings.Contains(comment, "persisting plan results"), "generation completion unexpectedly failed: %s", comment)
}

func TestPlanCommandRunner_AutoplanPersistenceFailureFailsClosed(t *testing.T) {
	vcsClient := setup(t)
	modelPull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "abc123",
		BaseBranch: "main",
	}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "gitea-autoplan",
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	policyCtx := projectCtx
	policyCtx.CommandName = command.PolicyCheck
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.AutoTrigger,
	}
	tmp := t.TempDir()
	closer := dbUpdater.Database.(interface{ Close() error })
	Ok(t, closer.Close())

	When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{projectCtx, policyCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	planCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Plan})

	Assert(t, ctx.CommandHasErrors, "expected PullStatus persistence failure to mark autoplan errored")
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Plan))
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Apply))
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Plan Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "persisting plan results"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "invalidating previous plan state before planning"), "got: %s", comment)
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.SuccessCommitStatus), Eq(command.Plan), Any[models.ProjectCounts]())
	projectCommandRunner.VerifyWasCalled(Never()).PolicyCheck(policyCtx)
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectCtx)
}

func TestPlanCommandRunner_AutoplanFinalPersistenceFailureLeavesGenerationNonApplyable(t *testing.T) {
	underlying := newTestBoltDB(t)
	database := &failNextPullStatusWriteDB{Database: underlying}
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
	}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "autoplan-replan",
		BaseRepo:    pull.BaseRepo,
		Pull:        pull,
		Steps:       []valid.Step{{StepName: "run", RunCommand: "custom-plan-command"}},
	}
	_, err := underlying.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(projectCtx.RepoRelDir, projectCtx.Workspace, projectCtx.ProjectName),
	})
	Ok(t, err)
	setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.AutoTrigger,
	}
	tmp := t.TempDir()
	When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	database.failNextUpdate = true

	planCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Plan})

	Assert(t, ctx.CommandHasErrors, "expected final autoplan persistence failure to fail the command")
	projectCommandRunner.VerifyWasCalledOnce().Plan(projectCtx)
	status, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "expected durable PullStatus")
	project := projectStatus(t, status, projectCtx.Workspace, projectCtx.RepoRelDir, projectCtx.ProjectName)
	Equals(t, models.ErroredPlanStatus, project.Status)
	Assert(t, project.PlanGeneration != "", "expected active generation marker")
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.SuccessCommitStatus), Eq(command.Plan), Any[models.ProjectCounts]())
}

func TestPlanCommandRunner_ManualPlanPersistenceFailureFailsClosed(t *testing.T) {
	vcsClient := setup(t)
	modelPull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "abc123",
		BaseBranch: "main",
	}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "infrastructure",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "manual-plan",
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	policyCtx := projectCtx
	policyCtx.CommandName = command.PolicyCheck
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	cmd := &events.CommentCommand{Name: command.Plan}
	tmp := t.TempDir()
	closer := dbUpdater.Database.(interface{ Close() error })
	Ok(t, closer.Close())

	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectCtx, policyCtx}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected PullStatus persistence failure to mark manual plan errored")
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Plan))
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Apply))
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Plan Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "persisting plan results"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "invalidating previous plan state before planning"), "got: %s", comment)
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.SuccessCommitStatus), Eq(command.Plan), Any[models.ProjectCounts]())
	projectCommandRunner.VerifyWasCalled(Never()).PolicyCheck(policyCtx)
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectCtx)
}

func TestPlanCommandRunner_FailedManagedReplanInvalidatesPreviousPlanGeneration(t *testing.T) {
	assertFailedReplanInvalidatesPreviousPlanGeneration(t, valid.DefaultPlanStage.Steps, valid.DefaultApplyStage.Steps, true, nil)
}

func TestPlanCommandRunner_FailedRunOnlyReplanInvalidatesPreviousPlanGeneration(t *testing.T) {
	runOnlyPlan := []valid.Step{{StepName: "run", RunCommand: "some-custom-plan-command custom-plan-path"}}
	assertFailedReplanInvalidatesPreviousPlanGeneration(t, runOnlyPlan, []valid.Step{{
		StepName:   "run",
		RunCommand: "some-custom-apply-command custom-plan-path",
	}}, false, nil)
}

func TestPlanCommandRunner_FailedReplanReportsManagedPlanCleanupFailure(t *testing.T) {
	assertFailedReplanInvalidatesPreviousPlanGeneration(
		t,
		valid.DefaultPlanStage.Steps,
		valid.DefaultApplyStage.Steps,
		true,
		errors.New("injected managed plan cleanup failure"),
	)
}

func TestPlanCommandRunner_FailedReplanReleasesOnlyLockCreatedByGeneration(t *testing.T) {
	RegisterMockTestingT(t)
	underlying := newTestBoltDB(t)
	database := &failNextPullStatusWriteDB{Database: underlying}
	planLocker := &trackingPlanGenerationLocker{}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	projectCtx := command.ProjectContext{
		CommandName:   command.Plan,
		RepoRelDir:    ".",
		Workspace:     events.DefaultWorkspace,
		ProjectName:   "locked-replan",
		BaseRepo:      pull.BaseRepo,
		Pull:          pull,
		Steps:         valid.DefaultPlanStage.Steps,
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	_, err := underlying.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(projectCtx.RepoRelDir, projectCtx.Workspace, projectCtx.ProjectName),
	})
	Ok(t, err)
	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.planLocker = planLocker
	})
	cmd := &events.CommentCommand{Name: command.Plan, ProjectName: projectCtx.ProjectName}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).Then(func([]Param) ReturnValues {
		planLocker.lock = &models.ProjectLock{
			Pull:      pull,
			Project:   models.NewProject(pull.BaseRepo.FullName, projectCtx.RepoRelDir, projectCtx.ProjectName),
			Workspace: projectCtx.Workspace,
		}
		return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
	})
	database.failNextUpdate = true

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected final persistence failure")
	Equals(t, 1, planLocker.unlockIfOwnedCalls)
	Equals(t, 0, planLocker.unlockByPullCalls)
	Assert(t, planLocker.lock == nil, "expected generation-created plan lock to be released")
}

func TestPlanCommandRunner_FailedTargetedReplanPreservesPreexistingPlanLock(t *testing.T) {
	RegisterMockTestingT(t)
	underlying := newTestBoltDB(t)
	database := &failNextPullStatusWriteDB{Database: underlying}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	projectCtx := command.ProjectContext{
		CommandName:   command.Plan,
		RepoRelDir:    ".",
		Workspace:     events.DefaultWorkspace,
		ProjectName:   "locked-replan",
		BaseRepo:      pull.BaseRepo,
		Pull:          pull,
		Steps:         valid.DefaultPlanStage.Steps,
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	preexistingLock := &models.ProjectLock{
		Pull:      pull,
		Project:   models.NewProject(pull.BaseRepo.FullName, projectCtx.RepoRelDir, projectCtx.ProjectName),
		Workspace: projectCtx.Workspace,
	}
	planLocker := &trackingPlanGenerationLocker{lock: preexistingLock}
	_, err := underlying.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(projectCtx.RepoRelDir, projectCtx.Workspace, projectCtx.ProjectName),
	})
	Ok(t, err)
	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.planLocker = planLocker
	})
	cmd := &events.CommentCommand{Name: command.Plan, ProjectName: projectCtx.ProjectName}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	database.failNextUpdate = true

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected final persistence failure")
	Equals(t, 0, planLocker.unlockIfOwnedCalls)
	Equals(t, 0, planLocker.unlockByPullCalls)
	Assert(t, planLocker.lock == preexistingLock, "expected preexisting plan lock to remain")
}

func TestPlanCommandRunner_FailedReplanDoesNotCleanNewerGeneration(t *testing.T) {
	RegisterMockTestingT(t)
	underlying := newTestBoltDB(t)
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	projectCtx := command.ProjectContext{
		CommandName:   command.Plan,
		RepoRelDir:    ".",
		Workspace:     events.DefaultWorkspace,
		ProjectName:   "superseded-replan",
		BaseRepo:      pull.BaseRepo,
		Pull:          pull,
		Steps:         valid.DefaultPlanStage.Steps,
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	_, err := underlying.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(projectCtx.RepoRelDir, projectCtx.Workspace, projectCtx.ProjectName),
	})
	Ok(t, err)
	database := &supersedePlanCompletionDB{
		Database: underlying,
		project: models.ProjectStatus{
			Workspace:   projectCtx.Workspace,
			RepoRelDir:  projectCtx.RepoRelDir,
			ProjectName: projectCtx.ProjectName,
		},
	}
	planLocker := &trackingPlanGenerationLocker{}
	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.planLocker = planLocker
	})
	cmd := &events.CommentCommand{Name: command.Plan, ProjectName: projectCtx.ProjectName}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).Then(func([]Param) ReturnValues {
		planLocker.lock = &models.ProjectLock{
			Pull:      pull,
			Project:   models.NewProject(pull.BaseRepo.FullName, projectCtx.RepoRelDir, projectCtx.ProjectName),
			Workspace: projectCtx.Workspace,
		}
		return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
	})

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected superseded final persistence to fail the replan")
	workingDir.(*mocks.MockWorkingDir).VerifyWasCalled(Never()).DeletePlan(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[string](), Any[string]())
	Equals(t, 0, planLocker.unlockIfOwnedCalls)
	Equals(t, 0, planLocker.unlockByPullCalls)
	Assert(t, planLocker.lock != nil, "expected the newer generation lock to remain")
	status, err := underlying.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "expected newer plan generation status")
	Equals(t, models.ErroredPlanStatus, status.Projects[0].Status)
	Equals(t, "newer-generation", status.Projects[0].PlanGeneration)
}

func assertFailedReplanInvalidatesPreviousPlanGeneration(t *testing.T, planSteps, applySteps []valid.Step, expectManagedPlanCleanup bool, cleanupError error) {
	t.Helper()
	RegisterMockTestingT(t)

	underlying := newTestBoltDB(t)
	database := &failNextPullStatusWriteDB{Database: underlying}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  ".",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "replanned-project",
		BaseRepo:    pull.BaseRepo,
		Pull:        pull,
		Steps:       planSteps,
	}
	_, err := underlying.UpdatePullWithResults(pull, []command.ProjectResult{
		plannedProjectResult(projectCtx.RepoRelDir, projectCtx.Workspace, projectCtx.ProjectName),
	})
	Ok(t, err)

	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
	})
	cmd := &events.CommentCommand{Name: command.Plan, ProjectName: projectCtx.ProjectName}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: pull.BaseRepo,
		Trigger:  command.CommentTrigger,
	}
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "new plan generation"},
	})
	if cleanupError != nil {
		When(workingDir.(*mocks.MockWorkingDir).DeletePlan(
			Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(projectCtx.Workspace), Eq(projectCtx.RepoRelDir), Eq(projectCtx.ProjectName))).ThenReturn(cleanupError)
	}
	database.failNextUpdate = true

	planCommandRunner.Run(ctx, cmd)

	Assert(t, ctx.CommandHasErrors, "expected failed PullStatus persistence to fail the replan")
	if expectManagedPlanCleanup {
		workingDir.(*mocks.MockWorkingDir).VerifyWasCalledOnce().DeletePlan(
			Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(projectCtx.Workspace), Eq(projectCtx.RepoRelDir), Eq(projectCtx.ProjectName))
	} else {
		workingDir.(*mocks.MockWorkingDir).VerifyWasCalled(Never()).DeletePlan(
			Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[string](), Any[string]())
	}
	if cleanupError != nil {
		_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
			Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
		Assert(t, strings.Contains(comment, "injected PullStatus persistence failure"), "got: %s", comment)
		Assert(t, strings.Contains(comment, cleanupError.Error()), "got: %s", comment)
	}

	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockApply := mocks.NewMockStepRunner()
	mockRun := mocks.NewMockCustomStepRunner()
	repoDir := t.TempDir()
	applyCtx := projectCtx
	applyCtx.CommandName = command.Apply
	applyCtx.Steps = applySteps
	applyCtx.Log = logging.NewNoopLogger(t)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(applyCtx.Workspace, applyCtx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("new unpersisted plan generation"), 0600))
	customPlanPath := filepath.Join(repoDir, "custom-plan-path")
	Ok(t, os.WriteFile(customPlanPath, []byte("new unpersisted custom plan generation"), 0600))
	When(mockWorkingDir.GetWorkingDir(applyCtx.Pull.BaseRepo, applyCtx.Pull, applyCtx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(applyCtx.Pull.BaseRepo, applyCtx.Pull, applyCtx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(applyCtx.Pull), Any[models.User](), Eq(applyCtx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		RunStepRunner:             mockRun,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: database},
	}

	result := runner.Apply(applyCtx)

	Assert(t, result.Error != nil, "expected failed replan generation to reject apply")
	Assert(t, strings.Contains(result.Error.Error(), "plan is incomplete"), "got: %s", result.Error)
	Assert(t, strings.Contains(result.Error.Error(), "run `atlantis plan`"), "got: %s", result.Error)
	mockApply.VerifyWasCalled(Never()).Run(
		Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
	mockRun.VerifyWasCalled(Never()).Run(
		Any[command.ProjectContext](),
		Any[*valid.CommandShell](),
		Any[string](),
		Any[string](),
		Any[map[string]string](),
		AnyBool(),
		Any[[]valid.PostProcessRunOutputOption](),
		Any[[]*regexp.Regexp](),
	)
}

type failNextPullStatusWriteDB struct {
	db.Database
	failNextUpdate bool
}

type beginGenerationAfterReplaceDB struct {
	db.Database
	project  models.ProjectStatus
	planPath string
	marker   []byte
	started  bool
}

type blockingPlanCompletionDB struct {
	db.Database
	started chan struct{}
	release chan struct{}
}

func (d *beginGenerationAfterReplaceDB) ReplacePullWithResults(pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	status, err := d.Database.ReplacePullWithResults(pull, results)
	if err != nil || d.started {
		return status, err
	}
	d.started = true
	if _, err := d.BeginPlanGeneration(pull, []models.ProjectStatus{d.project}, "generation-after-replace"); err != nil {
		return models.PullStatus{}, err
	}
	if err := os.WriteFile(d.planPath, d.marker, 0600); err != nil {
		return models.PullStatus{}, err
	}
	return status, nil
}

type supersedePlanCompletionDB struct {
	db.Database
	project models.ProjectStatus
}

type trackingPlanGenerationLocker struct {
	lock               *models.ProjectLock
	unlockIfOwnedCalls int
	unlockByPullCalls  int
}

type planMarkerApplyStepRunner struct {
	t        *testing.T
	planPath string
	marker   string
	calls    int
}

func (r *planMarkerApplyStepRunner) Run(command.ProjectContext, []string, string, map[string]string) (string, error) {
	r.calls++
	contents, err := os.ReadFile(r.planPath)
	if err != nil {
		return "", fmt.Errorf("reading final plan generation: %w", err)
	}
	Equals(r.t, r.marker, string(contents))
	return "applied matching generation", nil
}

func (l *trackingPlanGenerationLocker) TryLock(models.Project, string, models.PullRequest, models.User) (locking.TryLockResponse, error) {
	return locking.TryLockResponse{}, nil
}

func (l *trackingPlanGenerationLocker) Unlock(string) (*models.ProjectLock, error) {
	return nil, nil
}

func (l *trackingPlanGenerationLocker) UnlockIfOwnedByPull(_ models.Project, _ string, pullNum int) (*models.ProjectLock, error) {
	l.unlockIfOwnedCalls++
	if l.lock == nil || l.lock.Pull.Num != pullNum {
		return nil, nil
	}
	unlocked := l.lock
	l.lock = nil
	return unlocked, nil
}

func (l *trackingPlanGenerationLocker) List() (map[string]models.ProjectLock, error) {
	return nil, nil
}

func (l *trackingPlanGenerationLocker) UnlockByPull(string, int) ([]models.ProjectLock, error) {
	l.unlockByPullCalls++
	return nil, nil
}

func (l *trackingPlanGenerationLocker) GetLock(string) (*models.ProjectLock, error) {
	return l.lock, nil
}

func (d *blockingPlanCompletionDB) CompletePlanGeneration(pull models.PullRequest, generation string, results []command.ProjectResult) (models.PullStatus, error) {
	close(d.started)
	<-d.release
	return d.Database.CompletePlanGeneration(pull, generation, results)
}

func (d *supersedePlanCompletionDB) CompletePlanGeneration(pull models.PullRequest, _ string, _ []command.ProjectResult) (models.PullStatus, error) {
	if _, err := d.BeginPlanGeneration(pull, []models.ProjectStatus{d.project}, "newer-generation"); err != nil {
		return models.PullStatus{}, fmt.Errorf("starting newer plan generation: %w", err)
	}
	return models.PullStatus{}, errors.New("injected superseded PullStatus persistence failure")
}

func (d *failNextPullStatusWriteDB) UpdatePullWithResults(pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if d.failNextUpdate {
		d.failNextUpdate = false
		return models.PullStatus{}, errors.New("injected PullStatus persistence failure")
	}
	return d.Database.UpdatePullWithResults(pull, results)
}

func (d *failNextPullStatusWriteDB) CompletePlanGeneration(pull models.PullRequest, generation string, results []command.ProjectResult) (models.PullStatus, error) {
	if d.failNextUpdate {
		d.failNextUpdate = false
		return models.PullStatus{}, errors.New("injected PullStatus persistence failure")
	}
	return d.Database.CompletePlanGeneration(pull, generation, results)
}

func TestRunCommentCommand_IgnoredTargetedDirNoOp(t *testing.T) {
	cases := []struct {
		name        string
		commandName command.Name
		subName     string
	}{
		{
			name:        "plan",
			commandName: command.Plan,
		},
		{
			name:        "apply",
			commandName: command.Apply,
		},
		{
			name:        "approve_policies",
			commandName: command.ApprovePolicies,
		},
		{
			name:        "import",
			commandName: command.Import,
		},
		{
			name:        "version",
			commandName: command.Version,
		},
		{
			name:        "state rm",
			commandName: command.State,
			subName:     "rm",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vcsClient := setup(t)
			ch.FailOnPreWorkflowHookError = true
			pull := &github.PullRequest{State: github.Ptr("open")}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			cmd := &events.CommentCommand{Name: c.commandName, SubName: c.subName, RepoRelDir: "ignored"}

			When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
			When(projectCommandBuilder.ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(true)
			When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))

			ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

			preWorkflowHooksCommandRunner.(*mocks.MockPreWorkflowHooksCommandRunner).VerifyWasCalledOnce().RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())
			vcsClient.VerifyWasCalled(Never()).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
			commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
			commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
			postWorkflowHooksCommandRunner.(*mocks.MockPostWorkflowHooksCommandRunner).VerifyWasCalled(Never()).RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())
		})
	}
}

func TestRunCommentCommand_IgnoredTargetedDirSkipsValidationComments(t *testing.T) {
	cases := []struct {
		name      string
		cmd       *events.CommentCommand
		modelPull models.PullRequest
		headRepo  models.Repo
		setup     func(t *testing.T, vcsClient *vcsmocks.MockClient)
		verify    func(t *testing.T, vcsClient *vcsmocks.MockClient)
	}{
		{
			name: "team allowlist",
			cmd:  &events.CommentCommand{Name: command.Plan, RepoRelDir: "ignored"},
			setup: func(t *testing.T, vcsClient *vcsmocks.MockClient) {
				checker, err := command.NewTeamAllowlistChecker("allowed-team:apply")
				Ok(t, err)
				ch.TeamAllowlistChecker = checker
				When(vcsClient.GetTeamNamesForUser(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.User))).ThenReturn([]string{"blocked-team"}, nil)
			},
			verify: func(t *testing.T, vcsClient *vcsmocks.MockClient) {
				vcsClient.VerifyWasCalledOnce().GetTeamNamesForUser(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.User))
			},
		},
		{
			name: "var file allowlist",
			cmd:  &events.CommentCommand{Name: command.Plan, RepoRelDir: "ignored", Flags: []string{"-var-file", "/tmp/outside.tfvars"}},
			setup: func(t *testing.T, vcsClient *vcsmocks.MockClient) {
				checker, err := events.NewVarFileAllowlistChecker("")
				Ok(t, err)
				ch.VarFileAllowlistChecker = checker
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vcsClient := setup(t)
			if c.modelPull.BaseRepo.FullName == "" {
				c.modelPull = models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			}
			if c.headRepo.FullName == "" {
				c.headRepo = testdata.GithubRepo
			}

			if c.setup != nil {
				c.setup(t, vcsClient)
			}

			pull := &github.PullRequest{State: github.Ptr("open")}
			When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(c.modelPull, c.modelPull.BaseRepo, c.headRepo, nil)
			When(projectCommandBuilder.ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(true)

			ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, c.cmd)

			if c.verify != nil {
				c.verify(t, vcsClient)
			}
			preWorkflowHooksCommandRunner.(*mocks.MockPreWorkflowHooksCommandRunner).VerifyWasCalled(Never()).RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())
			vcsClient.VerifyWasCalled(Never()).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
			commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
			commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
			postWorkflowHooksCommandRunner.(*mocks.MockPostWorkflowHooksCommandRunner).VerifyWasCalled(Never()).RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())
		})
	}
}

func TestRunCommentCommand_IgnoredTargetedDirValidatesPullBeforeIgnoreCheck(t *testing.T) {
	cases := []struct {
		name        string
		modelPull   models.PullRequest
		headRepo    models.Repo
		setup       func()
		wantComment string
	}{
		{
			name: "fork pull request",
			modelPull: models.PullRequest{
				BaseRepo: testdata.GithubRepo,
				State:    models.OpenPullState,
				Num:      testdata.Pull.Num,
			},
			headRepo: models.Repo{
				Owner:    "forkrepo",
				FullName: "forkrepo/atlantis",
			},
			wantComment: fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s  or, to disable this message, set --%s", "allow-fork-prs-flag", ""),
		},
		{
			name: "closed pull request",
			modelPull: models.PullRequest{
				BaseRepo: testdata.GithubRepo,
				State:    models.ClosedPullState,
				Num:      testdata.Pull.Num,
			},
			headRepo:    testdata.GithubRepo,
			wantComment: "Atlantis commands can't be run on closed pull requests",
		},
		{
			name: "base branch mismatch",
			modelPull: models.PullRequest{
				BaseRepo:   testdata.GithubRepo,
				State:      models.OpenPullState,
				Num:        testdata.Pull.Num,
				BaseBranch: "release",
			},
			headRepo: testdata.GithubRepo,
			setup: func() {
				ch.GlobalCfg.Repos[0].BranchRegex = regexp.MustCompile("^main$")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vcsClient := setup(t)
			if c.setup != nil {
				c.setup()
			}

			pull := &github.PullRequest{State: github.Ptr("open")}
			When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(c.modelPull, c.modelPull.BaseRepo, c.headRepo, nil)
			When(projectCommandBuilder.ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(true)

			ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan, RepoRelDir: "ignored"})

			projectCommandBuilder.VerifyWasCalled(Never()).ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())
			preWorkflowHooksCommandRunner.(*mocks.MockPreWorkflowHooksCommandRunner).VerifyWasCalled(Never()).RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())
			if c.wantComment == "" {
				vcsClient.VerifyWasCalled(Never()).CreateComment(
					Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
			} else {
				vcsClient.VerifyWasCalledOnce().CreateComment(
					Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(c.modelPull.Num), Eq(c.wantComment), Eq(""))
			}
		})
	}
}

func TestRunCommentCommand_IgnoredTargetedDirNoHooksDoesNotUseLocalConfig(t *testing.T) {
	setup(t)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	cmd := &events.CommentCommand{Name: command.Apply, RepoRelDir: "ignored"}
	ignoreChecks := 0

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(args []Param) ReturnValues {
		ignoreChecks++
		return ReturnValues{ignoreChecks == 1}
	})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

	Equals(t, 1, ignoreChecks)
	preWorkflowHooksCommandRunner.(*mocks.MockPreWorkflowHooksCommandRunner).VerifyWasCalledOnce().RunPreHooks(Any[*command.Context](), Eq(cmd))
	projectCommandBuilder.VerifyWasCalled(Never()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestRunCommentCommand_IgnoredTargetedDirPreHooksCanGenerateExplicitConfig(t *testing.T) {
	setup(t)
	configuredPreHooks := &configuredPreWorkflowHooksCommandRunner{hasHooks: true}
	preWorkflowHooksCommandRunner = configuredPreHooks
	ch.PreWorkflowHooksCommandRunner = configuredPreHooks
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "ignored"}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "ignored",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	ignoreChecks := 0

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(args []Param) ReturnValues {
		ignoreChecks++
		return ReturnValues{ignoreChecks == 1}
	})
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Eq(cmd))).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

	Equals(t, 1, configuredPreHooks.calls)
	projectCommandBuilder.VerifyWasCalledOnce().BuildPlanCommands(Any[*command.Context](), Eq(cmd))
	projectCommandRunner.VerifyWasCalledOnce().Plan(projectCtx)
}

func TestRunCommentCommand_IgnoredTargetedDirNonFatalPreHookErrorCanGenerateExplicitConfig(t *testing.T) {
	setup(t)
	configuredPreHooks := &configuredPreWorkflowHooksCommandRunner{hasHooks: true, err: errors.New("hook error")}
	preWorkflowHooksCommandRunner = configuredPreHooks
	ch.PreWorkflowHooksCommandRunner = configuredPreHooks
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "ignored"}
	projectCtx := command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  "ignored",
		Workspace:   events.DefaultWorkspace,
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	ignoreChecks := 0

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(projectCommandBuilder.ShouldIgnoreTargetedDir(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(args []Param) ReturnValues {
		ignoreChecks++
		return ReturnValues{ignoreChecks == 1}
	})
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Eq(cmd))).ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, cmd)

	Equals(t, 1, configuredPreHooks.calls)
	Equals(t, 2, ignoreChecks)
	projectCommandBuilder.VerifyWasCalledOnce().BuildPlanCommands(Any[*command.Context](), Eq(cmd))
	projectCommandRunner.VerifyWasCalledOnce().Plan(projectCtx)
}

func TestRunCommentCommandApprovePolicy_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an approve_policy command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	approvePoliciesCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.ApprovePolicies})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunCommentCommandUnlock_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an unlock command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	unlockCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Unlock})
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.PendingCommitStatus), Any[command.Name]())
}

func TestRunCommentCommandImport_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an import command is run on a pull request and SilenceNoProjects is enabled, we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	importCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Import})
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestImportOrStateRm_DiscardedPlanStatusAllowsLaterGenericApply(t *testing.T) {
	cases := []struct {
		name       string
		cmd        events.CommentCommand
		projectCmd command.ProjectContext
		output     command.ProjectCommandOutput
	}{
		{
			name: "import",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
		},
		{
			name: "state rm",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setup(t)
			logger := logging.NewNoopLogger(t)
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
			_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
				{
					Command:     command.Plan,
					RepoRelDir:  tc.projectCmd.RepoRelDir,
					Workspace:   tc.projectCmd.Workspace,
					ProjectName: tc.projectCmd.ProjectName,
					ProjectCommandOutput: command.ProjectCommandOutput{
						PlanSuccess: &models.PlanSuccess{},
					},
				},
			})
			Ok(t, err)

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logger,
				Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}
			tc.projectCmd.BaseRepo = testdata.GithubRepo
			tc.projectCmd.Pull = modelPull

			switch tc.cmd.Name {
			case command.Import:
				When(pullReqStatusFetcher.FetchPullStatus(logger, modelPull)).ThenReturn(models.PullReqStatus{}, nil)
				When(projectCommandBuilder.BuildImportCommands(ctx, &tc.cmd)).ThenReturn([]command.ProjectContext{tc.projectCmd}, nil)
				When(projectCommandRunner.Import(tc.projectCmd)).ThenReturn(tc.output)
				importCommandRunner.Run(ctx, &tc.cmd)
			case command.State:
				When(projectCommandBuilder.BuildStateRmCommands(ctx, &tc.cmd)).ThenReturn([]command.ProjectContext{tc.projectCmd}, nil)
				When(projectCommandRunner.StateRm(tc.projectCmd)).ThenReturn(tc.output)
				stateCommandRunner.Run(ctx, &tc.cmd)
			}

			pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
			Ok(t, err)
			Assert(t, pullStatus != nil, "expected PullStatus")
			Equals(t, 1, len(pullStatus.Projects))
			Equals(t, models.DiscardedPlanStatus, pullStatus.Projects[0].Status)

			applyCtx := &command.Context{
				Log:        logger,
				Pull:       modelPull,
				PullStatus: pullStatus,
			}
			Ok(t, events.ValidatePlansForApply(applyCtx, nil))
		})
	}
}

func TestImportCommandRunner_DiscardedPlanDBFailureIsUserVisible(t *testing.T) {
	vcsClient := setup(t)
	logger := logging.NewNoopLogger(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	projectCmd := command.ProjectContext{
		CommandName: command.Import,
		RepoRelDir:  "dir1",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "projA",
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	cmd := events.CommentCommand{Name: command.Import, ProjectName: "projA"}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}

	closer := dbUpdater.Database.(interface{ Close() error })
	Ok(t, closer.Close())

	When(pullReqStatusFetcher.FetchPullStatus(logger, modelPull)).ThenReturn(models.PullReqStatus{}, nil)
	When(projectCommandBuilder.BuildImportCommands(ctx, &cmd)).ThenReturn([]command.ProjectContext{projectCmd}, nil)
	When(projectCommandRunner.Import(projectCmd)).ThenReturn(command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}})

	importCommandRunner.Run(ctx, &cmd)

	Assert(t, ctx.CommandHasErrors, "expected discarded plan DB failure to mark command errored")
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("import")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Import Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "writing discarded plan status"), "got: %s", comment)
}

func TestStateCommandRunner_DiscardedPlanDBFailureIsUserVisible(t *testing.T) {
	vcsClient := setup(t)
	logger := logging.NewNoopLogger(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	projectCmd := command.ProjectContext{
		CommandName: command.State,
		SubCommand:  "rm",
		RepoRelDir:  "dir1",
		Workspace:   events.DefaultWorkspace,
		ProjectName: "projA",
		BaseRepo:    testdata.GithubRepo,
		Pull:        modelPull,
	}
	cmd := events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projA"}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}

	closer := dbUpdater.Database.(interface{ Close() error })
	Ok(t, closer.Close())

	When(projectCommandBuilder.BuildStateRmCommands(ctx, &cmd)).ThenReturn([]command.ProjectContext{projectCmd}, nil)
	When(projectCommandRunner.StateRm(projectCmd)).ThenReturn(command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}})

	stateCommandRunner.Run(ctx, &cmd)

	Assert(t, ctx.CommandHasErrors, "expected discarded plan DB failure to mark command errored")
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("state")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "State Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "writing discarded plan status"), "got: %s", comment)
}

func TestImportOrStateRm_DoesNotDiscardPlanStatusOnErrorOrFailure(t *testing.T) {
	cases := []struct {
		name       string
		cmd        events.CommentCommand
		projectCmd command.ProjectContext
		output     command.ProjectCommandOutput
	}{
		{
			name: "import error",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{Error: errors.New("import error")},
		},
		{
			name: "import failure",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{Failure: "import failed"},
		},
		{
			name: "state rm error",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{Error: errors.New("state rm error")},
		},
		{
			name: "state rm failure",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{Failure: "state rm failed"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setup(t)
			logger := logging.NewNoopLogger(t)
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
			_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
				{
					Command:     command.Plan,
					RepoRelDir:  tc.projectCmd.RepoRelDir,
					Workspace:   tc.projectCmd.Workspace,
					ProjectName: tc.projectCmd.ProjectName,
					ProjectCommandOutput: command.ProjectCommandOutput{
						PlanSuccess: &models.PlanSuccess{},
					},
				},
			})
			Ok(t, err)

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logger,
				Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}
			cmd := tc.cmd
			projectCmd := tc.projectCmd
			projectCmd.BaseRepo = testdata.GithubRepo
			projectCmd.Pull = modelPull

			switch cmd.Name {
			case command.Import:
				When(pullReqStatusFetcher.FetchPullStatus(logger, modelPull)).ThenReturn(models.PullReqStatus{}, nil)
				When(projectCommandBuilder.BuildImportCommands(ctx, &cmd)).ThenReturn([]command.ProjectContext{projectCmd}, nil)
				When(projectCommandRunner.Import(projectCmd)).ThenReturn(tc.output)
				importCommandRunner.Run(ctx, &cmd)
			case command.State:
				When(projectCommandBuilder.BuildStateRmCommands(ctx, &cmd)).ThenReturn([]command.ProjectContext{projectCmd}, nil)
				When(projectCommandRunner.StateRm(projectCmd)).ThenReturn(tc.output)
				stateCommandRunner.Run(ctx, &cmd)
			}

			pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
			Ok(t, err)
			Assert(t, pullStatus != nil, "expected PullStatus")
			Equals(t, 1, len(pullStatus.Projects))
			Equals(t, models.PlannedPlanStatus, pullStatus.Projects[0].Status)
		})
	}
}

func TestImportOrStateRm_DiscardsOnlyExistingPullStatusProject(t *testing.T) {
	cases := []struct {
		name       string
		cmd        events.CommentCommand
		projectCmd command.ProjectContext
		output     command.ProjectCommandOutput
	}{
		{
			name: "import",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projB"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projB",
			},
			output: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
		},
		{
			name: "state rm",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projB"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projB",
			},
			output: command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setup(t)
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
			_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
				plannedProjectResult("dir1", events.DefaultWorkspace, "projA"),
				plannedProjectResult("dir1", events.DefaultWorkspace, "projB"),
			})
			Ok(t, err)

			runImportOrStateRmResult(t, modelPull, tc.cmd, tc.projectCmd, tc.output)

			pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
			Ok(t, err)
			Equals(t, models.PlannedPlanStatus, projectStatus(t, pullStatus, events.DefaultWorkspace, "dir1", "projA").Status)
			Equals(t, models.DiscardedPlanStatus, projectStatus(t, pullStatus, events.DefaultWorkspace, "dir1", "projB").Status)
		})
	}
}

func TestImportOrStateRm_DoesNotCreateDiscardedStatusWithoutPullStatus(t *testing.T) {
	cases := []struct {
		name       string
		cmd        events.CommentCommand
		projectCmd command.ProjectContext
		output     command.ProjectCommandOutput
	}{
		{
			name: "import",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
		},
		{
			name: "state rm",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setup(t)
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}

			runImportOrStateRmResult(t, modelPull, tc.cmd, tc.projectCmd, tc.output)

			pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
			Ok(t, err)
			Assert(t, pullStatus == nil, "expected no PullStatus to be created")
		})
	}
}

func TestImportOrStateRm_DoesNotCreateDiscardedStatusForMissingProject(t *testing.T) {
	cases := []struct {
		name       string
		cmd        events.CommentCommand
		projectCmd command.ProjectContext
		output     command.ProjectCommandOutput
	}{
		{
			name: "import",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projB"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projB",
			},
			output: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
		},
		{
			name: "state rm",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projB"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projB",
			},
			output: command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setup(t)
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
			_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
				plannedProjectResult("dir1", events.DefaultWorkspace, "projA"),
			})
			Ok(t, err)

			runImportOrStateRmResult(t, modelPull, tc.cmd, tc.projectCmd, tc.output)

			pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
			Ok(t, err)
			Equals(t, 1, len(pullStatus.Projects))
			Equals(t, models.PlannedPlanStatus, projectStatus(t, pullStatus, events.DefaultWorkspace, "dir1", "projA").Status)
		})
	}
}

func TestImportOrStateRm_DoesNotDiscardStaleHeadPullStatus(t *testing.T) {
	cases := []struct {
		name       string
		cmd        events.CommentCommand
		projectCmd command.ProjectContext
		output     command.ProjectCommandOutput
	}{
		{
			name: "import",
			cmd:  events.CommentCommand{Name: command.Import, ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.Import,
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
		},
		{
			name: "state rm",
			cmd:  events.CommentCommand{Name: command.State, SubName: "rm", ProjectName: "projA"},
			projectCmd: command.ProjectContext{
				CommandName: command.State,
				SubCommand:  "rm",
				RepoRelDir:  "dir1",
				Workspace:   events.DefaultWorkspace,
				ProjectName: "projA",
			},
			output: command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setup(t)
			stalePull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "old123"}
			currentPull := stalePull
			currentPull.HeadCommit = "new123"
			_, err := dbUpdater.Database.UpdatePullWithResults(stalePull, []command.ProjectResult{
				plannedProjectResult("dir1", events.DefaultWorkspace, "projA"),
			})
			Ok(t, err)

			runImportOrStateRmResult(t, currentPull, tc.cmd, tc.projectCmd, tc.output)

			pullStatus, err := dbUpdater.Database.GetPullStatus(currentPull)
			Ok(t, err)
			Assert(t, pullStatus != nil, "expected stale PullStatus to remain")
			Equals(t, "old123", pullStatus.Pull.HeadCommit)
			Equals(t, models.PlannedPlanStatus, projectStatus(t, pullStatus, events.DefaultWorkspace, "dir1", "projA").Status)
		})
	}
}

func runImportOrStateRmResult(t *testing.T, pull models.PullRequest, cmd events.CommentCommand, projectCmd command.ProjectContext, output command.ProjectCommandOutput) {
	t.Helper()
	logger := logging.NewNoopLogger(t)
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
		Pull:     pull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	projectCmd.BaseRepo = testdata.GithubRepo
	projectCmd.Pull = pull

	switch cmd.Name {
	case command.Import:
		When(pullReqStatusFetcher.FetchPullStatus(logger, pull)).ThenReturn(models.PullReqStatus{}, nil)
		When(projectCommandBuilder.BuildImportCommands(ctx, &cmd)).ThenReturn([]command.ProjectContext{projectCmd}, nil)
		When(projectCommandRunner.Import(projectCmd)).ThenReturn(output)
		importCommandRunner.Run(ctx, &cmd)
	case command.State:
		When(projectCommandBuilder.BuildStateRmCommands(ctx, &cmd)).ThenReturn([]command.ProjectContext{projectCmd}, nil)
		When(projectCommandRunner.StateRm(projectCmd)).ThenReturn(output)
		stateCommandRunner.Run(ctx, &cmd)
	default:
		t.Fatalf("unsupported command %q", cmd.Name)
	}
}

func plannedProjectResult(repoRelDir, workspace, projectName string) command.ProjectResult {
	return command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  repoRelDir,
		Workspace:   workspace,
		ProjectName: projectName,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{},
		},
	}
}

type assertPlanLockDB struct {
	db.Database
	t            *testing.T
	locker       events.WorkingDirLocker
	repoFullName string
	pullNum      int
	called       *bool
}

func (a assertPlanLockDB) UpdatePullWithResults(pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	*a.called = true
	Assert(a.t, a.locker.HasCommandLock(a.repoFullName, a.pullNum, command.Plan), "expected plan lock during pull status write")
	return a.Database.UpdatePullWithResults(pull, results)
}

func (a assertPlanLockDB) BeginPlanGeneration(pull models.PullRequest, projects []models.ProjectStatus, generation string) (models.PullStatus, error) {
	*a.called = true
	Assert(a.t, a.locker.HasCommandLock(a.repoFullName, a.pullNum, command.Plan), "expected plan lock during plan generation invalidation")
	return a.Database.BeginPlanGeneration(pull, projects, generation)
}

func (a assertPlanLockDB) CompletePlanGeneration(pull models.PullRequest, generation string, results []command.ProjectResult) (models.PullStatus, error) {
	*a.called = true
	Assert(a.t, a.locker.HasCommandLock(a.repoFullName, a.pullNum, command.Plan), "expected plan lock during plan generation completion")
	return a.Database.CompletePlanGeneration(pull, generation, results)
}

func projectStatus(t *testing.T, pullStatus *models.PullStatus, workspace, repoRelDir, projectName string) models.ProjectStatus {
	t.Helper()
	for _, project := range pullStatus.Projects {
		if project.Workspace == workspace && project.RepoRelDir == repoRelDir && project.ProjectName == projectName {
			return project
		}
	}
	t.Fatalf("missing project status for dir %q workspace %q project %q", repoRelDir, workspace, projectName)
	return models.ProjectStatus{}
}

type failUnlockByPullLocker struct {
	t *testing.T
}

func (l failUnlockByPullLocker) TryLock(models.Project, string, models.PullRequest, models.User) (locking.TryLockResponse, error) {
	return locking.TryLockResponse{}, nil
}

func (l failUnlockByPullLocker) Unlock(string) (*models.ProjectLock, error) {
	return nil, nil
}

func (l failUnlockByPullLocker) List() (map[string]models.ProjectLock, error) {
	return nil, nil
}

func (l failUnlockByPullLocker) UnlockByPull(string, int) ([]models.ProjectLock, error) {
	l.t.Fatalf("no-project cleanup must not call UnlockByPull")
	return nil, nil
}

func (l failUnlockByPullLocker) UnlockIfOwnedByPull(models.Project, string, int) (*models.ProjectLock, error) {
	return nil, nil
}

func (l failUnlockByPullLocker) GetLock(string) (*models.ProjectLock, error) {
	return nil, nil
}

func installPlanCommandRunnerLocker(vcsClient *vcsmocks.MockClient, locker locking.Locker, silenceNoProjects bool) {
	planCommandRunner = events.NewPlanCommandRunner(
		false,
		false,
		vcsClient,
		pendingPlanFinder,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		commitUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		cancellationTracker,
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		autoMerger,
		1,
		silenceNoProjects,
		dbUpdater.Database,
		locker,
		false,
		pullReqStatusFetcher,
		false,
	)
	ch.CommentCommandRunnerByCmd[command.Plan] = planCommandRunner
}

func TestRunCommentCommand_DisableApplyAllDisabled(t *testing.T) {
	t.Log("if \"atlantis apply\" is run and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	applyCommandRunner.DisableApplyAll = true
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, modelPull.Num, &events.CommentCommand{Name: command.Apply})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num),
		Eq("**Error:** Running `atlantis apply` without flags is disabled. You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags."), Eq("apply"))
}

func TestRunCommentCommand_DisableAutoplan(t *testing.T) {
	t.Log("if \"DisableAutoplan\" is true, auto plans are disabled and we are silencing return and do not comment with error")
	setup(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "main"}

	ch.DisableAutoplan = true
	defer func() { ch.DisableAutoplan = false }()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(commitUpdater.UpdateCombinedCount(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())).ThenReturn(nil)
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)
	projectCommandBuilder.VerifyWasCalled(Never()).BuildAutoplanCommands(Any[*command.Context]())
}

func TestRunCommentCommand_DisableAutoplanLabel(t *testing.T) {
	t.Log("if \"DisableAutoplanLabel\" is present and pull request has that label, auto plans are disabled and we are silencing return and do not comment with error")
	vcsClient := setup(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "main"}

	ch.DisableAutoplanLabel = "disable-auto-plan"
	defer func() { ch.DisableAutoplanLabel = "" }()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(ch.VCSClient.GetPullLabels(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn([]string{"disable-auto-plan", "need-help"}, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)
	projectCommandBuilder.VerifyWasCalled(Never()).BuildAutoplanCommands(Any[*command.Context]())
	vcsClient.VerifyWasCalledOnce().GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
}

func TestRunCommentCommand_DisableAutoplanLabel_PullNotLabeled(t *testing.T) {
	t.Log("if \"DisableAutoplanLabel\" is present but pull request doesn't have that label, auto plans run")
	vcsClient := setup(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "main"}

	ch.DisableAutoplanLabel = "disable-auto-plan"
	defer func() { ch.DisableAutoplanLabel = "" }()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(ch.VCSClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn(nil, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)
	projectCommandBuilder.VerifyWasCalled(Once()).BuildAutoplanCommands(Any[*command.Context]())
	vcsClient.VerifyWasCalledOnce().GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
}

func TestRunCommentCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.Ptr("closed"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.ClosedPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq("Atlantis commands can't be run on closed pull requests"), Eq(""))
}

func TestRunCommentCommand_MatchedBranch(t *testing.T) {
	t.Log("if a command is run on a pull request which matches base branches run plan successfully")
	vcsClient := setup(t)

	ch.GlobalCfg.Repos = append(ch.GlobalCfg.Repos, valid.Repo{
		IDRegex:     regexp.MustCompile(".*"),
		BranchRegex: regexp.MustCompile("^main$"),
	})
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "main"}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq("Ran Plan for 0 projects:"), Eq("plan"))
}

func TestRunCommentCommand_UnmatchedBranch(t *testing.T) {
	t.Log("if a command is run on a pull request which doesn't match base branches do not comment with error")
	vcsClient := setup(t)

	ch.GlobalCfg.Repos = append(ch.GlobalCfg.Repos, valid.Repo{
		IDRegex:     regexp.MustCompile(".*"),
		BranchRegex: regexp.MustCompile("^main$"),
	})
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "foo"}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunUnlockCommand_VCSComment(t *testing.T) {
	testCases := []struct {
		name    string
		prState *string
	}{
		{
			name:    "PR open",
			prState: github.Ptr("open"),
		},

		{
			name:    "PR closed",
			prState: github.Ptr("closed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("if an unlock command is run on a pull request in state %s, atlantis should"+
				" invoke the delete command and comment on PR accordingly", *tc.prState)

			vcsClient := setup(t)
			pull := &github.PullRequest{
				State: tc.prState,
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
				Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo,
				testdata.GithubRepo, nil)

			ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num,
				&events.CommentCommand{Name: command.Unlock})

			deleteLockCommand.VerifyWasCalledOnce().DeleteLocksByPull(Any[logging.SimpleLogging](),
				Eq(testdata.GithubRepo.FullName), Eq(testdata.Pull.Num))
			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num),
				Eq("All Atlantis locks for this PR have been unlocked and plans discarded"), Eq("unlock"))
		})
	}
}

func TestRunUnlockCommandFail_VCSComment(t *testing.T) {
	t.Log("if unlock PR command is run and delete fails, atlantis should" +
		" invoke comment on PR with error message")

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
	When(deleteLockCommand.DeleteLocksByPull(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo.FullName),
		Eq(testdata.Pull.Num))).ThenReturn(0, errors.New("err"))

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num,
		&events.CommentCommand{Name: command.Unlock})

	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("Failed to delete PR locks"), Eq("unlock"))
}

func TestRunUnlockCommandFail_DisableUnlockLabel(t *testing.T) {
	t.Log("if PR has label equal to disable-unlock-label unlock should fail")

	doNotUnlock := "do-not-unlock"

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
	When(deleteLockCommand.DeleteLocksByPull(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo.FullName),
		Eq(testdata.Pull.Num))).ThenReturn(0, errors.New("err"))
	When(ch.VCSClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(modelPull))).ThenReturn([]string{doNotUnlock, "need-help"}, nil)

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num,
		&events.CommentCommand{Name: command.Unlock})

	vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(testdata.Pull.Num), Eq("Not allowed to unlock PR with "+doNotUnlock+" label"), Eq("unlock"))
}

func TestRunUnlockCommandFail_GetLabelsFail(t *testing.T) {
	t.Log("if GetPullLabels fails do not unlock PR")

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
	When(deleteLockCommand.DeleteLocksByPull(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo.FullName),
		Eq(testdata.Pull.Num))).ThenReturn(0, errors.New("err"))
	When(ch.VCSClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(modelPull))).ThenReturn(nil, errors.New("err"))

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num,
		&events.CommentCommand{Name: command.Unlock})

	vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num),
		Eq("Failed to retrieve PR labels... Not unlocking"), Eq("unlock"))
}

func TestRunUnlockCommandDoesntRetrieveLabelsIfDisableUnlockLabelNotSet(t *testing.T) {
	t.Log("if disable-unlock-label is not set do not call GetPullLabels")

	doNotUnlock := "do-not-unlock"

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
	When(deleteLockCommand.DeleteLocksByPull(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo.FullName),
		Eq(testdata.Pull.Num))).ThenReturn(0, errors.New("err"))
	When(ch.VCSClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo),
		Eq(modelPull))).ThenReturn([]string{doNotUnlock, "need-help"}, nil)
	unlockCommandRunner.DisableUnlockLabel = ""

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num,
		&events.CommentCommand{Name: command.Unlock})

	vcsClient.VerifyWasCalled(Never()).GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
}

func TestRunAutoplanCommand_DeletePlans(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
}

func TestRunAutoplan_NoProjectsWritesCurrentEmptyPullStatus(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "old-project",
			Workspace:  events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunAutoplan_NoProjectsMissingPullDirWritesEmptyPullStatus(t *testing.T) {
	setup(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn("", os.ErrNotExist)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunAutoplan_NoProjectsRetainsOldPlanFilesAndClearsPullStatus(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	oldPlanDir := filepath.Join(tmp, "old-project")
	Ok(t, os.MkdirAll(oldPlanDir, 0700))
	oldPlanPath := filepath.Join(oldPlanDir, runtime.GetPlanFilename(events.DefaultWorkspace, ""))
	Ok(t, os.WriteFile(oldPlanPath, []byte("old plan"), 0600))
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "old-project",
			Workspace:  events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{}, nil)
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	contents, err := os.ReadFile(oldPlanPath)
	Ok(t, err)
	Equals(t, "old plan", string(contents))
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunAutoplan_NoProjectsDoesNotUnlockActiveLocks(t *testing.T) {
	vcsClient := setup(t)
	installPlanCommandRunnerLocker(vcsClient, failUnlockByPullLocker{t: t}, false)

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{}, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
}

func TestRunAutoplan_NoProjectsEmptyPullStatusWriteFailureIsUserVisible(t *testing.T) {
	vcsClient := setup(t)
	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.AutoTrigger,
	}
	closer := dbUpdater.Database.(interface{ Close() error })
	Ok(t, closer.Close())

	When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	planCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Plan})

	Assert(t, ctx.CommandHasErrors, "expected empty PullStatus write failure to mark autoplan errored")
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq(models.FailedCommitStatus), Eq(command.Plan))
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Any[string](), Eq("plan")).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Plan Error"), "got: %s", comment)
	Assert(t, strings.Contains(comment, "writing empty plan status"), "got: %s", comment)
}

func TestRunAutoplanCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_False(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = false
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Eq(models.PendingCommitStatus), Eq(command.Plan))
	commitUpdater.VerifyWasCalled(Never()).UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Eq(models.FailedCommitStatus), Any[command.Name]())
}

func TestRunAutoplanCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_True(t *testing.T) {
	vcsClient := setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = true
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	// gomock will fail if lockingLocker.UnlockByPull is called unexpectedly (no EXPECT set)
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Eq(models.FailedCommitStatus), Eq(command.Plan))

	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()

	Assert(t, strings.Contains(comment, "Error: Pre-workflow hook failed"), fmt.Sprintf("comment should be about a pre-workflow hook failure but was %q", comment))
}

func TestRunCommentCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_False(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = false
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
}

func TestRunCommentCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_True(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = true
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())
	// gomock will fail if lockingLocker.UnlockByPull is called unexpectedly (no EXPECT set)
}

func TestRunGenericPlanCommand_DeletePlans(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	projectCtx := command.ProjectContext{
		CommandName:         command.Plan,
		ExecutionOrderGroup: 0,
		ProjectName:         "TestProject",
		Workspace:           "default",
		BaseRepo:            testdata.GithubRepo,
		Pull:                testdata.Pull,
	}
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{projectCtx}, nil)
	When(projectCommandRunner.Plan(projectCtx)).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: "true",
		},
	})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
}

func TestRunSpecificPlanCommandDoesnt_DeletePlans(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan, ProjectName: "default"})
	pendingPlanFinder.VerifyWasCalled(Never()).Find(tmp)
}

// Test that if one plan fails and we are using automerge, that
// we delete the plans.
func TestRunAutoplanCommandWithError_DeletePlans(t *testing.T) {
	vcsClient := setup(t)

	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName:      command.Plan,
				AutomergeEnabled: true, // Setting this manually, since this tests bypasses automerge param reconciliation logic and otherwise defaults to false.
			},
			{
				CommandName:      command.Plan,
				AutomergeEnabled: true, // Setting this manually, since this tests bypasses automerge param reconciliation logic and otherwise defaults to false.
			},
		}, nil)
	callCount := 0
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).Then(func(_ []Param) ReturnValues {
		if callCount == 0 {
			// The first call, we return a successful result.
			callCount++
			return ReturnValues{
				command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			}
		}
		// The second call, we return a failed result.
		return ReturnValues{
			command.ProjectCommandOutput{
				Error: errors.New("err"),
			},
		}
	})

	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	// gets called twice: the first time before the plan starts, the second time after the plan errors
	pendingPlanFinder.VerifyWasCalled(Times(2)).Find(tmp)

	vcsClient.VerifyWasCalled(Times(0)).DiscardReviews(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
}

func TestRunGenericPlanCommand_DiscardApprovals(t *testing.T) {
	vcsClient := setup(t, func(testConfig *TestConfig) {
		testConfig.discardApprovalOnPlan = true
	})

	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalled(Never()).Find(Any[string]())

	vcsClient.VerifyWasCalledOnce().DiscardReviews(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
}

func TestApplyMergeablityWhenPolicyCheckFails(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with failing policy check then apply is not performed")
	setup(t)
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}

	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	_, _ = boltDB.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command: command.PolicyCheck,
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: fmt.Errorf("failing policy"),
			},
			ProjectName: "default",
			Workspace:   "default",
			RepoRelDir:  ".",
		},
	})

	When(ch.VCSClient.PullIsMergeable(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq("atlantis-test"), Eq([]string{}))).ThenReturn(models.MergeableStatus{
		IsMergeable: true,
	}, nil)

	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(args []Param) ReturnValues {
		return ReturnValues{
			[]command.ProjectContext{
				{
					CommandName:       command.Apply,
					ProjectName:       "default",
					Workspace:         "default",
					RepoRelDir:        ".",
					ProjectPlanStatus: models.ErroredPolicyCheckStatus,
				},
			},
			nil,
		}
	})

	When(workingDir.GetPullDir(testdata.GithubRepo, modelPull)).ThenReturn(tmp, nil)
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, &modelPull, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
}

func TestApplyWithAutoMerge_VSCMerge(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge then a VCS merge is performed")

	vcsClient, modelPull := setupApplyWithAutoMerge(t)

	pullOptions := models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
	}

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
	vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Eq(modelPull), Eq(pullOptions))
}

func TestApplyWithAutoMerge_UsesGlobalMergeMethod(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge and no --auto-merge-method" +
		" comment flag, the server-side default merge method is used")

	vcsClient, modelPull := setupApplyWithAutoMerge(t)
	autoMerger.GlobalAutomergeMethod = "squash"
	t.Cleanup(func() { autoMerger.GlobalAutomergeMethod = "" })

	pullOptions := models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
		MergeMethod:               "squash",
	}

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
	vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Eq(modelPull), Eq(pullOptions))
}

func TestApplyWithAutoMerge_CommentFlagOverridesGlobalMergeMethod(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge and a --auto-merge-method" +
		" comment flag, the comment flag overrides the server-side default")

	vcsClient, modelPull := setupApplyWithAutoMerge(t)
	autoMerger.GlobalAutomergeMethod = "squash"
	t.Cleanup(func() { autoMerger.GlobalAutomergeMethod = "" })

	pullOptions := models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
		MergeMethod:               "rebase",
	}

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply, AutoMergeMethod: "rebase"})
	vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Eq(modelPull), Eq(pullOptions))
}

func TestApplyWithAutoMerge_DisableAutomergeLabelStillMergesWhenLabelMissing(t *testing.T) {
	t.Log("if disable-automerge-label is configured but the pull request does not have that label, automerge is performed")

	disableAutomergeLabel := "no-auto-merge"
	vcsClient, modelPull := setupApplyWithAutoMerge(t, func(testConfig *TestConfig) {
		testConfig.DisableAutomergeLabel = disableAutomergeLabel
	})
	When(vcsClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn([]string{"safe-to-merge"}, nil)

	pullOptions := models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
	}

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	vcsClient.VerifyWasCalledOnce().GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
	vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Eq(modelPull), Eq(pullOptions))
}

func TestApplyWithAutoMerge_DisableAutomergeLabelSkipsMergeWhenLabelPresent(t *testing.T) {
	t.Log("if disable-automerge-label is configured and the pull request has that label, automerge is not performed")

	disableAutomergeLabel := "no-auto-merge"
	vcsClient, modelPull := setupApplyWithAutoMerge(t, func(testConfig *TestConfig) {
		testConfig.DisableAutomergeLabel = disableAutomergeLabel
	})
	When(vcsClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn([]string{disableAutomergeLabel, "need-help"}, nil)

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	vcsClient.VerifyWasCalledOnce().GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
	vcsClient.VerifyWasCalled(Never()).MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
}

func TestApplyWithAutoMerge_DisableAutomergeLabelSkipsMergeWhenLabelsCannotBeFetched(t *testing.T) {
	t.Log("if disable-automerge-label is configured and pull request labels cannot be fetched, automerge is not performed")

	disableAutomergeLabel := "no-auto-merge"
	vcsClient, modelPull := setupApplyWithAutoMerge(t, func(testConfig *TestConfig) {
		testConfig.DisableAutomergeLabel = disableAutomergeLabel
	})
	When(vcsClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn(nil, errors.New("err"))

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	vcsClient.VerifyWasCalledOnce().GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
	vcsClient.VerifyWasCalled(Never()).MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
}

func setupApplyWithAutoMerge(t *testing.T, options ...func(testConfig *TestConfig)) (*vcsmocks.MockClient, models.PullRequest) {
	t.Helper()

	vcsClient := setup(t, options...)
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	_, err := dbUpdater.Database.UpdatePullWithResults(modelPull, nil)
	Ok(t, err)
	autoMerger.GlobalAutomerge = true
	autoMerger.GlobalAutomergeMethod = ""
	t.Cleanup(func() {
		autoMerger.GlobalAutomerge = false
		autoMerger.GlobalAutomergeMethod = ""
	})

	return vcsClient, modelPull
}

func TestRunApply_DiscardedProjects(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge and at least one project" +
		" has a discarded plan, automerge should not take place")
	vcsClient := setup(t)
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() {
		boltDB.Close()
	})
	Ok(t, err)
	dbUpdater.Database = boltDB
	applyCommandRunner.Database = boltDB
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo
	_, err = boltDB.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: ".",
			Workspace:  "default",
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "tf-output",
					LockURL:         "lock-url",
				},
			},
		},
	})
	Ok(t, err)
	Ok(t, boltDB.UpdateProjectStatus(pull, "default", ".", models.DiscardedPlanStatus))
	ghPull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(ghPull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(ghPull))).ThenReturn(pull, pull.BaseRepo, testdata.GithubRepo, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(tmp, nil)
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, &pull, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

	vcsClient.VerifyWasCalled(Never()).MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
}

func TestRunCommentCommand_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := setup(t)
	drainer.ShutdownBlocking()
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("Atlantis server is shutting down, please try again later."), Eq(""))
}

func TestRunCommentCommand_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occurred")
	setup(t)
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenPanic(
		"panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	githubGetter.VerifyWasCalledOnce().GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}

func TestRunAutoplanCommand_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := setup(t)
	drainer.ShutdownBlocking()
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("Atlantis server is shutting down, please try again later."), Eq("plan"))
}

func TestRunAutoplanCommand_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occurred")
	setup(t)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	projectCommandBuilder.VerifyWasCalledOnce().BuildAutoplanCommands(Any[*command.Context]())
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}
