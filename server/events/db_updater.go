package events

import (
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/models"
)

type DBUpdater struct {
	DB *db.BoltDB
}

func (c *DBUpdater) updateDB(ctx *CommandContext, pull models.PullRequest, results []models.ProjectResult) (models.PullStatus, error) {
	// Filter out results that errored due to the directory not existing. We
	// don't store these in the database because they would never be "apply-able"
	// and so the pull request would always have errors.
	var filtered []models.ProjectResult
	for _, r := range results {
		if _, ok := r.Error.(DirNotExistErr); ok {
			ctx.Log.Debug("ignoring error result from project at dir %q workspace %q because it is dir not exist error", r.RepoRelDir, r.Workspace)
			continue
		}
		filtered = append(filtered, r)
	}
	ctx.Log.Debug("updating DB with pull results")
	return c.DB.UpdatePullWithResults(pull, filtered)
}
