// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

// Result is the result of running a Command.
type Result struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
	// PlansDeleted used to be true if all plans created during this command
	// were deleted, which happened when automerging was enabled and a project
	// errored. Plans are now kept on partial failure, so it is always false.
	//
	// Deprecated: kept only so the API response shape stays stable; do not
	// read or set it.
	PlansDeleted bool
}

// HasErrors returns true if there were any errors during the execution,
// even if it was only in one project.
func (c Result) HasErrors() bool {
	if c.Error != nil || c.Failure != "" {
		return true
	}
	for _, r := range c.ProjectResults {
		if !r.IsSuccessful() {
			return true
		}
	}
	return false
}
