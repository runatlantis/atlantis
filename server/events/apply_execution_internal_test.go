// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDurableApply_DirectoryErrorReleasesOnlyExecution(t *testing.T) {
	storage, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { Ok(t, storage.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "head", BaseRepo: models.Repo{FullName: "owner/repo"}}
	project := command.ProjectContext{CommandName: command.Apply, Workspace: "default", RepoRelDir: "missing", ProjectName: "a", PlanGeneration: "G1", AcceptedPlanGeneration: "G1"}
	_, err = storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{project}, true, command.NoClaim{})
	Ok(t, err)
	accepted, err := storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, ProjectName: project.ProjectName, PlanGeneration: "G1", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	Ok(t, err)
	updater := &DBUpdater{Database: storage}
	runner := &ApplyCommandRunner{dbUpdater: updater}
	ctx := &command.Context{Pull: pull, PullStatus: &accepted, Log: logging.NewNoopLogger(t)}
	projects := []command.ProjectContext{project}
	Ok(t, runner.beginDurableApply(ctx, projects))
	Assert(t, projects[0].ApplyExecutionID != "", "execution must be admitted before running")
	Assert(t, errors.Is(runner.beginDurableApply(ctx, []command.ProjectContext{project}), db.ErrApplyAlreadyStarted), "a second command must not execute")
	directoryError := DirNotExistErr{RepoRelDir: project.RepoRelDir}
	result := RunOneProjectCmd(func(command.ProjectContext) command.ProjectCommandOutput {
		return command.ProjectCommandOutput{Error: directoryError}
	}, projects[0])
	Equals(t, projects[0].ApplyExecutionID, result.ApplyExecutionID)
	restored, err := updater.updateDB(ctx, pull, []command.ProjectResult{result})
	Ok(t, err)
	Equals(t, accepted, restored)
	Equals(t, directoryError, result.Error)
	Ok(t, applyResultStatusUpdateError(command.Result{ProjectResults: []command.ProjectResult{result}}, restored, pull, pull, &accepted))
	Ok(t, runner.beginDurableApply(ctx, []command.ProjectContext{project}))
}
