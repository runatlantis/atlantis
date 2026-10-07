// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"testing"
	"time"

	"github.com/google/go-github/v92/github"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
)

// fakePullStatusFetcher returns the pull status stored for each pull number.
type fakePullStatusFetcher map[int]*models.PullStatus

func (f fakePullStatusFetcher) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	return f[pull.Num], nil
}

func TestDefaultStackedPullPlanner_PlanNextPull(t *testing.T) {
	merged := models.PullRequest{Num: 1, BaseRepo: testdata.GithubRepo}
	nextPull := models.PullRequest{
		Num:        2,
		HeadCommit: "sha2",
		Author:     "author",
		State:      models.OpenPullState,
		BaseRepo:   testdata.GithubRepo,
	}

	cases := map[string]struct {
		// stacks are the stacks returned by each fetch.
		stacks        [][]models.StackedPull
		allowDraftPRs bool
		statuses      fakePullStatusFetcher
		expPlanned    bool
	}{
		"plans the next pull request": {
			stacks: [][]models.StackedPull{{
				{Num: 1, Merged: true},
				{Num: 2, Open: true},
				{Num: 3, Open: true},
			}},
			expPlanned: true,
		},
		"skips closed pull requests above the merged one": {
			stacks: [][]models.StackedPull{{
				{Num: 1, Merged: true},
				{Num: 4},
				{Num: 2, Open: true},
			}},
			expPlanned: true,
		},
		"retries while GitHub still reports the merged pull request as open": {
			stacks: [][]models.StackedPull{
				{{Num: 1, Open: true}, {Num: 2, Open: true}},
				{{Num: 1, Merged: true}, {Num: 2, Open: true}},
			},
			expPlanned: true,
		},
		"does not plan when a pull request below is still open": {
			stacks: [][]models.StackedPull{{
				{Num: 5, Open: true},
				{Num: 1, Merged: true},
				{Num: 2, Open: true},
			}},
		},
		"does not plan when the merged pull request is the top of the stack": {
			stacks: [][]models.StackedPull{{
				{Num: 6, Merged: true},
				{Num: 1, Merged: true},
			}},
		},
		"does not plan when the pull request is not stacked": {
			stacks: [][]models.StackedPull{nil},
		},
		"does not plan draft pull requests": {
			stacks: [][]models.StackedPull{{
				{Num: 1, Merged: true},
				{Num: 2, Open: true, Draft: true},
			}},
		},
		"plans draft pull requests when allowed": {
			stacks: [][]models.StackedPull{{
				{Num: 1, Merged: true},
				{Num: 2, Open: true, Draft: true},
			}},
			allowDraftPRs: true,
			expPlanned:    true,
		},
		"does not plan again when already planned at the head commit": {
			stacks: [][]models.StackedPull{{
				{Num: 1, Merged: true},
				{Num: 2, Open: true},
			}},
			statuses: fakePullStatusFetcher{2: {
				Pull:     nextPull,
				Projects: []models.ProjectStatus{{Status: models.PlannedPlanStatus}},
			}},
		},
		"plans again when the previous plan errored": {
			stacks: [][]models.StackedPull{{
				{Num: 1, Merged: true},
				{Num: 2, Open: true},
			}},
			statuses: fakePullStatusFetcher{2: {
				Pull:     nextPull,
				Projects: []models.ProjectStatus{{Status: models.ErroredPlanStatus}},
			}},
			expPlanned: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			logger := logging.NewNoopLogger(t)
			stackGetter := mocks.NewMockGithubStackGetter()
			pullGetter := mocks.NewMockGithubPullGetter()
			eventParser := mocks.NewMockEventParsing()
			commandRunner := mocks.NewMockCommandRunner()

			stub := When(stackGetter.GetPullRequestStack(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(1)))
			for _, stack := range c.stacks {
				stub = stub.ThenReturn(stack, nil)
			}
			ghPull := &github.PullRequest{Number: new(2)}
			When(pullGetter.GetPullRequest(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(2))).ThenReturn(ghPull, nil)
			When(eventParser.ParseGithubPull(Any[logging.SimpleLogging](), Eq(ghPull))).ThenReturn(nextPull, testdata.GithubRepo, testdata.GithubRepo, nil)

			statuses := c.statuses
			if statuses == nil {
				statuses = fakePullStatusFetcher{}
			}
			planner := &events.DefaultStackedPullPlanner{
				GithubPullGetter:  pullGetter,
				GithubStackGetter: stackGetter,
				EventParser:       eventParser,
				CommandRunner:     commandRunner,
				PullStatusFetcher: statuses,
				AllowDraftPRs:     c.allowDraftPRs,
				StackRetryDelay:   time.Millisecond,
			}
			planner.PlanNextPull(logger, testdata.GithubRepo, merged)

			times := Never()
			if c.expPlanned {
				times = Once()
			}
			commandRunner.VerifyWasCalled(times).RunAutoplanCommand(Eq(testdata.GithubRepo), Eq(testdata.GithubRepo), Eq(nextPull), Eq(models.User{Username: "author"}))
			stackGetter.VerifyWasCalled(Times(len(c.stacks))).GetPullRequestStack(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(1))
		})
	}
}
