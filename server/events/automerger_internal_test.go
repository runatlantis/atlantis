// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
)

// newAutomergeTestContext returns a command.Context sufficient for exercising
// AutoMerger.automerge with a pull request that has a single applied project.
func newAutomergeTestContext(t *testing.T) (*command.Context, models.PullStatus) {
	ctx := &command.Context{
		Log: logging.NewNoopLogger(t),
		Pull: models.PullRequest{
			Num:      1,
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}
	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{RepoRelDir: ".", Workspace: "default", Status: models.AppliedPlanStatus},
		},
	}
	return ctx, pullStatus
}

// fastRetryAutoMerger builds an AutoMerger with negligible backoff so retry
// tests don't sleep for whole seconds.
func fastRetryAutoMerger(vcsClient *vcsmocks.MockClient, retryCount int) *AutoMerger {
	return &AutoMerger{
		VCSClient:           vcsClient,
		AutomergeRetryCount: retryCount,
		retryBackoffMin:     time.Millisecond,
		retryBackoffMax:     time.Millisecond,
	}
}

func anyMergePullArgs() (logging.SimpleLogging, models.PullRequest, models.PullRequestOptions) {
	return Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]()
}

// TestAutoMerger_automerge_RetriesUntilSuccess verifies that a transient merge
// failure is retried and, once the merge succeeds, no failure comment is posted.
func TestAutoMerger_automerge_RetriesUntilSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	// Fail twice, then succeed on the third attempt.
	When(vcsClient.MergePull(anyMergePullArgs())).
		ThenReturn(errors.New("required status checks are expected")).
		ThenReturn(errors.New("required status checks are expected")).
		ThenReturn(nil)

	am := fastRetryAutoMerger(vcsClient, 3)
	ctx, pullStatus := newAutomergeTestContext(t)

	am.automerge(ctx, pullStatus, false, "")

	l, p, o := anyMergePullArgs()
	vcsClient.VerifyWasCalled(Times(3)).MergePull(l, p, o)
	// Only the initial "automerging" comment should be posted, not a failure one.
	cl, cr, cn, cc, ck := anyCreateCommentArgs()
	vcsClient.VerifyWasCalled(Once()).CreateComment(cl, cr, cn, cc, ck)
}

// TestAutoMerger_automerge_RetriesExhausted verifies that once retries are
// exhausted a failure comment is posted and MergePull is called exactly
// AutomergeRetryCount+1 times.
func TestAutoMerger_automerge_RetriesExhausted(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.MergePull(anyMergePullArgs())).ThenReturn(errors.New("boom"))

	am := fastRetryAutoMerger(vcsClient, 2)
	ctx, pullStatus := newAutomergeTestContext(t)

	am.automerge(ctx, pullStatus, false, "")

	l, p, o := anyMergePullArgs()
	vcsClient.VerifyWasCalled(Times(3)).MergePull(l, p, o)
	// Initial "automerging" comment plus a failure comment.
	cl, cr, cn, cc, ck := anyCreateCommentArgs()
	vcsClient.VerifyWasCalled(Times(2)).CreateComment(cl, cr, cn, cc, ck)
}

// TestAutoMerger_automerge_NoRetryByDefault verifies that with the default
// retry count of 0 the merge is attempted exactly once, preserving the original
// behavior.
func TestAutoMerger_automerge_NoRetryByDefault(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.MergePull(anyMergePullArgs())).ThenReturn(errors.New("boom"))

	am := &AutoMerger{VCSClient: vcsClient}
	ctx, pullStatus := newAutomergeTestContext(t)

	am.automerge(ctx, pullStatus, false, "")

	l, p, o := anyMergePullArgs()
	vcsClient.VerifyWasCalled(Once()).MergePull(l, p, o)
}

func anyCreateCommentArgs() (logging.SimpleLogging, models.Repo, int, string, string) {
	return Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()
}

// TestAutoMerger_retryBackoff verifies the delay grows exponentially, stays
// within [min, cap] once jitter is applied, and never exceeds the maximum.
func TestAutoMerger_retryBackoff(t *testing.T) {
	am := &AutoMerger{
		retryBackoffMin: 2 * time.Second,
		retryBackoffMax: 30 * time.Second,
	}

	// Uncapped exponential ceilings per attempt: 2s, 4s, 8s, ... capped at 30s.
	caps := map[int]time.Duration{
		1: 2 * time.Second,
		2: 4 * time.Second,
		3: 8 * time.Second,
		4: 16 * time.Second,
		5: 30 * time.Second,
		6: 30 * time.Second,
	}
	// Sample repeatedly since jitter is random.
	for attempt, ceiling := range caps {
		for range 100 {
			got := am.retryBackoff(attempt)
			if got < am.retryBackoffMin {
				t.Fatalf("attempt %d: backoff %s below min %s", attempt, got, am.retryBackoffMin)
			}
			if got > ceiling {
				t.Fatalf("attempt %d: backoff %s above expected ceiling %s", attempt, got, ceiling)
			}
		}
	}
}

// TestAutoMerger_retryBackoff_Defaults verifies the package defaults are used
// when no overrides are set and that a min greater than max is clamped.
func TestAutoMerger_retryBackoff_Defaults(t *testing.T) {
	am := &AutoMerger{}
	if got := am.retryBackoff(1); got != defaultAutomergeRetryMinBackoff {
		t.Fatalf("expected first backoff to equal default min %s, got %s", defaultAutomergeRetryMinBackoff, got)
	}

	// max below min should be clamped up to min, yielding a fixed delay.
	clamped := &AutoMerger{retryBackoffMin: 5 * time.Second, retryBackoffMax: time.Second}
	if got := clamped.retryBackoff(3); got != 5*time.Second {
		t.Fatalf("expected clamped backoff of 5s, got %s", got)
	}
}
