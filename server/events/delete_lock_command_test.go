package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/db"
	lockmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, we return the error")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	_, err := dlc.DeleteLock("id")
	ErrEquals(t, "err", err)
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we return nil")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	lock, err := dlc.DeleteLock("id")
	Ok(t, err)
	Assert(t, lock == nil, "lock was not nil")
}

func TestDeleteLock_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	lock, err := dlc.DeleteLock("id")
	Ok(t, err)
	Assert(t, lock != nil, "lock was nil")
}

func TestDeleteLock_Success(t *testing.T) {
	t.Log("Delete lock deletes successfully the working dir")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	workingDir := events.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{
		Pull:      pull,
		Workspace: "workspace",
		Project: models.Project{
			Path:         "path",
			RepoFullName: "owner/repo",
		},
	}, nil)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	db, err := db.New(tmp)
	Ok(t, err)
	dlc := events.DefaultDeleteLockCommand{
		Locker:           l,
		Logger:           logging.NewNoopLogger(t),
		DB:               db,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
	}
	lock, err := dlc.DeleteLock("id")
	Ok(t, err)
	Assert(t, lock != nil, "lock was nil")
	workingDir.VerifyWasCalledOnce().DeleteForWorkspace(pull.BaseRepo, pull, "workspace")
}

func TestDeleteLocksByPull_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	ErrEquals(t, "err", err)
}

func TestDeleteLocksByPull_None(t *testing.T) {
	t.Log("If there is no lock at that ID there is no error")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	Ok(t, err)
}

func TestDeleteLocksByPull_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{{}}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	Ok(t, err)
}
