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
	"context"
	"errors"
	"fmt"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/v4"
	"regexp"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/command/policies"
	"github.com/runatlantis/atlantis/server/events/vcs"
	lyft_vcs "github.com/runatlantis/atlantis/server/events/vcs/lyft"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"

	. "github.com/petergtz/pegomock"
	lockingmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	eventmocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/vcs/markdown"
	. "github.com/runatlantis/atlantis/testing"
)

var projectCommandBuilder *mocks.MockProjectCommandBuilder
var projectCommandRunner *mocks.MockProjectCommandRunner
var ch events.DefaultCommandRunner
var fa events.ForceApplyCommandRunner
var workingDir events.WorkingDir
var pendingPlanFinder *mocks.MockPendingPlanFinder
var drainer *events.Drainer
var deleteLockCommand *mocks.MockDeleteLockCommand
var vcsUpdater *mocks.MockVCSStatusUpdater
var staleCommandChecker *mocks.MockStaleCommandChecker

// TODO: refactor these into their own unit tests.
// these were all split out from default command runner in an effort to improve
// readability however the tests were kept as is.
var dbUpdater *events.DBUpdater
var pullUpdater events.OutputUpdater
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
	vcsClient := vcsmocks.NewMockClient()
	githubClient := vcsmocks.NewMockIGithubClient()
	logger = logging.NewNoopCtxLogger(t)
	projectCommandRunner = mocks.NewMockProjectCommandRunner()
	workingDir = mocks.NewMockWorkingDir()
	pendingPlanFinder = mocks.NewMockPendingPlanFinder()
	vcsUpdater = mocks.NewMockVCSStatusUpdater()
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

	pullUpdater = &events.PullOutputUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     &markdown.Renderer{},
	}

	parallelPoolSize := 1
	policyCheckCommandRunner = events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		vcsUpdater,
		projectCommandRunner,
		parallelPoolSize,
	)

	planCommandRunner = events.NewPlanCommandRunner(
		vcsClient,
		pendingPlanFinder,
		workingDir,
		vcsUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		parallelPoolSize,
	)

	pullReqStatusFetcher := lyft_vcs.NewSQBasedPullStatusFetcher(githubClient, vcs.NewLyftPullMergeabilityChecker("atlantis"))

	applyCommandRunner = events.NewApplyCommandRunner(
		vcsClient,
		false,
		applyLockChecker,
		vcsUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		pullUpdater,
		dbUpdater,
		parallelPoolSize,
		pullReqStatusFetcher,
	)
	featureAllocator, featureAllocatorErr := feature.NewStringSourcedAllocator(logger)
	assert.NoError(t, featureAllocatorErr)

	approvePoliciesCommandRunner = events.NewApprovePoliciesCommandRunner(vcsUpdater, projectCommandBuilder, projectCommandRunner, pullUpdater, dbUpdater, &policies.CommandOutputGenerator{
		PrjCommandRunner:  projectCommandRunner,
		PrjCommandBuilder: projectCommandBuilder,
	}, featureAllocator)

	unlockCommandRunner = events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		parallelPoolSize,
	)

	commentCommandRunnerByCmd := map[command.Name]command.Runner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
	}

	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(nil)

	globalCfg := valid.NewGlobalCfg("somedir")
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	staleCommandChecker = mocks.NewMockStaleCommandChecker()

	ch = events.DefaultCommandRunner{
		VCSClient:                     vcsClient,
		CommentCommandRunnerByCmd:     commentCommandRunnerByCmd,
		Logger:                        logging.NewNoopCtxLogger(t),
		GlobalCfg:                     globalCfg,
		StatsScope:                    scope,
		Drainer:                       drainer,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		PullStatusFetcher:             defaultBoltDB,
		StaleCommandChecker:           staleCommandChecker,
		VCSStatusUpdater:              vcsUpdater,
	}
	return vcsClient
}

func TestRunCommentCommand_ForkPRDisabled(t *testing.T) {
	t.Log("if a command is run on a forked pull request and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	ctx := context.Background()
	modelPull := models.PullRequest{
		BaseRepo: fixtures.GithubRepo,
		State:    models.OpenPullState,
	}

	headRepo := fixtures.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	ch.RunCommentCommand(ctx, fixtures.GithubRepo, headRepo, modelPull, fixtures.User, fixtures.Pull.Num, nil, time.Now(), 0)
	commentMessage := "Atlantis commands can't be run on fork pull requests."
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, commentMessage, "")
}

func TestRunCommentCommand_PreWorkflowHookError(t *testing.T) {
	t.Log("if pre workflow hook errors out stop the execution")
	for _, cmd := range []command.Name{command.Plan, command.Apply} {
		t.Run(cmd.String(), func(t *testing.T) {
			RegisterMockTestingT(t)
			_ = setup(t)
			ctx := context.Background()
			modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, Num: fixtures.Pull.Num, State: models.OpenPullState}
			preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

			When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
			When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(fmt.Errorf("catastrophic error"))

			ch.PreWorkflowHooksCommandRunner = preWorkflowHooksCommandRunner

			ch.RunCommentCommand(ctx, fixtures.GithubRepo, modelPull.BaseRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: cmd}, time.Now(), 0)
			_, _, _, status, cmdName, _, _ := vcsUpdater.VerifyWasCalledOnce().UpdateCombined(matchers.AnyContextContext(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsVcsStatus(), matchers.AnyCommandName(), AnyString(), AnyString()).GetCapturedArguments()
			Equals(t, models.FailedVCSStatus, status)
			Equals(t, cmd, cmdName)
		})
	}

	t.Run("if pre workflow hook errors out for approve policies, set the policy check status to failed", func(t *testing.T) {
		RegisterMockTestingT(t)
		_ = setup(t)
		ctx := context.Background()
		modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, Num: fixtures.Pull.Num, State: models.OpenPullState}
		preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

		When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
		When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(fmt.Errorf("catastrophic error"))

		ch.PreWorkflowHooksCommandRunner = preWorkflowHooksCommandRunner

		ch.RunCommentCommand(ctx, fixtures.GithubRepo, modelPull.BaseRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.ApprovePolicies}, time.Now(), 0)
		_, _, _, status, cmdName, _, _ := vcsUpdater.VerifyWasCalledOnce().UpdateCombined(matchers.AnyContextContext(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsVcsStatus(), matchers.AnyCommandName(), AnyString(), AnyString()).GetCapturedArguments()
		Equals(t, models.FailedVCSStatus, status)
		Equals(t, command.PolicyCheck, cmdName)
	})
}

func TestRunCommentCommandApprovePolicy_NoProjects_SilenceEnabled(t *testing.T) {
	t.Log("if an approve_policy command is run on a pull request and SilenceNoProjects is enabled and we are silencing all comments if the modified files don't have a matching project")
	vcsClient := setup(t)
	ctx := context.Background()
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState}
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.ApprovePolicies}, time.Now(), 0)
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
	vcsUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsVcsStatus(models.SuccessVCSStatus),
		matchers.EqModelsCommandName(command.PolicyCheck),
		EqInt(0),
		EqInt(0),
		AnyString(),
	)
}

func TestRunCommentCommand_DisableApplyAllDisabled(t *testing.T) {
	t.Log("if \"atlantis apply\" is run and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	ctx := context.Background()
	applyCommandRunner.DisableApplyAll = true
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, modelPull.Num, &command.Comment{Name: command.Apply}, time.Now(), 0)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "**Error:** Running `atlantis apply` without flags is disabled. You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags.", "apply")
}

func TestForceApplyRunCommentCommandRunner_CommentWhenEnabled(t *testing.T) {
	t.Log("if \"atlantis apply --force\" is run and this is enabled atlantis should" +
		" comment with a warning")
	vcsClient := setup(t)
	ctx := context.Background()

	fa = events.ForceApplyCommandRunner{
		CommandRunner: &ch,
		VCSClient:     vcsClient,
	}

	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}

	fa.RunCommentCommand(ctx, fixtures.GithubRepo, models.Repo{}, models.PullRequest{}, fixtures.User, modelPull.Num, &command.Comment{Name: command.Apply, ForceApply: true}, time.Now(), 0)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "⚠️ WARNING ⚠️\n\n You have bypassed all apply requirements for this PR 🚀 . This can have unpredictable consequences 🙏🏽 and should only be used in an emergency 🆘 .\n\n 𝐓𝐡𝐢𝐬 𝐚𝐜𝐭𝐢𝐨𝐧 𝐰𝐢𝐥𝐥 𝐛𝐞 𝐚𝐮𝐝𝐢𝐭𝐞𝐝.\n", "")
}

func TestRunCommentCommand_DisableDisableAutoplan(t *testing.T) {
	t.Log("if \"DisableAutoplan is true\" are disabled and we are silencing return and do not comment with error")
	setup(t)
	ctx := context.Background()
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

	ch.RunAutoplanCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	projectCommandBuilder.VerifyWasCalled(Never()).BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())
}

func TestRunCommentCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	ctx := context.Background()
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.ClosedPullState, Num: fixtures.Pull.Num}

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, nil, time.Now(), 0)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Atlantis commands can't be run on closed pull requests", "")
}

func TestRunCommentCommand_MatchedBranch(t *testing.T) {
	t.Log("if a command is run on a pull request which matches base branches run plan successfully")
	vcsClient := setup(t)
	ctx := context.Background()

	ch.GlobalCfg.Repos = append(ch.GlobalCfg.Repos, valid.Repo{
		IDRegex:     regexp.MustCompile(".*"),
		BranchRegex: regexp.MustCompile("^main$"),
	})
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, BaseBranch: "main"}
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.Plan}, time.Now(), 0)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Ran Plan for 0 projects:\n\n\n\n", "plan")
}

func TestRunCommentCommand_UnmatchedBranch(t *testing.T) {
	t.Log("if a command is run on a pull request which doesn't match base branches do not comment with error")
	vcsClient := setup(t)
	ctx := context.Background()

	ch.GlobalCfg.Repos = append(ch.GlobalCfg.Repos, valid.Repo{
		IDRegex:     regexp.MustCompile(".*"),
		BranchRegex: regexp.MustCompile("^main$"),
	})
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, BaseBranch: "foo"}

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.Plan}, time.Now(), 0)
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunUnlockCommand_VCSComment(t *testing.T) {
	t.Log("if unlock PR command is run, atlantis should" +
		" invoke the delete command and comment on PR accordingly")

	vcsClient := setup(t)
	ctx := context.Background()
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.Unlock}, time.Now(), 0)

	deleteLockCommand.VerifyWasCalledOnce().DeleteLocksByPull(fixtures.GithubRepo.FullName, fixtures.Pull.Num)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "All Atlantis locks for this PR have been unlocked and plans discarded", "unlock")
}

func TestRunUnlockCommandFail_VCSComment(t *testing.T) {
	t.Log("if unlock PR command is run and delete fails, atlantis should" +
		" invoke comment on PR with error message")

	vcsClient := setup(t)
	ctx := context.Background()
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(deleteLockCommand.DeleteLocksByPull(fixtures.GithubRepo.FullName, fixtures.Pull.Num)).ThenReturn(0, errors.New("err"))
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.Unlock}, time.Now(), 0)

	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Failed to delete PR locks", "unlock")
}

func TestRunAutoplanCommand_PreWorkflowHookError(t *testing.T) {
	t.Log("if pre workflow hook errors out stop the execution")
	setup(t)
	ctx := context.Background()
	preWorkflowHooksCommandRunner = mocks.NewMockPreWorkflowHooksCommandRunner()

	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
	When(preWorkflowHooksCommandRunner.RunPreHooks(matchers.AnyContextContext(), matchers.AnyPtrToEventsCommandContext())).ThenReturn(fmt.Errorf("catastrophic error"))
	When(projectCommandRunner.Plan(matchers.AnyCommandProjectContext())).ThenReturn(command.ProjectResult{PlanSuccess: &models.PlanSuccess{}})

	ch.PreWorkflowHooksCommandRunner = preWorkflowHooksCommandRunner

	ch.RunAutoplanCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	_, _, _, status, cmdName, _, _ := vcsUpdater.VerifyWasCalledOnce().UpdateCombined(matchers.AnyContextContext(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsVcsStatus(), matchers.AnyCommandName(), AnyString(), AnyString()).GetCapturedArguments()
	Equals(t, models.FailedVCSStatus, status)
	Equals(t, command.Plan, cmdName)
}

func TestFailedApprovalCreatesFailedStatusUpdate(t *testing.T) {
	t.Log("if \"atlantis approve_policies\" is run by non policy owner policy check status fails.")
	setup(t)
	ctx := context.Background()
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB

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

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.ApprovePolicies}, time.Now(), 0)
	vcsUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsVcsStatus(models.SuccessVCSStatus),
		matchers.EqModelsCommandName(command.PolicyCheck),
		EqInt(0),
		EqInt(0),
		AnyString(),
	)
}

func TestApprovedPoliciesUpdateFailedPolicyStatus(t *testing.T) {
	t.Log("if \"atlantis approve_policies\" is run by policy owner all policy checks are approved.")
	setup(t)
	ctx := context.Background()
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB

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
				Command: command.PolicyCheck,
				PolicyCheckSuccess: &models.PolicyCheckSuccess{
					PolicyCheckOutput: "Hello World",
				},
			},
		}
	})

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, fixtures.Pull.Num, &command.Comment{
		Name: command.ApprovePolicies,
	}, time.Now(), 0)
	vcsUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		matchers.AnyContextContext(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		matchers.EqModelsVcsStatus(models.SuccessVCSStatus),
		matchers.EqModelsCommandName(command.PolicyCheck),
		EqInt(1),
		EqInt(1),
		AnyString(),
	)
}

func TestApplyMergeablityWhenPolicyCheckFails(t *testing.T) {
	t.Log("if \"atlantis apply\" is run with failing policy check then apply is not performed")
	setup(t)
	ctx := context.Background()
	tmp, cleanup := TempDir(t)
	defer cleanup()
	boltDB, err := db.New(tmp)
	Ok(t, err)
	dbUpdater.DB = boltDB
	modelPull := models.PullRequest{
		BaseRepo: fixtures.GithubRepo,
		State:    models.OpenPullState,
		Num:      fixtures.Pull.Num,
	}

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
	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, modelPull, fixtures.User, fixtures.Pull.Num, &command.Comment{Name: command.Apply}, time.Now(), 0)
}

func TestRunCommentCommand_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := setup(t)
	ctx := context.Background()
	drainer.ShutdownBlocking()
	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, models.PullRequest{}, fixtures.User, fixtures.Pull.Num, nil, time.Now(), 0)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Atlantis server is shutting down, please try again later.", "")
}

func TestRunCommentCommand_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occurred")
	setup(t)
	ctx := context.Background()
	ch.RunCommentCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, models.PullRequest{}, fixtures.User, fixtures.Pull.Num, nil, time.Now(), 0)
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}

func TestRunAutoplanCommand_DrainOngoing(t *testing.T) {
	t.Log("if drain is ongoing then a message should be displayed")
	vcsClient := setup(t)
	ctx := context.Background()
	drainer.ShutdownBlocking()
	ch.RunAutoplanCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "Atlantis server is shutting down, please try again later.", "plan")
}

func TestRunAutoplanCommand_DrainNotOngoing(t *testing.T) {
	t.Log("if drain is not ongoing then remove ongoing operation must be called even if panic occurred")
	setup(t)
	ctx := context.Background()
	fixtures.Pull.BaseRepo = fixtures.GithubRepo
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(false)
	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).ThenPanic("panic test - if you're seeing this in a test failure this isn't the failing test")
	ch.RunAutoplanCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	projectCommandBuilder.VerifyWasCalledOnce().BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())
	Equals(t, 0, drainer.GetStatus().InProgressOps)
}

func TestRunCommentCommand_DropStaleRequest(t *testing.T) {
	t.Log("if comment command is stale then entire request should be dropped")
	vcsClient := setup(t)
	ctx := context.Background()
	modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(true)

	ch.RunCommentCommand(ctx, fixtures.GithubRepo, models.Repo{}, models.PullRequest{}, fixtures.User, modelPull.Num, &command.Comment{Name: command.Apply}, time.Now(), 0)
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunAutoplanCommand_DropStaleRequest(t *testing.T) {
	t.Log("if autoplan command is stale then entire request should be dropped")
	vcsClient := setup(t)
	ctx := context.Background()
	When(staleCommandChecker.CommandIsStale(matchers.AnyPtrToModelsCommandContext())).ThenReturn(true)

	ch.RunAutoplanCommand(ctx, fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	vcsClient.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString(), AnyString())
}

func TestRunPRRCommand_RunPRReviewCommand(t *testing.T) {
	staleChecker := &testStaleCommandChecker{}
	policyCommandRunner := &testPolicyCommandRunner{}
	preWorkflowHooksRunner := &testPreWorkflowHooksRunnerRunner{}
	ch = events.DefaultCommandRunner{
		Drainer:                       &events.Drainer{},
		Logger:                        logging.NewNoopCtxLogger(t),
		GlobalCfg:                     valid.NewGlobalCfg("somedir"),
		StatsScope:                    tally.NewTestScope("atlantis", map[string]string{}),
		PullStatusFetcher:             &testPullStatusFetcher{},
		StaleCommandChecker:           staleChecker,
		PreWorkflowHooksCommandRunner: preWorkflowHooksRunner,
		PolicyCommandRunner:           policyCommandRunner,
	}
	ctx := context.Background()
	ch.RunPRReviewCommand(ctx, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	assert.True(t, policyCommandRunner.wasCalled)
}

func TestRunPRRCommand_RunPRReviewCommand_StaleCommand(t *testing.T) {
	staleChecker := &testStaleCommandChecker{
		stale: true,
	}
	policyCommandRunner := &testPolicyCommandRunner{}
	ch = events.DefaultCommandRunner{
		Drainer:             &events.Drainer{},
		Logger:              logging.NewNoopCtxLogger(t),
		GlobalCfg:           valid.NewGlobalCfg("somedir"),
		StatsScope:          tally.NewTestScope("atlantis", map[string]string{}),
		PullStatusFetcher:   &testPullStatusFetcher{},
		StaleCommandChecker: staleChecker,
		PolicyCommandRunner: policyCommandRunner,
	}
	ctx := context.Background()
	ch.RunPRReviewCommand(ctx, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	assert.False(t, policyCommandRunner.wasCalled)
}

func TestRunPRRCommand_RunPRReviewCommand_HooksError(t *testing.T) {
	staleChecker := &testStaleCommandChecker{}
	preWorkflowHooksRunner := &testPreWorkflowHooksRunnerRunner{
		error: assert.AnError,
	}
	policyCommandRunner := &testPolicyCommandRunner{}
	vcsStatusUpdater := &testVCSStatusUpdater{}
	ch = events.DefaultCommandRunner{
		Drainer:                       &events.Drainer{},
		Logger:                        logging.NewNoopCtxLogger(t),
		GlobalCfg:                     valid.NewGlobalCfg("somedir"),
		StatsScope:                    tally.NewTestScope("atlantis", map[string]string{}),
		PullStatusFetcher:             &testPullStatusFetcher{},
		StaleCommandChecker:           staleChecker,
		PreWorkflowHooksCommandRunner: preWorkflowHooksRunner,
		VCSStatusUpdater:              vcsStatusUpdater,
		PolicyCommandRunner:           policyCommandRunner,
	}
	ctx := context.Background()
	ch.RunPRReviewCommand(ctx, fixtures.GithubRepo, fixtures.Pull, fixtures.User, time.Now(), 0)
	assert.False(t, policyCommandRunner.wasCalled)
}

type testVCSStatusUpdater struct {
	output string
	error  error
}

func (u testVCSStatusUpdater) UpdateCombined(context.Context, models.Repo, models.PullRequest, models.VCSStatus, fmt.Stringer, string, string) (string, error) {
	return u.output, u.error
}

func (u testVCSStatusUpdater) UpdateCombinedCount(context.Context, models.Repo, models.PullRequest, models.VCSStatus, fmt.Stringer, int, int, string) (string, error) {
	return u.output, u.error
}

func (u testVCSStatusUpdater) UpdateProject(context.Context, command.ProjectContext, fmt.Stringer, models.VCSStatus, string, string) (string, error) {
	return u.output, u.error
}

type testPullStatusFetcher struct {
	error error
	pull  *models.PullStatus
}

func (p testPullStatusFetcher) GetPullStatus(_ models.PullRequest) (*models.PullStatus, error) {
	return p.pull, p.error
}

type testPreWorkflowHooksRunnerRunner struct {
	error error
}

func (h testPreWorkflowHooksRunnerRunner) RunPreHooks(context.Context, *command.Context) error {
	return h.error
}

type testStaleCommandChecker struct {
	stale bool
}

func (s testStaleCommandChecker) CommandIsStale(_ *command.Context) bool {
	return s.stale
}

type testPolicyCommandRunner struct {
	wasCalled bool
}

func (r *testPolicyCommandRunner) Run(_ *command.Context) {
	r.wasCalled = true
}
