package yaml_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
	yamlv2 "gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         yaml.Spec
		expErr      string
	}{
		{
			description: "no data",
			input:       "",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "yaml nil",
			input:       "~",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "invalid key",
			input:       "invalid: key",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 1: field invalid not found in struct yaml.Spec",
		},
		{
			description: "version set",
			input:       "version: 2",
			exp: yaml.Spec{
				Version:   Int(2),
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "projects key without value",
			input:       "projects:",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "workflows key without value",
			input:       "workflows:",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "projects with a map",
			input:       "projects:\n  key: value",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!map into []yaml.Project",
		},
		{
			description: "projects with a scalar",
			input:       "projects: value",
			exp: yaml.Spec{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `value` into []yaml.Project",
		},
		{
			description: "should use values if set",
			input: `
version: 2
projects:
- dir: mydir
  workspace: myworkspace
  workflow: default
  terraform_version: v0.11.0
  auto_plan:
    enabled: false
    when_modified: []
  apply_requirements: [mergeable]
workflows:
  default:
    plan:
      steps: []
    apply:
     steps: []`,
			exp: yaml.Spec{
				Version: Int(2),
				Projects: []yaml.Project{
					{
						Dir:              String("mydir"),
						Workspace:        String("myworkspace"),
						Workflow:         String("default"),
						TerraformVersion: String("v0.11.0"),
						AutoPlan: &yaml.AutoPlan{
							WhenModified: []string{},
							Enabled:      Bool(false),
						},
						ApplyRequirements: []string{"mergeable"},
					},
				},
				Workflows: map[string]yaml.Workflow{
					"default": {
						Apply: &yaml.Stage{
							Steps: []yaml.Step{},
						},
						Plan: &yaml.Stage{
							Steps: []yaml.Step{},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var conf yaml.Spec
			err := yamlv2.UnmarshalStrict([]byte(c.input), &conf)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, conf)
		})
	}
}
