package vcs

// Declare all package dependencies here

func NewPullMergeabilityChecker(commitStatusPrefix string) MergeabilityChecker {
	statusFilters := newValidStatusFilters(commitStatusPrefix)
	checksFilters := newValidChecksFilters()

	return &PullMergeabilityChecker{
		supplementalChecker: newSupplementalMergeabilityChecker(statusFilters, checksFilters),
	}
}

func newValidStatusFilters(commitStatusPrefix string) []ValidStatusFilter {
	titleMatcher := StatusTitleMatcher{TitlePrefix: commitStatusPrefix}
	applyStatusFilter := &ApplyStatusFilter{
		statusTitleMatcher: titleMatcher,
	}

	return []ValidStatusFilter{
		SuccessStateFilter, applyStatusFilter,
	}
}

func newValidChecksFilters() []ValidChecksFilter {
	return []ValidChecksFilter{
		SuccessConclusionFilter,
	}
}

func newSupplementalMergeabilityChecker(
	statusFilters []ValidStatusFilter, 
	checksFilters []ValidChecksFilter,
) MergeabilityChecker {
	return &SupplementalMergabilityChecker{
		statusFilter: statusFilters,
		checksFilters: checksFilters,
	}
}