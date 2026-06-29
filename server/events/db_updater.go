// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type DBUpdater struct {
	Database db.Database
}

func (c *DBUpdater) updateDB(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	// Filter out results that errored due to the directory not existing. We
	// don't store these in the database because they would never be "apply-able"
	// and so the pull request would always have errors.
	var filtered []command.ProjectResult
	for _, r := range results {
		if _, ok := r.Error.(DirNotExistErr); ok {
			ctx.Log.Debug("ignoring error result from project at dir %q workspace %q because it is dir not exist error", r.RepoRelDir, r.Workspace)
			continue
		}
		filtered = append(filtered, r)
	}
	ctx.Log.Debug("updating DB with pull results")
	return c.Database.UpdatePullWithResults(pull, filtered)
}

func (c *DBUpdater) replaceDB(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if err := c.Database.DeletePullStatus(pull); err != nil {
		return models.PullStatus{}, err
	}
	return c.updateDB(ctx, pull, results)
}

func (c *DBUpdater) updateDBForDiscardedPlans(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) error {
	var discarded []command.ProjectResult
	for _, res := range results {
		if res.Error != nil || res.Failure != "" {
			continue
		}
		if res.ImportSuccess == nil && res.StateRmSuccess == nil {
			continue
		}
		discarded = append(discarded, res)
	}
	if len(discarded) == 0 {
		return nil
	}
	_, err := c.updateDB(ctx, pull, discarded)
	return err
}
