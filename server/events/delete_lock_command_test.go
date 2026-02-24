// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

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
	workingDir := events.NewMockWorkingDir(ctrl)
	workingDir.EXPECT().DeletePlan(gomock.Any(), pull.BaseRepo, pull, workspace, path, projectName)
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
}

func TestDeleteLocksByPull_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir(ctrl)
	l.EXPECT().UnlockByPull(repoName, pullNum).Return(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	ErrEquals(t, "err", err)
}

func TestDeleteLocksByPull_None(t *testing.T) {
	t.Log("If there is no lock at that ID there is no error")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir(ctrl)
	l.EXPECT().UnlockByPull(repoName, pullNum).Return([]models.ProjectLock{}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
}

func TestDeleteLocksByPull_SingleSuccess(t *testing.T) {
	t.Log("If a single lock is successfully deleted")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	path := "."
	workspace := "default"
	projectName := "projectname"

	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir(ctrl)
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
	workingDir.EXPECT().DeletePlan(gomock.Any(), pull.BaseRepo, pull, workspace, path, projectName)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
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

	ctrl := gomock.NewController(t)
	l := lockmocks.NewMockLocker(ctrl)
	workingDir := events.NewMockWorkingDir(ctrl)
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
	workingDir.EXPECT().DeletePlan(logger, pull.BaseRepo, pull, workspace, path1, projectName)
	workingDir.EXPECT().DeletePlan(logger, pull.BaseRepo, pull, workspace, path2, projectName)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
}
