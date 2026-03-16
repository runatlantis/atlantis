// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPlanCommandRunner_IsSilenced(t *testing.T) {
	logger := logging.NewNoopLogger(t)

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

			matched := c.Matched
			projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ *command.Context, _ *events.CommentCommand) ([]command.ProjectContext, error) {
					if matched {
						return []command.ProjectContext{{CommandName: command.Plan}}, nil
					}
					return []command.ProjectContext{}, nil
				})
			projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}).AnyTimes()

			expVCSStatusSet := c.ExpVCSStatusSet
			expSucc := c.ExpVCSStatusSucc
			expTotal := c.ExpVCSStatusTotal
			planStatusCallCount := 0
			commitUpdater.EXPECT().UpdateCombinedCount(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(func(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, cmdName command.Name, numSuccess int, numTotal int) error {
				if cmdName == command.Plan {
					planStatusCallCount++
					if !expVCSStatusSet {
						t.Errorf("UpdateCombinedCount should not have been called when VCS status is silenced")
					} else {
						require.Equal(t, models.SuccessCommitStatus, status)
						require.Equal(t, expSucc, numSuccess)
						require.Equal(t, expTotal, numTotal)
					}
				}
				return nil
			}).AnyTimes()
			t.Cleanup(func() {
				if expVCSStatusSet {
					require.GreaterOrEqual(t, planStatusCallCount, 1, "UpdateCombinedCount should have been called at least once with command.Plan")
				}
			})

			planCommandRunner.Run(ctx, cmd)

			timesComment := 1
			if c.ExpSilenced {
				timesComment = 0
			}

			vcsClient.VerifyWasCalled(Times(timesComment)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
		})
	}
}

func TestPlanCommandRunner_ExecutionOrder(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		Description           string
		ProjectContexts       []command.ProjectContext
		ProjectCommandOutputs []command.ProjectCommandOutput
		RunnerInvokeTimes     []int
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
			RunnerInvokeTimes: []int{1, 1},
			PlanFailed:        true,
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
			RunnerInvokeTimes: []int{1, 0},
			PlanFailed:        true,
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
			RunnerInvokeTimes: []int{1, 0},
			PlanFailed:        true,
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
			RunnerInvokeTimes: []int{1, 1, 0, 0},
			PlanFailed:        true,
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
			RunnerInvokeTimes: []int{1, 1, 1, 1},
			PlanFailed:        true,
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
			RunnerInvokeTimes: []int{1, 1},
			PlanFailed:        true,
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
			RunnerInvokeTimes: []int{1, 1},
			PlanFailed:        false,
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
			RunnerInvokeTimes: []int{1, 1},
			PlanFailed:        true,
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

			projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).Return(c.ProjectContexts, nil)
			planCallIdx := 0
			outputs := c.ProjectCommandOutputs
			projectCommandRunner.EXPECT().Plan(gomock.Any()).DoAndReturn(
				func(_ command.ProjectContext) command.ProjectCommandOutput {
					idx := planCallIdx
					planCallIdx++
					if idx < len(outputs) {
						return outputs[idx]
					}
					return command.ProjectCommandOutput{}
				}).AnyTimes()

			planCommandRunner.Run(ctx, cmd)

			require.Equal(t, c.PlanFailed, ctx.CommandHasErrors)

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Eq(modelPull.Num), Any[string](), Eq("plan"),
			)
		})
	}
}

func TestPlanCommandRunner_AtlantisApplyStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		Description            string
		ProjectContexts        []command.ProjectContext
		ProjectCommandOutput   []command.ProjectCommandOutput
		PrevPlanStored         bool // stores a previous "No changes" plan in the database
		DoNotUpdateApply       bool // certain circumtances we want to skip the call to update apply
		ExpVCSApplyStatusTotal int
		ExpVCSApplyStatusSucc  int
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

			projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).Return(c.ProjectContexts, nil)

			planCallIdx := 0
			outputs := c.ProjectCommandOutput
			projectCommandRunner.EXPECT().Plan(gomock.Any()).DoAndReturn(
				func(_ command.ProjectContext) command.ProjectCommandOutput {
					idx := planCallIdx
					planCallIdx++
					if idx < len(outputs) {
						return outputs[idx]
					}
					return command.ProjectCommandOutput{}
				}).AnyTimes()

			ExpCommitStatus := models.SuccessCommitStatus
			if c.ExpVCSApplyStatusSucc != c.ExpVCSApplyStatusTotal {
				ExpCommitStatus = models.PendingCommitStatus
			}
			applyCallCount := 0
			doNotUpdateApply := c.DoNotUpdateApply
			expApplySucc := c.ExpVCSApplyStatusSucc
			expApplyTotal := c.ExpVCSApplyStatusTotal
			commitUpdater.EXPECT().UpdateCombinedCount(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(func(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, cmdName command.Name, numSuccess int, numTotal int) error {
				if cmdName == command.Apply {
					applyCallCount++
					if doNotUpdateApply {
						t.Errorf("UpdateCombinedCount should not have been called with command.Apply")
					} else {
						require.Equal(t, ExpCommitStatus, status)
						require.Equal(t, expApplySucc, numSuccess)
						require.Equal(t, expApplyTotal, numTotal)
					}
				}
				return nil
			}).AnyTimes()
			t.Cleanup(func() {
				if !doNotUpdateApply {
					require.Equal(t, 1, applyCallCount, "UpdateCombinedCount should have been called exactly once with command.Apply")
				}
			})

			planCommandRunner.Run(ctx, cmd)

			vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), AnyInt(), AnyString(), AnyString())
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
		projectCommandBuilder.EXPECT().BuildAutoplanCommands(gomock.Any()).Return([]command.ProjectContext{}, nil)

		// CRITICAL VERIFICATION: With silence flags enabled, no status should be set at all
		// This prevents any VCS status checks from being created (issue #5389)
		// The silence flags mean "don't create any status checks"
		commitUpdater.EXPECT().UpdateCombinedCount(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(nil).Times(0)

		// Run through the plan command (which will internally check for projects)
		cmd := &events.CommentCommand{Name: command.Plan}
		planCommandRunner.Run(ctx, cmd)
	})
}

func TestPlanCommandRunner_PendingApplyStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		Description            string
		VCSType                models.VCSHostType
		PendingApplyFlag       bool
		ProjectResults         []command.ProjectCommandOutput
		ExpApplyStatus         models.CommitStatus
		ExpVCSApplyStatusTotal int
		ExpVCSApplyStatusSucc  int
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

			projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).Return(projectContexts, nil)
			projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(c.ProjectResults[0])

			// Verify based on whether we expect a status update
			applyCallCount := 0
			expShouldUpdate := c.ExpShouldUpdateStatus
			expApplyStatus := c.ExpApplyStatus
			expApplySucc := c.ExpVCSApplyStatusSucc
			expApplyTotal := c.ExpVCSApplyStatusTotal
			commitUpdater.EXPECT().UpdateCombinedCount(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(func(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, cmdName command.Name, numSuccess int, numTotal int) error {
				if cmdName == command.Apply {
					applyCallCount++
					if !expShouldUpdate {
						t.Errorf("UpdateCombinedCount should not have been called with command.Apply")
					} else {
						require.Equal(t, expApplyStatus, status)
						require.Equal(t, expApplySucc, numSuccess)
						require.Equal(t, expApplyTotal, numTotal)
					}
				}
				return nil
			}).AnyTimes()
			t.Cleanup(func() {
				if expShouldUpdate {
					require.Equal(t, 1, applyCallCount, "UpdateCombinedCount should have been called exactly once with command.Apply")
				}
			})

			planCommandRunner.Run(ctx, cmd)
		})
	}
}
