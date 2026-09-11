// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-github/v88/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/require"
)

func TestPlanCommandRunner_IsSilenced(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description       string
		Matched           bool
		Targeted          bool
		VCSStatusSilence  bool
		PrevPlanStored    bool // stores a 1/1 passing plan in the database
		ExpVCSStatusSet   bool
		ExpVCSStatusTotal int
		ExpVCSStatusSucc  int
		ExpSilenced       bool
	}{
		{
			Description:     "When planning, don't comment but set the 0/0 VCS status",
			ExpVCSStatusSet: true,
			ExpSilenced:     true,
		},
		{
			Description:     "When planning with any previous plans, don't comment but set the 0/0 VCS status",
			PrevPlanStored:  true,
			ExpVCSStatusSet: true,
			ExpSilenced:     true,
		},
		{
			Description:     "When planning with unmatched target, don't comment but set the 0/0 VCS status",
			Targeted:        true,
			ExpVCSStatusSet: true,
			ExpSilenced:     true,
		},
		{
			Description:       "When planning with unmatched target and any previous plans, don't comment and maintain VCS status",
			Targeted:          true,
			PrevPlanStored:    true,
			ExpVCSStatusSet:   true,
			ExpSilenced:       true,
			ExpVCSStatusSucc:  1,
			ExpVCSStatusTotal: 1,
		},
		{
			Description:      "When planning with silenced VCS status, don't set any status",
			VCSStatusSilence: true,
			ExpVCSStatusSet:  false, // Silence means no status updates at all
			ExpSilenced:      true,
		},
		{
			Description:       "When planning with matching projects, comment as usual",
			Matched:           true,
			ExpVCSStatusSet:   true,
			ExpSilenced:       false,
			ExpVCSStatusSucc:  1,
			ExpVCSStatusTotal: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			// create an empty DB
			tmp := t.TempDir()
			db, err := boltdb.New(tmp)
			t.Cleanup(func() {
				db.Close()
			})
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.SilenceNoProjects = true
				tc.silenceVCSStatusNoProjects = c.VCSStatusSilence
				tc.database = db
			})

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			cmd := &events.CommentCommand{Name: command.Plan}
			if c.Targeted {
				cmd.RepoRelDir = "mydir"
			}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}
			if c.PrevPlanStored {
				_, err = db.UpdatePullWithResults(modelPull, []command.ProjectResult{
					{
						Command:    command.Plan,
						RepoRelDir: "prevdir",
						Workspace:  "default",
						ProjectCommandOutput: command.ProjectCommandOutput{
							PlanSuccess: &models.PlanSuccess{},
						},
					},
				})
				Ok(t, err)
			}

			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).Then(func(args []Param) ReturnValues {
				if c.Matched {
					return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
				}
				return ReturnValues{[]command.ProjectContext{}, nil}
			})
			When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

			planCommandRunner.Run(ctx, cmd)

			timesComment := 1
			if c.ExpSilenced {
				timesComment = 0
			}

			vcsClient.VerifyWasCalled(Times(timesComment)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
			if c.ExpVCSStatusSet {
				commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Eq[models.CommitStatus](models.SuccessCommitStatus),
					Eq[command.Name](command.Plan),
					Eq(models.ProjectCounts{Success: c.ExpVCSStatusSucc, Total: c.ExpVCSStatusTotal}),
				)
			} else {
				commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.CommitStatus](),
					Eq[command.Name](command.Plan),
					Any[models.ProjectCounts](),
				)
			}
		})
	}
}

func TestPlanCommandRunner_IgnoredTargetedDirNoOp(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := setup(t)
	planCommandRunner.DiscardApprovalOnPlan = true
	scopeNull := metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis")
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "ignored"}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    scopeNull,
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}

	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{}, events.ErrIgnoredTargetedDir)

	planCommandRunner.Run(ctx, cmd)
	Assert(t, ctx.CommandSkipped, "expected ignored targeted dir to mark the command skipped")

	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
	commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
	vcsClient.VerifyWasCalled(Never()).DiscardReviews(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
	projectCommandRunner.VerifyWasCalled(Never()).Plan(Any[command.ProjectContext]())
}

func TestPlanCommandRunner_ExecutionOrder(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description           string
		ProjectContexts       []command.ProjectContext
		ProjectCommandOutputs []command.ProjectCommandOutput
		RunnerInvokeMatch     []*EqMatcher
		PrevPlanStored        bool
		PlanFailed            bool
	}{
		{
			Description: "When first plan fails, the second don't run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					Workspace:                 "first",
					ProjectName:               "First",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					Workspace:                 "second",
					ProjectName:               "Second",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Error: errors.New("shabang"),
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			PlanFailed: true,
		},
		{
			Description: "When first fails, the second will not run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					ProjectName:               "Second",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},

				{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Never(),
			},
			PlanFailed: true,
		},
		{
			Description: "When first fails by autorun, the second will not run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:               command.Plan,
					AutoplanEnabled:           true,
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					AutoplanEnabled:           true,
					ExecutionOrderGroup:       1,
					ProjectName:               "Second",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},

				{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Never(),
			},
			PlanFailed: true,
		},
		{
			Description: "When both in a group of two succeeds, the following two will run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					ProjectName:               "Second",
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					ProjectName:               "Third",
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					ProjectName:               "Fourth",
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Error: errors.New("shabang"),
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
				Never(),
				Never(),
			},
			PlanFailed: true,
		},
		{
			Description: "When one out of two fails, the following two will not run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelPlanEnabled:       true,
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					ProjectName:               "Second",
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					ProjectName:               "Third",
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					AbortOnExecutionOrderFail: true,
					ProjectName:               "Fourth",
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},

				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},

				{
					Error: errors.New("shabang"),
				},

				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
				Once(),
				Once(),
			},
			PlanFailed: true,
		},
		{
			Description: "Don't block when parallel is not set",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					AbortOnExecutionOrderFail: true,
				},
				{
					CommandName:               command.Plan,
					ExecutionOrderGroup:       1,
					ProjectName:               "Second",
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			PlanFailed: true,
		},
		{
			Description: "All project finished successfully",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:         command.Plan,
					ExecutionOrderGroup: 0,
					ProjectName:         "First",
				},
				{
					CommandName:         command.Plan,
					ExecutionOrderGroup: 1,
					ProjectName:         "Second",
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			PlanFailed: false,
		},
		{
			Description: "Don't block when abortOnExecutionOrderFail is not set",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:         command.Plan,
					ExecutionOrderGroup: 0,
					ProjectName:         "First",
				},
				{
					CommandName:         command.Plan,
					ExecutionOrderGroup: 1,
					ProjectName:         "Second",
				},
			},
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			PlanFailed: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			// vcsClient := setup(t)

			tmp := t.TempDir()
			db, err := boltdb.New(tmp)
			t.Cleanup(func() {
				db.Close()
			})
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.database = db
			})

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")

			pull := &github.PullRequest{
				State: github.Ptr("open"),
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			cmd := &events.CommentCommand{Name: command.Plan}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(c.ProjectContexts, nil)
			// When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).Then(func(args []Param) ReturnValues {
			// 	return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
			// })
			for i := range c.ProjectContexts {
				When(projectCommandRunner.Plan(c.ProjectContexts[i])).ThenReturn(c.ProjectCommandOutputs[i])
			}

			planCommandRunner.Run(ctx, cmd)

			for i := range c.ProjectContexts {
				projectCommandRunner.VerifyWasCalled(c.RunnerInvokeMatch[i]).Plan(c.ProjectContexts[i])
			}

			require.Equal(t, c.PlanFailed, ctx.CommandHasErrors)

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Eq(modelPull.Num), Any[string](), Eq("plan"),
			)
		})
	}
}

func TestPlanCommandRunner_AtlantisApplyStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description            string
		ProjectContexts        []command.ProjectContext
		ProjectCommandOutput   []command.ProjectCommandOutput
		PrevPlanStored         bool // stores a previous "No changes" plan in the database
		DoNotUpdateApply       bool // certain circumstances we want to skip the call to update apply
		ExpVCSApplyStatusTotal int
		ExpVCSApplyStatusSucc  int
		ExpVCSApplyNoChanges   int
	}{
		{
			Description: "When planning with changes, do not change the apply status",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 0 to add, 0 to change, 1 to destroy.",
					},
				},
			},
			DoNotUpdateApply: true,
		},
		{
			Description: "When planning with no changes, set the 1/1 apply status",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
			},
			ExpVCSApplyStatusTotal: 1,
			ExpVCSApplyStatusSucc:  1,
			ExpVCSApplyNoChanges:   1,
		},
		{
			Description: "When planning with no changes and previous plan with no changes do not set the apply status",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 0 to add, 0 to change, 1 to destroy.",
					},
				},
			},
			DoNotUpdateApply: true,
			PrevPlanStored:   true,
		},
		{
			Description: "When planning with no changes and previous 'No changes' plan, set the 2/2 apply status",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
			},
			PrevPlanStored:         true,
			ExpVCSApplyStatusTotal: 2,
			ExpVCSApplyStatusSucc:  2,
			ExpVCSApplyNoChanges:   2,
		},
		{
			Description: "When planning again with changes following a previous 'No changes' plan do not set the apply status",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "prevdir",
					Workspace:   "default",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 0 to add, 0 to change, 1 to destroy.",
					},
				},
			},
			DoNotUpdateApply: true,
			PrevPlanStored:   true,
		},
		{
			Description: "When planning again with changes following a previous 'No changes' plan, while another plan with 'No changes' do not set the apply status.",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "prevdir",
					Workspace:   "default",
				},
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 0 to add, 0 to change, 1 to destroy.",
					},
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
			},
			DoNotUpdateApply: true,
			PrevPlanStored:   true,
		},
		{
			Description: "When planning again with no changes following a previous 'No changes' plan, while another plan also with 'No changes', set the 2/2 apply status.",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "prevdir",
					Workspace:   "default",
				},
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			},
			ProjectCommandOutput: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
			},
			PrevPlanStored:         true,
			ExpVCSApplyStatusTotal: 2,
			ExpVCSApplyStatusSucc:  2,
			ExpVCSApplyNoChanges:   2,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			// create an empty DB
			tmp := t.TempDir()
			db, err := boltdb.New(tmp)
			t.Cleanup(func() {
				db.Close()
			})
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.database = db
			})

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			cmd := &events.CommentCommand{Name: command.Plan}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			if c.PrevPlanStored {
				_, err = db.UpdatePullWithResults(modelPull, []command.ProjectResult{
					{
						Command:    command.Plan,
						RepoRelDir: "prevdir",
						Workspace:  "default",
						ProjectCommandOutput: command.ProjectCommandOutput{
							PlanSuccess: &models.PlanSuccess{
								TerraformOutput: "No changes. Your infrastructure matches the configuration.",
							},
						},
					},
				})
				Ok(t, err)
			}

			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(c.ProjectContexts, nil)

			for i := range c.ProjectContexts {
				When(projectCommandRunner.Plan(c.ProjectContexts[i])).ThenReturn(c.ProjectCommandOutput[i])
			}

			planCommandRunner.Run(ctx, cmd)

			vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), AnyInt(), AnyString(), AnyString())

			ExpCommitStatus := models.SuccessCommitStatus
			if c.ExpVCSApplyStatusSucc != c.ExpVCSApplyStatusTotal {
				ExpCommitStatus = models.PendingCommitStatus
			}
			if c.DoNotUpdateApply {
				commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.CommitStatus](),
					Eq[command.Name](command.Apply),
					Any[models.ProjectCounts](),
				)
			} else {
				commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Eq[models.CommitStatus](ExpCommitStatus),
					Eq[command.Name](command.Apply),
					Eq(models.ProjectCounts{Success: c.ExpVCSApplyStatusSucc, Total: c.ExpVCSApplyStatusTotal, NoChanges: c.ExpVCSApplyNoChanges}),
				)
			}
		})
	}
}

// TestPlanCommandRunner_SilenceFlagsClearsPendingStatus tests that when silence flags are enabled
// and no projects are found, the pending status that was set earlier is cleared.
// This is a regression test for issue #5389 where PRs were getting stuck with pending status.
func TestPlanCommandRunner_SilenceFlagsClearsPendingStatus(t *testing.T) {
	// Test the specific scenario from issue #5389:
	// When silence flags are enabled and no projects match when_modified patterns,
	// the pending status should be cleared instead of leaving the PR stuck.

	// This test ensures that even when ATLANTIS_SILENCE_VCS_STATUS_NO_PLANS and
	// ATLANTIS_SILENCE_VCS_STATUS_NO_PROJECTS are true, we still update the status
	// to clear any pending state that was set earlier (e.g., in command_runner.go)

	t.Run("silence flags with no projects should not set any status", func(t *testing.T) {
		RegisterMockTestingT(t)

		_ = setup(t, func(tc *TestConfig) {
			tc.SilenceNoProjects = true
			tc.silenceVCSStatusNoProjects = true // This is the key flag
			tc.silenceVCSStatusNoPlans = true    // This is the key flag
		})

		modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
		scopeNull := metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis")

		ctx := &command.Context{
			User:     testdata.User,
			Log:      logging.NewNoopLogger(t),
			Scope:    scopeNull,
			Pull:     modelPull,
			HeadRepo: testdata.GithubRepo,
			Trigger:  command.AutoTrigger,
		}

		// Mock no projects found (simulating when_modified patterns not matching)
		When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{}, nil)

		// This is the key test: when both conditions are true:
		// 1. Silence flags are enabled
		// 2. No projects are found
		// We should NOT set any VCS status at all

		// The plan runner is now configured with silence flags
		// When it finds no projects, it should not set any VCS status
		// because silence means no status checks at all

		// Run through the plan command (which will internally check for projects)
		cmd := &events.CommentCommand{Name: command.Plan}
		planCommandRunner.Run(ctx, cmd)

		// CRITICAL VERIFICATION: With silence flags enabled, no status should be set at all
		// This prevents any VCS status checks from being created (issue #5389)
		// The silence flags mean "don't create any status checks"
		commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
			Any[logging.SimpleLogging](),
			Any[models.Repo](),
			Any[models.PullRequest](),
			Any[models.CommitStatus](),
			Any[command.Name](),
			Any[models.ProjectCounts](),
		)
	})
}
func TestPlanCommandRunner_AutoplanFetchesPullStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	t.Run("autoplan fetches pull request status", func(t *testing.T) {
		tmp := t.TempDir()
		db, err := boltdb.New(tmp)
		t.Cleanup(func() {
			db.Close()
		})
		Ok(t, err)

		_ = setup(t, func(tc *TestConfig) {
			tc.database = db
		})

		scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")
		modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

		ctx := &command.Context{
			User:     testdata.User,
			Log:      logging.NewNoopLogger(t),
			Scope:    scopeNull,
			Pull:     modelPull,
			HeadRepo: testdata.GithubRepo,
			Trigger:  command.AutoTrigger,
		}

		expectedStatus := models.PullReqStatus{
			MergeableStatus: models.MergeableStatus{IsMergeable: true},
			ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
		}

		When(pullReqStatusFetcher.FetchPullStatus(Any[logging.SimpleLogging](), Eq(modelPull))).ThenReturn(expectedStatus, nil)
		When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{}, nil)

		cmd := &events.CommentCommand{Name: command.Plan}
		planCommandRunner.Run(ctx, cmd)

		// Verify FetchPullStatus was called
		pullReqStatusFetcher.VerifyWasCalledOnce().FetchPullStatus(Any[logging.SimpleLogging](), Eq(modelPull))

		// Verify the status was set on the context
		require.True(t, ctx.PullRequestStatus.MergeableStatus.IsMergeable, "PullRequestStatus.MergeableStatus.IsMergeable must be true")
		require.True(t, ctx.PullRequestStatus.ApprovalStatus.IsApproved, "PullRequestStatus.ApprovalStatus.IsApproved must be true")
	})

	t.Run("autoplan continues when FetchPullStatus returns error", func(t *testing.T) {
		tmp := t.TempDir()
		db, err := boltdb.New(tmp)
		t.Cleanup(func() {
			db.Close()
		})
		Ok(t, err)

		_ = setup(t, func(tc *TestConfig) {
			tc.database = db
		})

		scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")
		modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

		ctx := &command.Context{
			User:     testdata.User,
			Log:      logging.NewNoopLogger(t),
			Scope:    scopeNull,
			Pull:     modelPull,
			HeadRepo: testdata.GithubRepo,
			Trigger:  command.AutoTrigger,
		}

		When(pullReqStatusFetcher.FetchPullStatus(Any[logging.SimpleLogging](), Eq(modelPull))).ThenReturn(models.PullReqStatus{}, errors.New("api error"))
		When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{
			{CommandName: command.Plan},
		}, nil)
		When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}})

		cmd := &events.CommentCommand{Name: command.Plan}
		planCommandRunner.Run(ctx, cmd)

		// Verify FetchPullStatus was called despite returning an error
		pullReqStatusFetcher.VerifyWasCalledOnce().FetchPullStatus(Any[logging.SimpleLogging](), Eq(modelPull))

		// Verify autoplan continued (Plan was still executed)
		projectCommandRunner.VerifyWasCalledOnce().Plan(Any[command.ProjectContext]())

		// Verify status defaults to false when FetchPullStatus errors
		require.False(t, ctx.PullRequestStatus.MergeableStatus.IsMergeable, "IsMergeable must default to false on error")
	})
}

func TestPlanCommandRunner_PendingApplyStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description            string
		VCSType                models.VCSHostType
		PendingApplyFlag       bool
		ProjectResults         []command.ProjectCommandOutput
		ExpApplyStatus         models.CommitStatus
		ExpVCSApplyStatusTotal int
		ExpVCSApplyStatusSucc  int
		ExpVCSApplyNoChanges   int
		ExpShouldUpdateStatus  bool
	}{
		{
			Description:      "GitLab with flag enabled and unapplied plans should set pending status",
			VCSType:          models.Gitlab,
			PendingApplyFlag: true,
			ProjectResults: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
			},
			ExpApplyStatus:         models.PendingCommitStatus,
			ExpVCSApplyStatusTotal: 1,
			ExpVCSApplyStatusSucc:  0,
			ExpShouldUpdateStatus:  true,
		},
		{
			Description:      "GitLab with flag disabled and unapplied plans should NOT update apply status",
			VCSType:          models.Gitlab,
			PendingApplyFlag: false,
			ProjectResults: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
			},
			ExpShouldUpdateStatus: false,
		},
		{
			Description:      "GitHub with flag enabled should NOT update apply status (default behavior)",
			VCSType:          models.Github,
			PendingApplyFlag: true,
			ProjectResults: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
			},
			ExpShouldUpdateStatus: false,
		},
		{
			Description:      "GitLab with all plans applied should set success status",
			VCSType:          models.Gitlab,
			PendingApplyFlag: true,
			ProjectResults: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
			},
			ExpApplyStatus:         models.SuccessCommitStatus,
			ExpVCSApplyStatusTotal: 1,
			ExpVCSApplyStatusSucc:  1,
			ExpVCSApplyNoChanges:   1,
			ExpShouldUpdateStatus:  true,
		},
		{
			Description:      "Bitbucket with flag enabled should NOT update apply status",
			VCSType:          models.BitbucketCloud,
			PendingApplyFlag: true,
			ProjectResults: []command.ProjectCommandOutput{
				{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
			},
			ExpShouldUpdateStatus: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			tmp := t.TempDir()
			db, err := boltdb.New(tmp)
			t.Cleanup(func() {
				db.Close()
			})
			Ok(t, err)

			_ = setup(t, func(tc *TestConfig) {
				tc.database = db
				tc.PendingApplyStatus = c.PendingApplyFlag
			})

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")

			// Create repo with the appropriate VCS type
			repo := testdata.GithubRepo
			repo.VCSHost = models.VCSHost{
				Type: c.VCSType,
			}

			modelPull := models.PullRequest{
				BaseRepo: repo,
				State:    models.OpenPullState,
				Num:      testdata.Pull.Num,
			}

			cmd := &events.CommentCommand{Name: command.Plan}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: repo,
				Trigger:  command.CommentTrigger,
			}

			projectContexts := []command.ProjectContext{
				{
					CommandName: command.Plan,
					RepoRelDir:  "mydir",
				},
			}

			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(projectContexts, nil)
			When(projectCommandRunner.Plan(projectContexts[0])).ThenReturn(c.ProjectResults[0])

			planCommandRunner.Run(ctx, cmd)

			// Verify based on whether we expect a status update
			if c.ExpShouldUpdateStatus {
				commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Eq[models.CommitStatus](c.ExpApplyStatus),
					Eq[command.Name](command.Apply),
					Eq(models.ProjectCounts{Success: c.ExpVCSApplyStatusSucc, Total: c.ExpVCSApplyStatusTotal, NoChanges: c.ExpVCSApplyNoChanges}),
				)
			} else {
				// Verify that UpdateCombinedCount was NOT called for Apply command
				commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.CommitStatus](),
					Eq[command.Name](command.Apply),
					Any[models.ProjectCounts](),
				)
			}
		})
	}
}

// prelockProject builds a plan ProjectContext for the pre-locking tests. The
// project name doubles as its directory so lock keys stay easy to read.
func prelockProject(name string, mode valid.RepoLocksMode) command.ProjectContext {
	return command.ProjectContext{
		CommandName:   command.Plan,
		BaseRepo:      testdata.GithubRepo,
		RepoRelDir:    name,
		Workspace:     "default",
		ProjectName:   name,
		RepoLocksMode: mode,
	}
}

// prelockCommandContext returns a plan command context for the pre-locking
// tests, with the VCS lookups the command runner performs already stubbed.
func prelockCommandContext(t *testing.T, logger logging.SimpleLogging) (*command.Context, *events.CommentCommand) {
	pull := &github.PullRequest{State: github.Ptr("open")}
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

	When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num))).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Eq(pull))).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	return &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}, &events.CommentCommand{Name: command.Plan}
}

// TestPlanCommandRunner_PreLockOrdering asserts that with
// --lock-all-projects-before-plan every lock is acquired before the first plan
// runs, and that projects which do not lock on plan are left alone.
func TestPlanCommandRunner_PreLockOrdering(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description string
		Enabled     bool
		Projects    []command.ProjectContext
		ExpSequence []string
	}{
		{
			Description: "every lock is taken before the first plan runs",
			Enabled:     true,
			Projects: []command.ProjectContext{
				prelockProject("a", valid.RepoLocksOnPlanMode),
				prelockProject("b", valid.RepoLocksOnPlanMode),
			},
			ExpSequence: []string{"lock a", "lock b", "plan a", "plan b"},
		},
		{
			Description: "projects that do not lock on plan are not pre-locked",
			Enabled:     true,
			Projects: []command.ProjectContext{
				prelockProject("a", valid.RepoLocksOnPlanMode),
				prelockProject("b", valid.RepoLocksOnApplyMode),
				prelockProject("c", valid.RepoLocksDisabledMode),
			},
			ExpSequence: []string{"lock a", "plan a", "plan b", "plan c"},
		},
		{
			Description: "duplicate lock keys are only locked once",
			Enabled:     true,
			Projects: []command.ProjectContext{
				prelockProject("a", valid.RepoLocksOnPlanMode),
				prelockProject("a", valid.RepoLocksOnPlanMode),
			},
			ExpSequence: []string{"lock a", "plan a", "plan a"},
		},
		{
			Description: "locking stays interleaved when the flag is off",
			Enabled:     false,
			Projects: []command.ProjectContext{
				prelockProject("a", valid.RepoLocksOnPlanMode),
				prelockProject("b", valid.RepoLocksOnPlanMode),
			},
			ExpSequence: []string{"plan a", "plan b"},
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			tmp := t.TempDir()
			database, err := boltdb.New(tmp)
			t.Cleanup(func() { database.Close() })
			Ok(t, err)

			setup(t, func(tc *TestConfig) {
				tc.database = database
				tc.lockAllProjectsBeforePlan = c.Enabled
			})

			var sequence []string

			When(projectLocker.TryLock(
				Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
				Any[string](), Any[models.Project](), Any[bool](),
			)).Then(func(params []Param) ReturnValues {
				project := params[4].(models.Project)
				workspace := params[3].(string)
				sequence = append(sequence, "lock "+project.ProjectName)
				return ReturnValues{
					&events.TryLockResponse{
						LockAcquired: true,
						UnlockFn:     func() error { return nil },
						LockKey:      models.GenerateLockKey(project, workspace),
					},
					nil,
				}
			})

			// Register one stub per distinct project. Two equal ProjectContext
			// values (the "duplicate lock keys" case) must only be stubbed once:
			// pegomock matches the second `When` call against the first project's
			// already-registered stub and invokes it right there, polluting
			// `sequence` before the run even starts.
			registeredPlanStubs := make(map[string]bool)
			for i := range c.Projects {
				projCtx := c.Projects[i]
				stubKey := fmt.Sprintf("%s/%s/%s", projCtx.RepoRelDir, projCtx.Workspace, projCtx.ProjectName)
				if registeredPlanStubs[stubKey] {
					continue
				}
				registeredPlanStubs[stubKey] = true
				When(projectCommandRunner.Plan(projCtx)).Then(func(_ []Param) ReturnValues {
					sequence = append(sequence, "plan "+projCtx.ProjectName)
					return ReturnValues{command.ProjectCommandOutput{
						PlanSuccess: &models.PlanSuccess{TerraformOutput: "no changes"},
					}}
				})
			}

			ctx, cmd := prelockCommandContext(t, logger)
			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(c.Projects, nil)

			planCommandRunner.Run(ctx, cmd)

			require.Equal(t, c.ExpSequence, sequence)
		})
	}
}

// TestPlanCommandRunner_PreLockAbortsWhenLockUnavailable asserts that a single
// unavailable lock stops the whole run before any plan starts, and that the
// locks this run had already taken are released again.
func TestPlanCommandRunner_PreLockAbortsWhenLockUnavailable(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	t.Cleanup(func() { database.Close() })
	Ok(t, err)

	lockingClient := locking.NewClient(database)

	// A competing pull request already holds the lock on project b.
	competitor := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num + 1,
	}
	blockedProject := models.NewProject(testdata.GithubRepo.FullName, "b", "b")
	held, err := lockingClient.TryLock(blockedProject, "default", competitor, testdata.User)
	Ok(t, err)
	require.True(t, held.LockAcquired, "the competing pull request should hold the lock")

	// The project locker renders the failure message, so it needs its own VCS
	// client to build the link to the blocking pull request.
	lockVCSClient := vcsmocks.NewMockClient()
	When(lockVCSClient.MarkdownPullLink(Any[models.PullRequest]())).ThenReturn("competing-pull", nil)

	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.lockAllProjectsBeforePlan = true
		tc.projectLocker = &events.DefaultProjectLocker{
			Locker:         lockingClient,
			NoOpLocker:     locking.NewNoOpLocker(),
			VCSClient:      lockVCSClient,
			ExecutableName: "atlantis",
		}
		tc.planLockingLocker = lockingClient
	})

	projectA := prelockProject("a", valid.RepoLocksOnPlanMode)
	projectB := prelockProject("b", valid.RepoLocksOnPlanMode)

	ctx, cmd := prelockCommandContext(t, logger)
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectA, projectB}, nil)

	planCommandRunner.Run(ctx, cmd)

	// Nothing planned -- not even project a, whose lock was free.
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectA)
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectB)
	require.True(t, ctx.CommandHasErrors)

	// The lock this run took for project a is released, and the competing pull
	// request keeps the one it already had.
	locks, err := lockingClient.List()
	Ok(t, err)
	require.Len(t, locks, 1)
	remaining, ok := locks[models.GenerateLockKey(blockedProject, "default")]
	require.True(t, ok, "the competing pull request should keep its lock")
	require.Equal(t, competitor.Num, remaining.Pull.Num)

	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Eq(ctx.Pull.Num), Any[string](), Eq("plan"),
	)
}

// TestPlanCommandRunner_PreLockReleasesUnplannedProjects covers the clean-up
// that makes pre-locking safe: when a run stops part way through, the projects
// that never produced a plan must not stay locked. Here an execution order
// group fails with abort_on_execution_order_fail, so the second group never
// runs at all -- the same shape as a run stopped by `atlantis cancel`.
func TestPlanCommandRunner_PreLockReleasesUnplannedProjects(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	t.Cleanup(func() { database.Close() })
	Ok(t, err)

	lockingClient := locking.NewClient(database)
	lockVCSClient := vcsmocks.NewMockClient()

	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.lockAllProjectsBeforePlan = true
		tc.projectLocker = &events.DefaultProjectLocker{
			Locker:         lockingClient,
			NoOpLocker:     locking.NewNoOpLocker(),
			VCSClient:      lockVCSClient,
			ExecutableName: "atlantis",
		}
		tc.planLockingLocker = lockingClient
	})

	projectA := prelockProject("a", valid.RepoLocksOnPlanMode)
	projectA.ParallelPlanEnabled = true
	projectA.AbortOnExecutionOrderFail = true
	projectA.ExecutionOrderGroup = 0

	projectB := prelockProject("b", valid.RepoLocksOnPlanMode)
	projectB.ParallelPlanEnabled = true
	projectB.AbortOnExecutionOrderFail = true
	projectB.ExecutionOrderGroup = 1

	When(projectCommandRunner.Plan(projectA)).ThenReturn(command.ProjectCommandOutput{
		Error: errors.New("terraform init failed"),
	})

	ctx, cmd := prelockCommandContext(t, logger)
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectA, projectB}, nil)

	planCommandRunner.Run(ctx, cmd)

	// Group 1 was skipped because group 0 failed.
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectB)

	// Neither project produced a plan, so neither may still hold a lock.
	locks, err := lockingClient.List()
	Ok(t, err)
	require.Empty(t, locks, "projects that produced no plan must not stay locked")
}

// TestPlanCommandRunner_PreLockOrderingAutoplan asserts that pre-locking also
// takes effect on the autoplan (auto-triggered) path, not just the
// comment-triggered one covered by TestPlanCommandRunner_PreLockOrdering.
func TestPlanCommandRunner_PreLockOrderingAutoplan(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	setup(t, func(tc *TestConfig) {
		tc.lockAllProjectsBeforePlan = true
	})

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logger,
		Scope:    metricstest.NewLoggingScope(t, logger, "atlantis"),
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.AutoTrigger,
	}

	projectA := prelockProject("a", valid.RepoLocksOnPlanMode)
	projectB := prelockProject("b", valid.RepoLocksOnPlanMode)

	When(pullReqStatusFetcher.FetchPullStatus(Any[logging.SimpleLogging](), Eq(modelPull))).ThenReturn(models.PullReqStatus{}, nil)
	When(projectCommandBuilder.BuildAutoplanCommands(ctx)).ThenReturn([]command.ProjectContext{projectA, projectB}, nil)

	var sequence []string
	When(projectLocker.TryLock(
		Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), Any[bool](),
	)).Then(func(params []Param) ReturnValues {
		project := params[4].(models.Project)
		workspace := params[3].(string)
		sequence = append(sequence, "lock "+project.ProjectName)
		return ReturnValues{
			&events.TryLockResponse{
				LockAcquired: true,
				UnlockFn:     func() error { return nil },
				LockKey:      models.GenerateLockKey(project, workspace),
			},
			nil,
		}
	})

	for _, projCtx := range []command.ProjectContext{projectA, projectB} {
		projCtx := projCtx
		When(projectCommandRunner.Plan(projCtx)).Then(func(_ []Param) ReturnValues {
			sequence = append(sequence, "plan "+projCtx.ProjectName)
			return ReturnValues{command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{TerraformOutput: "no changes"},
			}}
		})
	}

	planCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Plan})

	require.Equal(t, []string{"lock a", "lock b", "plan a", "plan b"}, sequence)
}

// TestPlanCommandRunner_PreLockAbortMessageUsesWorkspaceForUnnamedProjects
// covers projectDisplayName's fallback branch: a project with no configured
// name is identified by directory + workspace instead of a bare name.
func TestPlanCommandRunner_PreLockAbortMessageUsesWorkspaceForUnnamedProjects(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	t.Cleanup(func() { database.Close() })
	Ok(t, err)

	lockingClient := locking.NewClient(database)

	// A competing pull request already holds the lock, so pre-locking aborts
	// immediately and the failure message must name this (unnamed) project.
	competitor := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num + 1}
	blockedProject := models.NewProject(testdata.GithubRepo.FullName, "unnamed-dir", "")
	held, err := lockingClient.TryLock(blockedProject, "default", competitor, testdata.User)
	Ok(t, err)
	require.True(t, held.LockAcquired)

	lockVCSClient := vcsmocks.NewMockClient()
	When(lockVCSClient.MarkdownPullLink(Any[models.PullRequest]())).ThenReturn("competing-pull", nil)

	vcsClient := setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.lockAllProjectsBeforePlan = true
		tc.projectLocker = &events.DefaultProjectLocker{
			Locker:         lockingClient,
			NoOpLocker:     locking.NewNoOpLocker(),
			VCSClient:      lockVCSClient,
			ExecutableName: "atlantis",
		}
		tc.planLockingLocker = lockingClient
	})

	unnamedProject := command.ProjectContext{
		CommandName:   command.Plan,
		BaseRepo:      testdata.GithubRepo,
		RepoRelDir:    "unnamed-dir",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	// A second project whose lock is never even attempted (unnamedProject
	// fails first): its result becomes the "skipped" message, which embeds
	// projectDisplayName(unnamedProject) -- the branch under test.
	otherProject := prelockProject("other", valid.RepoLocksOnPlanMode)

	ctx, cmd := prelockCommandContext(t, logger)
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(
		[]command.ProjectContext{unnamedProject, otherProject}, nil)

	planCommandRunner.Run(ctx, cmd)

	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Eq(ctx.Pull.Num), Any[string](), Eq("plan"),
	).GetCapturedArguments()
	require.Contains(t, comment, "`unnamed-dir` (workspace `default`)")
}

// failingDatabase wraps a real db.Database, forcing TryLock and
// UpdatePullWithResults to fail. This lets a test exercise both
// preLockProjects' TryLock-error branch and handlePreLockAbort's DB-write
// failure branch, without disturbing any other database operation (like the
// unlock calls deletePlansAndPlanLocks makes before pre-locking even starts).
type failingDatabase struct {
	db.Database
}

func (f *failingDatabase) TryLock(models.ProjectLock) (bool, models.ProjectLock, error) {
	return false, models.ProjectLock{}, errors.New("simulated lock backend failure")
}

func (f *failingDatabase) UpdatePullWithResults(models.PullRequest, []command.ProjectResult) (models.PullStatus, error) {
	return models.PullStatus{}, errors.New("simulated database write failure")
}

// TestPlanCommandRunner_PreLockAcquireErrorUpdatesCommitStatusEvenIfDBWriteFails
// covers two paths TestPlanCommandRunner_PreLockAbortsWhenLockUnavailable
// doesn't: TryLock returning a real error (not just LockAcquired: false), and
// handlePreLockAbort's fallback when writing the aborted result to the DB
// itself also fails -- the commit status must still move off Pending instead
// of being left stuck there forever.
func TestPlanCommandRunner_PreLockAcquireErrorUpdatesCommitStatusEvenIfDBWriteFails(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	t.Cleanup(func() { boltDB.Close() })
	Ok(t, err)

	failingDB := &failingDatabase{Database: boltDB}
	lockingClient := locking.NewClient(failingDB)
	lockVCSClient := vcsmocks.NewMockClient()

	setup(t, func(tc *TestConfig) {
		tc.database = failingDB
		tc.lockAllProjectsBeforePlan = true
		tc.projectLocker = &events.DefaultProjectLocker{
			Locker:         lockingClient,
			NoOpLocker:     locking.NewNoOpLocker(),
			VCSClient:      lockVCSClient,
			ExecutableName: "atlantis",
		}
		tc.planLockingLocker = lockingClient
	})

	projectA := prelockProject("a", valid.RepoLocksOnPlanMode)

	ctx, cmd := prelockCommandContext(t, logger)
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectA}, nil)

	planCommandRunner.Run(ctx, cmd)

	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectA)
	require.True(t, ctx.CommandHasErrors)

	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(ctx.Pull), Eq(models.FailedCommitStatus), Eq(command.Plan))
	commitUpdater.VerifyWasCalledOnce().UpdateCombined(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(ctx.Pull), Eq(models.FailedCommitStatus), Eq(command.Apply))
}

// panicOnLockLocker wraps a ProjectLocker and panics instead of delegating
// when asked to lock panicForProject, to test panic recovery in
// preLockProjects.
type panicOnLockLocker struct {
	inner           events.ProjectLocker
	panicForProject string
}

func (p *panicOnLockLocker) TryLock(log logging.SimpleLogging, pull models.PullRequest, user models.User, workspace string, project models.Project, repoLocking bool) (*events.TryLockResponse, error) {
	if project.ProjectName == p.panicForProject {
		panic("simulated backend panic")
	}
	return p.inner.TryLock(log, pull, user, workspace, project, repoLocking)
}

// TestPlanCommandRunner_PreLockPanicReleasesAcquiredLocks asserts that a panic
// partway through preLockProjects still releases every lock already acquired
// in that run before the panic propagates.
func TestPlanCommandRunner_PreLockPanicReleasesAcquiredLocks(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	t.Cleanup(func() { database.Close() })
	Ok(t, err)

	lockingClient := locking.NewClient(database)
	lockVCSClient := vcsmocks.NewMockClient()

	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.lockAllProjectsBeforePlan = true
		tc.projectLocker = &panicOnLockLocker{
			inner: &events.DefaultProjectLocker{
				Locker:         lockingClient,
				NoOpLocker:     locking.NewNoOpLocker(),
				VCSClient:      lockVCSClient,
				ExecutableName: "atlantis",
			},
			panicForProject: "b",
		}
		tc.planLockingLocker = lockingClient
	})

	projectA := prelockProject("a", valid.RepoLocksOnPlanMode)
	projectB := prelockProject("b", valid.RepoLocksOnPlanMode)

	ctx, cmd := prelockCommandContext(t, logger)
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{projectA, projectB}, nil)

	require.Panics(t, func() {
		planCommandRunner.Run(ctx, cmd)
	})

	locks, err := lockingClient.List()
	Ok(t, err)
	require.Empty(t, locks, "the lock acquired before the panic must be released, not left behind")
}

// failingUnlockLocker wraps a real locking.Locker and forces
// UnlockIfOwnedByPull to fail for one specific project, but only once that
// project is actually locked -- otherwise the run's pre-existing "clean up
// any stale locks from a previous run" step (which happens before pre-locking
// even starts, and unlocks nothing yet) would itself spuriously fail.
type failingUnlockLocker struct {
	locking.Locker
	failProjectName string
}

func (f *failingUnlockLocker) UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error) {
	if project.ProjectName == f.failProjectName {
		if lock, err := f.Locker.GetLock(models.GenerateLockKey(project, workspace)); err == nil && lock != nil {
			return nil, errors.New("simulated unlock failure")
		}
	}
	return f.Locker.UnlockIfOwnedByPull(project, workspace, pullNum)
}

// TestPlanCommandRunner_ReleaseUnplannedLocksContinuesPastUnlockFailure
// asserts that when releasing the locks of projects that produced no plan,
// one project's unlock failure doesn't stop the others from being released.
func TestPlanCommandRunner_ReleaseUnplannedLocksContinuesPastUnlockFailure(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	t.Cleanup(func() { database.Close() })
	Ok(t, err)

	lockingClient := locking.NewClient(database)
	lockVCSClient := vcsmocks.NewMockClient()

	setup(t, func(tc *TestConfig) {
		tc.database = database
		tc.lockAllProjectsBeforePlan = true
		tc.projectLocker = &events.DefaultProjectLocker{
			Locker:         lockingClient,
			NoOpLocker:     locking.NewNoOpLocker(),
			VCSClient:      lockVCSClient,
			ExecutableName: "atlantis",
		}
		tc.planLockingLocker = &failingUnlockLocker{Locker: lockingClient, failProjectName: "b"}
	})

	projectZ := prelockProject("z", valid.RepoLocksOnPlanMode)
	projectZ.ParallelPlanEnabled = true
	projectZ.AbortOnExecutionOrderFail = true
	projectZ.ExecutionOrderGroup = 0

	projectA := prelockProject("a", valid.RepoLocksOnPlanMode)
	projectA.ParallelPlanEnabled = true
	projectA.AbortOnExecutionOrderFail = true
	projectA.ExecutionOrderGroup = 1

	projectB := prelockProject("b", valid.RepoLocksOnPlanMode)
	projectB.ParallelPlanEnabled = true
	projectB.AbortOnExecutionOrderFail = true
	projectB.ExecutionOrderGroup = 1

	projectC := prelockProject("c", valid.RepoLocksOnPlanMode)
	projectC.ParallelPlanEnabled = true
	projectC.AbortOnExecutionOrderFail = true
	projectC.ExecutionOrderGroup = 1

	When(projectCommandRunner.Plan(projectZ)).ThenReturn(command.ProjectCommandOutput{
		Error: errors.New("terraform init failed"),
	})

	ctx, cmd := prelockCommandContext(t, logger)
	When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(
		[]command.ProjectContext{projectZ, projectA, projectB, projectC}, nil)

	planCommandRunner.Run(ctx, cmd)

	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectA)
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectB)
	projectCommandRunner.VerifyWasCalled(Never()).Plan(projectC)

	locks, err := lockingClient.List()
	Ok(t, err)
	_, aLocked := locks[models.GenerateLockKey(models.NewProject(testdata.GithubRepo.FullName, "a", "a"), "default")]
	_, bLocked := locks[models.GenerateLockKey(models.NewProject(testdata.GithubRepo.FullName, "b", "b"), "default")]
	_, cLocked := locks[models.GenerateLockKey(models.NewProject(testdata.GithubRepo.FullName, "c", "c"), "default")]
	require.False(t, aLocked, "project a must be released despite project b's simulated unlock failure")
	require.True(t, bLocked, "project b's lock remains because its unlock was simulated to fail")
	require.False(t, cLocked, "project c must be released despite project b's simulated unlock failure")
}
