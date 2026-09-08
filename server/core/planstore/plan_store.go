// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/utils"
)

// ReposDir is the directory beneath a data dir or a plan store dir under which
// the per-repository trees are laid out.
const ReposDir = "repos"

// PullDir returns the directory holding everything stored for a single pull
// request beneath root, where root is either a data dir or a plan store dir.
func PullDir(root, repoFullName string, pullNum int) string {
	return filepath.Join(root, ReposDir, repoFullName, strconv.Itoa(pullNum))
}

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
	// where the set of planned projects is unknown. Targeted apply does not
	// call this; DefaultProjectCommandRunner.doApply calls Load before plan
	// validation so the local .tfplan exists after a re-clone.
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
type LocalPlanStore struct {
	// SeparatePlanDir is the plan store directory when it is configured to be
	// distinct from the data dir. Plans written there outlive the checkout, so
	// a working dir lost to a restart can still be recovered without an
	// external store. Empty means plans live inside the checkout and die with
	// it, which leaves nothing to recover from.
	SeparatePlanDir string
}

func (s *LocalPlanStore) Save(_ command.ProjectContext, _ string) error {
	return nil
}

func (s *LocalPlanStore) Load(_ command.ProjectContext, _ string) error {
	return nil
}

func (s *LocalPlanStore) Remove(_ command.ProjectContext, planPath string) error {
	return utils.RemoveIgnoreNonExistent(planPath)
}

// ListWorkspaces reports the workspaces that still have plans on disk in a
// separate plan store dir. Without one there is no inventory to read, because
// the plans went away with the checkout.
func (s *LocalPlanStore) ListWorkspaces(owner, repo string, pullNum int) ([]string, error) {
	if s.SeparatePlanDir == "" {
		return nil, nil
	}
	pullDir := PullDir(s.SeparatePlanDir, filepath.Join(owner, repo), pullNum)
	entries, err := os.ReadDir(pullDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var workspaces []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		hasPlan, err := containsPlanFile(filepath.Join(pullDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if hasPlan {
			workspaces = append(workspaces, entry.Name())
		}
	}
	return workspaces, nil
}

// RestorePlans has nothing to fetch: with a separate plan store dir the plans
// are already on disk, untouched by the loss of the checkout. It only needs to
// report that recovery is possible at all.
func (s *LocalPlanStore) RestorePlans(_, _, _ string, _ int) error {
	if s.SeparatePlanDir == "" {
		return ErrRestoreNotSupported
	}
	return nil
}

// containsPlanFile reports whether dir holds a plan file that apply could
// actually use. Plans under .terragrunt-cache are skipped to match
// PendingPlanFinder, which ignores them (#487).
func containsPlanFile(dir string) (bool, error) {
	found := false
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".terragrunt-cache" {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".tfplan" {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return found, nil
}

func (s *LocalPlanStore) DeleteForPull(_, _ string, _ int) error {
	return nil // no-op: working dir deletion handles local files
}

func (s *LocalPlanStore) DeletePlanForProject(_, _ string, _ int, _, _, _ string) error {
	return nil // no-op: local plan deleted by WorkingDir.DeletePlan
}
