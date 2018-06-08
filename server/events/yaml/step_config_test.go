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
		exp         yaml.StepConfig
		expErr      string
	}{

		//Single string.
		{
			description: "should parse just init",
			input:       `init`,
			exp: yaml.StepConfig{
				StepType: "init",
			},
		},
		{
			description: "should parse just plan",
			input:       `plan`,
			exp: yaml.StepConfig{
				StepType: "plan",
			},
		},
		{
			description: "should parse just apply",
			input:       `apply`,
			exp: yaml.StepConfig{
				StepType: "apply",
			},
		},

		// With extra_args.
		{
			description: "should parse init with extra_args",
			input: `
init:
  extra_args: [arg1, arg2]`,
			exp: yaml.StepConfig{
				StepType:  "init",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "should parse plan with extra_args",
			input: `
plan:
  extra_args: [arg1, arg2]`,
			exp: yaml.StepConfig{
				StepType:  "plan",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "should parse apply with extra_args",
			input: `
apply:
  extra_args: [arg1, arg2]`,
			exp: yaml.StepConfig{
				StepType:  "apply",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},

		// extra_args with non-strings.
		{
			description: "should convert non-string extra_args into strings",
			input: `
init:
  extra_args: [1]`,
			exp: yaml.StepConfig{
				StepType:  "init",
				ExtraArgs: []string{"1"},
			},
		},
		{
			description: "should convert non-string extra_args into strings",
			input: `
plan:
  extra_args: [true]`,
			exp: yaml.StepConfig{
				StepType:  "plan",
				ExtraArgs: []string{"true"},
			},
		},

		// Custom run step.
		{
			description: "should allow for custom run steps",
			input: `
run: echo my command`,
			exp: yaml.StepConfig{
				StepType: "run",
				Run:      []string{"echo", "my", "command"},
			},
		},
		{
			description: "should split words correctly in run step",
			input: `
run: echo 'my command'`,
			exp: yaml.StepConfig{
				StepType: "run",
				Run:      []string{"echo", "my command"},
			},
		},

		// Invalid steps
		{
			description: "should error when element is a map",
			input: `
key1: val
key2: val`,
			expErr: "each step can have only one map key, you probably have something like:\nsteps:\n  - key1: val\n    key2: val",
		},
		{
			description: "should error when unrecognized step is used",
			input: `
invalid: val
`,
			expErr: "unsupported step \"invalid\"",
		},
		{
			description: "should error when unrecognized step is used",
			input: `
invalid:
  extra_args: []
`,
			expErr: "unsupported step \"invalid\"",
		},
		{
			description: "should error when unrecognized step is used",
			input: `
run: []`,
			expErr: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!seq into string",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got yaml.StepConfig
			err := yamlv2.Unmarshal([]byte(c.input), &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)
		})
	}
}
