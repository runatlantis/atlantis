package terraform

import (
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	job_runner "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job/step/cmd"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job/step/env"
	"go.temporal.io/sdk/workflow"
)

type workflowActivities struct {
	*activities.Terraform
	*activities.Github
}

// jobRunner runs a deploy plan/apply job
type jobRunner interface {
	Run(ctx workflow.Context, job job.Job, localRoot *root.LocalRoot) (string, error)
}

type PlanStatus int
type PlanReview struct {
	Status PlanStatus
}

const (
	Approved PlanStatus = iota
	Rejected
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
	Activities *workflowActivities
	JobRunner  jobRunner
	Request    Request
}

func newRunner(ctx workflow.Context, request Request) *Runner {
	var a *workflowActivities

	cmdStepRunner := cmd.Runner{
		Activity: a,
	}
	return &Runner{
		Activities: a,
		Request:    request,
		JobRunner: job_runner.NewRunner(
			&cmdStepRunner,
			&env.Runner{
				CmdRunner: cmdStepRunner,
			},
		),
	}
}

func (r *Runner) Run(ctx workflow.Context) error {
	// Download files into disk
	var fetchRootResponse activities.FetchRootResponse
	err := workflow.ExecuteActivity(ctx, r.Activities.FetchRoot, activities.FetchRootRequest{
		Repo:         r.Request.Repo,
		Root:         r.Request.Root,
		DeploymentId: r.Request.DeploymentId,
	}).Get(ctx, &fetchRootResponse)
	if err != nil {
		return errors.Wrap(err, "executing GH repo clone")
	}

	_, err = r.JobRunner.Run(ctx, r.Request.Root.Plan, fetchRootResponse.LocalRoot)
	if err != nil {
		return errors.Wrap(err, "running plan job")
	}

	// Wait for plan review signal
	var planReview PlanReview
	signalChan := workflow.GetSignalChannel(ctx, "planreview-repo-steps")
	more := signalChan.Receive(ctx, &planReview)
	if !more {
		return errors.New("plan review signal channel cancelled")
	}
	if planReview.Status == Rejected {
		return nil
	}

	// Run apply steps
	_, err = r.JobRunner.Run(ctx, r.Request.Root.Apply, fetchRootResponse.LocalRoot)
	if err != nil {
		return errors.Wrap(err, "running apply job")
	}

	// Cleanup
	var cleanupResponse activities.CleanupResponse
	err = workflow.ExecuteActivity(ctx, r.Activities.Cleanup, activities.CleanupRequest{
		LocalRoot: fetchRootResponse.LocalRoot,
	}).Get(ctx, &cleanupResponse)
	if err != nil {
		return errors.Wrap(err, "cleaning up")
	}
	return nil
}
