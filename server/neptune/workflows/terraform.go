package workflows

import (
	"net/url"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/workflow"
)

// Export anything that callers need such as requests, signals, etc.
type TerraformRequest = terraform.Request

type TerraformActivities struct {
	activities.Terraform
}

func NewTerraformActivities(scope tally.Scope, serverURL *url.URL) (*TerraformActivities, error) {
	terraformActivities := activities.NewTerraform(serverURL)
	return &TerraformActivities{
		Terraform: *terraformActivities,
	}, nil
}

func Terraform(ctx workflow.Context, request TerraformRequest) error {
	return terraform.Workflow(ctx, request)
}
