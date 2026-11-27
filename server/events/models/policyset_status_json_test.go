package models_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPolicySetStatus_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name          string
		json          string
		expApprovals  models.PolicySetApprovals
		expName       string
		expPassed     bool
		expHashes     []string
		expErrContain string
	}{
		{
			name:         "approvals as a list",
			json:         `{"PolicySetName":"zero_trust","Passed":true,"Approvals":[{"Approver":"alice","Hashes":["h1"]}],"Hashes":["h1"]}`,
			expName:      "zero_trust",
			expPassed:    true,
			expApprovals: models.PolicySetApprovals{{Approver: "alice", Hashes: []string{"h1"}}},
			expHashes:    []string{"h1"},
		},
		{
			name:      "approvals as a count", // pre sticky-approvals server
			json:      `{"PolicySetName":"zero_trust_policies","Passed":true,"Approvals":0}`,
			expName:   "zero_trust_policies",
			expPassed: true,
		},
		{
			name:      "a non-zero count does not revive approvals",
			json:      `{"PolicySetName":"zero_trust","Passed":false,"Approvals":3,"Hashes":["h1"]}`,
			expName:   "zero_trust",
			expHashes: []string{"h1"},
		},
		{
			name:    "approvals null",
			json:    `{"PolicySetName":"zero_trust","Approvals":null}`,
			expName: "zero_trust",
		},
		{
			name:    "approvals absent",
			json:    `{"PolicySetName":"zero_trust"}`,
			expName: "zero_trust",
		},
		{
			name:         "empty list",
			json:         `{"PolicySetName":"zero_trust","Approvals":[]}`,
			expName:      "zero_trust",
			expApprovals: models.PolicySetApprovals{},
		},
		{
			name:          "approvals as a string",
			json:          `{"PolicySetName":"zero_trust","Approvals":"two"}`,
			expErrContain: "neither a list nor a count",
		},
		{
			name:          "approvals as a bool",
			json:          `{"PolicySetName":"zero_trust","Approvals":true}`,
			expErrContain: "neither a list nor a count",
		},
		{
			name:          "approvals as an object",
			json:          `{"PolicySetName":"zero_trust","Approvals":{"alice":1}}`,
			expErrContain: "neither a list nor a count",
		},
		{
			name:          "malformed json is still rejected",
			json:          `{"PolicySetName":"zero_trust","Approvals":[`,
			expErrContain: "unexpected end of JSON input",
		},
		{
			name:          "malformed list element",
			json:          `{"PolicySetName":"zero_trust","Approvals":[{"Approver":5}]}`,
			expErrContain: "cannot unmarshal",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var status models.PolicySetStatus
			err := json.Unmarshal([]byte(c.json), &status)

			if c.expErrContain != "" {
				ErrContains(t, c.expErrContain, err)
				return
			}
			Ok(t, err)
			Equals(t, c.expName, status.PolicySetName)
			Equals(t, c.expPassed, status.Passed)
			Equals(t, c.expApprovals, status.Approvals)
			Equals(t, c.expHashes, status.Hashes)
		})
	}
}

// A count must leave the policy set unapproved, not credit an unnamed approver.
func TestPolicySetStatus_LegacyCountYieldsNoApprovals(t *testing.T) {
	var status models.PolicySetStatus
	Ok(t, json.Unmarshal([]byte(`{"Approvals":5,"Hashes":["h1"]}`), &status))
	Equals(t, 0, status.GetCurApprovals())
}

// The stored pull that could not be read.
func TestPullStatus_UnmarshalLegacyPolicyApprovals(t *testing.T) {
	stored := `{"Projects":[{"Workspace":"default","RepoRelDir":"tf/cloudflare-billing",` +
		`"ProjectName":"cloudflare-billing",` +
		`"PolicyStatus":[{"PolicySetName":"zero_trust_policies","Passed":true,"Approvals":0}],` +
		`"Status":7}],"Pull":{"Num":14599,"HeadBranch":"branch","BaseBranch":"master","Author":"someone"}}`

	var pull models.PullStatus
	Ok(t, json.Unmarshal([]byte(stored), &pull))
	Equals(t, 1, len(pull.Projects))
	Equals(t, 1, len(pull.Projects[0].PolicyStatus))
	Equals(t, "zero_trust_policies", pull.Projects[0].PolicyStatus[0].PolicySetName)
	Equals(t, true, pull.Projects[0].PolicyStatus[0].Passed)
	Equals(t, 14599, pull.Pull.Num)
}

// The wire format must be unchanged.
func TestPolicySetStatus_RoundTrip(t *testing.T) {
	original := models.PolicySetStatus{
		PolicySetName:   "zero_trust",
		Passed:          true,
		Approvals:       models.PolicySetApprovals{{Approver: "alice", Hashes: []string{"h1", "h2"}}},
		Hashes:          []string{"h1", "h2"},
		PolicyItemRegex: ".*",
	}

	encoded, err := json.Marshal(original)
	Ok(t, err)

	var decoded models.PolicySetStatus
	Ok(t, json.Unmarshal(encoded, &decoded))
	Equals(t, original, decoded)
}
