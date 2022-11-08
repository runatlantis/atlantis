package vcs

// Declare all package dependencies here

func NewPullMergeabilityChecker(vcsStatusPrefix string) MergeabilityChecker {
	statusFilters := newValidStatusFilters(vcsStatusPrefix)
	checksFilters := newValidChecksFilters(vcsStatusPrefix)

	return &PullMergeabilityChecker{
		supplementalChecker: newSupplementalMergeabilityChecker(statusFilters, checksFilters),
	}
}

func newValidStatusFilters(vcsStatusPrefix string) []ValidStatusFilter {
	return []ValidStatusFilter{
		SuccessStateFilter,
	}
}

func newValidChecksFilters(vcsStatusPrefix string) []ValidChecksFilter {
	titleMatcher := StatusTitleMatcher{TitlePrefix: vcsStatusPrefix}
	applyChecksFilter := &ApplyChecksFilter{
		statusTitleMatcher: titleMatcher,
	}
	return []ValidChecksFilter{
		SuccessConclusionFilter, SkippedConclusionFilter, applyChecksFilter,
	}
}

func newSupplementalMergeabilityChecker(
	statusFilters []ValidStatusFilter,
	checksFilters []ValidChecksFilter,
) MergeabilityChecker {
	return &SupplementalMergabilityChecker{
		statusFilter:  statusFilters,
		checksFilters: checksFilters,
	}
}
