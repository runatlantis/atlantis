package command_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// Test PolicyCleared and PolicySummary
func TestPolicyCheckResults_PolicyFuncs(t *testing.T) {
	cases := []struct {
		description      string
		policySetsConfig valid.PolicySets
		policySetStatus  []models.PolicySetStatus
		policyClearedExp bool
	}{
		{
			description: "single policy set, not passed",
			policySetsConfig: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 1,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Passed:        false,
					Approvals:     0,
				},
			},
			policyClearedExp: false,
		},
		{
			description: "single policy set, passed",
			policySetsConfig: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 1,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Passed:        true,
					Approvals:     0,
				},
			},
			policyClearedExp: true,
		},
		{
			description: "single policy set, fully approved",
			policySetsConfig: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 1,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Passed:        false,
					Approvals:     1,
				},
			},
			policyClearedExp: true,
		},
		{
			description: "multiple policy sets, different states.",
			policySetsConfig: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 2,
					},
					{
						Name:         "policy2",
						ApproveCount: 1,
					},
					{
						Name:         "policy3",
						ApproveCount: 1,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Passed:        false,
					Approvals:     0,
				},
				{
					PolicySetName: "policy2",
					Passed:        false,
					Approvals:     1,
				},
				{
					PolicySetName: "policy3",
					Passed:        true,
					Approvals:     0,
				},
			},
			policyClearedExp: false,
		},
		{
			description: "multiple policy sets, all cleared.",
			policySetsConfig: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 2,
					},
					{
						Name:         "policy2",
						ApproveCount: 1,
					},
					{
						Name:         "policy3",
						ApproveCount: 1,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Passed:        false,
					Approvals:     2,
				},
				{
					PolicySetName: "policy2",
					Passed:        false,
					Approvals:     1,
				},
				{
					PolicySetName: "policy3",
					Passed:        true,
					Approvals:     0,
				},
			},
			policyClearedExp: true,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			pcs := command.ProjectContext{
				ProjectPolicyStatus: c.policySetStatus,
				PolicySets:          c.policySetsConfig,
			}
			Equals(t, c.policyClearedExp, pcs.PolicyCleared())
		})
	}
}
