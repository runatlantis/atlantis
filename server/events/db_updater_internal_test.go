// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

func TestDBUpdater_StaleApplyResultDoesNotOverwriteNewerPullStatus(t *testing.T) {
	assertDBUpdaterStaleApplyPreservesNewerPullStatus(t, command.ProjectCommandOutput{
		Error: errors.New("mergeable requirement failed"),
	})
}

func TestDBUpdater_StaleApplyFailurePreservesNewerPullStatus(t *testing.T) {
	assertDBUpdaterStaleApplyPreservesNewerPullStatus(t, command.ProjectCommandOutput{
		Failure: "mergeable requirement failed",
	})
}

func TestDBUpdater_StaleApplyFailurePreservesNewerPullStatusWhenBaseDiffers(t *testing.T) {
	assertDBUpdaterSameHeadDifferentBaseApplyFailurePreservesCurrentBase(t)
}

func assertDBUpdaterSameHeadDifferentBaseApplyFailurePreservesCurrentBase(t *testing.T) {
	t.Helper()
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &DBUpdater{Database: database}
	stalePull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	currentPull := stalePull
	currentPull.BaseBranch = "release"
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
				Failure: "mergeable requirement failed",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pullStatus.Pull.BaseBranch != currentPull.BaseBranch {
		t.Fatalf("expected DBUpdater to preserve current base %q, got %q", currentPull.BaseBranch, pullStatus.Pull.BaseBranch)
	}
	project := findProjectInPullStatus(&pullStatus, DefaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}

func assertDBUpdaterStaleApplyPreservesNewerPullStatus(t *testing.T, output command.ProjectCommandOutput) {
	t.Helper()
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
			Command:              command.Apply,
			Workspace:            DefaultWorkspace,
			RepoRelDir:           "dirA",
			ProjectName:          "projA",
			ProjectCommandOutput: output,
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

func TestDBUpdater_SameHeadApplyFailureWritesErroredApplyStatus(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &DBUpdater{Database: database}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
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

	pullStatus, err := updater.updateDB(&command.Context{Log: logging.NewNoopLogger(t)}, pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   DefaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Failure: "mergeable requirement failed",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	project := findProjectInPullStatus(&pullStatus, DefaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.ErroredApplyStatus {
		t.Fatalf("expected current project status %q, got %q", models.ErroredApplyStatus, project.Status)
	}
}

func TestDBUpdater_SameHeadSameBaseApplyFailureWritesErroredApplyStatus(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &DBUpdater{Database: database}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "release",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
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

	pullStatus, err := updater.updateDB(&command.Context{Log: logging.NewNoopLogger(t)}, pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   DefaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Failure: "mergeable requirement failed",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	project := findProjectInPullStatus(&pullStatus, DefaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.ErroredApplyStatus {
		t.Fatalf("expected current project status %q, got %q", models.ErroredApplyStatus, project.Status)
	}
}

func TestDBUpdater_SameHeadDifferentBaseApplyFailureDoesNotOverwriteCurrentBaseStatus(t *testing.T) {
	assertDBUpdaterSameHeadDifferentBaseApplyFailurePreservesCurrentBase(t)
}

func TestDBUpdater_BaseRetargetStaleCommandDoesNotWriteApplyError(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &DBUpdater{Database: database}
	stalePull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	currentPull := stalePull
	currentPull.BaseBranch = "release"
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
				Error: fmt.Errorf("%w: pull request base branch changed", errStaleCommandHead),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pullStatus.Pull.BaseBranch != currentPull.BaseBranch {
		t.Fatalf("expected DBUpdater to preserve current base %q, got %q", currentPull.BaseBranch, pullStatus.Pull.BaseBranch)
	}
	project := findProjectInPullStatus(&pullStatus, DefaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}

func TestDBUpdater_SameHeadApplyErrorDoesNotTriggerStaleResultGuard(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &DBUpdater{Database: database}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
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

	pullStatus, err := updater.updateDB(&command.Context{Log: logging.NewNoopLogger(t)}, pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   DefaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: errors.New("apply failed"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	project := findProjectInPullStatus(&pullStatus, DefaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.ErroredApplyStatus {
		t.Fatalf("expected current project status %q, got %q", models.ErroredApplyStatus, project.Status)
	}
}

func TestDBUpdater_StaleApplyResultGuardRunsBeforeDirNotExistFiltering(t *testing.T) {
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
				Error: DirNotExistErr{RepoRelDir: "dirA"},
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
