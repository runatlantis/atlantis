package raw_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v3"
)

func TestStage_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Stage
	}{
		{
			description: "empty",
			input:       "",
			exp: raw.Stage{
				Steps: nil,
			},
		},
		{
			description: "all fields set",
			input: `
steps: [step1]
`,
			exp: raw.Stage{
				Steps: []raw.Step{
					{
						Key: String("step1"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a raw.Stage
			err := yaml.UnmarshalStrict([]byte(c.input), &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}

func TestStage_Validate(t *testing.T) {
	// Should validate each step.
	s := raw.Stage{
		Steps: []raw.Step{
			{
				Key: String("invalid"),
			},
		},
	}
	validation.ErrorTag = "yaml"
	ErrEquals(t, "steps: (0: \"invalid\" is not a valid step type, maybe you omitted the 'run' key.).", s.Validate())

	// Empty steps should validate.
	Ok(t, (raw.Stage{}).Validate())
}

func TestStage_ToValid(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Stage
		exp         valid.Stage
	}{
		{
			description: "nothing set",
			input:       raw.Stage{},
			exp: valid.Stage{
				Steps: nil,
			},
		},
		{
			description: "fields set",
			input: raw.Stage{
				Steps: []raw.Step{
					{
						Key: String("init"),
					},
				},
			},
			exp: valid.Stage{
				Steps: []valid.Step{
					{
						StepName: "init",
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}
