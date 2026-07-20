// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

// MergeMethod is the strategy Atlantis asks the VCS host to use when it
// automerges a pull request. Each provider exposes its own set of native merge
// strategies (GitHub has merge/rebase/squash, Gitea additionally has
// fast-forward-only, Bitbucket has merge_commit/squash/fast_forward, ...), so
// Atlantis normalises them to the common set below and each client translates
// the value onto its provider's API. Not every provider can perform every
// method; a client returns an error when asked for one it does not support.
type MergeMethod string

const (
	// MergeMethodMerge merges the pull request with a merge commit.
	MergeMethodMerge MergeMethod = "merge"
	// MergeMethodRebase rebases the pull request onto the base branch.
	MergeMethodRebase MergeMethod = "rebase"
	// MergeMethodSquash squashes the pull request's commits into one.
	MergeMethodSquash MergeMethod = "squash"
	// MergeMethodFastForward fast-forwards the base branch to the pull
	// request, i.e. merges without a merge commit.
	MergeMethodFastForward MergeMethod = "fast-forward"
)

func (m MergeMethod) String() string {
	return string(m)
}
