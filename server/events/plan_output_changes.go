// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/models"
)

// planOutputChanges is the subset of `terraform show -json <planfile>` that
// Atlantis reads. Terraform emits an entry for every root module output, with
// actions ["no-op"] when the plan does not change it.
type planOutputChanges struct {
	OutputChanges map[string]planOutputChange `json:"output_changes"`
}

type planOutputChange struct {
	Actions []string `json:"actions"`
}

// ParsePlanOutputChanges extracts the planned output changes from the JSON
// representation of a Terraform plan. It returns nil when the plan has no
// output changes section, which is indistinguishable from a configuration
// without outputs.
func ParsePlanOutputChanges(planJSON []byte) (map[string]models.OutputChange, error) {
	var parsed planOutputChanges
	if err := json.Unmarshal(planJSON, &parsed); err != nil {
		return nil, fmt.Errorf("parsing plan JSON: %w", err)
	}
	if len(parsed.OutputChanges) == 0 {
		return nil, nil
	}
	changes := make(map[string]models.OutputChange, len(parsed.OutputChanges))
	for name, change := range parsed.OutputChanges {
		changes[name] = models.OutputChange{Actions: change.Actions}
	}
	return changes, nil
}

// ChangedOutputNames returns the names of the outputs the plan changes. An
// empty result means the plan leaves every output unchanged (or the plan had
// no outputs).
func ChangedOutputNames(changes map[string]models.OutputChange) map[string]bool {
	changed := make(map[string]bool)
	for name, change := range changes {
		if change.Changed() {
			changed[name] = true
		}
	}
	return changed
}
