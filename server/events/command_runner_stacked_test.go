// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v92/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
)

func TestRunAutoplanCommand_StackAwarePlanning(t *testing.T) {
	stackedPull := models.PullRequest{
		Num:        2,
		BaseRepo:   testdata.GithubRepo,
		BaseBranch: "feature-1",
		Stack:      &models.PullRequestStack{Number: 1, Position: 2, Size: 3, BaseBranch: "main"},
	}
	stack := []models.StackedPull{
		{Num: 1, Open: true},
		{Num: 2, Open: true},
		{Num: 3, Open: true},
	}

	cases := map[string]struct {
		disabled    bool
		pull        func(models.PullRequest) models.PullRequest
		stack       []models.StackedPull
		stackErr    error
		expDeferred bool
	}{
		"defers autoplan while a pull request below is open": {
			stack:       stack,
			expDeferred: true,
		},
		"plans when the pull requests below are merged": {
			stack: []models.StackedPull{
				{Num: 1, Merged: true},
				{Num: 2, Open: true},
			},
		},
		"plans when the pull request targets the base of the stack": {
			pull: func(p models.PullRequest) models.PullRequest {
				p.BaseBranch = "main"
				return p
			},
			stack: stack,
		},
		"plans the bottom of the stack": {
			pull: func(p models.PullRequest) models.PullRequest {
				p.Stack = &models.PullRequestStack{Number: 1, Position: 1, Size: 3, BaseBranch: "main"}
				return p
			},
			stack: stack,
		},
		"plans when the stack can't be fetched": {
			stackErr: errors.New("err"),
		},
		"plans when the stack doesn't match the position of the pull request": {
			stack: []models.StackedPull{
				{Num: 2, Open: true},
				{Num: 1, Open: true},
			},
		},
		"plans when stack aware planning is disabled": {
			disabled: true,
			stack:    stack,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			vcsClient := setup(t)
			stackGetter := mocks.NewMockGithubStackGetter()
			ch.StackAwarePlanning = !c.disabled
			ch.GithubStackGetter = stackGetter
			t.Cleanup(func() {
				ch.StackAwarePlanning = false
				ch.GithubStackGetter = nil
			})

			pull := stackedPull
			if c.pull != nil {
				pull = c.pull(pull)
			}
			When(stackGetter.GetPullRequestStack(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(2))).ThenReturn(c.stack, c.stackErr)
			When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).
				ThenReturn([]command.ProjectContext{{CommandName: command.Plan}}, nil)

			ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, pull, testdata.User)

			deferredComment := "Autoplan deferred: this pull request is 2 of 3 in its stack and will be planned automatically" +
				" once the pull requests below it are merged (waiting on #1). Run `atlantis plan` to plan it now."
			if c.expDeferred {
				projectCommandBuilder.VerifyWasCalled(Never()).BuildAutoplanCommands(Any[*command.Context]())
				vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(2), Eq(deferredComment), Eq("plan"))
			} else {
				projectCommandBuilder.VerifyWasCalledOnce().BuildAutoplanCommands(Any[*command.Context]())
				vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(2), Eq(deferredComment), Eq("plan"))
			}
		})
	}
}

func TestApplyWithAutoMerge_Stacked(t *testing.T) {
	cases := map[string]struct {
		stack       []models.StackedPull
		statuses    fakePullStatusFetcher
		expMerge    bool
		expComments []string
	}{
		"merges the bottom of the stack": {
			stack: []models.StackedPull{
				{Num: 1, HeadCommit: "sha1", Merged: true},
				{Num: 0, Open: true},
			},
			expMerge:    true,
			expComments: []string{"Automatically merging because all plans have been successfully applied."},
		},
		"merges together with applied pull requests below": {
			stack: []models.StackedPull{
				{Num: 1, HeadCommit: "sha1", Open: true},
				{Num: 0, Open: true},
			},
			statuses: fakePullStatusFetcher{1: {
				Pull:     models.PullRequest{Num: 1, HeadCommit: "sha1"},
				Projects: []models.ProjectStatus{{Status: models.AppliedPlanStatus}},
			}},
			expMerge: true,
			expComments: []string{"Automatically merging because all plans have been successfully applied." +
				" GitHub will also merge #1 below it in the stack."},
		},
		"skips when a pull request below has not been applied": {
			stack: []models.StackedPull{
				{Num: 1, HeadCommit: "sha1", Open: true},
				{Num: 0, Open: true},
			},
			statuses: fakePullStatusFetcher{1: {
				Pull:     models.PullRequest{Num: 1, HeadCommit: "sha1"},
				Projects: []models.ProjectStatus{{Status: models.PlannedPlanStatus}},
			}},
			expComments: []string{"Automerge skipped: GitHub merges stacked pull requests together with the ones below them," +
				" and #1 below this one in the stack has not been applied"},
		},
		"skips when a pull request below was applied at an older commit": {
			stack: []models.StackedPull{
				{Num: 1, HeadCommit: "sha1-new", Open: true},
				{Num: 0, Open: true},
			},
			statuses: fakePullStatusFetcher{1: {
				Pull:     models.PullRequest{Num: 1, HeadCommit: "sha1"},
				Projects: []models.ProjectStatus{{Status: models.AppliedPlanStatus}},
			}},
			expComments: []string{"Automerge skipped: GitHub merges stacked pull requests together with the ones below them," +
				" and #1 below this one in the stack has not been applied"},
		},
		"skips when a pull request below is unknown to Atlantis": {
			stack: []models.StackedPull{
				{Num: 1, HeadCommit: "sha1", Open: true},
				{Num: 0, Open: true},
			},
			expComments: []string{"Automerge skipped: GitHub merges stacked pull requests together with the ones below them," +
				" and #1 below this one in the stack has not been applied"},
		},
		"skips when the stack doesn't match the position of the pull request": {
			stack: []models.StackedPull{
				{Num: 0, Open: true},
				{Num: 1, HeadCommit: "sha1", Merged: true},
			},
			expComments: []string{"Automerge skipped: pull request #0 is at position 2 of its stack but was listed at position 1"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			vcsClient, modelPull := setupApplyWithAutoMerge(t)
			stackGetter := mocks.NewMockGithubStackGetter()
			autoMerger.GithubStackGetter = stackGetter
			autoMerger.PullStatusFetcher = c.statuses
			t.Cleanup(func() {
				autoMerger.GithubStackGetter = nil
				autoMerger.PullStatusFetcher = nil
			})

			modelPull.Stack = &models.PullRequestStack{Number: 1, Position: 2, Size: 2, BaseBranch: "main"}
			When(eventParsing.ParseGithubPull(Any[logging.SimpleLogging](), Any[*github.PullRequest]())).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)
			When(stackGetter.GetPullRequestStack(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num))).ThenReturn(c.stack, nil)

			ch.RunCommentCommand(testdata.GithubRepo, &testdata.GithubRepo, nil, testdata.User, testdata.Pull.Num, &events.CommentCommand{Name: command.Apply})

			times := Never()
			if c.expMerge {
				times = Once()
			}
			vcsClient.VerifyWasCalled(times).MergePull(Any[logging.SimpleLogging](), Eq(modelPull), Any[models.PullRequestOptions]())
			for _, comment := range c.expComments {
				vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(modelPull.Num), Eq(comment), Eq("apply"))
			}
		})
	}
}
