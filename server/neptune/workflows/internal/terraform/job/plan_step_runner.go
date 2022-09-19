package job

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"go.temporal.io/sdk/workflow"
)

type planActivities interface {
	TerraformPlan(ctx context.Context, request activities.TerraformPlanRequest) (activities.TerraformPlanResponse, error)
}

type PlanStepRunner struct {
	Activity planActivities
}

func (r *PlanStepRunner) Run(executionContext *job.ExecutionContext, _ *root.LocalRoot, step job.Step) (string, error) {
	args, err := terraform.NewArgumentList(step.ExtraArgs)

	if err != nil {
		return "", errors.Wrapf(err, "creating argument list")
	}
	var resp activities.TerraformPlanResponse
	err = workflow.ExecuteActivity(executionContext.Context, r.Activity.TerraformPlan, activities.TerraformPlanRequest{
		Args:      args,
		Envs:      executionContext.Envs,
		TfVersion: executionContext.TfVersion,
		JobID:     executionContext.JobID,
		Path:      executionContext.Path,
	}).Get(executionContext, &resp)
	if err != nil {
		return "", errors.Wrap(err, "running terraform plan activity")
	}
	return resp.Output, nil
}
