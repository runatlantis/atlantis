package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v66/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

func TestApplyCommandRunner_IsLocked(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

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
				State: github.String("open"),
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			When(githubGetter.GetPullRequest(logger, testdata.GithubRepo, testdata.Pull.Num)).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(logger, pull)).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(applyLockChecker.CheckApplyLock()).ThenReturn(locking.ApplyCommandLock{Locked: c.ApplyLocked}, c.ApplyLockError)
			applyCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Apply})

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(c.ExpComment), Eq("apply"))
		})
	}
}

func TestApplyCommandRunner_IsSilenced(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

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

			applyCommandRunner.Run(ctx, cmd)

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
					Eq[command.Name](command.Apply),
					Eq(c.ExpVCSStatusSucc),
					Eq(c.ExpVCSStatusTotal),
				)
			} else {
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

func TestApplyCommandRunner_ExecutionOrder(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description       string
		ProjectContexts   []command.ProjectContext
		ProjectResults    []command.ProjectResult
		RunnerInvokeMatch []*EqMatcher
		ExpComment        string
	}{
		{
			Description: "When first apply fails, the second don't run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelApplyEnabled:       true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					ProjectName:                "Second",
					ParallelApplyEnabled:       true,
					AbortOnExcecutionOrderFail: true,
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
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "When first apply fails, the second not will run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelApplyEnabled:       true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					ProjectName:                "Second",
					ParallelApplyEnabled:       true,
					AbortOnExcecutionOrderFail: true,
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
			ExpComment: "Ran Apply for dir: `` workspace: ``\n\n**Apply Error**\n```\nshabang\n```",
		},
		{
			Description: "When both in a group of two succeeds, the following two will run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelApplyEnabled:       true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        0,
					ProjectName:                "Second",
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					ProjectName:                "Third",
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					ProjectName:                "Fourth",
					AbortOnExcecutionOrderFail: true,
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
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "When one out of two fails, the following two will not run",
			ProjectContexts: []command.ProjectContext{
				{
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					ParallelApplyEnabled:       true,
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        0,
					ProjectName:                "Second",
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					ProjectName:                "Third",
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					AbortOnExcecutionOrderFail: true,
					ProjectName:                "Fourth",
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
					ExecutionOrderGroup:        0,
					ProjectName:                "First",
					AbortOnExcecutionOrderFail: true,
				},
				{
					ExecutionOrderGroup:        1,
					ProjectName:                "Second",
					AbortOnExcecutionOrderFail: true,
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
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
		{
			Description: "Don't block when abortOnExcecutionOrderFail is not set",
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
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. dir: `` workspace: ``\n1. dir: `` workspace: ``\n---\n\n### 1. dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### " +
				"2. dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			pull := &github.PullRequest{
				State: github.String("open"),
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

			When(githubGetter.GetPullRequest(logger, testdata.GithubRepo, testdata.Pull.Num)).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(logger, pull)).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).ThenReturn(c.ProjectContexts, nil)
			for i := range c.ProjectContexts {
				When(projectCommandRunner.Apply(c.ProjectContexts[i])).ThenReturn(c.ProjectResults[i])
			}

			applyCommandRunner.Run(ctx, cmd)

			for i := range c.ProjectContexts {
				projectCommandRunner.VerifyWasCalled(c.RunnerInvokeMatch[i]).Apply(c.ProjectContexts[i])

			}

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(c.ExpComment), Eq("apply"),
			)
		})
	}
}
