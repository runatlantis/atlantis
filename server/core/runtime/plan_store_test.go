// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"os"
	"path/filepath"
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
