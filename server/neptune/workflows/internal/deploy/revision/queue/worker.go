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
	state WorkerState
}

// Work pops work off the queue and if the queue is empty,
// it waits for the queue to be non-empty or a cancelation signal
func (w *Worker) Work(ctx workflow.Context) {
	// set to complete once we return else callers could think we are still working based on the 'working' state.
	defer func() {
		w.state = CompleteWorkerState
	}()

	var latestDeployment *root.DeploymentInfo
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

		latestDeployment, err = w.fetchLatestDeployment(ctx, msg, latestDeployment)
		if err != nil {
			logger.Error(ctx, "unable to fetch latest deployment for root: %s, skipping deploy: %s", msg.Root.Name, err.Error())
			continue
		}

		commitDirection, err := w.getDeployRequestCommitDirection(ctx, msg, latestDeployment)
		if err != nil {
			logger.Error(ctx, "unable to determine depoy request commit direction: %s", err.Error())
			continue
		}

		if !w.isValidRevision(msg, commitDirection) {
			logger.Info(ctx, fmt.Sprintf("Deploy Request Revision: %s is not valid, moving to next one", msg.Revision))
			w.updateCheckRun(ctx, msg, github.CheckRunFailure, DirectionBehindSummary)
			continue
		}

		if err = w.processRevision(ctx, msg, latestDeployment, commitDirection); err != nil {
			logger.Error(ctx, "failed to process revision, moving to next one: %s", err.Error())
			continue
		}

		err = w.TerraformWorkflowRunner.Run(ctx, msg)
		if err != nil {
			logger.Error(ctx, "failed to deploy revision, moving to next one")
			continue
		}
		latestDeployment = w.buildLatestDeployment(msg)

		// TODO: Persist deployment on shutdown if it fails instead of blocking
		if err = w.persistLatestDeployment(ctx, latestDeployment); err != nil {
			logger.Error(ctx, "failed to persist latest deploy job")
		}
	}
}

// TODO: Check if triger type is Manual for Diverged commits for FA
func (w *Worker) isValidRevision(deployRequest terraform.DeploymentInfo, commitDirection activities.DiffDirection) bool {
	return commitDirection != activities.DirectionBehind
}

func (w *Worker) getDeployRequestCommitDirection(ctx workflow.Context, deployRequest terraform.DeploymentInfo, latestDeployment *root.DeploymentInfo) (activities.DiffDirection, error) {
	// root being deployed for the first time
	if latestDeployment == nil {
		return activities.DirectionAhead, nil
	}

	var compareCommitResp activities.CompareCommitResponse
	err := workflow.ExecuteActivity(ctx, w.Activities.CompareCommit, activities.CompareCommitRequest{
		DeployRequestRevision:  deployRequest.Revision,
		LatestDeployedRevision: latestDeployment.Revision,
		Repo:                   deployRequest.Repo,
	}).Get(ctx, &compareCommitResp)
	if err != nil {
		return activities.DiffDirection(""), errors.Wrap(err, "comparing committ")
	}

	return compareCommitResp.CommitComparison, nil
}

func (w *Worker) buildLatestDeployment(deployRequest terraform.DeploymentInfo) *root.DeploymentInfo {
	return &root.DeploymentInfo{
		Version:    DeploymentInfoVersion,
		ID:         deployRequest.ID.String(),
		CheckRunID: deployRequest.CheckRunID,
		Revision:   deployRequest.Revision,
		Root:       deployRequest.Root,
		Repo:       deployRequest.Repo,
	}
}

func (w *Worker) processRevision(ctx workflow.Context, deployRequest terraform.DeploymentInfo, latestDeployment *root.DeploymentInfo, commitComparison activities.DiffDirection) error {

	switch commitComparison {
	case activities.DirectionIdentical:
		logger.Info(ctx, fmt.Sprintf("Deployed Revision: %s is identical to the Deploy Request Revision: %s", latestDeployment.Revision, deployRequest.Revision))
		return nil

	// TODO: Handle Force Applies
	case activities.DirectionDiverged:
		logger.Info(ctx, fmt.Sprintf("Deployed Revision: %s is divergent from the Deploy Request Revision: %s.", latestDeployment.Revision, deployRequest.Revision))
		return nil

	case activities.DirectionAhead:
		var logMsg string
		if latestDeployment == nil {
			logMsg = fmt.Sprintf("Deploying root: %s at revision: %s for the first time", deployRequest.Root.Name, deployRequest.Revision)
		} else {
			logMsg = fmt.Sprintf("Deployed Revision: %s is ahead of the Deploy Request Revision: %s", latestDeployment.Revision, deployRequest.Revision)
		}
		logger.Info(ctx, logMsg)
		return nil

	default:
		return fmt.Errorf("Invalid commit comparison response: %s", commitComparison)
	}
}

func (w *Worker) fetchLatestDeployment(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo, latestDeployment *root.DeploymentInfo) (*root.DeploymentInfo, error) {
	// Skip fetching latest deployment it it's already in memory
	if latestDeployment != nil {
		return latestDeployment, nil
	}
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

func (w *Worker) persistLatestDeployment(ctx workflow.Context, deploymentInfo *root.DeploymentInfo) error {
	err := workflow.ExecuteActivity(ctx, w.Activities.StoreLatestDeployment, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: deploymentInfo,
	}).Get(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "persisting deployment info")
	}
	return nil
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
		logger.Error(ctx, "unable to update checkrun", err.Error())
	}
}
