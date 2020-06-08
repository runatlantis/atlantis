package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	dlc := events.DeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	deleteLockResult, _, _ := dlc.DeleteLock("id")
	Equals(t, models.DeleteLockFail, deleteLockResult)
}
