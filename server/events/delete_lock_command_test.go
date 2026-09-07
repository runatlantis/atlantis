// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	lockmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/core/planstore"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, we return the error")
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	l.EXPECT().GetLock("id").Return(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{Locker: l}
	_, _, err := dlc.DeleteLock(logger, "id")
	ErrEquals(t, "err", err)
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we return nil")
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	l.EXPECT().GetLock("id").Return(nil, nil)
	dlc := events.DefaultDeleteLockCommand{Locker: l}
	lock, discarded, err := dlc.DeleteLock(logger, "id")
	Ok(t, err)
	Assert(t, lock == nil, "lock was not nil")
	Assert(t, !discarded, "missing lock must not report discard")
}

func TestDeleteLock_Success(t *testing.T) {
	t.Log("Delete lock deletes successfully the plan file")
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t) // needed for pegomock WorkingDir mock
	workspace := "workspace"
	path := "path"
	projectName := ""
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	expectedLock := &models.ProjectLock{
		Pull:      pull,
		Workspace: workspace,
		Project: models.Project{
			Path:         path,
			RepoFullName: pull.BaseRepo.FullName,
		},
	}
	l.EXPECT().GetLock("id").Return(expectedLock, nil)
	l.EXPECT().UnlockIfOwnedByPull(expectedLock.Project, workspace, pull.Num).Return(expectedLock, nil)
	workingDir := events.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)
	_, err = db.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: workspace, RepoRelDir: path, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	Ok(t, err)
	dlc := events.DefaultDeleteLockCommand{
		Locker:           l,
		Database:         db,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
	}
	lock, discarded, err := dlc.DeleteLock(logger, "id")
	Ok(t, err)
	Assert(t, lock != nil, "lock was nil")
	Assert(t, discarded, "successful plan must be durably discarded")
	workingDir.VerifyWasCalledOnce().DeletePlan(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace),
		Eq(path), Eq(projectName))
}

func TestDeleteLocksByPull_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t) // needed for pegomock WorkingDir mock
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir()
	l.EXPECT().UnlockByPull(repoName, pullNum).Return(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, models.PullRequest{BaseRepo: models.Repo{FullName: repoName}, Num: pullNum})
	ErrEquals(t, "err", err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_None(t *testing.T) {
	t.Log("If there is no lock at that ID there is no error")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t) // needed for pegomock WorkingDir mock
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir()
	l.EXPECT().UnlockByPull(repoName, pullNum).Return([]models.ProjectLock{}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, models.PullRequest{BaseRepo: models.Repo{FullName: repoName}, Num: pullNum})
	Ok(t, err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_SingleSuccess(t *testing.T) {
	t.Log("If a single lock is successfully deleted")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	path := "."
	workspace := "default"
	projectName := "projectname"

	RegisterMockTestingT(t) // needed for pegomock WorkingDir mock
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: repoName},
		Num:      pullNum,
	}
	l.EXPECT().UnlockByPull(repoName, pullNum).Return([]models.ProjectLock{
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         path,
				RepoFullName: pull.BaseRepo.FullName,
				ProjectName:  projectName,
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, models.PullRequest{BaseRepo: models.Repo{FullName: repoName}, Num: pullNum})
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace),
		Eq(path), Eq(projectName))
}

func TestDeleteLocksByPull_MultipleSuccess(t *testing.T) {
	t.Log("If multiple locks are successfully deleted")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	path1 := "path1"
	path2 := "path2"
	workspace := "default"
	projectName := ""

	RegisterMockTestingT(t) // needed for pegomock WorkingDir mock
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: repoName},
		Num:      pullNum,
	}
	l.EXPECT().UnlockByPull(repoName, pullNum).Return([]models.ProjectLock{
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         path1,
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         path2,
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, models.PullRequest{BaseRepo: models.Repo{FullName: repoName}, Num: pullNum})
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(logger, pull.BaseRepo, pull, workspace, path1, projectName)
	workingDir.VerifyWasCalled(Once()).DeletePlan(logger, pull.BaseRepo, pull, workspace, path2, projectName)
}

type rejectDiscardDatabase struct{ db.Database }

func (d rejectDiscardDatabase) DiscardPlanStatus(_ models.PullRequest, _ models.ProjectStatus, _ command.PublicationWriteMode) (bool, error) {
	return false, errors.New("durable discard write failed")
}

func TestDeleteLock_DurableStatusRequiredBeforeUnlock(t *testing.T) {
	for _, scenario := range []string{"missing status", "wrong project", "write failure", "already applied"} {
		t.Run(scenario, func(t *testing.T) {
			RegisterMockTestingT(t)
			storage, err := boltdb.New(t.TempDir())
			Ok(t, err)
			t.Cleanup(func() { Ok(t, storage.Close()) })
			pull := models.PullRequest{Num: 1, HeadCommit: "head", BaseRepo: models.Repo{FullName: "owner/repo"}}
			project := models.NewProject(pull.BaseRepo.FullName, "path", "selected")
			locker := locking.NewClient(storage)
			held, err := locker.TryLock(project, "default", pull, models.User{})
			Ok(t, err)
			if scenario != "missing status" {
				result := command.ProjectResult{Command: command.Plan, Workspace: "default", RepoRelDir: "path", ProjectName: "selected", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
				if scenario == "wrong project" {
					result.ProjectName = "sibling"
				}
				if scenario == "already applied" {
					result.Command = command.Apply
					result.ApplySuccess = "applied"
				}
				_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{result}, command.NoClaim{})
				Ok(t, err)
			}
			var database db.Database = storage
			if scenario == "write failure" {
				database = rejectDiscardDatabase{storage}
			}
			workingDir := events.NewMockWorkingDir()
			deleter := events.DefaultDeleteLockCommand{Locker: locker, Database: database, WorkingDir: workingDir}
			_, discarded, err := deleter.DeleteLock(logging.NewNoopLogger(t), held.LockKey)
			Assert(t, !discarded, "no scenario should report a plan discard")
			actualLock, readErr := locker.GetLock(held.LockKey)
			Ok(t, readErr)
			if scenario == "already applied" {
				Ok(t, err)
				Assert(t, actualLock == nil, "applied lock can be released")
				status, err := storage.GetPullStatus(pull)
				Ok(t, err)
				Equals(t, models.AppliedPlanStatus, status.Projects[0].Status)
			} else {
				Assert(t, err != nil, "durable failure must reject discard")
				Assert(t, actualLock != nil, "failed discard must preserve the lock")
				workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[string](), Any[string]())
			}
		})
	}
}

func TestDeleteLock_ReapsExactAcceptedGeneration(t *testing.T) {
	RegisterMockTestingT(t)
	root := t.TempDir()
	storage, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { Ok(t, storage.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "current", BaseRepo: models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"}}
	ctx := command.ProjectContext{BaseRepo: pull.BaseRepo, Pull: pull, Workspace: "default", RepoRelDir: "path", ProjectName: "selected", RequiresAtlantisManagedPlanFile: true, PlanGeneration: "G1", SavedPlanHash: new(string)}
	_, err = storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{ctx}, true, command.NoClaim{})
	Ok(t, err)
	store := &planstore.LocalPlanStore{}
	path := runtime.GetPlanFilePath(ctx, filepath.Join(planstore.PullDir(root, pull.BaseRepo.FullName, pull.Num), ctx.Workspace, ctx.RepoRelDir))
	Ok(t, os.MkdirAll(filepath.Dir(path), 0o700))
	Ok(t, os.WriteFile(path, []byte("accepted plan"), 0o600))
	Ok(t, store.Save(ctx, path))
	_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: ctx.Workspace, RepoRelDir: ctx.RepoRelDir, ProjectName: ctx.ProjectName, PlanGeneration: "G1", ManagedPlanHash: *ctx.SavedPlanHash, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	Ok(t, err)
	locker := locking.NewClient(storage)
	oldLockPull := pull
	oldLockPull.HeadCommit = "older lock head"
	held, err := locker.TryLock(models.NewProject(pull.BaseRepo.FullName, ctx.RepoRelDir, ctx.ProjectName), ctx.Workspace, oldLockPull, models.User{})
	Ok(t, err)
	deleter := events.DefaultDeleteLockCommand{Locker: locker, Database: storage, WorkingDir: events.NewMockWorkingDir(), PlanStore: store, DataDir: root}
	_, discarded, err := deleter.DeleteLock(logging.NewNoopLogger(t), held.LockKey)
	Ok(t, err)
	Assert(t, discarded, "exact current status may be discarded even if lock metadata is older")
	ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", *ctx.SavedPlanHash
	Assert(t, store.Load(ctx, path) != nil, "discard must reap the accepted artifact")
	_, err = os.Stat(path)
	Assert(t, os.IsNotExist(err), "convention cache must also be reaped")
}

func TestDeleteLocksByPull_ReapsHostedPlanWithoutLocalLock(t *testing.T) {
	RegisterMockTestingT(t)
	root := t.TempDir()
	storage, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { Ok(t, storage.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "current", BaseRepo: models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo", VCSHost: models.VCSHost{Hostname: "github.com", Type: models.Github}}}
	ctx := command.ProjectContext{BaseRepo: pull.BaseRepo, Pull: pull, Workspace: "default", RepoRelDir: "path", ProjectName: "selected", RequiresAtlantisManagedPlanFile: true, PlanGeneration: "G1", SavedPlanHash: new(string)}
	_, err = storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{ctx}, true, command.NoClaim{})
	Ok(t, err)
	store := &planstore.LocalPlanStore{}
	path := runtime.GetPlanFilePath(ctx, filepath.Join(planstore.PullDir(root, pull.BaseRepo.FullName, pull.Num), ctx.Workspace, ctx.RepoRelDir))
	Ok(t, os.MkdirAll(filepath.Dir(path), 0o700))
	Ok(t, os.WriteFile(path, []byte("accepted plan"), 0o600))
	Ok(t, store.Save(ctx, path))
	_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: ctx.Workspace, RepoRelDir: ctx.RepoRelDir, ProjectName: ctx.ProjectName, PlanGeneration: "G1", ManagedPlanHash: *ctx.SavedPlanHash, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	Ok(t, err)
	locker := locking.NewClient(storage)
	deleter := events.DefaultDeleteLockCommand{Locker: locker, Database: storage, WorkingDir: events.NewMockWorkingDir(), PlanStore: store, DataDir: root}
	numLocks, err := deleter.DeleteLocksByPull(logging.NewNoopLogger(t), pull)
	Ok(t, err)
	Equals(t, 0, numLocks)
	status, err := storage.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)
	Equals(t, "", status.Projects[0].AcceptedPlanGeneration)
	ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", *ctx.SavedPlanHash
	Assert(t, store.Load(ctx, path) != nil, "discard must reap the accepted artifact")
	_, err = os.Stat(path)
	Assert(t, os.IsNotExist(err), "convention cache must also be reaped")
}

func TestDeleteLock_ReapsSeparatePlanDirectory(t *testing.T) {
	RegisterMockTestingT(t)
	root := t.TempDir()
	shared := t.TempDir()
	storage, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { Ok(t, storage.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "current", BaseRepo: models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"}}
	ctx := command.ProjectContext{BaseRepo: pull.BaseRepo, Pull: pull, Workspace: "default", RepoRelDir: "path", ProjectName: "selected", LocalSharePlanDir: shared, RequiresAtlantisManagedPlanFile: true, PlanGeneration: "G1", SavedPlanHash: new(string)}
	_, err = storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{ctx}, true, command.NoClaim{})
	Ok(t, err)
	store := &planstore.LocalPlanStore{}
	path := runtime.GetPlanFilePath(ctx, filepath.Join(planstore.PullDir(root, pull.BaseRepo.FullName, pull.Num), ctx.Workspace, ctx.RepoRelDir))
	Ok(t, os.MkdirAll(filepath.Dir(path), 0o700))
	Ok(t, os.WriteFile(path, []byte("accepted plan"), 0o600))
	Ok(t, store.Save(ctx, path))
	_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: ctx.Workspace, RepoRelDir: ctx.RepoRelDir, ProjectName: ctx.ProjectName, PlanGeneration: "G1", ManagedPlanHash: *ctx.SavedPlanHash, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	Ok(t, err)
	locker := locking.NewClient(storage)
	oldLockPull := pull
	oldLockPull.HeadCommit = "older lock head"
	held, err := locker.TryLock(models.NewProject(pull.BaseRepo.FullName, ctx.RepoRelDir, ctx.ProjectName), ctx.Workspace, oldLockPull, models.User{})
	Ok(t, err)
	deleter := events.DefaultDeleteLockCommand{Locker: locker, Database: storage, WorkingDir: events.NewMockWorkingDir(), PlanStore: store, DataDir: root, LocalSharePlanDir: shared}
	_, discarded, err := deleter.DeleteLock(logging.NewNoopLogger(t), held.LockKey)
	Ok(t, err)
	Assert(t, discarded, "exact current status may be discarded even if lock metadata is older")
	ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", *ctx.SavedPlanHash
	Assert(t, store.Load(ctx, path) != nil, "discard must reap the accepted artifact")
	_, err = os.Stat(path)
	Assert(t, os.IsNotExist(err), "convention cache must also be reaped")
}
