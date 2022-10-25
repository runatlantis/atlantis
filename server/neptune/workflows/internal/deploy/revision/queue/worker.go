package queue

import (
	"fmt"

	"github.com/pkg/errors"
	internalContext "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type queue interface {
	IsEmpty() bool
	CanPop() bool
	Pop() (terraform.DeploymentInfo, error)
	SetLockForMergedItems(ctx workflow.Context, state LockState)
}

type deployer interface {
	Deploy(ctx workflow.Context, requestedDeployment terraform.DeploymentInfo, latestDeployment *deployment.Info) (*deployment.Info, error)
}

type WorkerState string

const (
	WaitingWorkerState  WorkerState = "waiting"
	WorkingWorkerState  WorkerState = "working"
	CompleteWorkerState WorkerState = "complete"

	UnlockSignalName = "unlock"
)

type UnlockSignalRequest struct {
	User string
}

type Worker struct {
	Queue    queue
	Deployer deployer

	// mutable
	state WorkerState
}

type actionType string

const (
	canceled = "canceled"
	process  = "process"
	receive  = "receive"
)

// Work pops work off the queue and if the queue is empty,
// it waits for the queue to be non-empty or a cancelation signal
func (w *Worker) Work(ctx workflow.Context) {
	// set to complete once we return else callers could think we are still working based on the 'working' state.
	defer func() {
		w.state = CompleteWorkerState
	}()

	var previousDeployment *deployment.Info
	var currentAction actionType
	callback := func(f workflow.Future) {
		err := f.Get(ctx, nil)

		if temporal.IsCanceledError(err) {
			currentAction = canceled
			return
		}

		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unknown error %s, worker is shutting down", err.Error()))
			currentAction = canceled
			return
		}

		currentAction = process
	}
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(w.awaitWork(ctx), callback)

	var request UnlockSignalRequest
	selector.AddReceive(workflow.GetSignalChannel(ctx, UnlockSignalName), func(c workflow.ReceiveChannel, more bool) {
		_ = c.Receive(ctx, &request)
		currentAction = receive
	})

	for {
		if w.Queue.IsEmpty() {
			w.state = WaitingWorkerState
		}

		selector.Select(ctx)

		var currentDeployment *deployment.Info
		var err error
		switch currentAction {
		case canceled:
			logger.Info(ctx, "Received cancelled signal, worker is shutting down")
			return
		case process:
			logger.Info(ctx, "Processing... ")
			currentDeployment, err = w.deploy(ctx, previousDeployment)
		case receive:
			logger.Info(ctx, "Received unlock signal... ")
			w.Queue.SetLockForMergedItems(ctx, LockState{
				Status: UnlockedStatus,
			})
			continue
		default:
			logger.Warn(ctx, fmt.Sprintf("%s action not configured. This is probably a bug, skipping for now", currentAction))
			return
		}

		// from here on we know that we've processed an item so let's ensure we're adding an await future
		// back to our selector regardless of the outcome
		// Note: for validation errors we don't want to track this as a deploy
		if e, ok := err.(*ValidationError); ok {
			logger.Error(ctx, "deploy validation failed, moving on to next one", "err", e)
			selector.AddFuture(w.awaitWork(ctx), callback)
			continue
		}

		if err != nil {
			logger.Error(ctx, "failed to process revision, moving to next one", "err", err)
		}

		// we assume that regardless of error we should count this as a deploy
		previousDeployment = currentDeployment
		selector.AddFuture(w.awaitWork(ctx), callback)
	}
}

func (w *Worker) awaitWork(ctx workflow.Context) workflow.Future {
	future, settable := workflow.NewFuture(ctx)

	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.Await(ctx, func() bool {
			return w.Queue.CanPop()
		})

		settable.SetError(err)
	})

	return future
}

func (w *Worker) deploy(ctx workflow.Context, latestDeployment *deployment.Info) (*deployment.Info, error) {
	w.state = WorkingWorkerState

	msg, err := w.Queue.Pop()

	if err != nil {
		return nil, errors.Wrap(err, "popping off queue")
	}

	ctx = workflow.WithValue(ctx, internalContext.SHAKey, msg.Revision)
	ctx = workflow.WithValue(ctx, internalContext.DeploymentIDKey, msg.ID)
	return w.Deployer.Deploy(ctx, msg, latestDeployment)

}

func (w *Worker) GetState() WorkerState {
	return w.state
}
