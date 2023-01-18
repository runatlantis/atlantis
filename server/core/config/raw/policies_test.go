package raw_test

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v2"
)

func TestPolicySetsConfig_YAMLMarshalling(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.PolicySets
		expErr      string
	}{
		{
			description: "valid yaml",
			input: `
conftest_version: v1.0.0
organization: org
policy_sets:
- name: policy-name
  source: "local"
  path: "rel/path/to/policy-set"
`,
			exp: raw.PolicySets{
				Organization: "org",
				Version:      String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name",
						Source: valid.LocalPolicySet,
						Path:   "rel/path/to/policy-set",
					},
				},
			},
		},
		{
			description: "valid yaml with multiple paths",
			input: `
conftest_version: v1.0.0
organization: org
policy_sets:
- name: policy-name
  source: "local"
  paths: ["rel/path/to/policy-set", "rel/path/to/another/policy-set"]
`,
			exp: raw.PolicySets{
				Organization: "org",
				Version:      String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name",
						Source: valid.LocalPolicySet,
						Paths:  []string{"rel/path/to/policy-set", "rel/path/to/another/policy-set"},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got raw.PolicySets
			err := yaml.UnmarshalStrict([]byte(c.input), &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)

			_, err = yaml.Marshal(got)
			Ok(t, err)

			var got2 raw.PolicySets
			err = yaml.UnmarshalStrict([]byte(c.input), &got2)
			Ok(t, err)
			Equals(t, got2, got)
		})
	}
}

func TestPolicySets_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.PolicySets
		expErr      string
	}{
		// Valid inputs.
		{
			description: "policies",
			input: raw.PolicySets{
				Organization: "org",
				Version:      String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name-1",
						Owner:  "owner1",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
					{
						Name:   "policy-name-2",
						Owner:  "owner2",
						Path:   "rel/path/to/source",
						Source: valid.GithubPolicySet,
					},
					{
						Name:   "policy-name-3",
						Owner:  "owner3",
						Paths:  []string{"rel/path/to/source", "rel/diff/path/to/source"},
						Source: valid.LocalPolicySet,
					},
				},
			},
			expErr: "",
		},

		// Invalid inputs.
		{
			description: "empty elem",
			input:       raw.PolicySets{},
			expErr:      "policy_sets: cannot be empty; Declare policies that you would like to enforce.",
		},

		{
			description: "missing policy name",
			input: raw.PolicySets{
				PolicySets: []raw.PolicySet{
					{},
				},
			},
			expErr: "policy_sets: (0: (name: is required; owner: is required.).).",
		},
		{
			description: "invalid source type",
			input: raw.PolicySets{
				PolicySets: []raw.PolicySet{
					{
						Name:   "good-policy",
						Owner:  "owner1",
						Source: "invalid-source-type",
						Path:   "rel/path/to/source",
					},
				},
			},
			expErr: "policy_sets: (0: (source: only 'local' and 'github' source types are supported.).).",
		},
		{
			description: "empty string version",
			input: raw.PolicySets{
				Version: String(""),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name-1",
						Owner:  "owner1",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			expErr: "conftest_version: version \"\" could not be parsed: Malformed version: .",
		},
		{
			description: "invalid version",
			input: raw.PolicySets{
				Version: String("version123"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name-1",
						Owner:  "owner1",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			expErr: "conftest_version: version \"version123\" could not be parsed: Malformed version: version123.",
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

func TestPolicySets_ToValid(t *testing.T) {
	version, _ := version.NewVersion("v1.0.0")
	cases := []struct {
		description string
		input       raw.PolicySets
		exp         valid.PolicySets
	}{
		{
			description: "valid policies with owners",
			input: raw.PolicySets{
				Organization: "org",
				Version:      String("v1.0.0"),
				Owners: raw.PolicyOwners{
					Users: []string{
						"test",
					},
				},
				PolicySets: []raw.PolicySet{
					{
						Name:   "good-policy",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Organization: "org",
				Version:      version,
				Owners: valid.PolicyOwners{
					Users: []string{"test"},
				},
				PolicySets: []valid.PolicySet{
					{
						Name:   "good-policy",
						Path:   "rel/path/to/source",
						Source: "local",
					},
				},
			},
		},
		{
			description: "valid policies without owners",
			input: raw.PolicySets{
				Organization: "org",
				Version:      String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "good-policy",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Organization: "org",
				Version:      version,
				PolicySets: []valid.PolicySet{
					{
						Name:   "good-policy",
						Path:   "rel/path/to/source",
						Source: "local",
					},
				},
			},
		},
		{
			description: "valid policies with multiple paths",
			input: raw.PolicySets{
				Organization: "org",
				Version:      String("v1.0.0"),
				Owners: raw.PolicyOwners{
					Users: []string{
						"test",
					},
				},
				PolicySets: []raw.PolicySet{
					{
						Name:   "good-policy",
						Paths:  []string{"rel/path/to/source", "rel/path/to/source2"},
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Organization: "org",
				Version:      version,
				Owners: valid.PolicyOwners{
					Users: []string{"test"},
				},
				PolicySets: []valid.PolicySet{
					{
						Name:   "good-policy",
						Paths:  []string{"rel/path/to/source", "rel/path/to/source2"},
						Source: "local",
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
