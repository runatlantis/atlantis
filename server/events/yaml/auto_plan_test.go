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
			description: "omit unset fields",
			input:       "",
			exp: yaml.AutoPlan{
				Enabled:      nil,
				WhenModified: nil,
			},
		},
		{
			description: "all fields set",
			input: `
enabled: true
when_modified: ["something-else"]
`,
			exp: yaml.AutoPlan{
				Enabled:      Bool(true),
				WhenModified: []string{"something-else"},
			},
		},
		{
			description: "enabled false",
			input: `
enabled: false
when_modified: ["something-else"]
`,
			exp: yaml.AutoPlan{
				Enabled:      Bool(false),
				WhenModified: []string{"something-else"},
			},
		},
		{
			description: "modified elem empty",
			input: `
enabled: false
when_modified:
-
`,
			exp: yaml.AutoPlan{
				Enabled:      Bool(false),
				WhenModified: []string{""},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a yaml.AutoPlan
			err := yamlv2.UnmarshalStrict([]byte(c.input), &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}
