package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v66/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPlanCommandRunner_IsSilenced(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description       string
		Matched           bool
		Targeted          bool
		VCSStatusSilence  bool
		PrevPlanStored    bool // stores a 1/1 passing plan in the backend
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
			Description:      "When planning with silenced VCS status, don't do anything",
			VCSStatusSilence: true,
			ExpVCSStatusSet:  false,
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
			db, err := db.New(tmp)
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.SilenceNoProjects = true
				tc.silenceVCSStatusNoProjects = c.VCSStatusSilence
				tc.backend = db
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
						Command:     command.Plan,
						RepoRelDir:  "prevdir",
						Workspace:   "default",
						PlanSuccess: &models.PlanSuccess{},
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
		Description       string
		ProjectContexts   []command.ProjectContext
		ProjectResults    []command.ProjectResult
		RunnerInvokeMatch []*EqMatcher
		PrevPlanStored    bool
	}{
		{
			Description: "When first plan fails, the second don't run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					Workspace:                  "first",
					ProjectName:                "First",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					Workspace:                  "second",
					ProjectName:                "Second",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
		},
		{
			Description: "When first fails, the second will not run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					ProjectName:                "Second",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
				{
					Command:     command.Plan,
					PlanSuccess: &models.PlanSuccess{},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Never(),
			},
		},
		{
			Description: "When first fails by autorun, the second will not run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:                command.Plan,
					AutoplanEnabled:            true,
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					AutoplanEnabled:            true,
					ExecutionOrderGroup:        1,
					ProjectName:                "Second",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
				{
					Command:     command.Plan,
					PlanSuccess: &models.PlanSuccess{},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Never(),
			},
		},
		{
			Description: "When both in a group of two succeeds, the following two will run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					ProjectName:                "Second",
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					ProjectName:                "Third",
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					ProjectName:                "Fourth",
					AbortOnExcecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Command: command.Plan,
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
		},
		{
			Description: "When one out of two fails, the following two will not run",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelPlanEnabled:        true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					ProjectName:                "Second",
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					ProjectName:                "Third",
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					AbortOnExcecutionOrderFail: true,
					ProjectName:                "Fourth",
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
				{
					Command: command.Plan,
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
		},
		{
			Description: "Don't block when parallel is not set",
			ProjectContexts: []command.ProjectContext{
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					AbortOnExcecutionOrderFail: true,
				},
				{
					CommandName:                command.Plan,
					ExecutionOrderGroup:        1,
					ProjectName:                "Second",
					AbortOnExcecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
		},
		{
			Description: "Don't block when abortOnExcecutionOrderFail is not set",
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
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Plan,
					Error:   errors.New("shabang"),
				},
				{
					Command: command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "true",
					},
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			// vcsClient := setup(t)

			tmp := t.TempDir()
			db, err := db.New(tmp)
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.backend = db
			})

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			pull := &github.PullRequest{
				State: github.String("open"),
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
				When(projectCommandRunner.Plan(c.ProjectContexts[i])).ThenReturn(c.ProjectResults[i])
			}

			planCommandRunner.Run(ctx, cmd)

			for i := range c.ProjectContexts {
				projectCommandRunner.VerifyWasCalled(c.RunnerInvokeMatch[i]).Plan(c.ProjectContexts[i])
			}

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
		ProjectResults         []command.ProjectResult
		PrevPlanStored         bool // stores a previous "No changes" plan in the backend
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "mydir",
					Command:    command.Plan,
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "mydir",
					Command:    command.Plan,
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "mydir",
					Command:    command.Plan,
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "mydir",
					Command:    command.Plan,
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "prevdir",
					Workspace:  "default",
					Command:    command.Plan,
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "prevdir",
					Workspace:  "default",
					Command:    command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 0 to add, 0 to change, 1 to destroy.",
					},
				},
				{
					RepoRelDir: "mydir",
					Command:    command.Plan,
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
			ProjectResults: []command.ProjectResult{
				{
					RepoRelDir: "prevdir",
					Workspace:  "default",
					Command:    command.Plan,
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
				{
					RepoRelDir: "mydir",
					Command:    command.Plan,
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
			db, err := db.New(tmp)
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.backend = db
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
						PlanSuccess: &models.PlanSuccess{
							TerraformOutput: "No changes. Your infrastructure matches the configuration.",
						},
					},
				})
				Ok(t, err)
			}

			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).ThenReturn(c.ProjectContexts, nil)

			for i := range c.ProjectContexts {
				When(projectCommandRunner.Plan(c.ProjectContexts[i])).ThenReturn(c.ProjectResults[i])
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
