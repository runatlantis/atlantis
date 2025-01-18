package raw_test

import (
	"regexp"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestAutoDiscover_UnmarshalYAML(t *testing.T) {
	autoDiscoverEnabled := valid.AutoDiscoverEnabledMode
	ignoreString := "foobar"
	cases := []struct {
		description string
		input       string
		exp         raw.AutoDiscover
	}{
		{
			description: "omit unset fields",
			input:       "",
			exp: raw.AutoDiscover{
				Mode:   nil,
				Ignore: nil,
			},
		},
		{
			description: "all fields set",
			input: `
mode: enabled
ignore: foobar
`,
			exp: raw.AutoDiscover{
				Mode:   &autoDiscoverEnabled,
				Ignore: &ignoreString,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a raw.AutoDiscover
			err := unmarshalString(c.input, &a)
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
	regexWithoutSlahes := ".*"
	invalidRegexString := "/***/"
	validRegexString := "/.*/"
	validRegexPath := `/some\/path\//`
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
		{
			description: "ignore set to regex without slashes",
			input: raw.AutoDiscover{
				Mode:   &autoDiscoverAuto,
				Ignore: &regexWithoutSlahes,
			},
			errContains: String("regex must begin and end with a slash '/'"),
		},
		{
			description: "ignore set to broken regex",
			input: raw.AutoDiscover{
				Mode:   &autoDiscoverAuto,
				Ignore: &invalidRegexString,
			},
			errContains: String("error parsing regexp: missing argument to repetition operator: `*`"),
		},
		{
			description: "ignore set to valid regex",
			input: raw.AutoDiscover{
				Mode:   &autoDiscoverAuto,
				Ignore: &validRegexString,
			},
			errContains: nil,
		},
		{
			description: "ignore set to valid regex path",
			input: raw.AutoDiscover{
				Mode:   &autoDiscoverAuto,
				Ignore: &validRegexPath,
			},
			errContains: nil,
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
	ignoreString := ".*"
	ignoreRegex := regexp.MustCompile(".*")
	cases := []struct {
		description string
		input       raw.AutoDiscover
		exp         *valid.AutoDiscover
	}{
		{
			description: "nothing set",
			input:       raw.AutoDiscover{},
			exp: &valid.AutoDiscover{
				Mode:   valid.AutoDiscoverAutoMode,
				Ignore: nil,
			},
		},
		{
			description: "value set",
			input: raw.AutoDiscover{
				Mode:   &autoDiscoverEnabled,
				Ignore: &ignoreString,
			},
			exp: &valid.AutoDiscover{
				Mode:   valid.AutoDiscoverEnabledMode,
				Ignore: ignoreRegex,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}
