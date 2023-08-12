// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package locking_test

import (
	"errors"
	"testing"
	"time"

	"strings"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

var project = models.NewProject("owner/repo", "path")
var workspace = "workspace"
var pull = models.PullRequest{}
var user = models.User{}
var errExpected = errors.New("err")
var timeNow = time.Now().Local()
var pl = models.ProjectLock{Project: project, Pull: pull, User: user, Workspace: workspace, Time: timeNow}

func TestTryLock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.TryLock(Any[models.ProjectLock]())).ThenReturn(false, models.ProjectLock{}, nil, errExpected)
	t.Log("when the backend returns an error, TryLock should return that error")
	l := locking.NewClient(backend)
	_, err := l.TryLock(project, workspace, pull, user)
	Equals(t, err, err)
}

func TestTryLock_Success(t *testing.T) {
	RegisterMockTestingT(t)
	currLock := models.ProjectLock{}
	backend := mocks.NewMockBackend()
	When(backend.TryLock(Any[models.ProjectLock]())).ThenReturn(true, currLock, nil, nil)
	l := locking.NewClient(backend)
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace"}, r)
}

func TestUnlock_InvalidKey(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	l := locking.NewClient(backend)

	_, _, err := l.Unlock("invalidkey", true)
	Assert(t, err != nil, "expected err")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected err")
}

func TestUnlock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.Unlock(Any[models.Project](), Any[string](), Any[bool]())).ThenReturn(nil, nil, errExpected)
	l := locking.NewClient(backend)
	_, _, err := l.Unlock("owner/repo/path/workspace", false)
	Equals(t, err, err)
	backend.VerifyWasCalledOnce().Unlock(project, "workspace", false)
}

func TestUnlock(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.Unlock(Any[models.Project](), Any[string](), Eq(false))).ThenReturn(&pl, nil, nil)
	l := locking.NewClient(backend)
	lock, _, err := l.Unlock("owner/repo/path/workspace", false)
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestUnlock_UpdateQueue(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.Unlock(Any[models.Project](), Any[string](), Eq(true))).ThenReturn(&pl, nil, nil)
	l := locking.NewClient(backend)
	lock, _, err := l.Unlock("owner/repo/path/workspace", true)
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestList_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.List()).ThenReturn(nil, errExpected)
	l := locking.NewClient(backend)
	_, err := l.List()
	Equals(t, errExpected, err)
}

func TestList(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.List()).ThenReturn([]models.ProjectLock{pl}, nil)
	l := locking.NewClient(backend)
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{
		"owner/repo/path/workspace": pl,
	}, list)
}

func TestUnlockByPull(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.UnlockByPull("owner/repo", 1, true)).ThenReturn(nil, nil, errExpected)
	l := locking.NewClient(backend)
	_, _, err := l.UnlockByPull("owner/repo", 1, true)
	Equals(t, errExpected, err)
}

func TestGetLock_BadKey(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	l := locking.NewClient(backend)
	_, err := l.GetLock("invalidkey")
	Assert(t, err != nil, "err should not be nil")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected different err")
}

func TestGetLock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.GetLock(project, workspace)).ThenReturn(nil, errExpected)
	l := locking.NewClient(backend)
	_, err := l.GetLock("owner/repo/path/workspace")
	Equals(t, errExpected, err)
}

func TestGetLock(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.GetLock(project, workspace)).ThenReturn(&pl, nil)
	l := locking.NewClient(backend)
	lock, err := l.GetLock("owner/repo/path/workspace")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestTryLock_NoOpLocker(t *testing.T) {
	RegisterMockTestingT(t)
	currLock := models.ProjectLock{}
	l := locking.NewNoOpLocker()
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace"}, r)
}

func TestUnlock_NoOpLocker(t *testing.T) {
	l := locking.NewNoOpLocker()
	lock, _, err := l.Unlock("owner/repo/path/workspace", true)
	Ok(t, err)
	Equals(t, &models.ProjectLock{}, lock)
}

func TestList_NoOpLocker(t *testing.T) {
	l := locking.NewNoOpLocker()
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{}, list)
}

func TestUnlockByPull_NoOpLocker(t *testing.T) {
	l := locking.NewNoOpLocker()
	_, _, err := l.UnlockByPull("owner/repo", 1, true)
	Ok(t, err)
}

func TestGetLock_NoOpLocker(t *testing.T) {
	l := locking.NewNoOpLocker()
	lock, err := l.GetLock("owner/repo/path/workspace")
	Ok(t, err)
	var expected *models.ProjectLock
	Equals(t, expected, lock)
}

func TestApplyLocker(t *testing.T) {
	RegisterMockTestingT(t)
	applyLock := &command.Lock{
		CommandName: command.Apply,
		LockMetadata: command.LockMetadata{
			UnixTime: time.Now().Unix(),
		},
	}

	t.Run("LockApply", func(t *testing.T) {
		t.Run("backend errors", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			When(backend.LockCommand(Any[command.Name](), Any[time.Time]())).ThenReturn(nil, errExpected)
			l := locking.NewApplyClient(backend, false)
			lock, err := l.LockApply()
			Equals(t, errExpected, err)
			Assert(t, !lock.Locked, "exp false")
		})

		t.Run("can't lock if userConfig.DisableApply is set", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			l := locking.NewApplyClient(backend, true)
			_, err := l.LockApply()
			ErrEquals(t, "DisableApplyFlag is set; Apply commands are locked globally until flag is unset", err)

			backend.VerifyWasCalled(Never()).LockCommand(Any[command.Name](), Any[time.Time]())
		})

		t.Run("succeeds", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			When(backend.LockCommand(Any[command.Name](), Any[time.Time]())).ThenReturn(applyLock, nil)
			l := locking.NewApplyClient(backend, false)
			lock, _ := l.LockApply()
			Assert(t, lock.Locked, "exp lock present")
		})
	})

	t.Run("UnlockApply", func(t *testing.T) {
		t.Run("backend fails", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			When(backend.UnlockCommand(Any[command.Name]())).ThenReturn(errExpected)
			l := locking.NewApplyClient(backend, false)
			err := l.UnlockApply()
			Equals(t, errExpected, err)
		})

		t.Run("can't unlock if userConfig.DisableApply is set", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			l := locking.NewApplyClient(backend, true)
			err := l.UnlockApply()
			ErrEquals(t, "apply commands are disabled until DisableApply flag is unset", err)

			backend.VerifyWasCalled(Never()).UnlockCommand(Any[command.Name]())
		})

		t.Run("succeeds", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			When(backend.UnlockCommand(Any[command.Name]())).ThenReturn(nil)
			l := locking.NewApplyClient(backend, false)
			err := l.UnlockApply()
			Equals(t, nil, err)
		})

	})

	t.Run("CheckApplyLock", func(t *testing.T) {
		t.Run("fails", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			When(backend.CheckCommandLock(Any[command.Name]())).ThenReturn(nil, errExpected)
			l := locking.NewApplyClient(backend, false)
			lock, err := l.CheckApplyLock()
			Equals(t, errExpected, err)
			Equals(t, lock.Locked, false)
		})

		t.Run("when DisableApply flag is set always return a lock", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			l := locking.NewApplyClient(backend, true)
			lock, err := l.CheckApplyLock()
			Ok(t, err)
			Equals(t, lock.Locked, true)
			backend.VerifyWasCalled(Never()).CheckCommandLock(Any[command.Name]())
		})

		t.Run("UnlockCommand succeeds", func(t *testing.T) {
			backend := mocks.NewMockBackend()

			When(backend.CheckCommandLock(Any[command.Name]())).ThenReturn(applyLock, nil)
			l := locking.NewApplyClient(backend, false)
			lock, err := l.CheckApplyLock()
			Equals(t, nil, err)
			Assert(t, lock.Locked, "exp lock present")
		})
	})
}
