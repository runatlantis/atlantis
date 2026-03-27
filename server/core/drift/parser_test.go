// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/drift"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestParser_HasDrift(t *testing.T) {
	parser := drift.NewParser()

	cases := []struct {
		name     string
		plan     *models.PlanSuccess
		expected bool
	}{
		{
			name:     "nil plan",
			plan:     nil,
			expected: false,
		},
		{
			name: "no drift",
			plan: &models.PlanSuccess{
				TerraformOutput: "Plan: 0 to add, 0 to change, 0 to destroy.",
			},
			expected: false,
		},
		{
			name: "has additions",
			plan: &models.PlanSuccess{
				TerraformOutput: "Plan: 3 to add, 0 to change, 0 to destroy.",
			},
			expected: true,
		},
		{
			name: "has changes",
			plan: &models.PlanSuccess{
				TerraformOutput: "Plan: 0 to add, 2 to change, 0 to destroy.",
			},
			expected: true,
		},
		{
			name: "has destroys",
			plan: &models.PlanSuccess{
				TerraformOutput: "Plan: 0 to add, 0 to change, 1 to destroy.",
			},
			expected: true,
		},
		{
			name: "has imports",
			plan: &models.PlanSuccess{
				TerraformOutput: "Plan: 1 to import, 0 to add, 0 to change, 0 to destroy.",
			},
			expected: true,
		},
		{
			name: "mixed drift",
			plan: &models.PlanSuccess{
				TerraformOutput: "Plan: 2 to add, 3 to change, 1 to destroy.",
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := parser.HasDrift(c.plan)
			Equals(t, c.expected, result)
		})
	}
}

func TestParser_ParsePlanOutput(t *testing.T) {
	parser := drift.NewParser()

	t.Run("with drift", func(t *testing.T) {
		plan := &models.PlanSuccess{
			TerraformOutput: "Plan: 2 to add, 3 to change, 1 to destroy.",
		}
		result := parser.ParsePlanOutput(plan)

		Equals(t, true, result.HasDrift)
		Equals(t, 2, result.ToAdd)
		Equals(t, 3, result.ToChange)
		Equals(t, 1, result.ToDestroy)
	})

	t.Run("without drift", func(t *testing.T) {
		plan := &models.PlanSuccess{
			TerraformOutput: "Plan: 0 to add, 0 to change, 0 to destroy.",
		}
		result := parser.ParsePlanOutput(plan)

		Equals(t, false, result.HasDrift)
		Equals(t, 0, result.ToAdd)
		Equals(t, 0, result.ToChange)
		Equals(t, 0, result.ToDestroy)
	})

	t.Run("nil plan", func(t *testing.T) {
		result := parser.ParsePlanOutput(nil)
		Equals(t, models.DriftSummary{}, result)
	})
}

func TestParser_ParsePlanStats(t *testing.T) {
	parser := drift.NewParser()

	stats := models.PlanSuccessStats{
		Add:            5,
		Change:         3,
		Destroy:        2,
		Import:         1,
		ChangesOutside: true,
	}

	result := parser.ParsePlanStats(stats, "Plan: 5 to add, 3 to change, 2 to destroy, 1 to import")

	Equals(t, true, result.HasDrift)
	Equals(t, 5, result.ToAdd)
	Equals(t, 3, result.ToChange)
	Equals(t, 2, result.ToDestroy)
	Equals(t, 1, result.ToImport)
	Equals(t, true, result.ChangesOutside)
	Equals(t, "Plan: 5 to add, 3 to change, 2 to destroy, 1 to import", result.Summary)
}
