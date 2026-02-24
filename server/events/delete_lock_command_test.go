// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	lockmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
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
	l.EXPECT().Unlock("id").Return(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{Locker: l}
	_, err := dlc.DeleteLock(logger, "id")
	ErrEquals(t, "err", err)
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we return nil")
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	l.EXPECT().Unlock("id").Return(nil, nil)
	dlc := events.DefaultDeleteLockCommand{Locker: l}
	lock, err := dlc.DeleteLock(logger, "id")
	Ok(t, err)
	Assert(t, lock == nil, "lock was not nil")
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
	l.EXPECT().Unlock("id").Return(&models.ProjectLock{
		Pull:      pull,
		Workspace: workspace,
		Project: models.Project{
			Path:         path,
			RepoFullName: pull.BaseRepo.FullName,
		},
	}, nil)
	workingDir := events.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)
	dlc := events.DefaultDeleteLockCommand{
		Locker:           l,
		Database:         db,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
	}
	lock, err := dlc.DeleteLock(logger, "id")
	Ok(t, err)
	Assert(t, lock != nil, "lock was nil")
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
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
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
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_SingleSuccess(t *testing.T) {
	t.Log("If a single lock is successfully deleted, working dir is deleted by default")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	workspace := "default"

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
				Path:         ".",
				RepoFullName: pull.BaseRepo.FullName,
				ProjectName:  "projectname",
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeleteForWorkspace(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace))
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_MultipleSuccess(t *testing.T) {
	t.Log("If multiple locks are successfully deleted, working dirs are deleted by default")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	workspace := "default"

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
				Path:         "path1",
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         "path2",
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeleteForWorkspace(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace))
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_SingleSuccess_SkipWorkingDirDeletion(t *testing.T) {
	t.Log("If SkipWorkingDirDeletionOnUnlock is set, only plan files are deleted")
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
		Locker:                         l,
		WorkingDir:                     workingDir,
		SkipWorkingDirDeletionOnUnlock: true,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace),
		Eq(path), Eq(projectName))
	workingDir.VerifyWasCalled(Never()).DeleteForWorkspace(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())
}
