package yaml_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
	yamlv2 "gopkg.in/yaml.v2"
)

func TestAutoPlan_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         yaml.AutoPlan
	}{
		{
			description: "should use defaults",
			input: `
`,
			exp: yaml.AutoPlan{
				Enabled:      false,
				WhenModified: nil,
			},
		},
		{
			description: "should use all set fields",
			input: `
enabled: true
when_modified: ["something-else"]
`,
			exp: yaml.AutoPlan{
				Enabled:      true,
				WhenModified: []string{"something-else"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a yaml.AutoPlan
			err := yamlv2.Unmarshal([]byte(c.input), &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}
