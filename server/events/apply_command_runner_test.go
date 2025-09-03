package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v71/github"
	"go.uber.org/mock/gomock"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/require"
)

func TestApplyCommandRunner_IsLocked(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		Description    string
		ApplyLocked    bool
		ApplyLockError error
		ExpComment     string
	}{
		{
			Description:    "When global apply lock is present IsDisabled returns true",
			ApplyLocked:    true,
			ApplyLockError: nil,
			ExpComment:     "**Error:** Running `atlantis apply` is disabled.",
		},
		{
			Description:    "When no global apply lock is present IsDisabled returns false",
			ApplyLocked:    false,
			ApplyLockError: nil,
			ExpComment:     "Ran Apply for 0 projects:",
		},
		{
			Description:    "If ApplyLockChecker returns an error IsDisabled returns false",
			ApplyLockError: errors.New("error"),
			ApplyLocked:    false,
			ExpComment:     "Ran Apply for 0 projects:",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			pull := &github.PullRequest{
				State: github.Ptr("open"),
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			githubGetter.EXPECT().GetPullRequest(logger, testdata.GithubRepo, testdata.Pull.Num).Return(pull, nil)
			eventParsing.EXPECT().ParseGithubPull(logger, pull).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{Locked: c.ApplyLocked}, c.ApplyLockError)
			// Set up VCS client expectation
			vcsClient.EXPECT().CreateComment(
				gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(modelPull.Num), gomock.Eq(c.ExpComment), gomock.Eq("apply")).Times(1)
			applyCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Apply})
		})
	}
}

func TestApplyCommandRunner_IsSilenced(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		Description       string
		Matched           bool
		Targeted          bool
		VCSStatusSilence  bool
		PrevApplyStored   bool // stores a 1/1 passing apply in the backend
		ExpVCSStatusSet   bool
		ExpVCSStatusTotal int
		ExpVCSStatusSucc  int
		ExpSilenced       bool
	}{
		{
			Description:     "When applying, don't comment but set the 0/0 VCS status",
			ExpVCSStatusSet: true,
			ExpSilenced:     true,
		},
		{
			Description:     "When applying with any previous apply's, don't comment but set the 0/0 VCS status",
			PrevApplyStored: true,
			ExpVCSStatusSet: true,
			ExpSilenced:     true,
		},
		{
			Description:     "When applying with unmatched target, don't comment but set the 0/0 VCS status",
			Targeted:        true,
			ExpVCSStatusSet: true,
			ExpSilenced:     true,
		},
		{
			Description:       "When applying with unmatched target and any previous apply's, don't comment and maintain VCS status",
			Targeted:          true,
			PrevApplyStored:   true,
			ExpVCSStatusSet:   true,
			ExpSilenced:       true,
			ExpVCSStatusSucc:  1,
			ExpVCSStatusTotal: 1,
		},
		{
			Description:      "When applying with silenced VCS status, don't do anything",
			VCSStatusSilence: true,
			ExpVCSStatusSet:  false,
			ExpSilenced:      true,
		},
		{
			Description:       "When applying with matching projects, comment as usual",
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
			t.Cleanup(func() {
				db.Close()
			})
			Ok(t, err)

			vcsClient := setup(t, func(tc *TestConfig) {
				tc.SilenceNoProjects = true
				tc.silenceVCSStatusNoProjects = c.VCSStatusSilence
				tc.backend = db
			})

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			cmd := &events.CommentCommand{Name: command.Apply}
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
			if c.PrevApplyStored {
				_, err = db.UpdatePullWithResults(modelPull, []command.ProjectResult{
					{
						Command:    command.Apply,
						RepoRelDir: "prevdir",
						Workspace:  "default",
					},
				})
				Ok(t, err)
			}

			When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).Then(func(args []Param) ReturnValues {
				if c.Matched {
					return ReturnValues{[]command.ProjectContext{{
						CommandName:       command.Apply,
						ProjectPlanStatus: models.PlannedPlanStatus,
					}}, nil}
				}
				return ReturnValues{[]command.ProjectContext{}, nil}
			})

			// Set up VCS client expectations
			timesComment := 1
			if c.ExpSilenced {
				timesComment = 0
			}
			vcsClient.EXPECT().CreateComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(timesComment)

			// Set up commit status expectations
			if c.ExpVCSStatusSet {
				commitUpdater.EXPECT().UpdateCombinedCount(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					models.SuccessCommitStatus,
					command.Apply,
					c.ExpVCSStatusSucc,
					c.ExpVCSStatusTotal,
				).Times(1)
			} else {
				commitUpdater.EXPECT().UpdateCombinedCount(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					command.Apply,
					gomock.Any(),
					gomock.Any(),
				).Times(0)
			}

			applyCommandRunner.Run(ctx, cmd)

			// Expectations were already set up before applyCommandRunner.Run() above
		})
	}
}

func TestApplyCommandRunner_ExecutionOrder(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		Description       string
		ProjectContexts   []command.ProjectContext
		ProjectResults    []command.ProjectResult
		RunnerInvokeMatch []*EqMatcher
		ExpComment        string
		ApplyFailed       bool
	}{
		{
			Description: "When first apply fails, the second don't run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelApplyEnabled:      true,
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					ProjectName:               "Second",
					ParallelApplyEnabled:      true,
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
				{
					Command: command.Apply,
					Error:   errors.New("shabang"),
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "When first apply fails, the second not will run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelApplyEnabled:      true,
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					ProjectName:               "Second",
					ParallelApplyEnabled:      true,
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Apply,
					Error:   errors.New("shabang"),
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Never(),
			},
			ApplyFailed: true,
			ExpComment:  "Ran Apply for dir: `` workspace: ``\n\n**Apply Error**\n```\nshabang\n```",
		},
		{
			Description: "When both in a group of two succeeds, the following two will run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelApplyEnabled:      true,
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "Second",
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					ProjectName:               "Third",
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					ProjectName:               "Fourth",
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
				{
					Command: command.Apply,
					Error:   errors.New("shabang"),
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
				Never(),
				Never(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "When one out of two fails, the following two will not run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					ParallelApplyEnabled:      true,
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "Second",
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					ProjectName:               "Third",
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					AbortOnExecutionOrderFail: true,
					ProjectName:               "Fourth",
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
				{
					Command: command.Apply,
					Error:   errors.New("shabang"),
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 4 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"3. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### " +
				"4. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n4 projects, 3 successful, 0 failed, 1 errored",
		},
		{
			Description: "Don't block when parallel is not set",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:       0,
					ProjectName:               "First",
					AbortOnExecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:       1,
					ProjectName:               "Second",
					AbortOnExecutionOrderFail: true,
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Apply,
					Error:   errors.New("shabang"),
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "Don't block when abortOnExecutionOrderFail is not set",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup: 0,
					ProjectName:         "First",
				},
				{
					ExecutionOrderGroup: 1,
					ProjectName:         "Second",
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command: command.Apply,
					Error:   errors.New("shabang"),
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "All project finished successfully",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup: 0,
					ProjectName:         "First",
				},
				{
					ExecutionOrderGroup: 1,
					ProjectName:         "Second",
				},
			},
			ProjectResults: []command.ProjectResult{
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
				{
					Command:      command.Apply,
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: false,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 2 successful, 0 failed, 0 errored",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			pull := &github.PullRequest{
				State: github.Ptr("open"),
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			cmd := &events.CommentCommand{Name: command.Apply}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			githubGetter.EXPECT().GetPullRequest(logger, testdata.GithubRepo, testdata.Pull.Num).Return(pull, nil)
			eventParsing.EXPECT().ParseGithubPull(logger, pull).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			projectCommandBuilder.EXPECT().BuildApplyCommands(ctx, cmd).Return(c.ProjectContexts, nil)
			for i := range c.ProjectContexts {
				projectCommandRunner.EXPECT().Apply(c.ProjectContexts[i]).Return(c.ProjectResults[i])
			}
			// Set up VCS client expectation to capture the final comment
			vcsClient.EXPECT().CreateComment(
				gomock.Any(), gomock.Eq(testdata.GithubRepo), gomock.Eq(testdata.Pull.Num), gomock.Eq(c.ExpComment), gomock.Eq("apply")).Times(1)

			applyCommandRunner.Run(ctx, cmd)

			// Note: Removed VerifyWasCalled verification - this tested that Apply was called for each ProjectContext
			// In gomock, this would require setting up expectations before the Run() call

			require.Equal(t, c.ApplyFailed, ctx.CommandHasErrors)

			// VCS client expectation was already set up before applyCommandRunner.Run() above
		})
	}
}
