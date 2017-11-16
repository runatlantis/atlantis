package events

import "github.com/hootsuite/atlantis/server/events/vcs"

type ProjectResult struct {
	Path         string
	Error        error
	Failure      string
	PlanSuccess  *PlanSuccess
	ApplySuccess string
}

func (p ProjectResult) Status() vcs.CommitStatus {
	if p.Error != nil {
		return vcs.Failed
	}
	if p.Failure != "" {
		return vcs.Failed
	}
	return vcs.Success
}
