package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v2"
)

func TestPoliciesConfig_YAMLMarshalling(t *testing.T) {
	version := "v1.0.0"
	cases := []struct {
		description string
		input       string
		exp         raw.Policies
		expErr      string
	}{

		{
			description: "valid yaml",
			input: `
conftest_version: v1.0.0
policy_sets:
- name: policy-name
  source:
    type: "local"
    path: "rel/path/to/policy-set"
`,
			exp: raw.Policies{
				Version: version,
				PolicySets: []raw.PolicySet{
					{
						Name: "policy-name",
						Source: raw.PolicySetSource{
							Type: raw.LocalSourceType,
							Path: "rel/path/to/policy-set",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got raw.Policies
			err := yaml.UnmarshalStrict([]byte(c.input), &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)

			_, err = yaml.Marshal(got)
			Ok(t, err)

			var got2 raw.Policies
			err = yaml.UnmarshalStrict([]byte(c.input), &got2)
			Ok(t, err)
			Equals(t, got2, got)
		})
	}
}

func TestPolicies_Validate(t *testing.T) {
	version := "v1.0.0"
	cases := []struct {
		description string
		input       raw.Policies
		expErr      string
	}{
		// Valid inputs.
		{
			description: "policies",
			input: raw.Policies{
				Version: version,
				PolicySets: []raw.PolicySet{
					{
						Name: "policy-name-1",
						Source: raw.PolicySetSource{
							Path: "rel/path/to/source",
							Type: raw.LocalSourceType,
						},
					},
					{
						Name: "policy-name-2",
						Owners: []string{
							"john-doe",
							"jane-doe",
						},
						Source: raw.PolicySetSource{
							Path: "rel/path/to/source",
							Type: raw.GithubSourceType,
						},
					},
				},
			},
			expErr: "",
		},

		// Invalid inputs.
		{
			description: "empty elem",
			input:       raw.Policies{},
			expErr:      "policy_sets: cannot be empty; Declare policies that you would like to enforce.",
		},
		{
			description: "missing policy name and source path",
			input: raw.Policies{
				PolicySets: []raw.PolicySet{
					{},
				},
			},
			expErr: "policy_sets: (0: (name: is required; source: (path: is required.).).).",
		},
		{
			description: "invalid source type",
			input: raw.Policies{
				PolicySets: []raw.PolicySet{
					{
						Name: "good-policy",
						Source: raw.PolicySetSource{
							Type: "invalid-source-type",
							Path: "rel/path/to/source",
						},
					},
				},
			},
			expErr: "policy_sets: (0: (source: (type: only 'local' and 'github' source types are supported.).).).",
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

func TestPolicies_ToValid(t *testing.T) {
	version := "v1.0.0"
	cases := []struct {
		description string
		input       raw.Policies
		exp         valid.Policies
	}{
		{
			description: "valid policies",
			input: raw.Policies{
				Version: version,
				PolicySets: []raw.PolicySet{
					{
						Name: "good-policy",
						Owners: []string{
							"john-doe",
							"jane-doe",
						},
						Source: raw.PolicySetSource{
							Path: "rel/path/to/source",
							Type: raw.LocalSourceType,
						},
					},
				},
			},
			exp: valid.Policies{
				Version: version,
				PolicySets: []valid.PolicySet{
					{
						Name: "good-policy",
						Owners: []string{
							"john-doe",
							"jane-doe",
						},
						Source: valid.PolicySetSource{
							Path: "rel/path/to/source",
							Type: "local",
						},
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
