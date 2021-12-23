package vcs

import "github.com/runatlantis/atlantis/server/events/vcs/lyft"

// Declare all lyft package dependencies here

func NewLyftPullMergeabilityChecker(commitStatusPrefix string) MergeabilityChecker {
	statusFilters := newValidStatusFilters(commitStatusPrefix)

	statusFilters = append(statusFilters, lyft.NewSQFilter())
	checksFilters := newValidChecksFilters()

	supplementalChecker := newSupplementalMergeabilityChecker(statusFilters, checksFilters)
	supplementalChecker = lyft.NewOwnersStatusChecker(supplementalChecker)

	return &PullMergeabilityChecker{
		supplementalChecker: supplementalChecker,
	}
}
