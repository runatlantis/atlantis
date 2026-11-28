// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
)

func TestParsePlanOutputChanges(t *testing.T) {
	plan := []byte(`{
		"format_version": "1.2",
		"output_changes": {
			"unchanged": {"actions": ["no-op"], "before": "a", "after": "a"},
			"updated": {"actions": ["update"], "before": "a", "after": "b"},
			"created": {"actions": ["create"], "before": null, "after": "b"},
			"deleted": {"actions": ["delete"], "before": "b", "after": null},
			"unknown": {"actions": ["update"], "before": "a", "after": null, "after_unknown": true}
		}
	}`)

	changes, err := ParsePlanOutputChanges(plan)
	if err != nil {
		t.Fatalf("ParsePlanOutputChanges: %s", err)
	}
	if len(changes) != 5 {
		t.Fatalf("expected 5 outputs, got %d", len(changes))
	}

	changed := ChangedOutputNames(changes)
	want := []string{"created", "deleted", "unknown", "updated"}
	for _, name := range want {
		if !changed[name] {
			t.Fatalf("expected %q to be reported as changed, got %v", name, changed)
		}
	}
	if changed["unchanged"] {
		t.Fatalf("expected no-op output to be unchanged, got %v", changed)
	}
}

func TestParsePlanOutputChanges_NoOutputs(t *testing.T) {
	changes, err := ParsePlanOutputChanges([]byte(`{"format_version": "1.2"}`))
	if err != nil {
		t.Fatalf("ParsePlanOutputChanges: %s", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected no output changes, got %v", changes)
	}
}

func TestParsePlanOutputChanges_InvalidJSON(t *testing.T) {
	if _, err := ParsePlanOutputChanges([]byte(`not json`)); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestOutputChange_Changed(t *testing.T) {
	tests := []struct {
		name    string
		actions []string
		want    bool
	}{
		{name: "no actions", actions: nil, want: false},
		{name: "no-op", actions: []string{"no-op"}, want: false},
		{name: "create", actions: []string{"create"}, want: true},
		{name: "update", actions: []string{"update"}, want: true},
		{name: "delete", actions: []string{"delete"}, want: true},
		{name: "replace", actions: []string{"delete", "create"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := models.OutputChange{Actions: tt.actions}
			if got := change.Changed(); got != tt.want {
				t.Fatalf("Changed() = %v, want %v", got, tt.want)
			}
		})
	}
}
