// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package coordination

import (
	"errors"
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/core/coordination/boltdb"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

const defaultWorkspace = "default"

func TestPullStatusUpdater_StaleApplyResultDoesNotOverwriteNewerPullStatus(t *testing.T) {
	assertPullStatusUpdaterStaleApplyPreservesNewerPullStatus(t, command.ProjectCommandOutput{
		Error: errors.New("mergeable requirement failed"),
	})
}

func TestPullStatusUpdater_StaleApplyFailurePreservesNewerPullStatus(t *testing.T) {
	assertPullStatusUpdaterStaleApplyPreservesNewerPullStatus(t, command.ProjectCommandOutput{
		Failure: "mergeable requirement failed",
	})
}

func TestPullStatusUpdater_StaleApplyFailurePreservesNewerPullStatusWhenBaseDiffers(t *testing.T) {
	assertPullStatusUpdaterSameHeadDifferentBaseApplyFailurePreservesCurrentBase(t)
}

func assertPullStatusUpdaterSameHeadDifferentBaseApplyFailurePreservesCurrentBase(t *testing.T) {
	t.Helper()
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
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
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, stalePull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   defaultWorkspace,
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
		t.Fatalf("expected PullStatusUpdater to preserve current base %q, got %q", currentPull.BaseBranch, pullStatus.Pull.BaseBranch)
	}
	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}

func assertPullStatusUpdaterStaleApplyPreservesNewerPullStatus(t *testing.T, output command.ProjectCommandOutput) {
	t.Helper()
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
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
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, stalePull, []command.ProjectResult{
		{
			Command:              command.Apply,
			Workspace:            defaultWorkspace,
			RepoRelDir:           "dirA",
			ProjectName:          "projA",
			ProjectCommandOutput: output,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pullStatus.Pull.HeadCommit != currentPull.HeadCommit {
		t.Fatalf("expected PullStatusUpdater to preserve current head %q, got %q", currentPull.HeadCommit, pullStatus.Pull.HeadCommit)
	}
	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}

func TestPullStatusUpdater_SameHeadApplyFailureWritesErroredApplyStatus(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:     command.Plan,
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   defaultWorkspace,
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

	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.ErroredApplyStatus {
		t.Fatalf("expected current project status %q, got %q", models.ErroredApplyStatus, project.Status)
	}
}

func TestPullStatusUpdater_SameHeadSameBaseApplyFailureWritesErroredApplyStatus(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "release",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:     command.Plan,
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   defaultWorkspace,
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

	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.ErroredApplyStatus {
		t.Fatalf("expected current project status %q, got %q", models.ErroredApplyStatus, project.Status)
	}
}

func TestPullStatusUpdater_SameHeadDifferentBaseApplyFailureDoesNotOverwriteCurrentBaseStatus(t *testing.T) {
	assertPullStatusUpdaterSameHeadDifferentBaseApplyFailurePreservesCurrentBase(t)
}

func TestPullStatusUpdater_BaseRetargetStaleCommandDoesNotWriteApplyError(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
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
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, stalePull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   defaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: fmt.Errorf("%w: pull request base branch changed", command.ErrStaleCommandHead),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pullStatus.Pull.BaseBranch != currentPull.BaseBranch {
		t.Fatalf("expected PullStatusUpdater to preserve current base %q, got %q", currentPull.BaseBranch, pullStatus.Pull.BaseBranch)
	}
	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}

func TestPullStatusUpdater_SameHeadApplyErrorDoesNotTriggerStaleResultGuard(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:     command.Plan,
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   defaultWorkspace,
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

	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.ErroredApplyStatus {
		t.Fatalf("expected current project status %q, got %q", models.ErroredApplyStatus, project.Status)
	}
}

func TestPullStatusUpdater_StaleApplyResultGuardRunsBeforeDirNotExistFiltering(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	updater := &PullStatusUpdater{Store: database}
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
			Workspace:   defaultWorkspace,
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

	pullStatus, err := updater.Update(&command.Context{Log: logging.NewNoopLogger(t)}, stalePull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   defaultWorkspace,
			RepoRelDir:  "dirA",
			ProjectName: "projA",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: command.DirNotExistErr{RepoRelDir: "dirA"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pullStatus.Pull.HeadCommit != currentPull.HeadCommit {
		t.Fatalf("expected PullStatusUpdater to preserve current head %q, got %q", currentPull.HeadCommit, pullStatus.Pull.HeadCommit)
	}
	project := pullStatus.FindProject(defaultWorkspace, "dirA", "projA")
	if project == nil {
		t.Fatal("expected current project status to remain")
	}
	if project.Status != models.PlannedPlanStatus {
		t.Fatalf("expected current project status %q, got %q", models.PlannedPlanStatus, project.Status)
	}
}
