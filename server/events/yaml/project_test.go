package yaml_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
	yamlv2 "gopkg.in/yaml.v2"
)

func TestProject_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         yaml.Project
	}{
		{
			description: "omit unset fields",
			input:       "",
			exp: yaml.Project{
				Dir:               nil,
				Workspace:         nil,
				Workflow:          nil,
				TerraformVersion:  nil,
				AutoPlan:          nil,
				ApplyRequirements: nil,
			},
		},
		{
			description: "all fields set",
			input: `
dir: mydir
workspace: workspace
workflow: workflow
terraform_version: v0.11.0
auto_plan:
  when_modified: []
  enabled: false
apply_requirements:
- mergeable`,
			exp: yaml.Project{
				Dir:              String("mydir"),
				Workspace:        String("workspace"),
				Workflow:         String("workflow"),
				TerraformVersion: String("v0.11.0"),
				AutoPlan: &yaml.AutoPlan{
					WhenModified: []string{},
					Enabled:      Bool(false),
				},
				ApplyRequirements: []string{"mergeable"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var p yaml.Project
			err := yamlv2.UnmarshalStrict([]byte(c.input), &p)
			Ok(t, err)
			Equals(t, c.exp, p)
		})
	}
}
