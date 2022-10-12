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

const (
	UpdateCheckRunRetryCount = 5
	DeploymentInfoVersion    = "1.0.0"
	DirectionBehindSummary   = "This revision is behind the current revision and will not be deployed.  If this is intentional, revert the default branch to this revision to trigger a new deployment."
)

type terraformWorkflowRunner interface {
	Run(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) error
}

type dbActivities interface {
	FetchLatestDeployment(ctx context.Context, request activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error)
	StoreLatestDeployment(ctx context.Context, request activities.StoreLatestDeploymentRequest) error
}

type githubActivities interface {
	CompareCommit(ctx context.Context, request activities.CompareCommitRequest) (activities.CompareCommitResponse, error)
	UpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error)
}

type workerActivities interface {
	dbActivities
	githubActivities
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
	Queue                   *Queue
	TerraformWorkflowRunner terraformWorkflowRunner
	Activities              workerActivities

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

		var compareCommitResp activities.CompareCommitResponse
		err = workflow.ExecuteActivity(ctx, w.Activities.CompareCommit, activities.CompareCommitRequest{
			DeployRequestRevision:  msg.Revision,
			LatestDeployedRevision: w.LatestDeployment.Revision,
			Repo:                   msg.Repo,
		}).Get(ctx, &compareCommitResp)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unable to compare deploy request commit with the latest deployed commit: %s", err.Error()))
		}

		if !w.isValidRevision(ctx, msg, compareCommitResp) {
			continue
		}

		err = w.TerraformWorkflowRunner.Run(ctx, msg)
		if err != nil {
			logger.Error(ctx, "failed to deploy revision, moving to next one")
			continue
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

func (w *Worker) isValidRevision(ctx workflow.Context, msg terraform.DeploymentInfo, resp activities.CompareCommitResponse) bool {
	switch resp.CommitComparison {
	case activities.DirectionBehind:
		logger.Info(ctx, fmt.Sprintf("Deployed Revision: %s is ahead of the Deploy Request Revision: %s, moving to next one", w.LatestDeployment.Revision, msg.Revision))
		w.updateCheckRun(ctx, msg, github.CheckRunFailure, DirectionBehindSummary)
		return false

	case activities.DirectionIdentical:
		logger.Info(ctx, fmt.Sprintf("Deployed Revision: %s is identical to the Deploy Request Revision: %s", w.LatestDeployment.Revision, msg.Revision))
		return true

	// TODO: Handle Force Applies
	case activities.DirectionDiverged:
		logger.Info(ctx, fmt.Sprintf("Deployed Revision: %s is divergent from the Deploy Request Revision: %s.", w.LatestDeployment.Revision, msg.Revision))
		return true

	case activities.DirectionAhead:
		logger.Info(ctx, fmt.Sprintf("Deployed Revision: %s is ahead of the Deploy Request Revision: %s", w.LatestDeployment.Revision, msg.Revision))
		return true

	default:
		logger.Error(ctx, fmt.Sprintf("Invalid commit comparison response: %s", resp.CommitComparison))
		return false
	}
}

func (w *Worker) fetchLatestDeployment(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) (*root.DeploymentInfo, error) {
	var resp activities.FetchLatestDeploymentResponse
	err := workflow.ExecuteActivity(ctx, w.Activities.FetchLatestDeployment, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: deploymentInfo.Repo.GetFullName(),
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
		Repo:       deploymentInfo.Repo,
	}
	err := workflow.ExecuteActivity(ctx, w.Activities.StoreLatestDeployment, activities.StoreLatestDeploymentRequest{
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

// worker should not block on updating checkruns for invalid deploy requests so let's retry for UpdateCheckrunRetryCount only
func (w *Worker) updateCheckRun(ctx workflow.Context, deployRequest terraform.DeploymentInfo, state github.CheckRunState, summary string) {
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: UpdateCheckRunRetryCount,
	})

	err := workflow.ExecuteActivity(ctx, w.Activities.UpdateCheckRun, activities.UpdateCheckRunRequest{
		Title:   terraform.BuildCheckRunTitle(deployRequest.Root.Name),
		State:   state,
		Repo:    deployRequest.Repo,
		ID:      deployRequest.CheckRunID,
		Summary: summary,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error(ctx, "failed to update checkrun: %s", err.Error())
	}
}
