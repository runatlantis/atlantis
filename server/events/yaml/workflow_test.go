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
			description: "empty",
			input:       ``,
			exp: yaml.Workflow{
				Apply: nil,
				Plan:  nil,
			},
		},
		{
			description: "yaml null",
			input:       `~`,
			exp: yaml.Workflow{
				Apply: nil,
				Plan:  nil,
			},
		},
		{
			description: "only plan/apply set",
			input: `
plan:
apply:
`,
			exp: yaml.Workflow{
				Apply: nil,
				Plan:  nil,
			},
		},
		{
			description: "steps set to null",
			input: `
plan:
  steps: ~
apply:
  steps: ~`,
			exp: yaml.Workflow{
				Plan: &yaml.Stage{
					Steps: nil,
				},
				Apply: &yaml.Stage{
					Steps: nil,
				},
			},
		},
		{
			description: "steps set to empty slice",
			input: `
plan:
  steps: []
apply:
  steps: []`,
			exp: yaml.Workflow{
				Plan: &yaml.Stage{
					Steps: []yaml.Step{},
				},
				Apply: &yaml.Stage{
					Steps: []yaml.Step{},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var w yaml.Workflow
			err := yamlv2.UnmarshalStrict([]byte(c.input), &w)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, w)
		})
	}
}
