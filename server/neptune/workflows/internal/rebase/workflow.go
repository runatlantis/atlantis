package rebase

import (
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	workflowMetrics "github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	RebaseGithubRetryCount    = 3
	RebaseStartToCloseTimeout = 10 * time.Second
)

type Request struct {
	Repo github.Repo
	Root terraform.Root
}

type rebaseActivities struct {
	*activities.RevsionSetter
	*activities.Github
}

func Workflow(ctx workflow.Context, request Request) error {
	var r *rebaseActivities
	scope := workflowMetrics.NewScope(ctx, "rebase")

	// GH API calls should not hit ratelimit issues since we cap the TaskQueueActivitiesPerSecond for the rebase TQ
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: RebaseStartToCloseTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: RebaseGithubRetryCount,
		},
	})

	pullRebaser := PullRebaser{
		RebaseActivites: *r,
	}

	return pullRebaser.RebaseOpenPRsForRoot(ctx, request.Repo, request.Root, scope)
}
