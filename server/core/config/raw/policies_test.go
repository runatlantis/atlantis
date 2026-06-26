// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw_test

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v3"
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
policy_sets:
- name: policy-name
  source: "local"
  path: "rel/path/to/policy-set"
`,
			exp: raw.PolicySets{
				Version: String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name",
						Source: valid.LocalPolicySet,
						Path:   "rel/path/to/policy-set",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got raw.PolicySets
			err := unmarshalString(c.input, &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)

			_, err = yaml.Marshal(got)
			Ok(t, err)

			var got2 raw.PolicySets
			err = unmarshalString(c.input, &got2)
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
				Version: String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name-1",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
					{
						Name: "policy-name-2",
						Owners: raw.PolicyOwners{
							Users: []string{
								"john-doe",
								"jane-doe",
							},
						},
						Path:   "rel/path/to/source",
						Source: valid.GithubPolicySet,
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
			description: "missing policy name and source path",
			input: raw.PolicySets{
				PolicySets: []raw.PolicySet{
					{},
				},
			},
			expErr: "policy_sets: (0: (name: is required; path: is required.).).",
		},
		{
			description: "invalid source type",
			input: raw.PolicySets{
				PolicySets: []raw.PolicySet{
					{
						Name:   "good-policy",
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
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			expErr: "conftest_version: version \"\" could not be parsed: malformed version: .",
		},
		{
			description: "invalid version",
			input: raw.PolicySets{
				Version: String("version123"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name-1",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			expErr: "conftest_version: version \"version123\" could not be parsed: malformed version: version123.",
		},
		{
			description: "invalid top-level policy_item_regex",
			input: raw.PolicySets{
				PolicyItemRegex: String("[invalid"),
				PolicySets: []raw.PolicySet{
					{
						Name:   "policy-name-1",
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			expErr: "policy_item_regex: policy_item_regex \"[invalid\" is not a valid regular expression: error parsing regexp: missing closing ]: `[invalid`.",
		},
		{
			description: "invalid per-policy policy_item_regex",
			input: raw.PolicySets{
				PolicySets: []raw.PolicySet{
					{
						Name:            "policy-name-1",
						Path:            "rel/path/to/source",
						Source:          valid.LocalPolicySet,
						PolicyItemRegex: String("(unclosed"),
					},
				},
			},
			expErr: "policy_sets: (0: (policy_item_regex: policy_item_regex \"(unclosed\" is not a valid regular expression: error parsing regexp: missing closing ): `(unclosed`.).).",
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
				Version: String("v1.0.0"),
				Owners: raw.PolicyOwners{
					Users: []string{
						"test",
					},
					Teams: []string{
						"testteam",
					},
				},
				PolicySets: []raw.PolicySet{
					{
						Name: "good-policy",
						Owners: raw.PolicyOwners{
							Users: []string{
								"john-doe",
								"jane-doe",
							},
						},
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				PolicyItemRegex: valid.DefaultPolicyItemRegex,
				Owners: valid.PolicyOwners{
					Users: []string{"test"},
					Teams: []string{"testteam"},
				},
				PolicySets: []valid.PolicySet{
					{
						Name:            "good-policy",
						ApproveCount:    1,
						PolicyItemRegex: valid.DefaultPolicyItemRegex,
						Owners: valid.PolicyOwners{
							Users: []string{
								"john-doe",
								"jane-doe",
							},
						},
						Path:   "rel/path/to/source",
						Source: "local",
					},
				},
			},
		},
		{
			description: "valid policies without owners",
			input: raw.PolicySets{
				Version: String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name: "good-policy",
						Owners: raw.PolicyOwners{
							Users: []string{
								"john-doe",
								"jane-doe",
							},
						},
						Path:   "rel/path/to/source",
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				PolicyItemRegex: valid.DefaultPolicyItemRegex,
				PolicySets: []valid.PolicySet{
					{
						Name: "good-policy",
						Owners: valid.PolicyOwners{
							Users: []string{
								"john-doe",
								"jane-doe",
							},
						},
						Path:            "rel/path/to/source",
						Source:          "local",
						ApproveCount:    1,
						PolicyItemRegex: valid.DefaultPolicyItemRegex,
					},
				},
			},
		},
		{
			description: "sticky approvals enabled at top level",
			input: raw.PolicySets{
				Version:         String("v1.0.0"),
				StickyApprovals: true,
				PolicySets: []raw.PolicySet{
					{
						Name:   "sticky-policy",
						Path:   "rel/path",
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				StickyApprovals: true,
				PolicyItemRegex: valid.DefaultPolicyItemRegex,
				PolicySets: []valid.PolicySet{
					{
						Name:            "sticky-policy",
						Path:            "rel/path",
						Source:          "local",
						ApproveCount:    1,
						StickyApprovals: true,
						PolicyItemRegex: valid.DefaultPolicyItemRegex,
					},
				},
			},
		},
		{
			description: "policy set overrides top-level sticky approvals to false",
			input: raw.PolicySets{
				Version:         String("v1.0.0"),
				StickyApprovals: true,
				PolicySets: []raw.PolicySet{
					{
						Name:            "not-sticky",
						Path:            "rel/path",
						Source:          valid.LocalPolicySet,
						StickyApprovals: Bool(false),
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				StickyApprovals: true,
				PolicyItemRegex: valid.DefaultPolicyItemRegex,
				PolicySets: []valid.PolicySet{
					{
						Name:            "not-sticky",
						Path:            "rel/path",
						Source:          "local",
						ApproveCount:    1,
						StickyApprovals: false,
						PolicyItemRegex: valid.DefaultPolicyItemRegex,
					},
				},
			},
		},
		{
			description: "policy set enables sticky when top-level is off",
			input: raw.PolicySets{
				Version: String("v1.0.0"),
				PolicySets: []raw.PolicySet{
					{
						Name:            "sticky-override",
						Path:            "rel/path",
						Source:          valid.LocalPolicySet,
						StickyApprovals: Bool(true),
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				PolicyItemRegex: valid.DefaultPolicyItemRegex,
				PolicySets: []valid.PolicySet{
					{
						Name:            "sticky-override",
						Path:            "rel/path",
						Source:          "local",
						ApproveCount:    1,
						StickyApprovals: true,
						PolicyItemRegex: valid.DefaultPolicyItemRegex,
					},
				},
			},
		},
		{
			description: "custom policy_item_regex at top level",
			input: raw.PolicySets{
				Version:         String("v1.0.0"),
				PolicyItemRegex: String(`^FAIL.*`),
				PolicySets: []raw.PolicySet{
					{
						Name:   "fail-only",
						Path:   "rel/path",
						Source: valid.LocalPolicySet,
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				PolicyItemRegex: `^FAIL.*`,
				PolicySets: []valid.PolicySet{
					{
						Name:            "fail-only",
						Path:            "rel/path",
						Source:          "local",
						ApproveCount:    1,
						PolicyItemRegex: `^FAIL.*`,
					},
				},
			},
		},
		{
			description: "policy set overrides top-level policy_item_regex",
			input: raw.PolicySets{
				Version:         String("v1.0.0"),
				PolicyItemRegex: String(`^FAIL.*`),
				PolicySets: []raw.PolicySet{
					{
						Name:            "custom-regex",
						Path:            "rel/path",
						Source:          valid.LocalPolicySet,
						PolicyItemRegex: String(`^WARN.*`),
					},
				},
			},
			exp: valid.PolicySets{
				Version:         version,
				ApproveCount:    1,
				PolicyItemRegex: `^FAIL.*`,
				PolicySets: []valid.PolicySet{
					{
						Name:            "custom-regex",
						Path:            "rel/path",
						Source:          "local",
						ApproveCount:    1,
						PolicyItemRegex: `^WARN.*`,
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

func TestPolicySetsConfig_StickyApprovals_YAMLMarshalling(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.PolicySets
	}{
		{
			description: "sticky_policy_approvals true",
			input: `
sticky_policy_approvals: true
policy_sets:
- name: policy-name
  source: "local"
  path: "rel/path/to/policy-set"
`,
			exp: raw.PolicySets{
				StickyApprovals: true,
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
			description: "policy_item_regex at top level",
			input: `
policy_item_regex: "^FAIL.*"
policy_sets:
- name: policy-name
  source: "local"
  path: "rel/path/to/policy-set"
`,
			exp: raw.PolicySets{
				PolicyItemRegex: String("^FAIL.*"),
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
			description: "per-policy sticky and regex overrides",
			input: `
sticky_policy_approvals: true
policy_sets:
- name: policy-name
  source: "local"
  path: "rel/path/to/policy-set"
  sticky_policy_approvals: false
  policy_item_regex: "^WARN.*"
`,
			exp: raw.PolicySets{
				StickyApprovals: true,
				PolicySets: []raw.PolicySet{
					{
						Name:            "policy-name",
						Source:          valid.LocalPolicySet,
						Path:            "rel/path/to/policy-set",
						StickyApprovals: Bool(false),
						PolicyItemRegex: String("^WARN.*"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got raw.PolicySets
			err := unmarshalString(c.input, &got)
			Ok(t, err)
			Equals(t, c.exp, got)

			_, err = yaml.Marshal(got)
			Ok(t, err)
		})
	}
}
