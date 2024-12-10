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

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"

	"github.com/google/go-github/v66/github"
	. "github.com/petergtz/pegomock/v4"
	lockingmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
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
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
var postWorkflowHooksCommandRunner events.PostWorkflowHooksCommandRunner

type TestConfig struct {
	parallelPoolSize           int
	SilenceNoProjects          bool
	silenceVCSStatusNoPlans    bool
	silenceVCSStatusNoProjects bool
	StatusName                 string
	discardApprovalOnPlan      bool
	backend                    locking.Backend
	DisableUnlockLabel         string
}

func setup(t *testing.T, options ...func(testConfig *TestConfig)) *vcsmocks.MockClient {
	RegisterMockTestingT(t)

	// create an empty DB
	tmp := t.TempDir()
	defaultBoltDB, err := db.New(tmp)
	Ok(t, err)

	testConfig := &TestConfig{
		parallelPoolSize:      1,
		SilenceNoProjects:     false,
		StatusName:            "atlantis-test",
		discardApprovalOnPlan: false,
		backend:               defaultBoltDB,
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

	drainer = &events.Drainer{}
	deleteLockCommand = mocks.NewMockDeleteLockCommand()
	applyLockChecker = lockingmocks.NewMockApplyLockChecker()
	lockingLocker = lockingmocks.NewMockLocker()

	dbUpdater = &events.DBUpdater{
		Backend: testConfig.backend,
	}

	pullUpdater = &events.PullUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     events.NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false),
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
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		autoMerger,
		testConfig.parallelPoolSize,
		testConfig.SilenceNoProjects,
		testConfig.backend,
		lockingLocker,
		testConfig.discardApprovalOnPlan,
		pullReqStatusFetcher,
	)

	applyCommandRunner = events.NewApplyCommandRunner(
		vcsClient,
		false,
		applyLockChecker,
		commitUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		autoMerger,
		pullUpdater,
		dbUpdater,
		testConfig.backend,
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

	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	postWorkflowHooksCommandRunner = mocks.NewMockPostWorkflowHooksCommandRunner()

	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{})
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

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
		PullStatusFetcher:              testConfig.backend,
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
		vcsClient.VerifyWasCalled(Never()).GetTeamNamesForUser(testdata.GithubRepo, testdata.User)
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
		vcsClient.VerifyWasCalled(Never()).GetTeamNamesForUser(testdata.GithubRepo, testdata.User)
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
		Eq(0),
		Eq(0),
	)
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
		Eq(0),
		Eq(0),
	)
}

func TestRunCommentCommandApply_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an apply command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	applyCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
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
		Eq(0),
		Eq(0),
	)
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
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.PolicyCheck),
		Eq(0),
		Eq(0),
	)
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

func TestRunCommentCommand_DisableApplyAllDisabled(t *testing.T) {
	t.Log("if \"atlantis apply\" is run and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	applyCommandRunner.DisableApplyAll = true
	pull := &github.PullRequest{
		State: github.String("open"),
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
		State: github.String("closed"),
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
			prState: github.String("open"),
		},

		{
			name:    "PR closed",
			prState: github.String("closed"),
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
		State: github.String("open"),
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
		State: github.String("open"),
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
		State: github.String("open"),
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
		State: github.String("open"),
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
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
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
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
	lockingLocker.VerifyWasCalledOnce().UnlockByPull(testdata.Pull.BaseRepo.FullName, testdata.Pull.Num)
}

func TestRunAutoplanCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_False(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = false
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
	lockingLocker.VerifyWasCalledOnce().UnlockByPull(testdata.Pull.BaseRepo.FullName, testdata.Pull.Num)
}

func TestRunAutoplanCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_True(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB

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
	pendingPlanFinder.VerifyWasCalled(Never()).DeletePlans(Any[string]())
	lockingLocker.VerifyWasCalled(Never()).UnlockByPull(Any[string](), Any[int]())
}

func TestRunCommentCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_False(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	pull := &github.PullRequest{State: github.String("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = false
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
	lockingLocker.VerifyWasCalledOnce().UnlockByPull(testdata.Pull.BaseRepo.FullName, testdata.Pull.Num)
}

func TestRunCommentCommand_FailedPreWorkflowHook_FailOnPreWorkflowHookError_True(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(errors.New("err"))
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.FailOnPreWorkflowHookError = true
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalled(Never()).DeletePlans(Any[string]())
	lockingLocker.VerifyWasCalled(Never()).UnlockByPull(Any[string](), Any[int]())
}

func TestRunGenericPlanCommand_DeletePlans(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	pull := &github.PullRequest{State: github.String("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
	lockingLocker.VerifyWasCalledOnce().UnlockByPull(testdata.Pull.BaseRepo.FullName, testdata.Pull.Num)
}

func TestRunSpecificPlanCommandDoesnt_DeletePlans(t *testing.T) {
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan, ProjectName: "default"})
	pendingPlanFinder.VerifyWasCalled(Never()).DeletePlans(tmp)
}

// Test that if one plan fails and we are using automerge, that
// we delete the plans.
func TestRunAutoplanCommandWithError_DeletePlans(t *testing.T) {
	vcsClient := setup(t)

	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
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
				command.ProjectResult{
					PlanSuccess: &models.PlanSuccess{},
				},
			}
		}
		// The second call, we return a failed result.
		return ReturnValues{
			command.ProjectResult{
				Error: errors.New("err"),
			},
		}
	})

	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(tmp, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, testdata.Pull, testdata.User)
	// gets called twice: the first time before the plan starts, the second time after the plan errors
	pendingPlanFinder.VerifyWasCalled(Times(2)).DeletePlans(tmp)

	vcsClient.VerifyWasCalled(Times(0)).DiscardReviews(Any[models.Repo](), Any[models.PullRequest]())
}

func TestRunGenericPlanCommand_DiscardApprovals(t *testing.T) {
	vcsClient := setup(t, func(testConfig *TestConfig) {
		testConfig.discardApprovalOnPlan = true
	})

	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(tmp, nil)
	pull := &github.PullRequest{State: github.String("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
	testdata.Pull.BaseRepo = testdata.GithubRepo
	ch.RunCommentCommand(testdata.GithubRepo, nil, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Plan})
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
	lockingLocker.VerifyWasCalledOnce().UnlockByPull(testdata.Pull.BaseRepo.FullName, testdata.Pull.Num)

	vcsClient.VerifyWasCalledOnce().DiscardReviews(Any[models.Repo](), Any[models.PullRequest]())
}

func TestFailedApprovalCreatesFailedStatusUpdate(t *testing.T) {
	t.Log("if \"atlantis approve_policies\" is run by non policy owner policy check status fails.")
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.String("open"),
	}

	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	When(projectCommandBuilder.BuildApprovePoliciesCommands(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn([]command.ProjectContext{
		{
			CommandName: command.ApprovePolicies,
		},
		{
			CommandName: command.ApprovePolicies,
		},
	}, nil)

	When(workingDir.GetPullDir(testdata.GithubRepo, testdata.Pull)).ThenReturn(tmp, nil)

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, &testdata.Pull, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.ApprovePolicies})
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.PolicyCheck),
		Eq(0),
		Eq(2),
	)
}

func TestApprovedPoliciesUpdateFailedPolicyStatus(t *testing.T) {
	t.Log("if \"atlantis approve_policies\" is run by policy owner all policy checks are approved.")
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.String("open"),
	}

	modelPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	When(projectCommandBuilder.BuildApprovePoliciesCommands(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn([]command.ProjectContext{
		{
			CommandName: command.ApprovePolicies,
			PolicySets: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Users: []string{testdata.User.Username},
				},
			},
		},
	}, nil)

	When(workingDir.GetPullDir(testdata.GithubRepo, testdata.Pull)).ThenReturn(tmp, nil)
	When(projectCommandRunner.ApprovePolicies(Any[command.ProjectContext]())).Then(func(_ []Param) ReturnValues {
		return ReturnValues{
			command.ProjectResult{
				Command:            command.PolicyCheck,
				PolicyCheckResults: &models.PolicyCheckResults{},
			},
		}
	})

	ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, &testdata.Pull, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.ApprovePolicies})
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq[models.CommitStatus](models.SuccessCommitStatus),
		Eq[command.Name](command.PolicyCheck),
		Eq(1),
		Eq(1),
	)
}

func TestApplyMergeablityWhenPolicyCheckFails(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with failing policy check then apply is not performed")
	setup(t)
	tmp := t.TempDir()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.String("open"),
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
			Command:     command.PolicyCheck,
			Error:       fmt.Errorf("failing policy"),
			ProjectName: "default",
			Workspace:   "default",
			RepoRelDir:  ".",
		},
	})

	When(ch.VCSClient.PullIsMergeable(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull), Eq("atlantis-test"), Eq([]string{}))).ThenReturn(true, nil)

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

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
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
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.Backend = boltDB
	applyCommandRunner.Backend = boltDB
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo
	_, err = boltDB.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: ".",
			Workspace:  "default",
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "tf-output",
				LockURL:         "lock-url",
			},
		},
	})
	Ok(t, err)
	Ok(t, boltDB.UpdateProjectStatus(pull, "default", ".", models.DiscardedPlanStatus))
	ghPull := &github.PullRequest{
		State: github.String("open"),
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
