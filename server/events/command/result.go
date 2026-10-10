// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

// Result is the result of running a Command.
type Result struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
	// PlansDeleted is always false. Atlantis no longer deletes successful plans
	// when another project errors with automerge enabled.
	//
	// Deprecated: kept so the API response JSON keeps its existing shape.
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
