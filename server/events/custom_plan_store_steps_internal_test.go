// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRequiresAtlantisManagedPlanFileForCustomPlanStoreSteps(t *testing.T) {
	tests := []struct {
		name     string
		workflow valid.Workflow
		want     bool
	}{
		{
			name: "custom run steps only",
			workflow: valid.Workflow{
				Plan:  valid.Stage{Steps: []valid.Step{{StepName: "run"}}},
				Apply: valid.Stage{Steps: []valid.Step{{StepName: "run"}}},
			},
		},
		{
			name: "plan-producing run step",
			workflow: valid.Workflow{
				Plan: valid.Stage{Steps: []valid.Step{{StepName: "run", PlanStore: &valid.RunPlanStore{Mode: valid.RunPlanStoreSaveMode}}}},
			},
			want: true,
		},
		{
			name: "plan-consuming run step",
			workflow: valid.Workflow{
				Apply: valid.Stage{Steps: []valid.Step{{StepName: "run", PlanStore: &valid.RunPlanStore{Mode: valid.RunPlanStoreConsumeMode}}}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Equals(t, tt.want, requiresAtlantisManagedPlanFile(tt.workflow))
		})
	}
}
