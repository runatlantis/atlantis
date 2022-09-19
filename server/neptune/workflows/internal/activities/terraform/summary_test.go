package terraform_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
	"github.com/stretchr/testify/assert"
)

func TestSummary(t *testing.T) {
	plan := "{\"format_version\": \"1.0\",\"resource_changes\":[{\"change\":{\"actions\":[\"update\"]},\"address\":\"type.resource_update\"},{\"change\":{\"actions\":[\"create\"]},\"address\":\"type.resource_create\"}, {\"change\":{\"actions\":[\"delete\"]},\"address\":\"type.resource_delete\"}]}"

	t.Run("success", func(t *testing.T) {
		summary, err := terraform.NewPlanSummaryFromJSON([]byte(plan))
		assert.NoError(t, err)

		assert.Equal(t, terraform.PlanSummary{
			Creations: []terraform.ResourceSummary{
				{
					Address: "type.resource_create",
				},
			},
			Updates: []terraform.ResourceSummary{
				{
					Address: "type.resource_update",
				},
			},
			Deletions: []terraform.ResourceSummary{
				{
					Address: "type.resource_delete",
				},
			},
		}, summary)
	})

	t.Run("error", func(t *testing.T) {
		_, err := terraform.NewPlanSummaryFromJSON([]byte("{{"))
		assert.Error(t, err)
	})
}

func TestSummary_replace(t *testing.T) {
	plan := "{\"format_version\": \"1.0\",\"resource_changes\":[{\"change\":{\"actions\":[\"create\", \"delete\"]},\"address\":\"type.resource_replace\"}]}"
	summary, err := terraform.NewPlanSummaryFromJSON([]byte(plan))
	assert.NoError(t, err)

	assert.Equal(t, terraform.PlanSummary{
		Creations: []terraform.ResourceSummary{
			{
				Address: "type.resource_replace",
			},
		},
		Deletions: []terraform.ResourceSummary{
			{
				Address: "type.resource_replace",
			},
		},
	}, summary)
}
