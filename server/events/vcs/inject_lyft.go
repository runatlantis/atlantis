package vcs

import "github.com/runatlantis/atlantis/server/events/vcs/lyft"

// Declare all lyft package dependencies here

func NewLyftPullMergeabilityChecker(vcsStatusPrefix string) MergeabilityChecker {
	statusFilters := newValidStatusFilters(vcsStatusPrefix)
	statusFilters = append(statusFilters, lyft.NewSQFilter())

	checksFilters := newValidChecksFilters(vcsStatusPrefix)
	checksFilters = append(checksFilters, lyft.NewSQCheckFilter())

	supplementalChecker := newSupplementalMergeabilityChecker(statusFilters, checksFilters)
	supplementalChecker = lyft.NewOwnersStatusChecker(supplementalChecker)

	return &PullMergeabilityChecker{
		supplementalChecker: supplementalChecker,
	}
}
