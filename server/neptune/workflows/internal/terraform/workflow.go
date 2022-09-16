package terraform

import (
	"context"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/sideeffect"
	runner "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/workflow"
)

type githubActivities interface {
	FetchRoot(ctx context.Context, request activities.FetchRootRequest) (activities.FetchRootResponse, error)
}

type terraformActivities interface {
	Cleanup(ctx context.Context, request activities.CleanupRequest) (activities.CleanupResponse, error)
	GetWorkerInfo(ctx context.Context) (*activities.GetWorkerInfoResponse, error)
}

// jobRunner runs a deploy plan/apply job
type jobRunner interface {
	Run(ctx workflow.Context, localRoot *root.LocalRoot, jobInstance job.JobInstance) (string, error)
}

type PlanStatus int
type PlanReviewSignalRequest struct {
	Status PlanStatus

	// TODO: Output this info to the checks UI
	User string
}

const (
	Approved PlanStatus = iota
	Rejected
)

const (
	PlanReviewSignalName = "planreview"
)

func Workflow(ctx workflow.Context, request Request) error {
	options := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:       1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	sessionOptions := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 30 * time.Minute,
	}
	ctx, err := workflow.CreateSession(ctx, sessionOptions)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(ctx)

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
}

func newRunner(ctx workflow.Context, request Request) *Runner {
	var ta *activities.Terraform
	var ga *activities.Github

	cmdStepRunner := runner.CmdStepRunner{
		Activity: ta,
	}

	parent := workflow.GetInfo(ctx).ParentWorkflowExecution

	return &Runner{
		GithubActivities:    ga,
		TerraformActivities: ta,
		Request:             request,
		JobRunner: runner.NewRunner(
			&cmdStepRunner,
			&runner.EnvStepRunner{
				CmdStepRunner: cmdStepRunner,
			},
			&runner.InitStepRunner{
				Activity: ta,
			},
			&runner.PlanStepRunner{
				Activity: ta,
			},
			&runner.ApplyStepRunner{
				Activity: ta,
			},
		),
		RootFetcher: &RootFetcher{
			Request: request,
			Ga:      ga,
			Ta:      ta,
		},
		Store: state.NewWorkflowStore(
			func(s *state.Workflow) error {
				return workflow.SignalExternalWorkflow(ctx, parent.ID, parent.RunID, state.WorkflowStateChangeSignal, s).Get(ctx, nil)
			},
		),
	}
}

func (r *Runner) Plan(ctx workflow.Context, root *root.LocalRoot, serverURL *url.URL) error {
	jobID, err := sideeffect.GenerateUUID(ctx)
	if err != nil {
		return errors.Wrap(err, "generating job id")
	}

	if err := r.Store.InitPlanJob(jobID, serverURL); err != nil {
		return errors.Wrap(err, "initializing job")
	}

	if err := r.Store.UpdatePlanJobWithStatus(state.InProgressJobStatus); err != nil {
		return errors.Wrap(err, "updating job with in-progress status")
	}

	_, err = r.JobRunner.Run(ctx, root, job.JobInstance{
		Job:   r.Request.Root.Plan,
		JobID: jobID.String(),
	})

	if err != nil {
		if e := r.Store.UpdatePlanJobWithStatus(state.FailedJobStatus); e != nil {
			logger.Error(ctx, "unable to update job with failed status, job failed with error. ", err)
			return errors.Wrap(e, "updating job with failed status")
		}
		return errors.Wrap(err, "running job")
	}
	if err := r.Store.UpdatePlanJobWithStatus(state.SuccessJobStatus); err != nil {
		logger.Error(ctx, "unable to update job with success status")
		return errors.Wrap(err, "updating job with success status")
	}

	return nil
}

func (r *Runner) Apply(ctx workflow.Context, root *root.LocalRoot, serverURL *url.URL) error {
	jobID, err := sideeffect.GenerateUUID(ctx)

	if err != nil {
		return errors.Wrap(err, "generating job id")
	}

	if err := r.Store.InitApplyJob(jobID, serverURL); err != nil {
		return errors.Wrap(err, "initializing job")
	}

	// Wait for plan review signal
	var planReview PlanReviewSignalRequest
	signalChan := workflow.GetSignalChannel(ctx, PlanReviewSignalName)
	_ = signalChan.Receive(ctx, &planReview)

	if planReview.Status == Rejected {
		if err := r.Store.UpdateApplyJobWithStatus(state.RejectedJobStatus); err != nil {
			return errors.Wrap(err, "updating job with rejected status")
		}
		return nil
	}

	if err := r.Store.UpdateApplyJobWithStatus(state.InProgressJobStatus); err != nil {
		return errors.Wrap(err, "updating job with in-progress status")
	}

	_, err = r.JobRunner.Run(ctx, root, job.JobInstance{
		Job:   r.Request.Root.Apply,
		JobID: jobID.String(),
	})
	if err != nil {

		if err := r.Store.UpdateApplyJobWithStatus(state.FailedJobStatus); err != nil {
			return errors.Wrap(err, "updating job with failed status")
		}
		return errors.Wrap(err, "running job")
	}

	if err := r.Store.UpdateApplyJobWithStatus(state.SuccessJobStatus); err != nil {
		return errors.Wrap(err, "updating job with success status")
	}

	return nil
}

func (r *Runner) Run(ctx workflow.Context) error {
	var response *activities.GetWorkerInfoResponse
	err := workflow.ExecuteActivity(ctx, r.TerraformActivities.GetWorkerInfo).Get(ctx, &response)

	if err != nil {
		return errors.Wrap(err, "getting worker info")
	}

	root, cleanup, err := r.RootFetcher.Fetch(ctx)
	defer func() {
		err := cleanup()

		if err != nil {
			logger.Warn(ctx, "error cleaning up local root", "err", err)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "fetching root")
	}

	if err := r.Plan(ctx, root, response.ServerURL); err != nil {
		return errors.Wrap(err, "running plan job")
	}

	if err := r.Apply(ctx, root, response.ServerURL); err != nil {
		return errors.Wrap(err, "running apply job")
	}

	return nil
}
