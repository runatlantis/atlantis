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
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
	lyft_vcs "github.com/runatlantis/atlantis/server/events/vcs/lyft"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"

	"github.com/google/go-github/v31/github"
	. "github.com/petergtz/pegomock"
	lockingmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	eventmocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
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
var fa events.ForceApplyCommandRunner
var workingDir events.WorkingDir
var pendingPlanFinder *mocks.MockPendingPlanFinder
var drainer *events.Drainer
var deleteLockCommand *mocks.MockDeleteLockCommand
var commitUpdater *mocks.MockCommitStatusUpdater
var staleCommandChecker *mocks.MockStaleCommandChecker

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
var applyCommandRunner *events.ApplyCommandRunner
var unlockCommandRunner *events.UnlockCommandRunner
var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner

func setup(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	eventParsing = mocks.NewMockEventParsing()
	vcsClient := vcsmocks.NewMockClient()
	githubClient := vcsmocks.NewMockIGithubClient()
	githubGetter = mocks.NewMockGithubPullGetter()
	gitlabGetter = mocks.NewMockGitlabMergeRequestGetter()
	azuredevopsGetter = mocks.NewMockAzureDevopsPullGetter()
	logger = logging.NewNoopLogger(t)
	projectCommandRunner = mocks.NewMockProjectCommandRunner()
	workingDir = mocks.NewMockWorkingDir()
	pendingPlanFinder = mocks.NewMockPendingPlanFinder()
	commitUpdater = mocks.NewMockCommitStatusUpdater()
	tmp, cleanup := TempDir(t)
	defer cleanup()
	defaultBoltDB, err := db.New(tmp)
	Ok(t, err)

	drainer = &events.Drainer{}
	deleteLockCommand = eventmocks.NewMockDeleteLockCommand()
	applyLockChecker = lockingmocks.NewMockApplyLockChecker()

	dbUpdater = &events.DBUpdater{
		DB: defaultBoltDB,
	}

	pullUpdater = &events.PullUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     &events.MarkdownRenderer{},
	}

	autoMerger = &events.AutoMerger{
		VCSClient:       vcsClient,
		GlobalAutomerge: false,
	}

	parallelPoolSize := 1
	SilenceNoProjects := false
	policyCheckCommandRunner = events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		commitUpdater,
		projectCommandRunner,
		parallelPoolSize,
		false,
	)

	planCommandRunner = events.NewPlanCommandRunner(
		false,
		false,
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
		parallelPoolSize,
		SilenceNoProjects,
		defaultBoltDB,
	)

	pullReqStatusFetcher := lyft_vcs.NewSQBasedPullStatusFetcher(githubClient, vcs.NewLyftPullMergeabilityChecker("atlantis"))

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
		parallelPoolSize,
		SilenceNoProjects,
		false,
		pullReqStatusFetcher,
	)

	approvePoliciesCommandRunner = events.NewApprovePoliciesCommandRunner(
		commitUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		pullUpdater,
		dbUpdater,
		SilenceNoProjects,
		false,
	)

	unlockCommandRunner = events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
		SilenceNoProjects,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		parallelPoolSize,
		SilenceNoProjects,
	)

	commentCommandRunnerByCmd := map[command.Name]events.CommentCommandRunner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
	}

	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyPtrToEventsCommandContext())).ThenReturn(nil)

	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{})
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	staleCommandChecker = mocks.NewMockStaleCommandChecker()

	ch = events.DefaultCommandRunner{
		VCSClient:                     vcsClient,
		CommentCommandRunnerByCmd:     commentCommandRunnerByCmd,
		EventParser:                   eventParsing,
		GithubPullGetter:              githubGetter,
		GitlabMergeRequestGetter:      gitlabGetter,
		AzureDevopsPullGetter:         azuredevopsGetter,
		Logger:                        logger,
		GlobalCfg:                     globalCfg,
		StatsScope:                    scope,
		AllowForkPRs:                  false,
		AllowForkPRsFlag:              "allow-fork-prs-flag",
		Drainer:                       drainer,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		PullStatusFetcher:             defaultBoltDB,
		StaleCommandChecker:           staleCommandChecker,
	}
	return vcsClient
}

func TestRunCommentCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	vcsClient := setup(t)
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, 1, &events.CommentCommand{Name: command.Plan}, time.Now())
	_, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), fmt.Sprintf("comment should be about a goroutine panic but was %q", comment))
}

func TestRunCommentCommand_GithubPullErr(t *testing.T) {
	t.Log("if getting the github pull request fails an error should be logged")
	vcsClient := setup(t)
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "`Error: making pull request API call to GitHub: err`", "")
}

func TestRunCommentCommand_GitlabMergeRequestErr(t *testing.T) {
	t.Log("if getting the gitlab merge request fails an error should be logged")
	vcsClient := setup(t)
	When(gitlabGetter.GetMergeRequest(fixtures.GitlabRepo.FullName, fixtures.Pull.Num)).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(fixtures.GitlabRepo, &fixtures.GitlabRepo, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GitlabRepo, fixtures.Pull.Num, "`Error: making merge request API call to GitLab: err`", "")
}

func TestRunCommentCommand_GithubPullParseErr(t *testing.T) {
	t.Log("if parsing the returned github pull request fails an error should be logged")
	vcsClient := setup(t)
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(fixtures.Pull, fixtures.GithubRepo, fixtures.GitlabRepo, errors.New("err"))

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "`Error: extracting required fields from comment data: err`", "")
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
		BaseRepo: fixtures.GithubRepo,
		State:    models.OpenPullState,
	}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)

	headRepo := fixtures.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, headRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	commentMessage := fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s  or, to disable this message, set --%s", ch.AllowForkPRsFlag, ch.SilenceForkPRErrorsFlag)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, commentMessage, "")
}

func TestRunCommentCommand_ForkPRDisabled_SilenceEnabled(t *testing.T) {
	t.Log("if a command is run on a forked pull request and forks are disabled and we are silencing errors do not comment with error")
	vcsClient := setup(t)
	ch.AllowForkPRs = false // by default it's false so don't need to reset
	ch.SilenceForkPRErrors = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)

	headRepo := fixtures.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, headRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunCommentCommandPlan_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if a plan command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	planCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Plan}, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsCommitStatus(models.SuccessCommitStatus),
		matchers.EqModelsCommandName(command.Plan),
		EqInt(0),
		EqInt(0),
	)
}

func TestRunCommentCommandApply_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an apply command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	applyCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Apply}, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsCommitStatus(models.SuccessCommitStatus),
		matchers.EqModelsCommandName(command.Apply),
		EqInt(0),
		EqInt(0),
	)
}

func TestRunCommentCommandApprovePolicy_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an approve_policy command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	approvePoliciesCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.ApprovePolicies}, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsCommitStatus(models.SuccessCommitStatus),
		matchers.EqModelsCommandName(command.PolicyCheck),
		EqInt(0),
		EqInt(0),
	)
}

func TestRunCommentCommandUnlock_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an unlock command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	unlockCommandRunner.SilenceNoProjects = true
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Unlock}, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunCommentCommand_DisableApplyAllDisabled(t *testing.T) {
	t.Log("if \"atlantis apply\" is run and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	applyCommandRunner.DisableApplyAll = true
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, modelPull.Num, &events.CommentCommand{Name: command.Apply}, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "**Error:** Running `atlantis apply` without flags is disabled. You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags.", "apply")
}

func TestFeatureAwareRunCommentCommandRunner_CommentWhenEnabled(t *testing.T) {
	t.Log("if \"atlantis apply --force\" is run and this is enabled atlantis should" +
		" comment with a warning")
	vcsClient := setup(t)

	fa = events.ForceApplyCommandRunner{
		CommandRunner: &ch,
		VCSClient:     vcsClient,
		Logger:        logger,
	}

	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}

	fa.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, modelPull.Num, &events.CommentCommand{Name: command.Apply, ForceApply: true}, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "⚠️ WARNING ⚠️\n\n You have bypassed all apply requirements for this PR 🚀 . This can have unpredictable consequences 🙏🏽 and should only be used in an emergency 🆘 .\n\n 𝐓𝐡𝐢𝐬 𝐚𝐜𝐭𝐢𝐨𝐧 𝐰𝐢𝐥𝐥 𝐛𝐞 𝐚𝐮𝐝𝐢𝐭𝐞𝐝.\n", "")
}

func TestRunCommentCommand_DisableDisableAutoplan(t *testing.T) {
	t.Log("if \"DisableAutoplan is true\" are disabled and we are silencing return and do not comment with error")
	setup(t)
	ch.DisableAutoplan = true
	defer func() { ch.DisableAutoplan = false }()

	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)

	ch.RunAutoplanCommand(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())
}

func TestRunCommentCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("closed"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.ClosedPullState, Num: fixtures.Pull.Num}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Atlantis commands can't be run on closed pull requests", "")
}

func TestRunCommentCommand_MatchedBranch(t *testing.T) {
	t.Log("if a command is run on a pull request which matches base branches run plan successfully")
	vcsClient := setup(t)

	ch.GlobalCfg.Repos = append(ch.GlobalCfg.Repos, valid.Repo{
		IDRegex:     regexp.MustCompile(".*"),
		BranchRegex: regexp.MustCompile("^main$"),
	})
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, BaseBranch: "main"}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Plan}, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Ran Plan for 0 projects:\n\n\n\n", "plan")
}

func TestRunCommentCommand_UnmatchedBranch(t *testing.T) {
	t.Log("if a command is run on a pull request which doesn't match base branches do not comment with error")
	vcsClient := setup(t)

	ch.GlobalCfg.Repos = append(ch.GlobalCfg.Repos, valid.Repo{
		IDRegex:     regexp.MustCompile(".*"),
		BranchRegex: regexp.MustCompile("^main$"),
	})
	var pull github.PullRequest
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, BaseBranch: "foo"}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Plan}, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunUnlockCommand_VCSComment(t *testing.T) {
	t.Log("if unlock PR command is run, atlantis should" +
		" invoke the delete command and comment on PR accordingly")

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Unlock}, time.Now())

	deleteLockCommand.VerifyWasCalledOnce().DeleteLocksByPull(fixtures.GithubRepo.FullName, fixtures.Pull.Num)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "All Atlantis locks for this PR have been unlocked and plans discarded", "unlock")
}

func TestRunUnlockCommandFail_VCSComment(t *testing.T) {
	t.Log("if unlock PR command is run and delete fails, atlantis should" +
		" invoke comment on PR with error message")

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(deleteLockCommand.DeleteLocksByPull(fixtures.GithubRepo.FullName, fixtures.Pull.Num)).ThenReturn(0, errors.New("err"))
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Unlock}, time.Now())

	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Failed to delete PR locks", "unlock")
}

// Test that if one plan fails and we are using automerge, that
// we delete the plans.
func TestRunAutoplanCommand_DeletePlans(t *testing.T) {
	setup(t)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
			},
			{
				CommandName: command.Plan,
			},
		}, nil)
	callCount := 0
	When(projectCommandRunner.Plan(matchers.AnyModelsProjectCommandContext())).Then(func(_ []Param) ReturnValues {
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

	When(workingDir.GetPullDir(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).
		ThenReturn(tmp, nil)
	fixtures.Pull.BaseRepo = fixtures.GithubRepo
	ch.RunAutoplanCommand(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now())
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
}

func TestFailedApprovalCreatesFailedStatusUpdate(t *testing.T) {
	t.Log("if \"atlantis approve_policies\" is run by non policy owner policy check status fails.")
	setup(t)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.String("open"),
	}

	modelPull := models.PullRequest{
		BaseRepo: fixtures.GithubRepo,
		State:    models.OpenPullState,
		Num:      fixtures.Pull.Num,
	}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	When(projectCommandBuilder.BuildApprovePoliciesCommands(matchers.AnyPtrToEventsCommandContext(), matchers.AnyPtrToEventsCommentCommand())).ThenReturn([]command.ProjectContext{
		{
			CommandName: command.ApprovePolicies,
		},
		{
			CommandName: command.ApprovePolicies,
		},
	}, nil)

	When(workingDir.GetPullDir(fixtures.GithubRepo, fixtures.Pull)).ThenReturn(tmp, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, &fixtures.Pull, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.ApprovePolicies}, time.Now())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsCommitStatus(models.SuccessCommitStatus),
		matchers.EqModelsCommandName(command.PolicyCheck),
		EqInt(0),
		EqInt(0),
	)
}

func TestApprovedPoliciesUpdateFailedPolicyStatus(t *testing.T) {
	t.Log("if \"atlantis approve_policies\" is run by policy owner all policy checks are approved.")
	setup(t)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.String("open"),
	}

	modelPull := models.PullRequest{
		BaseRepo: fixtures.GithubRepo,
		State:    models.OpenPullState,
		Num:      fixtures.Pull.Num,
	}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	When(projectCommandBuilder.BuildApprovePoliciesCommands(matchers.AnyPtrToEventsCommandContext(), matchers.AnyPtrToEventsCommentCommand())).ThenReturn([]command.ProjectContext{
		{
			CommandName: command.ApprovePolicies,
			PolicySets: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Users: []string{fixtures.User.Username},
				},
			},
		},
	}, nil)

	When(workingDir.GetPullDir(fixtures.GithubRepo, fixtures.Pull)).ThenReturn(tmp, nil)
	When(projectCommandRunner.ApprovePolicies(matchers.AnyModelsProjectCommandContext())).Then(func(_ []Param) ReturnValues {
		return ReturnValues{
			command.ProjectResult{
				Command:            command.PolicyCheck,
				PolicyCheckSuccess: &models.PolicyCheckSuccess{},
			},
		}
	})

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, &fixtures.Pull, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.ApprovePolicies}, time.Now())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsCommitStatus(models.SuccessCommitStatus),
		matchers.EqModelsCommandName(command.PolicyCheck),
		EqInt(1),
		EqInt(1),
	)
}

func TestApplyMergeablityWhenPolicyCheckFails(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with failing policy check then apply is not performed")
	setup(t)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pull := &github.PullRequest{
		State: github.String("open"),
	}

	modelPull := models.PullRequest{
		BaseRepo: fixtures.GithubRepo,
		State:    models.OpenPullState,
		Num:      fixtures.Pull.Num,
	}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

	_, _ = boltDB.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:     command.PolicyCheck,
			Error:       fmt.Errorf("failing policy"),
			ProjectName: "default",
			Workspace:   "default",
			RepoRelDir:  ".",
		},
	})

	When(ch.VCSClient.PullIsMergeable(fixtures.GithubRepo, modelPull)).ThenReturn(true, nil)

	When(projectCommandBuilder.BuildApplyCommands(matchers.AnyPtrToEventsCommandContext(), matchers.AnyPtrToEventsCommentCommand())).Then(func(args []Param) ReturnValues {
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

	When(workingDir.GetPullDir(fixtures.GithubRepo, modelPull)).ThenReturn(tmp, nil)
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, &modelPull, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Apply}, time.Now())
}

func TestApplyWithAutoMerge_VSCMerge(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge then a VCS merge is performed")

	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()

	pullOptions := models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
	}

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Apply}, time.Now())
	vcsClient.VerifyWasCalledOnce().MergePull(modelPull, pullOptions)
}

func TestRunApply_DiscardedProjects(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with automerge and at least one project" +
		" has a discarded plan, automerge should not take place")
	vcsClient := setup(t)
	autoMerger.GlobalAutomerge = true
	defer func() { autoMerger.GlobalAutomerge = false }()
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB
	pull := fixtures.Pull
	pull.BaseRepo = fixtures.GithubRepo
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
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(ghPull, nil)
	When(eventParsing.ParseGithubPull(ghPull)).ThenReturn(pull, pull.BaseRepo, fixtures.GithubRepo, nil)
	When(workingDir.GetPullDir(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).
		ThenReturn(tmp, nil)
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, &pull, fixtures.User, fixtures.Pull.Num, &events.CommentCommand{Name: command.Apply}, time.Now())

	vcsClient.VerifyWasCalled(Never()).MergePull(matchers.AnyModelsPullRequest(), matchers.AnyModelsPullRequestOptions())
}

func TestRunCommentCommand_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := setup(t)
	drainer.ShutdownBlocking()
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Atlantis server is shutting down, please try again later.", "")
}

func TestRunCommentCommand_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occurred")
	setup(t)
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil, time.Now())
	githubGetter.VerifyWasCalledOnce().GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}

func TestRunAutoplanCommand_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := setup(t)
	drainer.ShutdownBlocking()
	ch.RunAutoplanCommand(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now())
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Atlantis server is shutting down, please try again later.", "plan")
}

func TestRunAutoplanCommand_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occurred")
	setup(t)
	fixtures.Pull.BaseRepo = fixtures.GithubRepo
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunAutoplanCommand(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now())
	projectCommandBuilder.VerifyWasCalledOnce().BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}

func TestRunCommentCommand_DropStaleRequest(t *testing.T) {
	t.Log("if comment command is stale then entire request should be dropped")
	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(true)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, modelPull.Num, &events.CommentCommand{Name: command.Apply}, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunAutoplanCommand_DropStaleRequest(t *testing.T) {
	t.Log("if autoplan command is stale then entire request should be dropped")
	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("open"),
	}
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(true)

	ch.RunAutoplanCommand(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now())
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}
