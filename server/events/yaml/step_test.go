package yaml_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
	yamlv2 "gopkg.in/yaml.v2"
)

func TestStepConfig_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         yaml.Step
		expErr      string
	}{

		// Single string.
		{
			description: "single string",
			input:       `astring`,
			exp: yaml.Step{
				Key: String("astring"),
			},
		},

		// map[string]map[string][]string i.e. extra_args style.
		{
			description: "extra_args style",
			input: `
key:
  mapValue: [arg1, arg2]`,
			exp: yaml.Step{
				Map: map[string]map[string][]string{
					"key": {
						"mapValue": {"arg1", "arg2"},
					},
				},
			},
		},
		{
			description: "extra_args style multiple keys",
			input: `
key:
  mapValue: [arg1, arg2]
  value2: []`,
			exp: yaml.Step{
				Map: map[string]map[string][]string{
					"key": {
						"mapValue": {"arg1", "arg2"},
						"value2":   {},
					},
				},
			},
		},
		{
			description: "extra_args style multiple top-level keys",
			input: `
key:
  val1: []
key2:
  val2: []`,
			exp: yaml.Step{
				Map: map[string]map[string][]string{
					"key": {
						"val1": {},
					},
					"key2": {
						"val2": {},
					},
				},
			},
		},

		// Run-step style
		{
			description: "run step",
			input: `
run: my command`,
			exp: yaml.Step{
				StringVal: map[string]string{
					"run": "my command",
				},
			},
		},
		{
			description: "run step multiple top-level keys",
			input: `
run: my command
key: value`,
			exp: yaml.Step{
				StringVal: map[string]string{
					"run": "my command",
					"key": "value",
				},
			},
		},

		// Empty
		{
			description: "empty",
			input:       "",
			exp: yaml.Step{
				Key:       nil,
				Map:       nil,
				StringVal: nil,
			},
		},

		// Errors
		{
			description: "extra args style no slice strings",
			input: `
key:
  value:
    another: map`,
			expErr: "yaml: unmarshal errors:\n  line 3: cannot unmarshal !!map into string",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got yaml.Step
			err := yamlv2.UnmarshalStrict([]byte(c.input), &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)
		})
	}
}
