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
		exp         yaml.Config
	}{
		{
			description: "should be empty if nothing set",
			input:       `~`,
			exp: yaml.Config{
				Version:   0,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "should use values if set",
			input: `
version: 2
projects:
- dir: mydir
  workspace: myworkspace
  workflow: default
workflows:
  default:
    plan:
      steps: []
    apply:
     steps: []`,
			exp: yaml.Config{
				Version: 2,
				Projects: []yaml.Project{
					{
						Dir:       "mydir",
						Workflow:  "default",
						Workspace: "myworkspace",
						AutoPlan: &yaml.AutoPlan{
							WhenModified: []string{"**/*.tf"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]yaml.Workflow{
					"default": {
						Apply: &yaml.Stage{
							Steps: []yaml.StepConfig{},
						},
						Plan: &yaml.Stage{
							Steps: []yaml.StepConfig{},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var conf yaml.Config
			err := yamlv2.Unmarshal([]byte(c.input), &conf)
			Ok(t, err)
			Equals(t, c.exp, conf)
		})
	}
}
