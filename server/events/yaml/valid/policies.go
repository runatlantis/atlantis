package valid

import (
	"github.com/hashicorp/go-version"
)

// PolicySets defines version of policy checker binary(conftest) and a list of
// PolicySet objects. PolicySets struct is used by PolicyCheck workflow to build
// context to enforce policies.
type PolicySets struct {
	Version    *version.Version
	PolicySets []PolicySet
}

type PolicySet struct {
	Source string
	Path   string
	Name   string
	Owners []string
}

func (p *PolicySets) HasPolicies() bool {
	return len(p.PolicySets) > 0
}
