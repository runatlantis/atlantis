// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type DBUpdater struct {
	Database db.Database
}

func (c *DBUpdater) updateDB(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if len(results) == 0 && pull.HeadCommit != "" {
		pullStatus, err := c.Database.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil && !pullStatusFreshForPull(pull, pullStatus.Pull) {
			ctx.Log.Debug("ignoring empty result from pull head %q base %q because current plan status is for head %q base %q", pull.HeadCommit, pull.BaseBranch, pullStatus.Pull.HeadCommit, pullStatus.Pull.BaseBranch)
			return *pullStatus, nil
		}
	}

	if staleApplyResultForCurrentPull(pull, results) {
		pullStatus, err := c.Database.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil && !pullStatusFreshForPull(pull, pullStatus.Pull) {
			ctx.Log.Debug("ignoring stale apply result from pull head %q base %q because current plan status is for head %q base %q", pull.HeadCommit, pull.BaseBranch, pullStatus.Pull.HeadCommit, pullStatus.Pull.BaseBranch)
			return *pullStatus, nil
		}
	}

	// Filter out results that errored due to the directory not existing. We
	// don't store these in the database because they would never be "apply-able"
	// and so the pull request would always have errors.
	var filtered []command.ProjectResult
	skippedStaleCommandHead := false
	for _, r := range results {
		if _, ok := r.Error.(DirNotExistErr); ok {
			ctx.Log.Debug("ignoring error result from project at dir %q workspace %q because it is dir not exist error", r.RepoRelDir, r.Workspace)
			continue
		}
		if errors.Is(r.Error, errStaleCommandHead) {
			ctx.Log.Debug("ignoring stale command-head result from project at dir %q workspace %q project %q", r.RepoRelDir, r.Workspace, r.ProjectName)
			skippedStaleCommandHead = true
			continue
		}
		filtered = append(filtered, r)
	}
	if skippedStaleCommandHead && len(filtered) == 0 {
		pullStatus, err := c.Database.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil {
			return *pullStatus, nil
		}
		return models.PullStatus{Pull: pull}, nil
	}
	ctx.Log.Debug("updating DB with pull results")
	return c.Database.UpdatePullWithResults(pull, filtered)
}

func staleApplyResultForCurrentPull(pull models.PullRequest, results []command.ProjectResult) bool {
	if pull.HeadCommit == "" {
		return false
	}
	for _, result := range results {
		if result.Command == command.Apply {
			return true
		}
	}
	return false
}

func (c *DBUpdater) replaceDB(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if err := c.Database.DeletePullStatus(pull); err != nil {
		return models.PullStatus{}, err
	}
	return c.updateDB(ctx, pull, results)
}

func (c *DBUpdater) updateDBForDiscardedPlans(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) error {
	pullStatus, err := c.Database.GetPullStatus(pull)
	if err != nil {
		return err
	}
	if pullStatus == nil || !pullStatusFreshForPull(pull, pullStatus.Pull) {
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
		proj := findProjectInPullStatus(pullStatus, res.Workspace, res.RepoRelDir, res.ProjectName)
		if proj == nil || !statusAllowedForDiscoveredPlan(proj.Status) {
			continue
		}
		discarded = append(discarded, res)
	}
	if len(discarded) == 0 {
		return nil
	}
	_, err = c.updateDB(ctx, pull, discarded)
	return err
}
