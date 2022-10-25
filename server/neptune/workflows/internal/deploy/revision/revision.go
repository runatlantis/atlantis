package revision

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	activity "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request/converter"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type idGenerator func(ctx workflow.Context) (uuid.UUID, error)

type NewRevisionRequest struct {
	Revision       string
	InitiatingUser request.User
	Root           request.Root
	Repo           request.Repo
	Tags           map[string]string
}

type Queue interface {
	Push(terraform.DeploymentInfo)
	GetLockState() queue.LockState
	SetLockForMergedItems(ctx workflow.Context, state queue.LockState)
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
	repo := converter.Repo(request.Repo)
	initiatingUser := converter.User(request.InitiatingUser)

	ctx := workflow.WithRetryPolicy(n.ctx, temporal.RetryPolicy{
		MaximumAttempts: 5,
	})

	// generate an id for this deployment and pass that to our check run
	id, err := n.idGenerator(ctx)

	if err != nil {
		logger.Error(ctx, "generating deployment id", "err", err)
	}

	resp := n.createCheckRun(ctx, id.String(), request.Revision, root, repo)

	// lock the queue on a manual deployment
	if root.Trigger == activity.ManualTrigger {
		n.queue.SetLockForMergedItems(ctx, queue.LockState{
			Status:   queue.LockedStatus,
			Revision: request.Revision,
		})
	}

	n.queue.Push(terraform.DeploymentInfo{
		ID:             id,
		Root:           root,
		Revision:       request.Revision,
		CheckRunID:     resp.ID,
		Repo:           repo,
		InitiatingUser: initiatingUser,
		Tags:           request.Tags,
	})
}

func (n *Receiver) createCheckRun(ctx workflow.Context, id, revision string, root activity.Root, repo github.Repo) activities.CreateCheckRunResponse {
	lock := n.queue.GetLockState()
	var actions []github.CheckRunAction
	var summary string

	if lock.Status == queue.LockedStatus && (root.Trigger == activity.MergeTrigger) {
		actions = append(actions, github.CreateUnlockAction())
		summary = fmt.Sprintf("This deploy is locked from a manual deployment for revision %s.  Unlock to proceed.", lock.Revision)
	}

	var resp activities.CreateCheckRunResponse
	err := workflow.ExecuteActivity(ctx, n.activities.CreateCheckRun, activities.CreateCheckRunRequest{
		Title:      terraform.BuildCheckRunTitle(root.Name),
		Sha:        revision,
		Repo:       repo,
		ExternalID: id,
		Summary:    summary,
		Actions:    actions,
	}).Get(ctx, &resp)

	// don't block on error here, we'll just try again later when we have our result.
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	return resp
}
