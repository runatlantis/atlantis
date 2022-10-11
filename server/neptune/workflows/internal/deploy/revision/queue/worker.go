package queue

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	internalContext "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const DeploymentInfoVersion = "1.0.0"

type terraformWorkflowRunner interface {
	Run(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) error
}

type dbActivities interface {
	FetchLatestDeployment(ctx context.Context, request activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error)
	StoreLatestDeployment(ctx context.Context, request activities.StoreLatestDeploymentRequest) error
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
	DbActivities            dbActivities
	Repo                    github.Repo

	// mutable
	state            WorkerState
	LatestDeployment *root.DeploymentInfo
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

		ctx := workflow.WithValue(ctx, internalContext.SHAKey, msg.Revision)
		ctx = workflow.WithValue(ctx, internalContext.DeploymentIDKey, msg.ID)

		// This should only happen on startup
		// If we fail to fetch the latest deployment, log and exit the workflow
		if w.LatestDeployment == nil {
			w.LatestDeployment, err = w.fetchLatestDeployment(ctx, msg)
			if err != nil {
				logger.Error(ctx, fmt.Sprint("Unable to fetch latest deployment, worker is shutting down", err.Error()))
				return
			}
		}

		// TODO: Validate revision before executing the workflow

		err = w.TerraformWorkflowRunner.Run(ctx, msg)
		if err != nil {
			logger.Error(ctx, "failed to deploy revision, moving to next one")
		}

		// TODO: Persist deployment on shutdown if it fails instead of blocking
		latestDeployment, err := w.persistLatestDeployment(ctx, msg)
		if err != nil {
			logger.Error(ctx, "failed to persist latest deploy job")
		}

		// Update the latest deployment in memory
		w.LatestDeployment = latestDeployment
	}
}

func (w *Worker) fetchLatestDeployment(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) (*root.DeploymentInfo, error) {
	// Fetch latest deployment
	var resp activities.FetchLatestDeploymentResponse
	err := workflow.ExecuteActivity(ctx, w.DbActivities.FetchLatestDeployment, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: w.Repo.GetFullName(),
		RootName:           deploymentInfo.Root.Name,
	}).Get(ctx, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "fetching latest deployment")
	}

	return resp.DeploymentInfo, nil
}

func (w *Worker) persistLatestDeployment(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) (*root.DeploymentInfo, error) {
	latestDeploymentInfo := root.DeploymentInfo{
		Version:    DeploymentInfoVersion,
		ID:         deploymentInfo.ID.String(),
		CheckRunID: deploymentInfo.CheckRunID,
		Revision:   deploymentInfo.Revision,
		Root:       deploymentInfo.Root,
		Repo:       w.Repo,
	}
	err := workflow.ExecuteActivity(ctx, w.DbActivities.StoreLatestDeployment, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: latestDeploymentInfo,
	}).Get(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "persisting deployment info")
	}
	return &latestDeploymentInfo, nil
}

func (w *Worker) GetState() WorkerState {
	return w.state
}
