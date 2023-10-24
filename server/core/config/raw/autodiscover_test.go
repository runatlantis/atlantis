package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v2"
)

func TestAutoDiscover_UnmarshalYAML(t *testing.T) {
	auto_discover_enabled := valid.AutoDiscoverEnabledMode
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
				Mode: &auto_discover_enabled,
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
	auto_discover_auto := valid.AutoDiscoverAutoMode
	auto_discover_enabled := valid.AutoDiscoverEnabledMode
	auto_discover_disabled := valid.AutoDiscoverDisabledMode
	random_string := valid.AutoDiscoverMode("random_string")
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
				Mode: &auto_discover_auto,
			},
			errContains: nil,
		},
		{
			description: "mode set to disabled",
			input: raw.AutoDiscover{
				Mode: &auto_discover_disabled,
			},
			errContains: nil,
		},
		{
			description: "mode set to enabled",
			input: raw.AutoDiscover{
				Mode: &auto_discover_enabled,
			},
			errContains: nil,
		},
		{
			description: "mode set to random string",
			input: raw.AutoDiscover{
				Mode: &random_string,
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
	auto_discover_enabled := valid.AutoDiscoverEnabledMode
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
				Mode: &auto_discover_enabled,
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
