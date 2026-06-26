// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v88/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/require"
)

func TestApplyCommandRunner_IsLocked(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description    string
		ApplyLocked    bool
		ApplyLockError error
		ExpComment     string
		ExpFailStatus  bool
		ExpHasErrors   bool
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
			Description:    "If ApplyLockChecker returns an error apply is rejected",
			ApplyLockError: errors.New("error"),
			ApplyLocked:    false,
			ExpComment:     "**Error:** Failed to check global apply lock. Running `atlantis apply` is not allowed until the lock backend is reachable.",
			ExpFailStatus:  true,
			ExpHasErrors:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t, func(tc *TestConfig) {
				tc.applyLockCheckerReturn = locking.ApplyCommandLock{Locked: c.ApplyLocked}
				tc.applyLockCheckerErr = c.ApplyLockError
			})

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")

			pull := &github.PullRequest{
				State: github.Ptr("open"),
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

			applyCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Apply})
			Equals(t, c.ExpHasErrors, ctx.CommandHasErrors)

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(c.ExpComment), Eq("apply"))
			if c.ExpFailStatus {
				commitUpdater.VerifyWasCalledOnce().UpdateCombined(
					Any[logging.SimpleLogging](),
					Eq(testdata.GithubRepo),
					Eq(modelPull),
					Eq(models.FailedCommitStatus),
					Eq(command.Apply),
				)
			} else {
				commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
					Any[logging.SimpleLogging](),
					Any[models.Repo](),
					Any[models.PullRequest](),
					Any[models.CommitStatus](),
					Eq(command.Apply),
				)
			}
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
		PrevApplyStored   bool // stores a 1/1 passing apply in the database
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
					Eq(models.ProjectCounts{Success: c.ExpVCSStatusSucc, Total: c.ExpVCSStatusTotal}),
				)
			} else {
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

func TestApplyCommandRunner_IgnoredTargetedDirNoOp(t *testing.T) {
	cases := []struct {
		description string
		setup       func(*TestConfig)
	}{
		{
			description: "global lock backend errors",
			setup: func(tc *TestConfig) {
				tc.applyLockCheckerErr = errors.New("lock backend down")
			},
		},
		{
			description: "global apply lock is active",
			setup: func(tc *TestConfig) {
				tc.applyLockCheckerReturn = locking.ApplyCommandLock{Locked: true}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			vcsClient := setup(t, c.setup)
			scopeNull := metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis")
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			cmd := &events.CommentCommand{Name: command.Apply, RepoRelDir: "ignored"}
			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{}, events.ErrIgnoredTargetedDir)

			applyCommandRunner.Run(ctx, cmd)
			Assert(t, ctx.CommandSkipped, "expected ignored targeted dir to mark the command skipped")

			vcsClient.VerifyWasCalled(Never()).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
			commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
			commitUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
			projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
		})
	}
}

func TestApplyCommandRunner_ExecutionOrder(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description           string
		ProjectContexts       []command.ProjectContext
		ProjectCommandOutputs []command.ProjectCommandOutput
		RunnerInvokeMatch     []*EqMatcher
		ExpComment            string
		ApplyFailed           bool
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					ApplySuccess: "Great success!",
				},
				{
					Error: errors.New("shabang"),
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. project: `First` dir: `` workspace: ``\n1. project: `Second` dir: `` workspace: ``\n---\n\n### 1. project: `First` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### 2. project: `Second` dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},
				{
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Never(),
			},
			ApplyFailed: true,
			ExpComment:  "Ran Apply for project: `First` dir: `` workspace: ``\n\n**Apply Error**\n```\nshabang\n```",
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					ApplySuccess: "Great success!",
				},
				{
					Error: errors.New("shabang"),
				},
				{
					ApplySuccess: "Great success!",
				},
				{
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
				"1. project: `First` dir: `` workspace: ``\n1. project: `Second` dir: `` workspace: ``\n---\n\n### 1. project: `First` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### 2. project: `Second` dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					ApplySuccess: "Great success!",
				},
				{
					ApplySuccess: "Great success!",
				},
				{
					Error: errors.New("shabang"),
				},
				{
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
				"1. project: `First` dir: `` workspace: ``\n1. project: `Second` dir: `` workspace: ``\n1. project: `Third` dir: `` workspace: ``\n1. project: `Fourth` dir: `` workspace: ``\n---\n\n### 1. project: `First` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### 2. project: `Second` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### 3. project: `Third` dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### 4. project: `Fourth` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n4 projects, 3 successful, 0 failed, 1 errored",
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},
				{
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. project: `First` dir: `` workspace: ``\n1. project: `Second` dir: `` workspace: ``\n---\n\n### 1. project: `First` dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### 2. project: `Second` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					Error: errors.New("shabang"),
				},
				{
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: true,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. project: `First` dir: `` workspace: ``\n1. project: `Second` dir: `` workspace: ``\n---\n\n### 1. project: `First` dir: `` workspace: ``\n**Apply Error**\n```\nshabang\n```\n\n---\n### 2. project: `Second` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 1 successful, 0 failed, 1 errored",
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
			ProjectCommandOutputs: []command.ProjectCommandOutput{
				{
					ApplySuccess: "Great success!",
				},
				{
					ApplySuccess: "Great success!",
				},
			},
			RunnerInvokeMatch: []*EqMatcher{
				Once(),
				Once(),
			},
			ApplyFailed: false,
			ExpComment: "Ran Apply for 2 projects:\n\n" +
				"1. project: `First` dir: `` workspace: ``\n1. project: `Second` dir: `` workspace: ``\n---\n\n### 1. project: `First` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### 2. project: `Second` dir: `` workspace: ``\n```diff\nGreat success!\n```\n\n---\n### Apply Summary\n\n2 projects, 2 successful, 0 failed, 0 errored",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")

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

			When(githubGetter.GetPullRequest(logger, testdata.GithubRepo, testdata.Pull.Num)).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(logger, pull)).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).ThenReturn(c.ProjectContexts, nil)
			for i := range c.ProjectContexts {
				When(projectCommandRunner.Apply(c.ProjectContexts[i])).ThenReturn(c.ProjectCommandOutputs[i])
			}

			applyCommandRunner.Run(ctx, cmd)

			for i := range c.ProjectContexts {
				projectCommandRunner.VerifyWasCalled(c.RunnerInvokeMatch[i]).Apply(c.ProjectContexts[i])
			}

			require.Equal(t, c.ApplyFailed, ctx.CommandHasErrors)

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(c.ExpComment), Eq("apply"),
			)
		})
	}
}

func TestApplyCommandRunner_NoChangesCount(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() { db.Close() })
	Ok(t, err)

	setup(t, func(tc *TestConfig) {
		tc.SilenceNoProjects = true
		tc.database = db
	})

	scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")
	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

	// Seed the DB: one applied project and one project that planned with no changes.
	_, err = db.UpdatePullWithResults(modelPull, []command.ProjectResult{
		{
			Command:    command.Apply,
			RepoRelDir: "dir1",
			Workspace:  "default",
		},
		{
			Command:    command.Plan,
			RepoRelDir: "dir2",
			Workspace:  "default",
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "No changes. Infrastructure is up-to-date.",
				},
			},
		},
	})
	Ok(t, err)

	pull := &github.PullRequest{State: github.Ptr("open")}
	When(githubGetter.GetPullRequest(logger, testdata.GithubRepo, testdata.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(logger, pull)).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

	// Target a dir that has no new projects so the runner re-reads the existing pull status.
	cmd := &events.CommentCommand{Name: command.Apply, RepoRelDir: "nomatch"}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    scopeNull,
		Pull:     modelPull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
	When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).ThenReturn([]command.ProjectContext{}, nil)

	applyCommandRunner.Run(ctx, cmd)

	// Verify that NoChanges=1 is passed correctly (not Any[int]()).
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[models.CommitStatus](),
		Eq[command.Name](command.Apply),
		Eq(models.ProjectCounts{Success: 2, Total: 2, NoChanges: 1}),
	)
}
