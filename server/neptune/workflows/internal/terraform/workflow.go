package terraform

import (
	"context"
	"fmt"
	"time"

	key "github.com/runatlantis/atlantis/server/neptune/context"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/sideeffect"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/gate"
	runner "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type githubActivities interface {
	GithubFetchRoot(ctx context.Context, request activities.FetchRootRequest) (activities.FetchRootResponse, error)
}

type terraformActivities interface {
	Cleanup(ctx context.Context, request activities.CleanupRequest) (activities.CleanupResponse, error)
	GetWorkerInfo(ctx context.Context) (*activities.GetWorkerInfoResponse, error)
}

// jobRunner runs a deploy plan/apply job
type jobRunner interface {
	Plan(ctx workflow.Context, localRoot *terraform.LocalRoot, jobID string) (activities.TerraformPlanResponse, error)
	Apply(ctx workflow.Context, localRoot *terraform.LocalRoot, jobID string, planFile string) error
}

const (
	StartToCloseTimeout = 24 * time.Hour
	HeartBeatTimeout    = 1 * time.Minute

	// 1 week timeout for review gate
	ReviewGateTimeout = 24 * time.Hour * 7
)

func Workflow(ctx workflow.Context, request Request) error {
	runner := newRunner(ctx, request)

	// blocking call
	return runner.Run(ctx)
}

type Runner struct {
	GithubActivities    githubActivities
	TerraformActivities terraformActivities
	JobRunner           jobRunner
	Request             Request
	Store               *state.WorkflowStore
	RootFetcher         *RootFetcher
	MetricsHandler      client.MetricsHandler
	ReviewGate          *gate.Review
}

func newRunner(ctx workflow.Context, request Request) *Runner {
	var ta *activities.Terraform
	var ga *activities.Github

	cmdStepRunner := runner.CmdStepRunner{
		Activity: ta,
	}

	parent := workflow.GetInfo(ctx).ParentWorkflowExecution

	metricsHandler := workflow.GetMetricsHandler(ctx)

	// We have critical things relying on this notification so this workflow provides guarantees around this. (ie. compliance auditing)  There should
	// be no situation where we are deploying while this is failing.
	store := state.NewWorkflowStore(
		func(s *state.Workflow) error {
			return workflow.SignalExternalWorkflow(ctx, parent.ID, parent.RunID, state.WorkflowStateChangeSignal, s).Get(ctx, nil)
		},
	)

	return &Runner{
		ReviewGate: &gate.Review{
			MetricsHandler: metricsHandler,
			Timeout:        ReviewGateTimeout,
			Client:         store,
		},
		GithubActivities:    ga,
		TerraformActivities: ta,
		Request:             request,
		JobRunner: runner.NewRunner(
			&cmdStepRunner,
			&runner.EnvStepRunner{
				CmdStepRunner: cmdStepRunner,
			},
			ta,
		),
		RootFetcher: &RootFetcher{
			Request: request,
			Ga:      ga,
			Ta:      ta,
		},
		MetricsHandler: metricsHandler,
		Store:          store,
	}
}

func (r *Runner) Plan(ctx workflow.Context, root *terraform.LocalRoot, serverURL fmt.Stringer) (activities.TerraformPlanResponse, error) {
	var response activities.TerraformPlanResponse
	jobID, err := sideeffect.GenerateUUID(ctx)
	if err != nil {
		return response, errors.Wrap(err, "generating job id")
	}

	// fail if we error here since all successive calls to update this will fail otherwise
	if err := r.Store.InitPlanJob(jobID, serverURL); err != nil {
		return response, errors.Wrap(err, "initializing job")
	}

	if err := r.Store.UpdatePlanJobWithStatus(state.InProgressJobStatus); err != nil {
		return response, newUpdateJobError(err, "unable to update job with in-progress status")
	}

	response, err = r.JobRunner.Plan(ctx, root, jobID.String())

	if err != nil {
		if e := r.Store.UpdatePlanJobWithStatus(state.FailedJobStatus); e != nil {
			// not returning UpdateJobError here since we want to surface the job failure itself
			logger.Error(ctx, "unable to update job with failed status, job failed with error. ", key.ErrKey, err)
		}
		return response, errors.Wrap(err, "running job")
	}
	if err := r.Store.UpdatePlanJobWithStatus(state.SuccessJobStatus, state.UpdateOptions{
		PlanSummary: response.Summary,
	}); err != nil {
		return response, newUpdateJobError(err, "unable to update job with success status")
	}

	return response, nil
}

func (r *Runner) Apply(ctx workflow.Context, root *terraform.LocalRoot, serverURL fmt.Stringer, planResponse activities.TerraformPlanResponse) error {
	jobID, err := sideeffect.GenerateUUID(ctx)

	if err != nil {
		return errors.Wrap(err, "generating job id")
	}

	// fail if we error here since all successive calls to update this will fail otherwise
	if err := r.Store.InitApplyJob(jobID, serverURL); err != nil {
		return errors.Wrap(err, "initializing job")
	}

	planStatus, err := r.ReviewGate.Await(ctx, root.Root, planResponse.Summary)
	if err != nil {
		logger.Error(ctx, "error waiting for plan review.", key.ErrKey, err)
		return newPlanRejectedError()
	}

	if planStatus == gate.Rejected {
		if err := r.Store.UpdateApplyJobWithStatus(state.RejectedJobStatus); err != nil {
			logger.Error(ctx, "unable to update job with rejected status.", key.ErrKey, err)
		}
		return newPlanRejectedError()
	}

	if err := r.Store.UpdateApplyJobWithStatus(state.InProgressJobStatus, state.UpdateOptions{
		StartTime: time.Now(),
	}); err != nil {
		return newUpdateJobError(err, "unable to update job with success status")
	}

	err = r.JobRunner.Apply(ctx, root, jobID.String(), planResponse.PlanFile)
	if err != nil {

		if err := r.Store.UpdateApplyJobWithStatus(state.FailedJobStatus, state.UpdateOptions{
			EndTime: time.Now(),
		}); err != nil {
			// not returning UpdateJobError here since we want to surface the job failure itself
			logger.Error(ctx, "unable to update job with failed status, job failed with error. ", key.ErrKey, err)
		}
		return errors.Wrap(err, "running job")
	}

	if err := r.Store.UpdateApplyJobWithStatus(state.SuccessJobStatus, state.UpdateOptions{
		EndTime: time.Now(),
	}); err != nil {
		return newUpdateJobError(err, "unable to update job with success status")
	}

	return nil
}

func (r *Runner) Run(ctx workflow.Context) error {
	var err error
	// make sure we are updating state on completion.
	defer func() {
		reason := state.SuccessfulCompletionReason

		if r := recover(); r != nil || err != nil {
			reason = state.InternalServiceError
		}

		// check for any timeouts that percolated up
		var timeoutErr *temporal.TimeoutError
		if errors.As(err, &timeoutErr) {
			switch timeoutErr.TimeoutType() {
			case enums.TIMEOUT_TYPE_HEARTBEAT, enums.TIMEOUT_TYPE_START_TO_CLOSE:
				reason = state.TimedOutError
			}
		}

		updateErr := r.Store.UpdateCompletion(state.WorkflowResult{
			Status: state.CompleteWorkflowStatus,
			Reason: reason,
		})
		if updateErr != nil {
			logger.Warn(ctx, "error updating completion status", key.ErrKey, err)
		}
	}()
	err = r.run(ctx)
	return err
}

func (r *Runner) run(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: StartToCloseTimeout,
		HeartbeatTimeout:    HeartBeatTimeout,
	})
	var response *activities.GetWorkerInfoResponse
	err := workflow.ExecuteActivity(ctx, r.TerraformActivities.GetWorkerInfo).Get(ctx, &response)
	if err != nil {
		return r.toExternalError(err, "getting worker info")
	}

	opts := workflow.GetActivityOptions(ctx)
	opts.TaskQueue = response.TaskQueue
	ctx = workflow.WithActivityOptions(ctx, opts)

	root, cleanup, err := r.RootFetcher.Fetch(ctx)
	if err != nil {
		return r.toExternalError(err, "fetching root")
	}
	defer func() {
		r.executeCleanup(ctx, cleanup)
	}()

	planResponse, err := r.Plan(ctx, root, response.ServerURL)
	if err != nil {
		return r.toExternalError(err, "running plan job")
	}

	if err := r.Apply(ctx, root, response.ServerURL, planResponse); err != nil {
		return r.toExternalError(err, "running apply job")
	}
	return nil
}

func (r *Runner) executeCleanup(ctx workflow.Context, handlers ...func(workflow.Context) error) {
	// create a new disconnected ctx since we want this run even in the event of
	// cancellation
	if temporal.IsCanceledError(ctx.Err()) {
		var cancel workflow.CancelFunc
		ctx, cancel = workflow.NewDisconnectedContext(ctx)
		defer cancel()
	}

	for _, h := range handlers {
		cleanupErr := h(ctx)

		if cleanupErr != nil {
			logger.Warn(ctx, "error cleaning up local root", key.ErrKey, cleanupErr)
		}
	}
}

type ApplicationError struct {
	ErrType string
	Msg     string
}

func (e ApplicationError) Error() string {
	return fmt.Sprintf("Type: %s, Msg: %s", e.ErrType, e.Msg)
}

func (e ApplicationError) ToTemporalApplicationError() error {
	return temporal.NewApplicationError(
		e.Msg,
		e.ErrType,
		e,
	)
}

// toExternalError allows callers of the workflow to handle specific error types
// in whatever fashion.
// we use errors.As here to ensure that we're accounting for wrapped errors
func (r *Runner) toExternalError(err error, msg string) error {
	var planRejected PlanRejectedError
	if errors.As(err, &planRejected) {
		e := ApplicationError{
			ErrType: planRejected.GetExternalType(),
			Msg:     errors.Wrap(err, msg).Error(),
		}
		return e.ToTemporalApplicationError()
	}

	var updateJobErr UpdateJobError
	if errors.As(err, &updateJobErr) {
		e := ApplicationError{
			ErrType: updateJobErr.GetExternalType(),
			// wrap original error to provide job type
			Msg: errors.Wrap(err, msg).Error(),
		}

		return e.ToTemporalApplicationError()
	}

	return errors.Wrap(err, msg)
}
