package queue

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"go.temporal.io/sdk/workflow"
)

type LockStateUpdater struct {
	Activities githubActivities
}

func (u *LockStateUpdater) UpdateQueuedRevisions(ctx workflow.Context, queue *Deploy) {
	lock := queue.GetLockState()
	infos := queue.GetOrderedMergedItems()

	var actions []github.CheckRunAction
	var summary string
	if lock.Status == LockedStatus {
		actions = append(actions, github.CreateUnlockAction())
		summary = fmt.Sprintf("This deploy is locked from a manual deployment for revision %s.  Unlock to proceed.", lock.Revision)
	}

	for _, i := range infos {
		err := workflow.ExecuteActivity(ctx, u.Activities.UpdateCheckRun, activities.UpdateCheckRunRequest{
			Title:   terraform.BuildCheckRunTitle(i.Root.Name),
			State:   github.CheckRunQueued,
			Repo:    i.Repo,
			ID:      i.CheckRunID,
			Summary: summary,
			Actions: actions,
		}).Get(ctx, nil)

		if err != nil {
			logger.Error(ctx, fmt.Sprintf("updating check run for revision %s", i.Revision), "err", err)
		}
	}
}
