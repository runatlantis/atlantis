package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v2"
)

func TestStepConfig_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Step
		expErr      string
	}{

		// Single string.
		{
			description: "single string",
			input:       `astring`,
			exp: raw.Step{
				Key: String("astring"),
			},
		},

		// MapType i.e. extra_args style.
		{
			description: "extra_args style",
			input: `
key:
  mapValue: [arg1, arg2]`,
			exp: raw.Step{
				Map: MapType{
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
			exp: raw.Step{
				Map: MapType{
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
			exp: raw.Step{
				Map: MapType{
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
			exp: raw.Step{
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
			exp: raw.Step{
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
			exp: raw.Step{
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
			var got raw.Step
			err := yaml.UnmarshalStrict([]byte(c.input), &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)
		})
	}
}

func TestStep_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Step
		expErr      string
	}{
		// Valid inputs.
		{
			description: "init step",
			input: raw.Step{
				Key: String("init"),
			},
			expErr: "",
		},
		{
			description: "plan step",
			input: raw.Step{
				Key: String("plan"),
			},
			expErr: "",
		},
		{
			description: "apply step",
			input: raw.Step{
				Key: String("apply"),
			},
			expErr: "",
		},
		{
			description: "init extra_args",
			input: raw.Step{
				Map: MapType{
					"init": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "plan extra_args",
			input: raw.Step{
				Map: MapType{
					"plan": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "var",
			input: raw.Step{
				Var: VarType{
					"var": {
						"name": "test",
						"command": "echo 123",
					},
				},
			},
			expErr: "",
		},
		{
			description: "apply extra_args",
			input: raw.Step{
				Map: MapType{
					"apply": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "run step",
			input: raw.Step{
				StringVal: map[string]string{
					"run": "my command",
				},
			},
			expErr: "",
		},

		// Invalid inputs.
		{
			description: "empty elem",
			input:       raw.Step{},
			expErr:      "step element is empty",
		},
		{
			description: "invalid step name",
			input: raw.Step{
				Key: String("invalid"),
			},
			expErr: "\"invalid\" is not a valid step type, maybe you omitted the 'run' key",
		},
		{
			description: "multiple keys in map",
			input: raw.Step{
				Map: MapType{
					"key1": nil,
					"key2": nil,
				},
			},
			expErr: "step element can only contain a single key, found 2: key1,key2",
		},
		{
			description: "multiple keys in string val",
			input: raw.Step{
				StringVal: map[string]string{
					"key1": "",
					"key2": "",
				},
			},
			expErr: "step element can only contain a single key, found 2: key1,key2",
		},
		{
			description: "invalid key in map",
			input: raw.Step{
				Map: MapType{
					"invalid": nil,
				},
			},
			expErr: "\"invalid\" is not a valid step type",
		},
		{
			description: "invalid key in string val",
			input: raw.Step{
				StringVal: map[string]string{
					"invalid": "",
				},
			},
			expErr: "\"invalid\" is not a valid step type",
		},
		{
			description: "non extra_arg key",
			input: raw.Step{
				Map: MapType{
					"init": {
						"invalid": nil,
					},
				},
			},
			expErr: "built-in steps only support a single extra_args key, found \"invalid\" in step init",
		},
		{
			// For atlantis.yaml v2, this wouldn't parse, but now there should
			// be no error.
			description: "unparseable shell command",
			input: raw.Step{
				StringVal: map[string]string{
					"run": "my 'c",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := c.input.Validate()
			if c.expErr == "" {
				Ok(t, err)
				return
			}
			ErrEquals(t, c.expErr, err)
		})
	}
}

func TestStep_ToValid(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Step
		exp         valid.Step
	}{
		{
			description: "init step",
			input: raw.Step{
				Key: String("init"),
			},
			exp: valid.Step{
				StepName: "init",
			},
		},
		{
			description: "plan step",
			input: raw.Step{
				Key: String("plan"),
			},
			exp: valid.Step{
				StepName: "plan",
			},
		},
		{
			description: "apply step",
			input: raw.Step{
				Key: String("apply"),
			},
			exp: valid.Step{
				StepName: "apply",
			},
		},
		{
			description: "var step",
			input: raw.Step{
				Var: VarType{
					"var": {
						"name": "test",
						"command": "echo 123",
					},
				},
			},
			exp: valid.Step{
				StepName:  "var",
				RunCommand: "echo 123",
				Variable: "test",
			},
		},
		{
			description: "init extra_args",
			input: raw.Step{
				Map: MapType{
					"init": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "init",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "plan extra_args",
			input: raw.Step{
				Map: MapType{
					"plan": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "plan",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "apply extra_args",
			input: raw.Step{
				Map: MapType{
					"apply": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "apply",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "run step",
			input: raw.Step{
				StringVal: map[string]string{
					"run": "my 'run command'",
				},
			},
			exp: valid.Step{
				StepName:   "run",
				RunCommand: "my 'run command'",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}

type MapType map[string]map[string][]string
type VarType map[string]map[string]string
