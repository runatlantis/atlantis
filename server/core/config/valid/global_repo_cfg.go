package valid

import "regexp"

// Repo is the final parsed version of server-side repo config.
type Repo struct {
	// ID is the exact match id of this config.
	// If IDRegex is set then this will be empty.
	ID string
	// IDRegex is the regex match for this config.
	// If ID is set then this will be nil.
	IDRegex                     *regexp.Regexp
	BranchRegex                 *regexp.Regexp
	ApplyRequirements           []string
	PreWorkflowHooks            []*PreWorkflowHook
	Workflow                    *Workflow
	PullRequestWorkflow         *Workflow
	DeploymentWorkflow          *Workflow
	AllowedWorkflows            []string
	AllowedPullRequestWorkflows []string
	AllowedDeploymentWorkflows  []string
	AllowedOverrides            []string
	AllowCustomWorkflows        *bool
	TemplateOverrides           map[string]string
	CheckoutStrategy            string
	RebaseEnabled               *bool
}

// IDMatches returns true if the repo ID otherID matches this config.
func (r Repo) IDMatches(otherID string) bool {
	if r.ID != "" {
		return r.ID == otherID
	}
	return r.IDRegex.MatchString(otherID)
}

// BranchMatches returns true if the branch other matches a branch regex (if preset).
func (r Repo) BranchMatches(other string) bool {
	if r.BranchRegex == nil {
		return true
	}
	return r.BranchRegex.MatchString(other)
}

// IDString returns a string representation of this config.
func (r Repo) IDString() string {
	if r.ID != "" {
		return r.ID
	}
	return "/" + r.IDRegex.String() + "/"
}
