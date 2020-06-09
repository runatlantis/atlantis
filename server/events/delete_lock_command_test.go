package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	lockmocks "github.com/runatlantis/atlantis/server/events/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _, _ := dlc.DeleteLock("id")
	Equals(t, models.DeleteLockFail, deleteLockResult)
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID return no lock found")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, nil)
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _, _ := dlc.DeleteLock("id")
	Equals(t, models.DeleteLockNotFound, deleteLockResult)
}

func TestDeleteLock_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _, _ := dlc.DeleteLock("id")
	Equals(t, models.DeleteLockSuccess, deleteLockResult)
}

func TestDeleteLocksByPull_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn(nil, errors.New("err"))
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _ := dlc.DeleteLocksByPull(repoName, pullNum)
	Equals(t, models.DeleteLockFail, deleteLockResult)
}

func TestDeleteLocksByPull_None(t *testing.T) {
	t.Log("If there is no lock at that ID return no locks found")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{}, nil)
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _ := dlc.DeleteLocksByPull(repoName, pullNum)
	Equals(t, models.DeleteLockNotFound, deleteLockResult)
}

func TestDeleteLocksByPull_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{models.ProjectLock{}}, nil)
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _ := dlc.DeleteLocksByPull(repoName, pullNum)
	Equals(t, models.DeleteLockSuccess, deleteLockResult)
}
