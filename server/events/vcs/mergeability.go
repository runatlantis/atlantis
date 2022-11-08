package vcs

import (
	"github.com/google/go-github/v45/github"
)

// ValidStatusFilter implementations filter any valid statuses,
// the definition of valid is up to the implementation.
type ValidStatusFilter interface {
	Filter(status []*github.RepoStatus) []*github.RepoStatus
}

// ApplyStatusFilter filters statuses that correspond to atlantis apply
type ApplyStatusFilter struct {
	statusTitleMatcher StatusTitleMatcher
}

func (d ApplyStatusFilter) Filter(statuses []*github.RepoStatus) []*github.RepoStatus {
	var filtered []*github.RepoStatus
	for _, status := range statuses {
		if d.statusTitleMatcher.MatchesCommand(status.GetContext(), "apply") {
			continue
		}

		filtered = append(filtered, status)
	}

	return filtered
}

// StateFilter filters statuses that match a given state.
type StateFilter string

func (d StateFilter) Filter(statuses []*github.RepoStatus) []*github.RepoStatus {
	var filtered []*github.RepoStatus
	for _, status := range statuses {
		if status.GetState() == string(d) {
			continue
		}

		filtered = append(filtered, status)
	}

	return filtered
}

var SuccessStateFilter StateFilter = "success"

type ValidChecksFilter interface {
	Filter(status []*github.CheckRun) []*github.CheckRun
}

// ApplyChecksFilter filters statuses that correspond to atlantis apply
type ApplyChecksFilter struct {
	statusTitleMatcher StatusTitleMatcher
}

func (d ApplyChecksFilter) Filter(checks []*github.CheckRun) []*github.CheckRun {
	var filtered []*github.CheckRun
	for _, check := range checks {
		if d.statusTitleMatcher.MatchesCommand(*check.Name, "apply") {
			continue
		}

		filtered = append(filtered, check)
	}

	return filtered
}

// ConclusionFilter filters checks that match a given conclusion
type ConclusionFilter string

func (c ConclusionFilter) Filter(checks []*github.CheckRun) []*github.CheckRun {
	var filtered []*github.CheckRun
	for _, check := range checks {
		if check.GetStatus() == "completed" && check.GetConclusion() == string(c) {
			continue
		}

		filtered = append(filtered, check)
	}

	return filtered
}

var SuccessConclusionFilter ConclusionFilter = "success"
var SkippedConclusionFilter ConclusionFilter = "skipped"

type MergeabilityChecker interface {
	Check(pull *github.PullRequest, statuses []*github.RepoStatus, checks []*github.CheckRun) bool
}

// SupplementalMergeabilityChecker is used to determine a more finegrained mergeability
// definition as Github's is purely based on green or not.
// This checker runs each status through a set of ValidStateFilters and ValidCheckFilters
// any leftover statuses or checks are considered invalid and mergeability fails
type SupplementalMergabilityChecker struct {
	statusFilter  []ValidStatusFilter
	checksFilters []ValidChecksFilter
}

func (c *SupplementalMergabilityChecker) Check(_ *github.PullRequest, statuses []*github.RepoStatus, checks []*github.CheckRun) bool {
	invalidStatuses := statuses
	for _, f := range c.statusFilter {
		invalidStatuses = f.Filter(invalidStatuses)
	}

	if len(invalidStatuses) > 0 {
		return false
	}

	invalidChecks := checks
	for _, f := range c.checksFilters {
		invalidChecks = f.Filter(invalidChecks)
	}

	return len(invalidChecks) <= 0
}

// PullMergeabilityChecker primarily uses the mergeable state from github PR and falls back to a supplement checker
// if that fails.
type PullMergeabilityChecker struct {
	supplementalChecker MergeabilityChecker
}

func (c *PullMergeabilityChecker) Check(pull *github.PullRequest, statuses []*github.RepoStatus, checks []*github.CheckRun) bool {
	state := pull.GetMergeableState()
	// We map our mergeable check to when the GitHub merge button is clickable.
	// This corresponds to the following states:
	// clean: No conflicts, all requirements satisfied.
	//        Merging is allowed (green box).
	// unstable: Failing/pending commit status that is not part of the required
	//           status checks. Merging is allowed (yellow box).
	// has_hooks: GitHub Enterprise only, if a repo has custom pre-receive
	//            hooks. Merging is allowed (green box).
	// See: https://github.com/octokit/octokit.net/issues/1763
	if state != "clean" && state != "unstable" && state != "has_hooks" {

		//blocked: Blocked by a failing/missing required status check.
		if state != "blocked" {
			return false
		}

		return c.supplementalChecker.Check(pull, statuses, checks)
	}
	return true
}
