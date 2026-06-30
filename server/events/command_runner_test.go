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
		lockingLocker,
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

func TestRunCommentCommandPlan_NoProjectsClearsOldPlanFilesAndPullStatus(t *testing.T) {
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
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{
		{RepoDir: tmp, RepoRelDir: "old-project", Workspace: events.DefaultWorkspace},
	}, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
	_, err = os.Stat(oldPlanPath)
	Assert(t, os.IsNotExist(err), "expected old plan file to be removed")
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunCommentCommandPlan_NoProjectsNonSilencedUsesSafeCleanup(t *testing.T) {
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
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{
		{RepoDir: tmp, RepoRelDir: "old-project", Workspace: events.DefaultWorkspace},
	}, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
	_, err = os.Stat(oldPlanPath)
	Assert(t, os.IsNotExist(err), "expected old plan file to be removed")
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunCommentCommandPlan_NoProjectsDoesNotUnlockActiveLocks(t *testing.T) {
	vcsClient := setup(t)
	installPlanCommandRunnerLocker(vcsClient, failUnlockByPullLocker{t: t}, true)

	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
}

func TestRunCommentCommandPlan_NoProjectsNonSilencedDoesNotUnlockByPull(t *testing.T) {
	vcsClient := setup(t)
	installPlanCommandRunnerLocker(vcsClient, failUnlockByPullLocker{t: t}, false)

	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(&pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
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

func TestRunAutoplan_NoProjectsClearsOldPlanFilesAndPullStatus(t *testing.T) {
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
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(pendingPlanFinder.Find(tmp)).ThenReturn([]events.PendingPlan{
		{RepoDir: tmp, RepoRelDir: "old-project", Workspace: events.DefaultWorkspace},
	}, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
	_, err = os.Stat(oldPlanPath)
	Assert(t, os.IsNotExist(err), "expected old plan file to be removed")
	pullStatus, err := dbUpdater.Database.GetPullStatus(modelPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected current empty PullStatus")
	Equals(t, "abc123", pullStatus.Pull.HeadCommit)
	Equals(t, 0, len(pullStatus.Projects))
}

func TestRunAutoplan_NoProjectsDoesNotUnlockActiveLocks(t *testing.T) {
	vcsClient := setup(t)
	installPlanCommandRunnerLocker(vcsClient, failUnlockByPullLocker{t: t}, false)

	tmp := t.TempDir()
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)

	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
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
	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)
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
	pendingPlanFinder.VerifyWasCalledOnce().Find(tmp)

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
