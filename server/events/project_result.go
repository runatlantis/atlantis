package events

import "github.com/hootsuite/atlantis/server/events/vcs"

// ProjectResult is the result of executing a plan/apply for a project.
type ProjectResult struct {
	Path         string
	Error        error
	Failure      string
	PlanSuccess  *PlanSuccess
	ApplySuccess string
}

// Status returns the vcs commit status of this project result.
func (p ProjectResult) Status() vcs.CommitStatus {
	if p.Error != nil {
		return vcs.Failed
	}
	if p.Failure != "" {
		return vcs.Failed
	}
	return vcs.Success
}
