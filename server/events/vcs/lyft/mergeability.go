package lyft

import (
	"github.com/google/go-github/v45/github"
)

const (
	SubmitQueueReadinessContext = "sq-ready-to-merge"
	OwnersStatusContext         = "_owners-check"
)

// NameStateFilter filters statuses that correspond to atlantis apply
type NameStateFilter struct {
	statusName string
	state      string
}

// MewSQFilter filters statuses matching submit queue if they are pending.
func NewSQFilter() NameStateFilter {
	return NameStateFilter{statusName: SubmitQueueReadinessContext, state: "pending"}
}

func (f NameStateFilter) Filter(statuses []*github.RepoStatus) []*github.RepoStatus {
	var filtered []*github.RepoStatus
	for _, status := range statuses {
		if status.GetState() == f.state && status.GetContext() == f.statusName {
			continue
		}

		filtered = append(filtered, status)
	}

	return filtered
}

type NameCheckFilter struct {
	checkName string
	status    string
}

func NewSQCheckFilter() NameCheckFilter {
	return NameCheckFilter{checkName: SubmitQueueReadinessContext, status: "queued"}
}

func (f NameCheckFilter) Filter(checks []*github.CheckRun) []*github.CheckRun {
	var filtered []*github.CheckRun
	for _, check := range checks {
		if check.GetStatus() == f.status && check.GetName() == f.checkName {
			continue
		}

		filtered = append(filtered, check)
	}

	return filtered
}

// This interface is brought into the package to prevent a cyclic dependency.
type MergeabilityChecker interface {
	Check(pull *github.PullRequest, statuses []*github.RepoStatus, checks []*github.CheckRun) bool
}

// OwnersStatusChecker delegates to an underlying mergeability checker iff the owners check already exists.
// since this check can be created at random we want to make sure it's present before doing anything.
// This doesn't check the state of the owners check and assumees that the state of all status checks are checked,
// further downstream.
type OwnersStatusChecker struct {
	delegate MergeabilityChecker
}

func NewOwnersStatusChecker(delegate MergeabilityChecker) *OwnersStatusChecker {
	return &OwnersStatusChecker{
		delegate: delegate,
	}
}

func (c *OwnersStatusChecker) Check(pull *github.PullRequest, statuses []*github.RepoStatus, checks []*github.CheckRun) bool {
	if status := findOwnersCheckStatus(statuses); status == nil {
		return false
	}

	return c.delegate.Check(pull, statuses, checks)
}

func findOwnersCheckStatus(statuses []*github.RepoStatus) *github.RepoStatus {
	for _, status := range statuses {
		if status.GetContext() == OwnersStatusContext {
			return status
		}
	}

	return nil
}
