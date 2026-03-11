// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/utils"
)

// PlanStore abstracts plan file persistence.
// In Phase 1, LocalPlanStore wraps current filesystem behavior (Save/Load are no-ops).
// In Phase 2, S3PlanStore will upload after plan and download before apply.
type PlanStore interface {
	// Save persists a plan file after terraform writes it to planPath.
	Save(ctx command.ProjectContext, planPath string) error
	// Load ensures a plan file exists at planPath before terraform reads it.
	Load(ctx command.ProjectContext, planPath string) error
	// Remove deletes a plan file (local + external) after apply/import/state-rm.
	Remove(ctx command.ProjectContext, planPath string) error
}

// LocalPlanStore implements PlanStore using the local filesystem.
// Save and Load are no-ops because terraform already reads/writes locally.
type LocalPlanStore struct{}

func (s *LocalPlanStore) Save(_ command.ProjectContext, _ string) error {
	return nil
}

func (s *LocalPlanStore) Load(_ command.ProjectContext, _ string) error {
	return nil
}

func (s *LocalPlanStore) Remove(_ command.ProjectContext, planPath string) error {
	return utils.RemoveIgnoreNonExistent(planPath)
}
