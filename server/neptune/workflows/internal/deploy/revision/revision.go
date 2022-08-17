package revision

import (
	"context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type NewRevisionRequest struct {
	Revision string
}

type Queue interface {
	Push(queue.Message)
}

type Activities interface {
	CreateCheckRun(ctx context.Context, request activities.CreateCheckRunRequest) (activities.CreateCheckRunResponse, error)
}

func NewReceiver(ctx workflow.Context, queue Queue, repo github.Repo, activities Activities) *Receiver {
	return &Receiver{
		queue:      queue,
		ctx:        ctx,
		repo:       repo,
		activities: activities,
	}
}

type Receiver struct {
	queue      Queue
	ctx        workflow.Context
	activities Activities
	repo       github.Repo
}

func (n *Receiver) Receive(c workflow.ReceiveChannel, more bool) {
	// more is false when the channel is closed, so let's just return right away
	if !more {
		return
	}

	var request NewRevisionRequest
	c.Receive(n.ctx, &request)

	ctx := workflow.WithRetryPolicy(n.ctx, temporal.RetryPolicy{
		MaximumAttempts: 5,
	})

	var resp activities.CreateCheckRunResponse

	err := workflow.ExecuteActivity(ctx, n.activities.CreateCheckRun, activities.CreateCheckRunRequest{
		Title: "atlantis/deploy",
		Sha:   request.Revision,
		Repo:  n.repo,
	}).Get(ctx, &resp)

	// don't block on error here, we'll just try again later when we have our result.
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	n.queue.Push(queue.Message{
		Revision:   request.Revision,
		CheckRunID: resp.ID,
	})

}
