package yaml_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
	yamlv2 "gopkg.in/yaml.v2"
)

func TestWorkflow_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         yaml.Workflow
		expErr      string
	}{
		{
			description: "should use defaults if set to null",
			input:       `~`,
			exp: yaml.Workflow{
				Apply: &yaml.Stage{
					Steps: []yaml.StepConfig{
						{
							StepType: "apply",
						},
					},
				},
				Plan: &yaml.Stage{
					Steps: []yaml.StepConfig{
						{
							StepType: "init",
						},
						{
							StepType: "plan",
						},
					},
				},
			},
		},
		{
			description: "should use set values",
			input: `
plan:
  steps:
  - plan
apply:
  steps: []
`,
			exp: yaml.Workflow{
				Apply: &yaml.Stage{
					Steps: []yaml.StepConfig{},
				},
				Plan: &yaml.Stage{
					Steps: []yaml.StepConfig{
						{
							StepType: "plan",
						},
					},
				},
			},
		},
		{
			description: "should use defaults for apply if only plan set",
			input: `
plan:
  steps: []`,
			exp: yaml.Workflow{
				Apply: &yaml.Stage{
					Steps: []yaml.StepConfig{
						{
							StepType: "apply",
						},
					},
				},
				Plan: &yaml.Stage{
					Steps: []yaml.StepConfig{},
				},
			},
		},
		{
			description: "should use defaults for plan if only apply set",
			input: `
apply:
  steps: []`,
			exp: yaml.Workflow{
				Apply: &yaml.Stage{
					Steps: []yaml.StepConfig{},
				},
				Plan: &yaml.Stage{
					Steps: []yaml.StepConfig{
						{
							StepType: "init",
						},
						{
							StepType: "plan",
						},
					},
				},
			},
		},
		{
			description: "should error if no steps key specified",
			input: `
apply:
- apply`,
			expErr: "missing \"steps\" key",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var w yaml.Workflow
			err := yamlv2.Unmarshal([]byte(c.input), &w)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, w)
		})
	}
}
