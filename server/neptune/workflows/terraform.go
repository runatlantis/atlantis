package workflows

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/config"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"go.temporal.io/sdk/workflow"
)

// Export anything that callers need such as requests, signals, etc.
type TerraformRequest = terraform.Request

type TerraformActivities struct {
	activities.Terraform
}

func NewTerraformActivities(config config.TerraformConfig, dataDir string, serverURL *url.URL) (*TerraformActivities, error) {
	terraformActivities, err := activities.NewTerraform(config, dataDir, serverURL)
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
