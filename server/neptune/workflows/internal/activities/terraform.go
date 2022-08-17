package activities

import "context"

type terraformActivities struct{}

// Terraform Init

type TerraformInitRequest struct {
}

func (t *terraformActivities) TerraformInit(ctx context.Context, request TerraformInitRequest) error {
	return nil
}

// Terraform Plan

type TerraformPlanRequest struct {
}

func (t *terraformActivities) TerraformPlan(ctx context.Context, request TerraformPlanRequest) error {
	return nil
}

// Terraform Apply

type TerraformApplyRequest struct {
}

func (t *terraformActivities) TerraformApply(ctx context.Context, request TerraformApplyRequest) error {
	return nil
}
