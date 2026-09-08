// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestGenerationCompletionExcludesMissingDirectory(t *testing.T) {
	storage, err := boltdb.New(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "head", BaseRepo: models.Repo{FullName: "owner/repo"}}
	projects := []command.ProjectContext{{Workspace: "default", RepoRelDir: "removed"}, {Workspace: "default", RepoRelDir: "remaining"}}
	_, err = storage.BeginPlanGeneration(pull, "G1", projects, true, command.NoClaim{})
	require.NoError(t, err)
	directoryError := DirNotExistErr{RepoRelDir: "removed"}
	results := []command.ProjectResult{
		{Command: command.Plan, Workspace: "default", RepoRelDir: "removed", PlanGeneration: "G1", ProjectCommandOutput: command.ProjectCommandOutput{Error: directoryError}},
		{Command: command.Plan, Workspace: "default", RepoRelDir: "remaining", PlanGeneration: "G1", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}},
	}
	updater := DBUpdater{Database: storage}
	ctx := &command.Context{Pull: pull, Log: logging.NewNoopLogger(t)}
	status, err := updater.updateDB(ctx, pull, results)
	require.NoError(t, err)
	require.Len(t, status.Projects, 1)
	require.Equal(t, "remaining", status.Projects[0].RepoRelDir)
	require.Equal(t, directoryError, results[0].Error, "original user-facing error must survive persistence")
	// An apply error deliberately excluded from persistence must not require
	// a missing project row or replace the original directory error.
	results[0].Command = command.Apply
	_, err = updater.updateDB(ctx, pull, results[:1])
	require.NoError(t, err)
	require.Equal(t, directoryError, results[0].Error)
}

type generationRestoreWorkingDir struct {
	WorkingDir
	root   string
	cloned []string
}

func (w *generationRestoreWorkingDir) GetWorkingDir(_ models.Repo, _ models.PullRequest, workspace string) (string, error) {
	path := filepath.Join(w.root, workspace)
	_, err := os.Stat(path)
	return path, err
}
func (w *generationRestoreWorkingDir) Clone(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, workspace string) (string, error) {
	w.cloned = append(w.cloned, workspace)
	path := filepath.Join(w.root, workspace)
	return path, os.MkdirAll(path, 0o700)
}

func TestGenerationRestoreUsesDurableInventoryAfterCheckoutLoss(t *testing.T) {
	storage, err := boltdb.New(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	planRoot := t.TempDir()
	store := &runtime.LocalPlanStore{SeparatePlanDir: planRoot}
	pull := models.PullRequest{Num: 1, HeadCommit: "same-head", BaseRepo: models.Repo{FullName: "owner/repo"}}
	project := command.ProjectContext{BaseRepo: pull.BaseRepo, Pull: pull, Workspace: "production", RepoRelDir: ".", RequiresAtlantisManagedPlanFile: true, LocalSharePlanDir: planRoot, SavedPlanHash: new(string)}
	_, err = storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{project}, true, command.NoClaim{})
	require.NoError(t, err)
	_, err = storage.BeginPlanGeneration(pull, "G2", []command.ProjectContext{project}, true, command.NoClaim{})
	require.NoError(t, err)
	path := runtime.GetPlanFilePath(project, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	project.PlanGeneration = "G2"
	require.NoError(t, os.WriteFile(path, []byte("accepted G2"), 0o600))
	require.NoError(t, store.Save(project, path))
	accepted, err := storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, PlanGeneration: "G2", ManagedPlanHash: *project.SavedPlanHash, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	require.NoError(t, err)
	project.PlanGeneration = "G1"
	require.NoError(t, os.WriteFile(path, []byte("late G1"), 0o600))
	require.NoError(t, store.Save(project, path))
	_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, PlanGeneration: "G1", ManagedPlanHash: *project.SavedPlanHash, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	require.ErrorIs(t, err, db.ErrPlanGenerationSuperseded)
	workingDir := &generationRestoreWorkingDir{root: t.TempDir()}
	builder := DefaultProjectCommandBuilder{WorkingDir: workingDir, PlanStore: store, LocalSharePlanDir: planRoot}
	ctx := &command.Context{Pull: pull, HeadRepo: pull.BaseRepo, PullStatus: &accepted}
	require.NoError(t, builder.restoreAcceptedGenerationPlans(ctx))
	require.Equal(t, []string{DefaultWorkspace, "production"}, workingDir.cloned)
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "accepted G2", string(contents))
}
