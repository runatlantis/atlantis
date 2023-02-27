package valid

import (
	"strings"

	version "github.com/hashicorp/go-version"
)

const (
	LocalPolicySet  string = "local"
	GithubPolicySet string = "github"
)

// PolicySets defines version of policy checker binary(conftest) and a list of
// PolicySet objects. PolicySets struct is used by PolicyCheck workflow to build
// context to enforce policies.
type PolicySets struct {
	Version      *version.Version
	Owners       PolicyOwners
	ApproveCount int
	PolicySets   []PolicySet
}

type PolicyOwners struct {
	Users []string
	Teams []string
}

type PolicySet struct {
	Source       string
	Path         string
	Name         string
	ApproveCount int
	Owners       PolicyOwners
}

func (p *PolicySets) HasPolicies() bool {
	return len(p.PolicySets) > 0
}

// Check if any level of policy owners includes teams
func (p *PolicySets) HasTeamOwners() bool {
	hasTeamOwners := len(p.Owners.Teams) > 0
	for _, policySet := range p.PolicySets {
		if len(policySet.Owners.Teams) > 0 {
			hasTeamOwners = true
		}
	}
	return hasTeamOwners
}

func (o *PolicyOwners) IsOwner(username string, userTeams []string) bool {
	for _, uname := range o.Users {
		if strings.EqualFold(uname, username) {
			return true
		}
	}

	for _, orgTeamName := range o.Teams {
		for _, userTeamName := range userTeams {
			if strings.EqualFold(orgTeamName, userTeamName) {
				return true
			}
		}
	}

	return false
}
