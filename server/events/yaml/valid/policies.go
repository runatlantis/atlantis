package valid

// Policies defines version of policy checker binary(conftest) and a list of
// PolicySet objects. Policies struct is used by PolicyCheck workflow to build
// context to enforce policies.
type Policies struct {
	Version    string
	PolicySets []PolicySet
}

type PolicySet struct {
	Source PolicySetSource
	Name   string
	Owners []string
}

type PolicySetSource struct {
	Type string
	Path string
}
