package valid_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPoliciesConfig_HasTeamOwners(t *testing.T) {
	cases := []struct {
		description string
		input       valid.PolicySets
		expResult   bool
	}{
		{
			description: "no team owners",
			input: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name: "policy1",
					},
				},
			},
			expResult: false,
		},
		{
			description: "has top-level team owner",
			input: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Teams: []string{
						"someteam",
					},
				},
				PolicySets: []valid.PolicySet{
					{
						Name: "policy1",
					},
				},
			},
			expResult: true,
		},
		{
			description: "has policy-level team owner",
			input: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name: "policy1",
						Owners: valid.PolicyOwners{
							Teams: []string{
								"someteam",
							},
						},
					},
				},
			},
			expResult: true,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			result := c.input.HasTeamOwners()
			Equals(t, c.expResult, result)
		})
	}
}

func TestPoliciesConfig_IsOwners(t *testing.T) {
	user := "testuser"
	userTeams := []string{"testuserteam"}

	cases := []struct {
		description string
		input       valid.PolicyOwners
		expResult   bool
	}{
		{
			description: "user is not owner",
			input: valid.PolicyOwners{
				Users: []string{
					"someotheruser",
				},
				Teams: []string{
					"someotherteam",
				},
			},
			expResult: false,
		},
		{
			description: "user is owner",
			input: valid.PolicyOwners{
				Users: []string{
					"testuser",
					"someotheruser",
				},
				Teams: []string{
					"someotherteam",
				},
			},
			expResult: true,
		},
		{
			description: "user is owner via team membership",
			input: valid.PolicyOwners{
				Users: []string{
					"someotheruser",
				},
				Teams: []string{
					"someotherteam",
					"testuserteam",
				},
			},
			expResult: true,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			result := c.input.IsOwner(user, userTeams)
			Equals(t, c.expResult, result)
		})
	}
}
