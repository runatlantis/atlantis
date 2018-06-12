package yaml_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
	yamlv2 "gopkg.in/yaml.v2"
)

func TestStage_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         yaml.Stage
	}{
		{
			description: "empty",
			input:       "",
			exp: yaml.Stage{
				Steps: nil,
			},
		},
		{
			description: "all fields set",
			input: `
steps: [step1]
`,
			exp: yaml.Stage{
				Steps: []yaml.Step{
					{
						Key: String("step1"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a yaml.Stage
			err := yamlv2.UnmarshalStrict([]byte(c.input), &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}
