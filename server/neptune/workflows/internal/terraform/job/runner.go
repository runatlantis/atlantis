package job

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"go.temporal.io/sdk/workflow"
)

type terraformActivities interface {
	TerraformInit(ctx context.Context, request activities.TerraformInitRequest) (activities.TerraformInitResponse, error)
	TerraformPlan(ctx context.Context, request activities.TerraformPlanRequest) (activities.TerraformPlanResponse, error)
	TerraformApply(ctx context.Context, request activities.TerraformApplyRequest) (activities.TerraformApplyResponse, error)
}

// stepRunner runs individual run steps
type stepRunner interface {
	Run(executionContext *job.ExecutionContext, localRoot *root.LocalRoot, step job.Step) (string, error)
}

type jobRunner struct {
	Activity      terraformActivities
	EnvStepRunner stepRunner
	CmdStepRunner stepRunner
}

func NewRunner(runStepRunner stepRunner, envStepRunner stepRunner) *jobRunner {
	return &jobRunner{
		CmdStepRunner: runStepRunner,
		EnvStepRunner: envStepRunner,
	}
}

func (r *jobRunner) Plan(ctx workflow.Context, localRoot *root.LocalRoot, jobID string) (activities.TerraformPlanResponse, error) {
	// Execution ctx for a job that handles setting up the env vars from the previous steps
	jobCtx := &job.ExecutionContext{
		Context:   ctx,
		Path:      localRoot.Path,
		Envs:      map[string]string{},
		TfVersion: localRoot.Root.TfVersion,
		JobID:     jobID,
	}

	var resp activities.TerraformPlanResponse

	for _, step := range localRoot.Root.Plan.Steps {
		var err error
		switch step.StepName {
		case "init":
			err = r.init(jobCtx, localRoot, step)
		case "plan":
			resp, err = r.plan(jobCtx, step.ExtraArgs)
		}

		if err != nil {
			return resp, errors.Wrapf(err, "running step %s", step.StepName)
		}

		err = r.runOptionalSteps(jobCtx, localRoot, step)

		if err != nil {
			return resp, errors.Wrapf(err, "running step %s", step.StepName)
		}
	}

	return resp, nil
}

func (r *jobRunner) Apply(ctx workflow.Context, localRoot *root.LocalRoot, jobID string, planFile string) error {
	// Execution ctx for a job that handles setting up the env vars from the previous steps
	jobCtx := &job.ExecutionContext{
		Context:   ctx,
		Path:      localRoot.Path,
		Envs:      map[string]string{},
		TfVersion: localRoot.Root.TfVersion,
		JobID:     jobID,
	}

	for _, step := range localRoot.Root.Apply.Steps {
		var err error
		switch step.StepName {
		case "apply":
			err = r.apply(jobCtx, planFile, step)
		}

		if err != nil {
			return errors.Wrapf(err, "running step %s", step.StepName)
		}

		err = r.runOptionalSteps(jobCtx, localRoot, step)
		if err != nil {
			return errors.Wrapf(err, "running step %s", step.StepName)
		}
	}

	return nil
}

func (r *jobRunner) apply(ctx *job.ExecutionContext, planFile string, step job.Step) error {
	args, err := terraform.NewArgumentList(step.ExtraArgs)

	if err != nil {
		return errors.Wrapf(err, "creating argument list")
	}
	var resp activities.TerraformApplyResponse
	err = workflow.ExecuteActivity(ctx, r.Activity.TerraformApply, activities.TerraformApplyRequest{
		Args:      args,
		Envs:      ctx.Envs,
		TfVersion: ctx.TfVersion,
		Path:      ctx.Path,
		JobID:     ctx.JobID,
		PlanFile:  planFile,
	}).Get(ctx, &resp)
	if err != nil {
		return errors.Wrap(err, "running terraform apply activity")
	}
	return nil
}

func (r *jobRunner) plan(ctx *job.ExecutionContext, extraArgs []string) (activities.TerraformPlanResponse, error) {
	var resp activities.TerraformPlanResponse

	args, err := terraform.NewArgumentList(extraArgs)
	if err != nil {
		return resp, errors.Wrapf(err, "creating argument list")
	}
	err = workflow.ExecuteActivity(ctx, r.Activity.TerraformPlan, activities.TerraformPlanRequest{
		Args:      args,
		Envs:      ctx.Envs,
		TfVersion: ctx.TfVersion,
		JobID:     ctx.JobID,
		Path:      ctx.Path,
	}).Get(ctx, &resp)
	if err != nil {
		return resp, errors.Wrap(err, "running terraform plan activity")
	}
	return resp, nil
}

func (r *jobRunner) init(ctx *job.ExecutionContext, localRoot *root.LocalRoot, step job.Step) error {
	args, err := terraform.NewArgumentList(step.ExtraArgs)

	if err != nil {
		return errors.Wrap(err, "creating argument list")
	}
	var resp activities.TerraformInitResponse
	err = workflow.ExecuteActivity(ctx.Context, r.Activity.TerraformInit, activities.TerraformInitRequest{
		Args:      args,
		Envs:      ctx.Envs,
		TfVersion: ctx.TfVersion,
		Path:      ctx.Path,
		JobID:     ctx.JobID,
	}).Get(ctx, &resp)
	if err != nil {
		return errors.Wrap(err, "running terraform init activity")
	}
	return nil
}

func (r *jobRunner) runOptionalSteps(ctx *job.ExecutionContext, localRoot *root.LocalRoot, step job.Step) error {
	switch step.StepName {
	case "run":
		_, err := r.CmdStepRunner.Run(ctx, localRoot, step)
		return err
	case "env":
		o, err := r.EnvStepRunner.Run(ctx, localRoot, step)
		ctx.Envs[step.EnvVarName] = o
		// output of env step doesn't need to be returned.
		return err
	}

	return nil
}
