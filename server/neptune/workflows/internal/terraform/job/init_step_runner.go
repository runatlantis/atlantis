package job

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"go.temporal.io/sdk/workflow"
)

type initActivities interface {
	TerraformInit(ctx context.Context, request activities.TerraformInitRequest) (activities.TerraformInitResponse, error)
}

type InitStepRunner struct {
	Activity initActivities
}

func (r *InitStepRunner) Run(executionContext *job.ExecutionContext, localRoot *root.LocalRoot, step job.Step) (string, error) {
	args, err := terraform.NewArgumentList(step.ExtraArgs)

	if err != nil {
		return "", errors.Wrapf(err, "creating argument list")
	}
	var resp activities.TerraformInitResponse
	err = workflow.ExecuteActivity(executionContext.Context, r.Activity.TerraformInit, activities.TerraformInitRequest{
		Args:      args,
		Envs:      executionContext.Envs,
		TfVersion: executionContext.TfVersion,
		Path:      executionContext.Path,
	}).Get(executionContext, &resp)
	if err != nil {
		return "", errors.Wrap(err, "running terraform init activity")
	}
	return resp.Output, nil
}
