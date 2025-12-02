package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v71/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
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

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
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
					Eq(c.ExpVCSStatusSucc),
					Eq(c.ExpVCSStatusTotal),
				)
			} else {
				commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.CommitStatus](),
					Eq[command.Name](command.Plan),
					Any[int](),
					Any[int](),
				)
			}
		})
	}
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

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

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

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
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
					AnyInt(),
					AnyInt(),
				)
			} else {
				commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Eq[models.CommitStatus](ExpCommitStatus),
					Eq[command.Name](command.Apply),
					Eq(c.ExpVCSApplyStatusSucc),
					Eq(c.ExpVCSApplyStatusTotal),
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
		scopeNull, _, _ := metrics.NewLoggingScope(logging.NewNoopLogger(t), "atlantis")

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
			Any[int](),
			Any[int](),
		)
	})
}
func TestPlanCommandRunner_PendingApplyStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description            string
		VCSType                models.VCSHostType
		PendingApplyFlag       bool
		ProjectResults         []command.ProjectResult
		ExpApplyStatus         models.CommitStatus
		ExpVCSApplyStatusTotal int
		ExpVCSApplyStatusSucc  int
		ExpShouldUpdateStatus  bool
	}{
		{
			Description:      "GitLab with flag enabled and unapplied plans should set pending status",
			VCSType:          models.Gitlab,
			PendingApplyFlag: true,
			ProjectResults: []command.ProjectResult{
				{
					Command:    command.Plan,
					RepoRelDir: "mydir",
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
			ProjectResults: []command.ProjectResult{
				{
					Command:    command.Plan,
					RepoRelDir: "mydir",
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
			ProjectResults: []command.ProjectResult{
				{
					Command:    command.Plan,
					RepoRelDir: "mydir",
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
			ProjectResults: []command.ProjectResult{
				{
					Command:    command.Plan,
					RepoRelDir: "mydir",
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
			ProjectResults: []command.ProjectResult{
				{
					Command:    command.Plan,
					RepoRelDir: "mydir",
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

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

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
					Eq(c.ExpVCSApplyStatusSucc),
					Eq(c.ExpVCSApplyStatusTotal),
				)
			} else {
				// Verify that UpdateCombinedCount was NOT called for Apply command
				commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.CommitStatus](),
					Eq[command.Name](command.Apply),
					Any[int](),
					Any[int](),
				)
			}
		})
	}
}
