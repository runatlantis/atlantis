package revision

import (
	"context"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request/converter"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type idGenerator func(ctx workflow.Context) (uuid.UUID, error)

type NewRevisionRequest struct {
	Revision string
	Root     request.Root
	Repo     request.Repo
}

type Queue interface {
	Push(terraform.DeploymentInfo)
}

type Activities interface {
	CreateCheckRun(ctx context.Context, request activities.CreateCheckRunRequest) (activities.CreateCheckRunResponse, error)
}

func NewReceiver(ctx workflow.Context, queue Queue, activities Activities, generator idGenerator) *Receiver {
	return &Receiver{
		queue:       queue,
		ctx:         ctx,
		activities:  activities,
		idGenerator: generator,
	}
}

type Receiver struct {
	queue       Queue
	ctx         workflow.Context
	activities  Activities
	idGenerator idGenerator
}

func (n *Receiver) Receive(c workflow.ReceiveChannel, more bool) {
	// more is false when the channel is closed, so let's just return right away
	if !more {
		return
	}

	var request NewRevisionRequest
	c.Receive(n.ctx, &request)

	root := converter.Root(request.Root)
	repo := github.Repo{
		Name:  request.Repo.Name,
		Owner: request.Repo.Owner,
		URL:   request.Repo.URL,
		Credentials: github.AppCredentials{
			InstallationToken: request.Repo.Credentials.InstallationToken,
		},
		Ref: github.Ref{
			Name: request.Repo.Ref.Name,
			Type: request.Repo.Ref.Type,
		},
	}

	ctx := workflow.WithRetryPolicy(n.ctx, temporal.RetryPolicy{
		MaximumAttempts: 5,
	})

	// generate an id for this deployment and pass that to our check run
	id, err := n.idGenerator(ctx)

	if err != nil {
		logger.Error(ctx, "generating deployment id", "err", err)
	}

	var resp activities.CreateCheckRunResponse
	err = workflow.ExecuteActivity(ctx, n.activities.CreateCheckRun, activities.CreateCheckRunRequest{
		Title:      terraform.BuildCheckRunTitle(root.Name),
		Sha:        request.Revision,
		Repo:       repo,
		ExternalID: id.String(),
	}).Get(ctx, &resp)

	// don't block on error here, we'll just try again later when we have our result.
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	n.queue.Push(terraform.DeploymentInfo{
		ID:         id,
		Root:       root,
		Revision:   request.Revision,
		CheckRunID: resp.ID,
		Repo:       repo,
	})
}
