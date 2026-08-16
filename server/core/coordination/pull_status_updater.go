// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package coordination

import (
	"errors"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// PullStatusUpdater applies write policy on top of PullStatusStore:
// stale or transient results are filtered out before they can overwrite
// fresher recorded status.
type PullStatusUpdater struct {
	Store PullStatusStore
}

// Update merges results into the pull's recorded status. Results from a
// stale head, from directories that no longer exist, or marked with
// command.ErrStaleCommandHead are dropped.
func (c *PullStatusUpdater) Update(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if len(results) == 0 && pull.HeadCommit != "" {
		pullStatus, err := c.Store.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil && !models.PullStatusFreshForPull(pull, pullStatus.Pull) {
			ctx.Log.Debug("ignoring empty result from pull head %q base %q because current plan status is for head %q base %q", pull.HeadCommit, pull.BaseBranch, pullStatus.Pull.HeadCommit, pullStatus.Pull.BaseBranch)
			return *pullStatus, nil
		}
	}

	if pull.HeadCommit != "" && command.HasApplyResult(results) {
		pullStatus, err := c.Store.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil && !models.PullStatusFreshForPull(pull, pullStatus.Pull) {
			ctx.Log.Debug("ignoring stale apply result from pull head %q base %q because current plan status is for head %q base %q", pull.HeadCommit, pull.BaseBranch, pullStatus.Pull.HeadCommit, pullStatus.Pull.BaseBranch)
			return *pullStatus, nil
		}
	}

	// Filter out results that errored due to the directory not existing. We
	// don't store these because they would never be "apply-able" and so the
	// pull request would always have errors.
	var filtered []command.ProjectResult
	skippedStaleCommandHead := false
	for _, r := range results {
		if _, ok := r.Error.(command.DirNotExistErr); ok {
			ctx.Log.Debug("ignoring error result from project at dir %q workspace %q because it is dir not exist error", r.RepoRelDir, r.Workspace)
			continue
		}
		if errors.Is(r.Error, command.ErrStaleCommandHead) {
			ctx.Log.Debug("ignoring stale command-head result from project at dir %q workspace %q project %q", r.RepoRelDir, r.Workspace, r.ProjectName)
			skippedStaleCommandHead = true
			continue
		}
		filtered = append(filtered, r)
	}
	if skippedStaleCommandHead && len(filtered) == 0 {
		pullStatus, err := c.Store.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil {
			return *pullStatus, nil
		}
		return models.PullStatus{Pull: pull}, nil
	}
	ctx.Log.Debug("updating pull status with results")
	return c.Store.UpdatePullWithResults(pull, filtered)
}

// Replace deletes the pull's recorded status and rebuilds it from results.
func (c *PullStatusUpdater) Replace(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if err := c.Store.DeletePullStatus(pull); err != nil {
		return models.PullStatus{}, err
	}
	return c.Update(ctx, pull, results)
}

// UpdateForDiscardedPlans records results whose successful import/state-rm
// invalidated a previously discovered plan.
func (c *PullStatusUpdater) UpdateForDiscardedPlans(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) error {
	pullStatus, err := c.Store.GetPullStatus(pull)
	if err != nil {
		return err
	}
	if pullStatus == nil || !models.PullStatusFreshForPull(pull, pullStatus.Pull) {
		return nil
	}

	var discarded []command.ProjectResult
	for _, res := range results {
		if res.Error != nil || res.Failure != "" {
			continue
		}
		if res.ImportSuccess == nil && res.StateRmSuccess == nil {
			continue
		}
		proj := pullStatus.FindProject(res.Workspace, res.RepoRelDir, res.ProjectName)
		if proj == nil || !models.StatusAllowsDiscoveredPlan(proj.Status) {
			continue
		}
		discarded = append(discarded, res)
	}
	if len(discarded) == 0 {
		return nil
	}
	_, err = c.Update(ctx, pull, discarded)
	return err
}
