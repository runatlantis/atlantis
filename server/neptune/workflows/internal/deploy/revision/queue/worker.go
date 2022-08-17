package queue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/steps"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type workerActivities interface {
	FetchLatestDeployment(ctx context.Context, request activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error)
	UpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error)
}

type WorkerState string

const (
	WaitingWorkerState  WorkerState = "waiting"
	WorkingWorkerState  WorkerState = "working"
	CompleteWorkerState WorkerState = "complete"
)

type Worker struct {
	Activities        workerActivities
	Queue             *Queue
	Repo              github.Repo
	Root              steps.Root
	TerraformWorkflow func(ctx workflow.Context, request terraform.Request) error

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

		revision := msg.Revision
		checkRunID := msg.CheckRunID
		ctx := workflow.WithValue(ctx, config.RevisionLogKey, revision)

		w.updateInProgress(ctx, checkRunID)
		err = w.work(ctx, revision)
		w.updateComplete(ctx, checkRunID)

		if err != nil {
			logger.Error(ctx, "failed to deploy revision, moving to next one")
		}
	}
}

func (w *Worker) GetState() WorkerState {
	return w.state
}

func (w *Worker) updateComplete(ctx workflow.Context, checkRunID int64) {
	// TODO: if we never created a check run, there was likely some issue, we should attempt to create it again.
	if checkRunID == 0 {
		logger.Info(ctx, "check run id is 0, skipping updating state to complete")
	}

	var resp activities.UpdateCheckRunResponse

	// intentionally infinitely retry since it's important we at least apply a completion status.
	// we might want to not block on this and track this in the background so as to not block future deploys
	_ = workflow.ExecuteActivity(ctx, w.Activities.UpdateCheckRun, activities.UpdateCheckRunRequest{
		Title: "atlantis/deploy",
		State: github.CheckRunComplete,
		// TODO: Add conclusion
		Repo: w.Repo,
		ID:   checkRunID,
	}).Get(ctx, &resp)
}

func (w *Worker) updateInProgress(ctx workflow.Context, checkRunID int64) {
	// TODO: if we never created a check run, there was likely some issue, we should attempt to create it again.
	if checkRunID == 0 {
		logger.Info(ctx, "check run id is 0, skipping updating state to in_progress")
		return
	}
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: 5,
	})

	var resp activities.UpdateCheckRunResponse

	err := workflow.ExecuteActivity(ctx, w.Activities.UpdateCheckRun, activities.UpdateCheckRunRequest{
		Title: "atlantis/deploy",
		State: github.CheckRunPending,
		Repo:  w.Repo,
		ID:    checkRunID,
	}).Get(ctx, &resp)

	if err != nil {
		logger.Error(ctx, err.Error())
	}
}

func (w *Worker) work(ctx workflow.Context, revision string) error {
	id, err := generateID(ctx)

	ctx = workflow.WithValue(ctx, config.DeploymentIDLogKey, id)

	logger.Info(ctx, "Generated id")

	if err != nil {
		return errors.Wrap(err, "generating id")
	}

	deployedRevision, err := w.fetchLatestDeployment(ctx)

	if err != nil {
		return errors.Wrap(err, "fetching latest deployment")
	}

	logger.Info(ctx, fmt.Sprintf("latest deployed revision %s", deployedRevision))

	// TODO: fill in the rest
	// Spin up a child workflow to handle Terraform operations
	//childWorkflowOptions := workflow.ChildWorkflowOptions{
	//	WorkflowID: id.String(),
	//}
	//ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)
	//terraformWorkflowRequest := terraform.Request{
	//	Repo: w.Repo,
	//	Root: w.Root,
	//}
	//err = workflow.ExecuteChildWorkflow(ctx, w.TerraformWorkflow, terraformWorkflowRequest).Get(ctx, nil)
	//if err != nil {
	//	return errors.Wrap(err, "executing child terraform workflow")
	//}
	return nil
}

func (w *Worker) fetchLatestDeployment(ctx workflow.Context) (string, error) {
	request := activities.FetchLatestDeploymentRequest{
		RepositoryURL: w.Repo.URL,
		RootName:      w.Root.Name,
	}

	var resp activities.FetchLatestDeploymentResponse
	err := workflow.ExecuteActivity(ctx, w.Activities.FetchLatestDeployment, request).Get(ctx, &resp)

	return resp.Revision, err
}

func generateID(ctx workflow.Context) (uuid.UUID, error) {
	// UUIDErr allows us to extract both the id and the err from the sideeffect
	// not sure if there is a better way to do this
	type UUIDErr struct {
		id  uuid.UUID
		err error
	}

	var result UUIDErr
	encodedResult := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		uuid, err := uuid.NewUUID()

		return UUIDErr{
			id:  uuid,
			err: err,
		}
	})

	err := encodedResult.Get(&result)

	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, "getting uuid from side effect")
	}

	return result.id, result.err
}
