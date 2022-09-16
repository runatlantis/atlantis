package workflows

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/config"

	terraform_model "github.com/runatlantis/atlantis/server/neptune/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"go.temporal.io/sdk/workflow"
)

// Export anything that callers need such as requests, signals, etc.
type TerraformRequest = terraform.Request

type TerraformPlanReviewSignalRequest = terraform.PlanReviewSignalRequest

type TerraformPlanReviewStatus = terraform.PlanStatus

const ApprovedPlanReviewStatus = terraform.Approved
const RejectedPlanReviewStatus = terraform.Rejected

const TerraformPlanReviewSignalName = terraform.PlanReviewSignalName

type TerraformActivities struct {
	activities.Terraform
}

type streamHandler interface {
	Stream(ctx context.Context, jobID string, ch <-chan terraform_model.Line) error
	Close(ctx context.Context, jobID string)
}

func NewTerraformActivities(config config.TerraformConfig, dataDir string, serverURL *url.URL, streamHandler streamHandler) (*TerraformActivities, error) {
	terraformActivities, err := activities.NewTerraform(config, dataDir, serverURL, streamHandler)
	if err != nil {
		return nil, errors.Wrap(err, "initializing terraform activities")
	}
	return &TerraformActivities{
		Terraform: *terraformActivities,
	}, nil
}

func Terraform(ctx workflow.Context, request TerraformRequest) error {
	return terraform.Workflow(ctx, request)
}
