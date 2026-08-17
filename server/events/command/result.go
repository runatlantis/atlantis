// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

// Result is the result of running a Command.
type Result struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
	// PlansDeleted is retained for rendering compatibility. It is true when all
	// plans created during this command were durably invalidated because
	// automerging requires every project plan to succeed. Physical artifacts may
	// remain until explicit cleanup.
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
