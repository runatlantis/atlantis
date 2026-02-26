// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

import "encoding/json"

// Result is the result of running a Command.
type Result struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
	// PlansDeleted is true if all plans created during this command were
	// deleted. This happens if automerging is enabled and one project has an
	// error since automerging requires all plans to succeed.
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

// MarshalJSON implements custom JSON marshaling to properly serialize the Error field.
// Go's error interface serializes as {} by default since it has no exported fields.
// This method converts the error to a string pointer for proper JSON output.
// The ProjectResults slice uses ProjectResult's MarshalJSON automatically.
func (c Result) MarshalJSON() ([]byte, error) {
	// Convert error to string pointer for proper serialization
	var errMsg *string
	if c.Error != nil {
		msg := c.Error.Error()
		errMsg = &msg
	}

	// Use type alias to avoid infinite recursion, then override Error field
	type Alias Result
	return json.Marshal(&struct {
		Error *string `json:"Error"`
		*Alias
	}{
		Error: errMsg,
		Alias: (*Alias)(&c),
	})
}
