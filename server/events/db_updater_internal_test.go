// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Internal tests for DBUpdater so that we can call the unexported updateDB method.
package events

import (
	"errors"
	"testing"

	dbmocks "github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

// buildDBContext returns a minimal *command.Context for DB-updater tests.
func buildDBContext(t *testing.T) *command.Context {
	t.Helper()
	return &command.Context{
		Log: logging.NewNoopLogger(t),
	}
}

// TestDBUpdater_UpdateDB_PassesFilteredResults verifies that updateDB removes
// DirNotExistErr results before forwarding to the database.
func TestDBUpdater_UpdateDB_PassesFilteredResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbmocks.NewMockDatabase(ctrl)
	updater := &DBUpdater{Database: mockDB}
	ctx := buildDBContext(t)

	pull := models.PullRequest{Num: 1}
	okResult := command.ProjectResult{RepoRelDir: "dir/ok", Workspace: "default"}
	dirErrResult := command.ProjectResult{
		RepoRelDir: "dir/missing",
		Workspace:  "default",
		ProjectCommandOutput: command.ProjectCommandOutput{
			Error: DirNotExistErr{RepoRelDir: "dir/missing"},
		},
	}

	wantPullStatus := models.PullStatus{Pull: pull}
	// Only okResult should be forwarded; dirErrResult must be filtered out.
	mockDB.EXPECT().
		UpdatePullWithResults(pull, []command.ProjectResult{okResult}).
		Return(wantPullStatus, nil)

	got, err := updater.updateDB(ctx, pull, []command.ProjectResult{okResult, dirErrResult})
	Ok(t, err)
	Equals(t, wantPullStatus, got)
}

// TestDBUpdater_UpdateDB_AllResultsOK verifies that when none of the results
// carry a DirNotExistErr, all results are forwarded to the database.
func TestDBUpdater_UpdateDB_AllResultsOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbmocks.NewMockDatabase(ctrl)
	updater := &DBUpdater{Database: mockDB}
	ctx := buildDBContext(t)

	pull := models.PullRequest{Num: 2}
	results := []command.ProjectResult{
		{RepoRelDir: "dir/a", Workspace: "default"},
		{RepoRelDir: "dir/b", Workspace: "staging"},
	}
	wantPullStatus := models.PullStatus{Pull: pull}

	mockDB.EXPECT().
		UpdatePullWithResults(pull, results).
		Return(wantPullStatus, nil)

	got, err := updater.updateDB(ctx, pull, results)
	Ok(t, err)
	Equals(t, wantPullStatus, got)
}

// TestDBUpdater_UpdateDB_AllDirNotExist verifies that when every result is a
// DirNotExistErr, an empty slice is forwarded to the database.
func TestDBUpdater_UpdateDB_AllDirNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbmocks.NewMockDatabase(ctrl)
	updater := &DBUpdater{Database: mockDB}
	ctx := buildDBContext(t)

	pull := models.PullRequest{Num: 3}
	results := []command.ProjectResult{
		{
			RepoRelDir: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: DirNotExistErr{RepoRelDir: "a"},
			},
		},
		{
			RepoRelDir: "b",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: DirNotExistErr{RepoRelDir: "b"},
			},
		},
	}
	wantPullStatus := models.PullStatus{}

	// Expect an empty (nil) slice since all were filtered.
	mockDB.EXPECT().
		UpdatePullWithResults(pull, []command.ProjectResult(nil)).
		Return(wantPullStatus, nil)

	got, err := updater.updateDB(ctx, pull, results)
	Ok(t, err)
	Equals(t, wantPullStatus, got)
}

// TestDBUpdater_UpdateDB_DatabaseError verifies that database errors are
// propagated back to the caller.
func TestDBUpdater_UpdateDB_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbmocks.NewMockDatabase(ctrl)
	updater := &DBUpdater{Database: mockDB}
	ctx := buildDBContext(t)

	pull := models.PullRequest{Num: 4}
	results := []command.ProjectResult{{RepoRelDir: "dir/x"}}
	wantErr := errors.New("database write failed")

	mockDB.EXPECT().
		UpdatePullWithResults(pull, results).
		Return(models.PullStatus{}, wantErr)

	_, err := updater.updateDB(ctx, pull, results)
	Assert(t, err != nil, "expected database error to be returned")
	Equals(t, wantErr, err)
}

// TestDBUpdater_UpdateDB_NonDirNotExistErrorKept verifies that a non-DirNotExistErr
// error result is NOT filtered – it must be forwarded to the database.
func TestDBUpdater_UpdateDB_NonDirNotExistErrorKept(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbmocks.NewMockDatabase(ctrl)
	updater := &DBUpdater{Database: mockDB}
	ctx := buildDBContext(t)

	pull := models.PullRequest{Num: 5}
	otherErr := errors.New("some other error")
	result := command.ProjectResult{
		RepoRelDir: "dir/y",
		ProjectCommandOutput: command.ProjectCommandOutput{
			Error: otherErr,
		},
	}

	wantPullStatus := models.PullStatus{}
	// The result with a generic error must NOT be filtered out.
	mockDB.EXPECT().
		UpdatePullWithResults(pull, []command.ProjectResult{result}).
		Return(wantPullStatus, nil)

	got, err := updater.updateDB(ctx, pull, []command.ProjectResult{result})
	Ok(t, err)
	Equals(t, wantPullStatus, got)
}

// TestDBUpdater_UpdateDB_EmptyResults verifies that an empty input slice
// results in UpdatePullWithResults being called with an empty (nil) slice.
func TestDBUpdater_UpdateDB_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbmocks.NewMockDatabase(ctrl)
	updater := &DBUpdater{Database: mockDB}
	ctx := buildDBContext(t)

	pull := models.PullRequest{Num: 6}
	wantPullStatus := models.PullStatus{}

	mockDB.EXPECT().
		UpdatePullWithResults(pull, []command.ProjectResult(nil)).
		Return(wantPullStatus, nil)

	got, err := updater.updateDB(ctx, pull, []command.ProjectResult{})
	Ok(t, err)
	Equals(t, wantPullStatus, got)
}
