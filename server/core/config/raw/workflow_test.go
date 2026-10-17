// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestWorkflow_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Workflow
		expErr      string
	}{
		{
			description: "empty",
			input:       ``,
			exp: raw.Workflow{
				Apply:       nil,
				PolicyCheck: nil,
				Plan:        nil,
			},
		},
		{
			description: "yaml null",
			input:       `~`,
			exp: raw.Workflow{
				Apply:       nil,
				PolicyCheck: nil,
				Plan:        nil,
			},
		},
		{
			description: "only plan/apply set",
			input: `
plan:
apply:
`,
			exp: raw.Workflow{
				Apply: nil,
				Plan:  nil,
			},
		},
		{
			description: "only plan/policy_check/apply set",
			input: `
plan:
policy_check:
apply:
`,
			exp: raw.Workflow{
				Apply:       nil,
				PolicyCheck: nil,
				Plan:        nil,
			},
		},
		{
			description: "steps set to null",
			input: `
plan:
  steps: ~
policy_check:
  steps: ~
apply:
  steps: ~`,
			exp: raw.Workflow{
				Plan: &raw.Stage{
					Steps: nil,
				},
				PolicyCheck: &raw.Stage{
					Steps: nil,
				},
				Apply: &raw.Stage{
					Steps: nil,
				},
			},
		},
		{
			description: "steps set to empty slice",
			input: `
plan:
  steps: []
policy_check:
  steps: []
apply:
  steps: []`,
			exp: raw.Workflow{
				Plan: &raw.Stage{
					Steps: []raw.Step{},
				},
				PolicyCheck: &raw.Stage{
					Steps: []raw.Step{},
				},
				Apply: &raw.Stage{
					Steps: []raw.Step{},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var w raw.Workflow
			err := unmarshalString(c.input, &w)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, w)
		})
	}
}

func TestWorkflow_Validate(t *testing.T) {
	// Should call the validate of Stage.
	w := raw.Workflow{
		Apply: &raw.Stage{
			Steps: []raw.Step{
				{
					Key: String("invalid"),
				},
			},
		},
	}
	validation.ErrorTag = "yaml"
	ErrEquals(t, "apply: (steps: (0: \"invalid\" is not a valid step type, maybe you omitted the 'run' key.).).", w.Validate())

	// Unset keys should validate.
	Ok(t, (raw.Workflow{}).Validate())
}

func TestWorkflow_ValidatePlanStoreStepPlacement(t *testing.T) {
	tests := []struct {
		name     string
		workflow raw.Workflow
		wantErr  string
	}{
		{
			name: "custom steps in matching stages",
			workflow: raw.Workflow{
				Plan:  &raw.Stage{Steps: []raw.Step{runPlanStoreStep("save")}},
				Apply: &raw.Stage{Steps: []raw.Step{runPlanStoreStep("consume")}},
			},
		},
		{
			name: "save in apply stage",
			workflow: raw.Workflow{
				Apply: &raw.Stage{Steps: []raw.Step{runPlanStoreStep("save")}},
			},
			wantErr: `apply: run step plan_store mode "save" is only valid in the plan stage`,
		},
		{
			name: "consume in plan stage",
			workflow: raw.Workflow{
				Plan: &raw.Stage{Steps: []raw.Step{runPlanStoreStep("consume")}},
			},
			wantErr: `plan: run step plan_store mode "consume" is only valid in the apply stage`,
		},
		{
			name: "duplicate save markers",
			workflow: raw.Workflow{
				Plan: &raw.Stage{Steps: []raw.Step{runPlanStoreStep("save"), runPlanStoreStep("save")}},
			},
			wantErr: `plan: run step plan_store mode "save" may only be configured once`,
		},
		{
			name: "duplicate consume markers",
			workflow: raw.Workflow{
				Apply: &raw.Stage{Steps: []raw.Step{runPlanStoreStep("consume"), runPlanStoreStep("consume")}},
			},
			wantErr: `apply: run step plan_store mode "consume" may only be configured once`,
		},
		{
			name: "store plan with built-in plan",
			workflow: raw.Workflow{
				Plan: &raw.Stage{Steps: []raw.Step{{Key: String("plan")}, runPlanStoreStep("save")}},
			},
			wantErr: `plan: run step plan_store mode "save" cannot be combined with the built-in "plan" step`,
		},
		{
			name: "remove plan with built-in apply",
			workflow: raw.Workflow{
				Apply: &raw.Stage{Steps: []raw.Step{{Key: String("apply")}, runPlanStoreStep("consume")}},
			},
			wantErr: `apply: run step plan_store mode "consume" cannot be combined with the built-in "apply" step`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Validate()
			if tt.wantErr == "" {
				Ok(t, err)
				return
			}
			ErrEquals(t, tt.wantErr, err)
		})
	}
}

func runPlanStoreStep(mode string) raw.Step {
	return raw.Step{CommandMap: EnvType{
		"run": {
			"command":    "custom",
			"plan_store": map[string]any{"mode": mode},
		},
	}}
}

func TestWorkflow_ToValid(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Workflow
		exp         valid.Workflow
	}{
		{
			description: "nothing set",
			input:       raw.Workflow{},
			exp: valid.Workflow{
				Apply:       valid.DefaultApplyStage,
				Plan:        valid.DefaultPlanStage,
				PolicyCheck: valid.DefaultPolicyCheckStage,
				Import:      valid.DefaultImportStage,
				StateRm:     valid.DefaultStateRmStage,
			},
		},
		{
			description: "fields set",
			input: raw.Workflow{
				Apply: &raw.Stage{
					Steps: []raw.Step{
						{
							Key: String("init"),
						},
					},
				},
				PolicyCheck: &raw.Stage{
					Steps: []raw.Step{
						{
							Key: String("policy_check"),
						},
					},
				},
				Plan: &raw.Stage{
					Steps: []raw.Step{
						{
							Key: String("init"),
						},
					},
				},
				Import: &raw.Stage{
					Steps: []raw.Step{
						{
							Key: String("import"),
						},
					},
				},
				StateRm: &raw.Stage{
					Steps: []raw.Step{
						{
							Key: String("state_rm"),
						},
					},
				},
			},
			exp: valid.Workflow{
				Apply: valid.Stage{
					Steps: []valid.Step{
						{
							StepName: "init",
						},
					},
				},
				PolicyCheck: valid.Stage{
					Steps: []valid.Step{
						{
							StepName: "policy_check",
						},
					},
				},
				Plan: valid.Stage{
					Steps: []valid.Step{
						{
							StepName: "init",
						},
					},
				},
				Import: valid.Stage{
					Steps: []valid.Step{
						{
							StepName: "import",
						},
					},
				},
				StateRm: valid.Stage{
					Steps: []valid.Step{
						{
							StepName: "state_rm",
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			c.exp.Name = "name"
			Equals(t, c.exp, c.input.ToValid("name"))
		})
	}
}
