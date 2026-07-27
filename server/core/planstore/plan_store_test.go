// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestLocalPlanStore_Save_IsNoop(t *testing.T) {
	store := &LocalPlanStore{}
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t)}
	err := store.Save(ctx, "/nonexistent/path/plan.tfplan")
	Ok(t, err)
}

func TestLocalPlanStore_Load_IsNoop(t *testing.T) {
	store := &LocalPlanStore{}
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t)}
	err := store.Load(ctx, "/nonexistent/path/plan.tfplan")
	Ok(t, err)
}

func TestLocalPlanStore_Remove_DeletesFile(t *testing.T) {
	store := &LocalPlanStore{}
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t)}

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "test.tfplan")
	err := os.WriteFile(planPath, []byte("plan"), 0600)
	Ok(t, err)

	err = store.Remove(ctx, planPath)
	Ok(t, err)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "plan file should be deleted")
}

func TestLocalPlanStore_Remove_NonexistentFile(t *testing.T) {
	store := &LocalPlanStore{}
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t)}

	err := store.Remove(ctx, "/nonexistent/path/plan.tfplan")
	Ok(t, err)
}

// Without a separate plan store dir the plans live in the checkout, so losing
// the checkout loses them and there is nothing to recover.
func TestLocalPlanStore_RestoreUnsupportedWithoutSeparateDir(t *testing.T) {
	store := &LocalPlanStore{}

	err := store.RestorePlans("", "owner", "repo", 1)
	Assert(t, errors.Is(err, ErrRestoreNotSupported), "expected ErrRestoreNotSupported, got %v", err)

	workspaces, err := store.ListWorkspaces("owner", "repo", 1)
	Ok(t, err)
	Equals(t, 0, len(workspaces))
}

// With a separate plan store dir the plans are already on disk, so restore is
// supported and is a no-op.
func TestLocalPlanStore_RestoreSupportedWithSeparateDir(t *testing.T) {
	store := &LocalPlanStore{SeparatePlanDir: t.TempDir()}

	Ok(t, store.RestorePlans("", "owner", "repo", 1))
	Ok(t, store.RestorePlans(filepath.Join(store.SeparatePlanDir, "anything"), "owner", "repo", 1))
}

func TestLocalPlanStore_ListWorkspaces(t *testing.T) {
	planStoreDir := t.TempDir()
	pullDir := PullDir(planStoreDir, "owner/repo", 42)

	writePlan := func(parts ...string) {
		path := filepath.Join(append([]string{pullDir}, parts...)...)
		Ok(t, os.MkdirAll(filepath.Dir(path), 0700))
		Ok(t, os.WriteFile(path, []byte("plan"), 0600))
	}

	writePlan("default", "project1", "default.tfplan")
	writePlan("staging", "staging.tfplan")
	// A workspace dir with no plan in it must not be reported.
	Ok(t, os.MkdirAll(filepath.Join(pullDir, "empty", "project1"), 0700))
	// Plans that PendingPlanFinder ignores must not be reported either.
	writePlan("cached", ".terragrunt-cache", "project1", "cached.tfplan")

	store := &LocalPlanStore{SeparatePlanDir: planStoreDir}
	workspaces, err := store.ListWorkspaces("owner", "repo", 42)
	Ok(t, err)

	sort.Strings(workspaces)
	Equals(t, []string{"default", "staging"}, workspaces)
}

// An unknown pull request is not an error, it just has nothing stored.
func TestLocalPlanStore_ListWorkspacesMissingPullDir(t *testing.T) {
	store := &LocalPlanStore{SeparatePlanDir: t.TempDir()}

	workspaces, err := store.ListWorkspaces("owner", "repo", 99)
	Ok(t, err)
	Equals(t, 0, len(workspaces))
}
