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

type applyActivities interface {
	TerraformApply(ctx context.Context, request activities.TerraformApplyRequest) (activities.TerraformApplyResponse, error)
}

type ApplyStepRunner struct {
	Activity applyActivities
}

func (r *ApplyStepRunner) Run(executionContext *job.ExecutionContext, _ *root.LocalRoot, step job.Step) (string, error) {
	args, err := terraform.NewArgumentList(step.ExtraArgs)

	if err != nil {
		return "", errors.Wrapf(err, "creating argument list")
	}
	var resp activities.TerraformApplyResponse
	err = workflow.ExecuteActivity(executionContext.Context, r.Activity.TerraformApply, activities.TerraformApplyRequest{
		Args:      args,
		Envs:      executionContext.Envs,
		TfVersion: executionContext.TfVersion,
		Path:      executionContext.Path,
		JobID:     executionContext.JobID,
	}).Get(executionContext, &resp)
	if err != nil {
		return "", errors.Wrap(err, "running terraform apply activity")
	}
	return resp.Output, nil
}
