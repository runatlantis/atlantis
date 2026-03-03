// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v83/github"
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
	"go.uber.org/mock/gomock"
)

func TestApplyCommandRunner_IsLocked(t *testing.T) {
	logger := logging.NewNoopLogger(t)

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
			vcsClient := setup(t, func(tc *TestConfig) {
				tc.applyLockCheckerReturn = locking.ApplyCommandLock{Locked: c.ApplyLocked}
				tc.applyLockCheckerErr = c.ApplyLockError
			})

			scopeNull := metricstest.NewLoggingScope(t, logger, "atlantis")

			pull := &github.PullRequest{
				State: github.Ptr("open"),
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			// Only set specific expectations for github getter/parser when not locked,
			// since locked cases return early before calling these.
			// The AnyTimes() defaults from setup() handle the locked case.
			if !c.ApplyLocked {
				githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(pull, nil).AnyTimes()
				eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Any()).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil).AnyTimes()
			}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			applyCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Apply})

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(c.ExpComment), Eq("apply"))
		})
	}
}

func TestApplyCommandRunner_IsSilenced(t *testing.T) {
	logger := logging.NewNoopLogger(t)

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

			matched := c.Matched
			projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ *command.Context, _ *events.CommentCommand) ([]command.ProjectContext, error) {
					if matched {
						return []command.ProjectContext{{
							CommandName:       command.Apply,
							ProjectPlanStatus: models.PlannedPlanStatus,
						}}, nil
					}
					return []command.ProjectContext{}, nil
				})

			if c.ExpVCSStatusSet {
				commitUpdater.EXPECT().UpdateCombinedCount(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Eq(models.SuccessCommitStatus),
					gomock.Eq(command.Apply),
					gomock.Eq(c.ExpVCSStatusSucc),
					gomock.Eq(c.ExpVCSStatusTotal),
				).Return(nil).Times(1)
			} else {
				commitUpdater.EXPECT().UpdateCombinedCount(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Eq(command.Apply),
					gomock.Any(),
					gomock.Any(),
				).Return(nil).Times(0)
			}

			applyCommandRunner.Run(ctx, cmd)

			timesComment := 1
			if c.ExpSilenced {
				timesComment = 0
			}

			vcsClient.VerifyWasCalled(Times(timesComment)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
		})
	}
}

func TestApplyCommandRunner_ExecutionOrder(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		Description           string
		ProjectContexts       []command.ProjectContext
		ProjectCommandOutputs []command.ProjectCommandOutput
		RunnerInvokeTimes     []int
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
			RunnerInvokeTimes: []int{1, 1},
			ApplyFailed:       true,
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
			RunnerInvokeTimes: []int{1, 0},
			ApplyFailed:       true,
			ExpComment:        "Ran Apply for project: `First` dir: `` workspace: ``\n\n**Apply Error**\n```\nshabang\n```",
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
			RunnerInvokeTimes: []int{1, 1, 0, 0},
			ApplyFailed:       true,
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
			RunnerInvokeTimes: []int{1, 1, 1, 1},
			ApplyFailed:       true,
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
			RunnerInvokeTimes: []int{1, 1},
			ApplyFailed:       true,
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
			RunnerInvokeTimes: []int{1, 1},
			ApplyFailed:       true,
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
			RunnerInvokeTimes: []int{1, 1},
			ApplyFailed:       false,
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

			githubGetter.EXPECT().GetPullRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(pull, nil).AnyTimes()
			eventParsing.EXPECT().ParseGithubPull(gomock.Any(), gomock.Any()).Return(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil).AnyTimes()

			projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).Return(c.ProjectContexts, nil)
			applyCallIndex := 0
			outputs := c.ProjectCommandOutputs
			projectCommandRunner.EXPECT().Apply(gomock.Any()).DoAndReturn(func(_ command.ProjectContext) command.ProjectCommandOutput {
				idx := applyCallIndex
				applyCallIndex++
				if idx < len(outputs) {
					return outputs[idx]
				}
				return command.ProjectCommandOutput{}
			}).AnyTimes()

			applyCommandRunner.Run(ctx, cmd)

			require.Equal(t, c.ApplyFailed, ctx.CommandHasErrors)

			vcsClient.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(c.ExpComment), Eq("apply"),
			)
		})
	}
}
