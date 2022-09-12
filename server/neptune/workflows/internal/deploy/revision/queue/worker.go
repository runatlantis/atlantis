package queue

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type terraformWorkflowRunner interface {
	Run(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) error
}

type WorkerState string

const (
	WaitingWorkerState  WorkerState = "waiting"
	WorkingWorkerState  WorkerState = "working"
	CompleteWorkerState WorkerState = "complete"
)

type Worker struct {
	Queue                   *Queue
	TerraformWorkflowRunner terraformWorkflowRunner

	// mutable
	state WorkerState
}

// Work pops work off the queue and if the queue is empty,
// it waits for the queue to be non-empty or a cancelation signal
func (w *Worker) Work(ctx workflow.Context) {
	// set to complete once we return else callers could think we are still working based on the 'working' state.
	defer func() {
		w.state = CompleteWorkerState
	}()

	for {
		if w.Queue.IsEmpty() {
			w.state = WaitingWorkerState
		}

		err := workflow.Await(ctx, func() bool {
			return !w.Queue.IsEmpty()
		})

		if temporal.IsCanceledError(err) {
			logger.Info(ctx, "Received cancelled signal, worker is shutting down")
			return
		}

		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unknown error %s, worker is shutting down", err.Error()))
			return
		}

		w.state = WorkingWorkerState

		msg := w.Queue.Pop()

		ctx := workflow.WithValue(ctx, context.SHAKey, msg.Revision)
		ctx = workflow.WithValue(ctx, context.DeploymentIDKey, msg.ID)

		err = w.TerraformWorkflowRunner.Run(ctx, msg)

		if err != nil {
			logger.Error(ctx, "failed to deploy revision, moving to next one")
		}
	}
}

func (w *Worker) GetState() WorkerState {
	return w.state
}
