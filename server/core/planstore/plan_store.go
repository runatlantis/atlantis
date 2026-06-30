// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"errors"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/utils"
)

// ErrRestoreNotSupported is returned by PlanStore implementations that do not
// support restoring plans (e.g. LocalPlanStore). Callers use errors.Is to
// distinguish this from actual restore failures.
var ErrRestoreNotSupported = errors.New("plan store does not support restore")

// PlanStore abstracts plan file persistence.
// LocalPlanStore wraps current filesystem behavior (Save/Load are no-ops).
// S3PlanStore uploads after plan and downloads before apply.
type PlanStore interface {
	// Save persists a plan file after terraform writes it to planPath.
	Save(ctx command.ProjectContext, planPath string) error
	// Load ensures a plan file exists at planPath before terraform reads it.
	Load(ctx command.ProjectContext, planPath string) error
	// Remove deletes a plan file (local + external) after apply/import/state-rm.
	Remove(ctx command.ProjectContext, planPath string) error
	// ListWorkspaces returns the distinct workspace names that have stored
	// plans for the given pull request. Used by the "apply all" path so that
	// every workspace can be cloned before RestorePlans writes plans into it
	// (Clone falls through to forceClone which os.RemoveAll's the target dir,
	// wiping any restored plan files). Implementations that don't support
	// restore should return (nil, nil).
	ListWorkspaces(owner, repo string, pullNum int) ([]string, error)
	// RestorePlans discovers and downloads all plans for a pull request into
	// pullDir. Only used by the "apply all" path (buildAllProjectCommandsByPlan)
	// where the set of planned projects is unknown. The single-project apply
	// path does not call this — it uses Load with an already-known key.
	// Callers must ensure each workspace directory is cloned (has a .git) before
	// invoking this; see ListWorkspaces.
	//
	// Capability probe: callers may invoke this with an empty pullDir to detect
	// whether the implementation supports restore at all. Implementations that
	// don't support restore MUST return ErrRestoreNotSupported. Implementations
	// that do support restore MUST treat empty pullDir as a no-op (return nil).
	RestorePlans(pullDir, owner, repo string, pullNum int) error
	// DeleteForPull removes all stored plan files for a pull request.
	// Called during PR close/merge cleanup.
	DeleteForPull(owner, repo string, pullNum int) error
	// DeletePlanForProject removes a specific project's plan from external storage.
	// Called when a single lock is deleted via the UI or API.
	DeletePlanForProject(owner, repo string, pullNum int, workspace, repoRelDir, projectName string) error
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

func (s *LocalPlanStore) ListWorkspaces(_, _ string, _ int) ([]string, error) {
	return nil, nil // no-op: local store has no remote inventory of workspaces
}

func (s *LocalPlanStore) RestorePlans(_, _, _ string, _ int) error {
	return ErrRestoreNotSupported
}

func (s *LocalPlanStore) DeleteForPull(_, _ string, _ int) error {
	return nil // no-op: working dir deletion handles local files
}

func (s *LocalPlanStore) DeletePlanForProject(_, _ string, _ int, _, _, _ string) error {
	return nil // no-op: local plan deleted by WorkingDir.DeletePlan
}
