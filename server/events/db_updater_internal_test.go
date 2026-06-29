// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

func TestDBUpdater_StaleApplyResultDoesNotOverwriteNewerPullStatus(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &DBUpdater{Database: database}
	stalePull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	currentPull := stalePull
	currentPull.HeadCommit = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	_, err = database.UpdatePullWithResults(currentPull, []command.ProjectResult{
		{
			Command:     command.Plan,
			Workspace:   DefaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	pullStatus, err := updater.updateDB(&command.Context{Log: logging.NewNoopLogger(t)}, stalePull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   DefaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: errors.New("mergeable requirement failed"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pullStatus.Pull.HeadCommit != currentPull.HeadCommit {
		t.Fatalf("expected DBUpdater to preserve current head %q, got %q", currentPull.HeadCommit, pullStatus.Pull.HeadCommit)
	}
	project := findProjectInPullStatus(&pullStatus, DefaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}
