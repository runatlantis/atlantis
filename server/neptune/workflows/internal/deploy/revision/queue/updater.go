package queue

import (
	"fmt"

	key "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/notifier"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/version"
	"go.temporal.io/sdk/workflow"
)

type CheckRunClient interface {
	CreateOrUpdate(ctx workflow.Context, deploymentID string, request notifier.GithubCheckRunRequest) (int64, error)
}

type LockStateUpdater struct {
	Activities          githubActivities
	GithubCheckRunCache CheckRunClient
}

func (u *LockStateUpdater) UpdateQueuedRevisions(ctx workflow.Context, queue *Deploy) {
	lock := queue.GetLockState()
	infos := queue.GetOrderedMergedItems()

	var actions []github.CheckRunAction
	var summary string
	state := github.CheckRunQueued
	if lock.Status == LockedStatus {
		actions = append(actions, github.CreateUnlockAction())
		state = github.CheckRunActionRequired
		summary = fmt.Sprintf("This deploy is locked from a manual deployment for revision %s.  Unlock to proceed.", lock.Revision)
	}

	for _, i := range infos {
		request := notifier.GithubCheckRunRequest{
			Title:   terraform.BuildCheckRunTitle(i.Root.Name),
			Sha:     i.Revision,
			State:   state,
			Repo:    i.Repo,
			Summary: summary,
			Actions: actions,
		}
		logger.Debug(ctx, fmt.Sprintf("Updating lock status for deployment id: %s", i.ID.String()))
		var err error

		version := workflow.GetVersion(ctx, version.CacheCheckRunSessions, workflow.DefaultVersion, 1)
		if version == workflow.DefaultVersion {
			err = workflow.ExecuteActivity(ctx, u.Activities.GithubUpdateCheckRun, activities.UpdateCheckRunRequest{
				Title:   request.Title,
				State:   request.State,
				Repo:    request.Repo,
				Summary: request.Summary,
				Actions: request.Actions,
				ID:      i.CheckRunID,
			}).Get(ctx, nil)
		} else {
			_, err = u.GithubCheckRunCache.CreateOrUpdate(ctx, i.ID.String(), request)
		}

		if err != nil {
			logger.Error(ctx, fmt.Sprintf("updating check run for revision %s", i.Revision), key.ErrKey, err)
		}
	}
}
