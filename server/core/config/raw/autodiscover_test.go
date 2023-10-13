package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v2"
)

func TestAutoDiscover_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Autodiscover
	}{
		{
			description: "omit unset fields",
			input:       "",
			exp: raw.Autodiscover{
				Enabled: nil,
			},
		},
		{
			description: "enabled true",
			input: `
enabled: true
`,
			exp: raw.Autodiscover{
				Enabled: Bool(true),
			},
		},
		{
			description: "enabled false",
			input: `
enabled: false
`,
			exp: raw.Autodiscover{
				Enabled: Bool(false),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a raw.Autodiscover
			err := yaml.UnmarshalStrict([]byte(c.input), &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}

func TestAutodiscover_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Autodiscover
	}{
		{
			description: "nothing set",
			input:       raw.Autodiscover{},
		},
		{
			description: "enabled false",
			input: raw.Autodiscover{
				Enabled: Bool(false),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Ok(t, c.input.Validate())
		})
	}
}

func TestAutodiscover_ToValid(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Autodiscover
		exp         valid.Autodiscover
	}{
		{
			description: "nothing set",
			input:       raw.Autodiscover{},
			exp: valid.Autodiscover{
				Enabled: true,
			},
		},
		{
			description: "enabled false",
			input: raw.Autodiscover{
				Enabled: Bool(false),
			},
			exp: valid.Autodiscover{
				Enabled: false,
			},
		},
		{
			description: "enabled true",
			input: raw.Autodiscover{
				Enabled: Bool(true),
			},
			exp: valid.Autodiscover{
				Enabled: true,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}
