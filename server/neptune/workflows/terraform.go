package workflows

import (
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

func Terraform(ctx workflow.Context, request TerraformRequest) error {
	return terraform.Workflow(ctx, request)
}
