// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"

	"github.com/google/go-github/v83/github"
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
var workingDir *mocks.MockWorkingDir
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
var preWorkflowHooksCommandRunner *mocks.MockPreWorkflowHooksCommandRunner
var postWorkflowHooksCommandRunner *mocks.MockPostWorkflowHooksCommandRunner
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
	PendingApplyStatus         bool
	applyLockCheckerReturn     locking.ApplyCommandLock
	applyLockCheckerErr        error
}

func setup(t *testing.T, options ...func(testConfig *TestConfig)) *vcsmocks.MockClient {
	// RegisterMockTestingT is still needed for vcsmocks (pegomock)
	RegisterMockTestingT(t)
	ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())

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

	projectCommandBuilder = mocks.NewMockProjectCommandBuilder(ctrl)
	eventParsing = mocks.NewMockEventParsing(ctrl)
	vcsClient := vcsmocks.NewMockClient()
	githubGetter = mocks.NewMockGithubPullGetter(ctrl)
	gitlabGetter = mocks.NewMockGitlabMergeRequestGetter(ctrl)
	azuredevopsGetter = mocks.NewMockAzureDevopsPullGetter(ctrl)
	logger := logging.NewNoopLogger(t)
	projectCommandRunner = mocks.NewMockProjectCommandRunner(ctrl)
	workingDir = mocks.NewMockWorkingDir(ctrl)
	pendingPlanFinder = mocks.NewMockPendingPlanFinder(ctrl)
	commitUpdater = mocks.NewMockCommitStatusUpdater(ctrl)
	pullReqStatusFetcher = vcsmocks.NewMockPullReqStatusFetcher()
	cancellationTracker = mocks.NewMockCancellationTracker(ctrl)

	drainer = &events.Drainer{}
	deleteLockCommand = mocks.NewMockDeleteLockCommand(ctrl)
	applyLockChecker = lockingmocks.NewMockApplyLockChecker(ctrl)
	lockingLocker = lockingmocks.NewMockLocker(ctrl)
	// Allow incidental calls to CheckApplyLock (called internally during apply operations).
	// Tests that need specific return values should set applyLockCheckerReturn/applyLockCheckerErr in TestConfig.
	applyLockChecker.EXPECT().CheckApplyLock().Return(testConfig.applyLockCheckerReturn, testConfig.applyLockCheckerErr).AnyTimes()
	// Allow incidental calls to UnlockByPull (called during plan operations to clean up locks)
	lockingLocker.EXPECT().UnlockByPull(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	// Allow incidental gomock calls that may or may not happen depending on test path.
	// These provide default no-op returns. Tests that need specific behavior will
	// register new EXPECT() calls which take priority (LIFO matching in gomock).
	projectCommandBuilder.EXPECT().BuildAutoplanCommands(gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildApprovePoliciesCommands(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildVersionCommands(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildImportCommands(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildStateRmCommands(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	projectCommandRunner.EXPECT().Apply(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	projectCommandRunner.EXPECT().ApprovePolicies(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	projectCommandRunner.EXPECT().PolicyCheck(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	projectCommandRunner.EXPECT().Version(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	projectCommandRunner.EXPECT().Import(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	projectCommandRunner.EXPECT().StateRm(gomock.Any()).Return(command.ProjectCommandOutput{}).AnyTimes()
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	gitlabGetter.EXPECT().GetMergeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	azuredevopsGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Any()).Return(models.PullRequest{}, models.Repo{}, models.Repo{}, nil).AnyTimes()
	commitUpdater.EXPECT().UpdateCombined(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	commitUpdater.EXPECT().UpdateCombinedCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	commitUpdater.EXPECT().UpdatePreWorkflowHook(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	commitUpdater.EXPECT().UpdatePostWorkflowHook(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deleteLockCommand.EXPECT().DeleteLocksByPull(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil).AnyTimes()
	pendingPlanFinder.EXPECT().DeletePlans(gomock.Any()).Return(nil).AnyTimes()
	pendingPlanFinder.EXPECT().Find(gomock.Any()).Return(nil, nil).AnyTimes()
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	workingDir.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workingDir.EXPECT().DeleteForWorkspace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	cancellationTracker.EXPECT().Cancel(gomock.Any()).AnyTimes()
	cancellationTracker.EXPECT().Clear(gomock.Any()).AnyTimes()
	cancellationTracker.EXPECT().IsCancelled(gomock.Any()).Return(false).AnyTimes()

	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner(ctrl)
	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	postWorkflowHooksCommandRunner = mocks.NewMockPostWorkflowHooksCommandRunner(ctrl)
	postWorkflowHooksCommandRunner.EXPECT().RunPostHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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

	planCommandRunner = events.NewPlanCommandRunner(
		testConfig.silenceVCSStatusNoPlans,
		testConfig.silenceVCSStatusNoProjects,
		vcsClient,
		pendingPlanFinder,
		workingDir,
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
		pullReqStatusFetcher,
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
		pullReqStatusFetcher,
		projectCommandBuilder,
		projectCommandRunner,
		testConfig.SilenceNoProjects,
	)

	commentCommandRunnerByCmd := map[command.Name]events.CommentCommandRunner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
		command.Import:          importCommandRunner,
	}

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).DoAndReturn(
		func(_ logging.SimpleLogging, _ models.Repo, _ int) (*github.PullRequest, error) {
			panic("panic test - if you're seeing this in a test failure this isn't the failing test")
		})
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, 1, &events.CommentCommand{Name: command.Plan})
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), fmt.Sprintf("comment should be about a goroutine panic but was %q", comment))
}

func TestRunCommentCommand_GithubPullErr(t *testing.T) {
	t.Log("if getting the github pull request fails an error should be logged")
	vcsClient := setup(t)
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(nil, errors.New("err"))
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("`Error: making pull request API call to GitHub: err`"), Eq(""))
}

func TestRunCommentCommand_GitlabMergeRequestErr(t *testing.T) {
	t.Log("if getting the gitlab merge request fails an error should be logged")
	vcsClient := setup(t)
	gitlabGetter.EXPECT().GetMergeRequest(gomock.Any(), gomock.Eq(testdata.GitlabRepo.FullName), gomock.Eq(testdata.Pull.Num)).Return(nil, errors.New("err"))
	ch.RunCommentCommand(testdata.GitlabRepo, &testdata.GitlabRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GitlabRepo), Eq(testdata.Pull.Num), Eq("`Error: making merge request API call to GitLab: err`"), Eq(""))
}

func TestRunCommentCommand_GithubPullParseErr(t *testing.T) {
	t.Log("if parsing the returned github pull request fails an error should be logged")
	vcsClient := setup(t)
	var pull github.PullRequest
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(testdata.Pull, testdata.GithubRepo, testdata.GitlabRepo, errors.New("err"))

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Eq("`Error: extracting required fields from comment data: err`"), Eq(""))
}

func TestRunCommentCommand_CommentPreWorkflowHookFailure(t *testing.T) {
	t.Log("if there is a pre-workflowhook failure with failure enabled it is commented back on the pull request")
	vcsClient := setup(t)
	ch.FailOnPreWorkflowHookError = true

	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(errors.New("err"))

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
		githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
		eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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
		githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
		eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)

	headRepo := testdata.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, headRepo, nil)

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)

	headRepo := testdata.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, headRepo, nil)

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunCommentCommandPlan_NoProjectsTarget_SilenceEnabled(t *testing.T) {
	// TODO
	t.Log("if a plan command is run against a project and SilenceNoProjects is enabled, we are silencing all comments if the project is not in the repo config")
	vcsClient := setup(t)
	planCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan, ProjectName: "meow"})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunCommentCommandApply_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an apply command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	applyCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunCommentCommandApprovePolicy_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an approve_policy command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	approvePoliciesCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Unlock})
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestRunCommentCommandImport_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an import command is run on a pull request and SilenceNoProjects is enabled, we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	importCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Import})
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)
}

func TestRunCommentCommand_DisableAutoplanLabel(t *testing.T) {
	t.Log("if \"DisableAutoplanLabel\" is present and pull request has that label, auto plans are disabled and we are silencing return and do not comment with error")
	vcsClient := setup(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "main"}

	ch.DisableAutoplanLabel = "disable-auto-plan"
	defer func() { ch.DisableAutoplanLabel = "" }()

	When(ch.VCSClient.GetPullLabels(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn([]string{"disable-auto-plan", "need-help"}, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)
	vcsClient.VerifyWasCalledOnce().GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))
}

func TestRunCommentCommand_DisableAutoplanLabel_PullNotLabeled(t *testing.T) {
	t.Log("if \"DisableAutoplanLabel\" is present but pull request doesn't have that label, auto plans run")
	vcsClient := setup(t)
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, BaseBranch: "main"}

	ch.DisableAutoplanLabel = "disable-auto-plan"
	defer func() { ch.DisableAutoplanLabel = "" }()

	When(ch.VCSClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull))).ThenReturn(nil, nil)

	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, modelPull, testdata.User)
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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(&pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(&pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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
			githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo),
				gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
			eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo,
				testdata.GithubRepo, nil)

			ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num,
				&events.CommentCommand{Name: command.Unlock})

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo),
		gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
	deleteLockCommand.EXPECT().DeleteLocksByPull(gomock.Any(), gomock.Eq(testdata.GithubRepo.FullName),
		gomock.Eq(testdata.Pull.Num)).Return(0, errors.New("err"))

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo),
		gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo),
		gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo),
		gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo,
		testdata.GithubRepo, nil)
	deleteLockCommand.EXPECT().DeleteLocksByPull(gomock.Any(), gomock.Eq(testdata.GithubRepo.FullName),
		gomock.Eq(testdata.Pull.Num)).Return(0, errors.New("err"))
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

	projectCommandBuilder.EXPECT().BuildAutoplanCommands(gomock.Any()).
		Return([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)
	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}).AnyTimes()
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
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

	projectCommandBuilder.EXPECT().BuildAutoplanCommands(gomock.Any()).
		Return([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
		}, nil)
	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return(tmp, nil)
	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = false
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
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

	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = true
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)

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

	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}).AnyTimes()
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return(tmp, nil)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = false
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
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

	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = true
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
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
	projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).
		Return([]command.ProjectContext{projectCtx}, nil)
	projectCommandRunner.EXPECT().Plan(gomock.Eq(projectCtx)).Return(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: "true",
		},
	})
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return(tmp, nil)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
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

	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}).AnyTimes()
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return(tmp, nil).AnyTimes()
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan, ProjectName: "default"})
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

	projectCommandBuilder.EXPECT().BuildAutoplanCommands(gomock.Any()).
		Return([]command.ProjectContext{
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
	projectCommandRunner.EXPECT().Plan(gomock.Any()).DoAndReturn(func(_ command.ProjectContext) command.ProjectCommandOutput {
		if callCount == 0 {
			// The first call, we return a successful result.
			callCount++
			return command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			}
		}
		// The second call, we return a failed result.
		return command.ProjectCommandOutput{
			Error: errors.New("err"),
		}
	}).AnyTimes()

	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).
		Return(tmp, nil).AnyTimes()
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)

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

	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}).AnyTimes()
	workingDir.EXPECT().GetPullDir(gomock.Any(), gomock.Any()).Return(tmp, nil)
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})

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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

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

	projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ *command.Context, _ *events.CommentCommand) ([]command.ProjectContext, error) {
			return []command.ProjectContext{
				{
					CommandName:       command.Apply,
					ProjectName:       "default",
					Workspace:         "default",
					RepoRelDir:        ".",
					ProjectPlanStatus: models.ErroredPolicyCheckStatus,
				},
			}, nil
		})

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, &modelPull, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
}

func TestApplyWithAutoMerge_VSCMerge(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge then a VCS merge is performed")

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.Ptr("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(pull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(pull)).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pullOptions := models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
	}

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})
	vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Eq(modelPull), Eq(pullOptions))
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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).Return(ghPull, nil)
	eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Eq(ghPull)).Return(pull, pull.BaseRepo, testdata.GithubRepo, nil)
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
	githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num)).DoAndReturn(
		func(_ logging.SimpleLogging, _ models.Repo, _ int) (*github.PullRequest, error) {
			panic("panic test - if you're seeing this in a test failure this isn't the failing test")
		})
	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
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
	projectCommandBuilder.EXPECT().BuildAutoplanCommands(gomock.Any()).DoAndReturn(
		func(_ *command.Context) ([]command.ProjectContext, error) {
			panic("panic test - if you're seeing this in a test failure this isn't the failing test")
		})
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}
