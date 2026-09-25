// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

// Default backoff bounds used between automerge retries when the AutoMerger
// does not override them (see retryBackoffMin/retryBackoffMax).
const (
	defaultAutomergeRetryMinBackoff = 2 * time.Second
	defaultAutomergeRetryMaxBackoff = 30 * time.Second
)

type AutoMerger struct {
	VCSClient             vcs.Client
	GlobalAutomerge       bool
	GlobalAutomergeMethod string
	// AutomergeRetryCount is the number of additional times a failed merge is
	// retried before giving up. Zero (the default) means the merge is attempted
	// exactly once, preserving the original behavior.
	AutomergeRetryCount int
	// retryBackoffMin and retryBackoffMax bound the wait between merge retries.
	// When zero the defaults above are used; they exist mainly so tests can use
	// short waits.
	retryBackoffMin time.Duration
	retryBackoffMax time.Duration
}

func (c *AutoMerger) automerge(ctx *command.Context, pullStatus models.PullStatus, deleteSourceBranchOnMerge bool, mergeMethod string) {
	// We only automerge if all projects have been successfully applied.
	for _, p := range pullStatus.Projects {
		if p.Status != models.AppliedPlanStatus {
			ctx.Log.Info("not automerging because project at dir %q, workspace %q has status %q", p.RepoRelDir, p.Workspace, p.Status.String())
			return
		}
	}

	// Comment that we're automerging the pull request.
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, automergeComment, command.Apply.String()); err != nil {
		ctx.Log.Err("failed to comment about automerge: %s", err)
		// Commenting isn't required so continue.
	}

	// Fall back to the server-side default merge method when the comment
	// command didn't specify one with --auto-merge-method.
	if mergeMethod == "" {
		mergeMethod = c.GlobalAutomergeMethod
	}

	// Make the API call to perform the merge, retrying on failure if configured.
	ctx.Log.Info("automerging pull request")
	var pullOptions models.PullRequestOptions
	pullOptions.DeleteSourceBranchOnMerge = deleteSourceBranchOnMerge
	pullOptions.MergeMethod = mergeMethod
	err := c.mergeWithRetry(ctx, pullOptions)

	if err != nil {
		ctx.Log.Err("automerging failed: %s", err)

		failureComment := fmt.Sprintf("Automerging failed:\n```\n%s\n```", err)
		if commentErr := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, failureComment, command.Apply.String()); commentErr != nil {
			ctx.Log.Err("failed to comment about automerge failing: %s", err)
		}
	}
}

// mergeWithRetry attempts to merge the pull request, retrying up to
// AutomergeRetryCount additional times on failure with an exponential backoff.
//
// Some VCS providers can briefly report a pull request as unmergeable right
// after Atlantis sets its apply commit status to success: with GitHub branch
// protection or repository rulesets the required status checks have not always
// propagated by the time the merge API call is made, so the merge is rejected
// (e.g. "required status checks are expected"). Such failures are transient and
// clear within a few seconds, so retrying works around them. When
// AutomergeRetryCount is 0 (the default) the merge is attempted exactly once and
// this behaves identically to a single MergePull call.
func (c *AutoMerger) mergeWithRetry(ctx *command.Context, pullOptions models.PullRequestOptions) error {
	maxAttempts := c.AutomergeRetryCount + 1
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = c.VCSClient.MergePull(ctx.Log, ctx.Pull, pullOptions); err == nil {
			return nil
		}
		if attempt == maxAttempts {
			break
		}
		sleep := c.retryBackoff(attempt)
		ctx.Log.Warn("automerge attempt %d/%d failed: %s; retrying in %s", attempt, maxAttempts, err, sleep)
		time.Sleep(sleep)
	}
	return err
}

// retryBackoff returns how long to wait before the given 1-indexed retry
// attempt. The delay starts at the minimum bound, doubles with each attempt,
// and is capped at the maximum bound. Full jitter is then applied (a random
// duration in [min, delay]) so repeated retries don't fire in lockstep.
func (c *AutoMerger) retryBackoff(attempt int) time.Duration {
	minBackoff, maxBackoff := c.backoffBounds()

	delay := minBackoff
	for i := 1; i < attempt && delay < maxBackoff; i++ {
		delay *= 2
	}
	if delay > maxBackoff {
		delay = maxBackoff
	}

	if delay > minBackoff {
		delay = minBackoff + rand.N(delay-minBackoff+1)
	}
	return delay
}

// backoffBounds returns the min and max wait between merge retries, applying
// the package defaults when the (test-only) overrides are unset and ensuring
// max is never below min.
func (c *AutoMerger) backoffBounds() (time.Duration, time.Duration) {
	minBackoff := c.retryBackoffMin
	if minBackoff <= 0 {
		minBackoff = defaultAutomergeRetryMinBackoff
	}
	maxBackoff := c.retryBackoffMax
	if maxBackoff <= 0 {
		maxBackoff = defaultAutomergeRetryMaxBackoff
	}
	if maxBackoff < minBackoff {
		maxBackoff = minBackoff
	}
	return minBackoff, maxBackoff
}

// automergeEnabled returns true if automerging is enabled in this context.
func (c *AutoMerger) automergeEnabled(projectCmds []command.ProjectContext) bool {
	// Use project automerge settings if projects exist; otherwise, use global automerge settings.
	automerge := c.GlobalAutomerge
	if len(projectCmds) > 0 {
		automerge = projectCmds[0].AutomergeEnabled
	}
	return automerge
}

// deleteSourceBranchOnMergeEnabled returns true if we should delete the source branch on merge in this context.
func (c *AutoMerger) deleteSourceBranchOnMergeEnabled(projectCmds []command.ProjectContext) bool {
	//check if this repo is configured for automerging.
	return (len(projectCmds) > 0 && projectCmds[0].DeleteSourceBranchOnMerge)
}
