package valid

import (
	"strings"

	"github.com/hashicorp/go-version"
)

const (
	LocalPolicySet  string = "local"
	GithubPolicySet string = "github"
)

// PolicySets defines version of policy checker binary(conftest) and a list of
// PolicySet objects. PolicySets struct is used by PolicyCheck workflow to build
// context to enforce policies.
type PolicySets struct {
	Version    *version.Version
	Owners     PolicyOwners
	PolicySets []PolicySet
}

type PolicyOwners struct {
	Users []string
}

type PolicySet struct {
	Source string
	Path   string
	Name   string
	Owners PolicyOwners
}

func (p *PolicySets) HasPolicies() bool {
	return len(p.PolicySets) > 0
}

func (p *PolicySets) IsOwner(username string) bool {
	for _, uname := range p.Owners.Users {
		if strings.EqualFold(uname, username) {
			return true
		}
	}

	return false
}
