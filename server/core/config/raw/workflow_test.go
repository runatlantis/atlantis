package raw_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v3"
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
			err := yaml.UnmarshalStrict([]byte(c.input), &w)
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
