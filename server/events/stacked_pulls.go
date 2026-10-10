// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_github_stack_getter.go GithubStackGetter

// GithubStackGetter fetches the stack a GitHub pull request belongs to.
type GithubStackGetter interface {
	// GetPullRequestStack returns the pull requests of the stack that pull
	// request pullNum belongs to, ordered from the bottom of the stack to the
	// top. It returns nil if the pull request is not part of a stack.
	GetPullRequestStack(logger logging.SimpleLogging, repo models.Repo, pullNum int) ([]models.StackedPull, error)
}

// pullsBelowInStack returns the pull requests below pull in stack, ordered
// from the bottom of the stack up.
func pullsBelowInStack(stack []models.StackedPull, pull models.PullRequest) ([]models.StackedPull, error) {
	for i, p := range stack {
		if p.Num != pull.Num {
			continue
		}
		// The order of the stack decides which pull requests GitHub merges
		// together, so refuse to guess if it disagrees with the position GitHub
		// reported for the pull request itself.
		if pull.Stack != nil && pull.Stack.Position != i+1 {
			return nil, fmt.Errorf("pull request #%d is at position %d of its stack but was listed at position %d", pull.Num, pull.Stack.Position, i+1)
		}
		return stack[:i], nil
	}
	return nil, fmt.Errorf("pull request #%d was not found in its stack", pull.Num)
}

// openStackedPulls returns the pull requests in pulls that are still open.
func openStackedPulls(pulls []models.StackedPull) []models.StackedPull {
	var open []models.StackedPull
	for _, p := range pulls {
		if p.Open {
			open = append(open, p)
		}
	}
	return open
}

// formatStackedPulls formats pulls as a list of pull request references, ex.
// "#1, #2".
func formatStackedPulls(pulls []models.StackedPull) string {
	refs := make([]string, len(pulls))
	for i, p := range pulls {
		refs[i] = fmt.Sprintf("#%d", p.Num)
	}
	return strings.Join(refs, ", ")
}

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_stacked_pull_planner.go StackedPullPlanner

// StackedPullPlanner plans the next pull request of a GitHub stack once the
// pull request below it has been merged.
type StackedPullPlanner interface {
	// PlanNextPull autoplans the pull request above merged in its stack, if
	// there is one and every pull request below it has been merged.
	PlanNextPull(logger logging.SimpleLogging, baseRepo models.Repo, merged models.PullRequest)
}

// DefaultStackedPullPlanner implements StackedPullPlanner.
type DefaultStackedPullPlanner struct {
	GithubPullGetter  GithubPullGetter
	GithubStackGetter GithubStackGetter
	EventParser       EventParsing
	CommandRunner     CommandRunner
	PullStatusFetcher PullStatusFetcher
	// AllowDraftPRs controls whether draft pull requests are autoplanned.
	AllowDraftPRs bool
	// StackRetryDelay is how long to wait before fetching the stack again
	// when GitHub still reports the merged pull request as open. Defaults to
	// one second and doubles with each attempt.
	StackRetryDelay time.Duration
}

const stackFetchAttempts = 4

func (p *DefaultStackedPullPlanner) PlanNextPull(logger logging.SimpleLogging, baseRepo models.Repo, merged models.PullRequest) {
	stack, err := p.fetchStackAfterMerge(logger, baseRepo, merged.Num)
	if err != nil {
		logger.Err("unable to fetch stack of merged pull request #%d: %s", merged.Num, err)
		return
	}
	if stack == nil {
		return
	}

	next, ok := nextStackedPullToPlan(stack, merged.Num)
	if !ok {
		logger.Debug("no pull request in the stack of #%d is ready to be planned", merged.Num)
		return
	}
	if next.Draft && !p.AllowDraftPRs {
		logger.Info("not planning draft pull request #%d now that #%d below it in the stack is merged", next.Num, merged.Num)
		return
	}

	ghPull, err := p.GithubPullGetter.GetPullRequest(logger, baseRepo, next.Num)
	if err != nil {
		logger.Err("unable to get pull request #%d: %s", next.Num, err)
		return
	}
	pull, pullBaseRepo, headRepo, err := p.EventParser.ParseGithubPull(logger, ghPull)
	if err != nil {
		logger.Err("unable to parse pull request #%d: %s", next.Num, err)
		return
	}
	if pull.State != models.OpenPullState {
		return
	}

	// GitHub rebases the rest of the stack when a pull request below is
	// merged, which can already have triggered an autoplan of this commit.
	status, err := p.PullStatusFetcher.GetPullStatus(pull)
	if err != nil {
		logger.Err("unable to fetch pull status of #%d: %s", pull.Num, err)
	} else if isPlannedAtCommit(status, pull.HeadCommit) {
		logger.Info("pull request #%d is already planned at %s, not planning it again", pull.Num, pull.HeadCommit)
		return
	}

	logger.Info("planning pull request #%d now that #%d below it in the stack is merged", pull.Num, merged.Num)
	p.CommandRunner.RunAutoplanCommand(pullBaseRepo, headRepo, pull, models.User{Username: pull.Author})
}

// fetchStackAfterMerge fetches the stack of the merged pull request num,
// retrying while GitHub still reports it as open.
func (p *DefaultStackedPullPlanner) fetchStackAfterMerge(logger logging.SimpleLogging, repo models.Repo, num int) ([]models.StackedPull, error) {
	delay := p.StackRetryDelay
	if delay <= 0 {
		delay = time.Second
	}
	var stack []models.StackedPull
	var err error
	for attempt := range stackFetchAttempts {
		if attempt > 0 {
			time.Sleep(delay)
			delay *= 2
		}
		stack, err = p.GithubStackGetter.GetPullRequestStack(logger, repo, num)
		if err != nil || !stackedPullIsOpen(stack, num) {
			return stack, err
		}
	}
	return nil, fmt.Errorf("GitHub still reports pull request #%d as open", num)
}

func stackedPullIsOpen(stack []models.StackedPull, num int) bool {
	for _, p := range stack {
		if p.Num == num {
			return p.Open
		}
	}
	return false
}

// nextStackedPullToPlan returns the first open pull request above merged in
// stack, if every pull request below it has been closed.
func nextStackedPullToPlan(stack []models.StackedPull, merged int) (models.StackedPull, bool) {
	above := false
	for i, p := range stack {
		if p.Num == merged {
			above = true
			continue
		}
		if !above || !p.Open {
			continue
		}
		if len(openStackedPulls(stack[:i])) > 0 {
			return models.StackedPull{}, false
		}
		return p, true
	}
	return models.StackedPull{}, false
}

// isPlannedAtCommit returns true if every project of status was planned
// successfully (or applied) at commit.
func isPlannedAtCommit(status *models.PullStatus, commit string) bool {
	if status == nil || status.Pull.HeadCommit != commit || len(status.Projects) == 0 {
		return false
	}
	for _, p := range status.Projects {
		if p.Status != models.PlannedPlanStatus && p.Status != models.PlannedNoChangesPlanStatus && p.Status != models.AppliedPlanStatus {
			return false
		}
	}
	return true
}
