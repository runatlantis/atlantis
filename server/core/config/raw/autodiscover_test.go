package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v3"
)

func TestAutoDiscover_UnmarshalYAML(t *testing.T) {
	autoDiscoverEnabled := valid.AutoDiscoverEnabledMode
	cases := []struct {
		description string
		input       string
		exp         raw.AutoDiscover
	}{
		{
			description: "omit unset fields",
			input:       "",
			exp: raw.AutoDiscover{
				Mode: nil,
			},
		},
		{
			description: "all fields set",
			input: `
mode: enabled
`,
			exp: raw.AutoDiscover{
				Mode: &autoDiscoverEnabled,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a raw.AutoDiscover
			err := yaml.UnmarshalStrict([]byte(c.input), &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}

func TestAutoDiscover_Validate(t *testing.T) {
	autoDiscoverAuto := valid.AutoDiscoverAutoMode
	autoDiscoverEnabled := valid.AutoDiscoverEnabledMode
	autoDiscoverDisabled := valid.AutoDiscoverDisabledMode
	randomString := valid.AutoDiscoverMode("random_string")
	cases := []struct {
		description string
		input       raw.AutoDiscover
		errContains *string
	}{
		{
			description: "nothing set",
			input:       raw.AutoDiscover{},
			errContains: nil,
		},
		{
			description: "mode set to auto",
			input: raw.AutoDiscover{
				Mode: &autoDiscoverAuto,
			},
			errContains: nil,
		},
		{
			description: "mode set to disabled",
			input: raw.AutoDiscover{
				Mode: &autoDiscoverDisabled,
			},
			errContains: nil,
		},
		{
			description: "mode set to enabled",
			input: raw.AutoDiscover{
				Mode: &autoDiscoverEnabled,
			},
			errContains: nil,
		},
		{
			description: "mode set to random string",
			input: raw.AutoDiscover{
				Mode: &randomString,
			},
			errContains: String("valid value"),
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			if c.errContains == nil {
				Ok(t, c.input.Validate())
			} else {
				ErrContains(t, *c.errContains, c.input.Validate())
			}
		})
	}
}

func TestAutoDiscover_ToValid(t *testing.T) {
	autoDiscoverEnabled := valid.AutoDiscoverEnabledMode
	cases := []struct {
		description string
		input       raw.AutoDiscover
		exp         *valid.AutoDiscover
	}{
		{
			description: "nothing set",
			input:       raw.AutoDiscover{},
			exp: &valid.AutoDiscover{
				Mode: valid.AutoDiscoverAutoMode,
			},
		},
		{
			description: "value set",
			input: raw.AutoDiscover{
				Mode: &autoDiscoverEnabled,
			},
			exp: &valid.AutoDiscover{
				Mode: valid.AutoDiscoverEnabledMode,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}
