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

	"go.uber.org/mock/gomock"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

var project = models.NewProject("owner/repo", "path", "")
var workspace = "workspace"
var pull = models.PullRequest{}
var user = models.User{}
var errExpected = errors.New("err")
var timeNow = time.Now().Local()
var pl = models.ProjectLock{Project: project, Pull: pull, User: user, Workspace: workspace, Time: timeNow}

func TestTryLock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().TryLock(gomock.Any()).Return(false, models.ProjectLock{}, errExpected)
	t.Log("when the backend returns an error, TryLock should return that error")
	l := locking.NewClient(backend)
	_, err := l.TryLock(project, workspace, pull, user)
	Equals(t, err, err)
}

func TestTryLock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	currLock := models.ProjectLock{}
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().TryLock(gomock.Any()).Return(true, currLock, nil)
	l := locking.NewClient(backend)
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace"}, r)
}

func TestUnlock_InvalidKey(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	l := locking.NewClient(backend)

	_, err := l.Unlock("invalidkey")
	Assert(t, err != nil, "expected err")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected err")
}

func TestUnlock_Err(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().Unlock(gomock.Any(), gomock.Any()).Return(nil, errExpected)
	l := locking.NewClient(backend)
	_, err := l.Unlock("owner/repo/path/workspace")
	Equals(t, err, err)
}

func TestUnlock(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().Unlock(gomock.Any(), gomock.Any()).Return(&pl, nil)
	l := locking.NewClient(backend)
	lock, err := l.Unlock("owner/repo/path/workspace")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestList_Err(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().List().Return(nil, errExpected)
	l := locking.NewClient(backend)
	_, err := l.List()
	Equals(t, errExpected, err)
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().List().Return([]models.ProjectLock{pl}, nil)
	l := locking.NewClient(backend)
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{
		"owner/repo/path/workspace": pl,
	}, list)
}

func TestUnlockByPull(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().UnlockByPull("owner/repo", 1).Return(nil, errExpected)
	l := locking.NewClient(backend)
	_, err := l.UnlockByPull("owner/repo", 1)
	Equals(t, errExpected, err)
}

func TestGetLock_BadKey(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	l := locking.NewClient(backend)
	_, err := l.GetLock("invalidkey")
	Assert(t, err != nil, "err should not be nil")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected different err")
}

func TestGetLock_Err(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().GetLock(project, workspace).Return(nil, errExpected)
	l := locking.NewClient(backend)
	_, err := l.GetLock("owner/repo/path/workspace")
	Equals(t, errExpected, err)
}

func TestGetLock(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	backend := mocks.NewMockBackend(ctrl)
	backend.EXPECT().GetLock(project, workspace).Return(&pl, nil)
	l := locking.NewClient(backend)
	lock, err := l.GetLock("owner/repo/path/workspace")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestTryLock_NoOpLocker(t *testing.T) {
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	currLock := models.ProjectLock{}
	l := locking.NewNoOpLocker()
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace"}, r)
}

func TestUnlock_NoOpLocker(t *testing.T) {
	l := locking.NewNoOpLocker()
	lock, err := l.Unlock("owner/repo/path/workspace")
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
	_, err := l.UnlockByPull("owner/repo", 1)
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
	ctrl := gomock.NewController(t); defer ctrl.Finish()
	applyLock := &command.Lock{
		CommandName: command.Apply,
		LockMetadata: command.LockMetadata{
			UnixTime: time.Now().Unix(),
		},
	}

	t.Run("LockApply", func(t *testing.T) {
		t.Run("backend errors", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			backend.EXPECT().LockCommand(gomock.Any(), gomock.Any()).Return(nil, errExpected)
			l := locking.NewApplyClient(backend, false, false)
			lock, err := l.LockApply()
			Equals(t, errExpected, err)
			Assert(t, !lock.Locked, "exp false")
		})

		t.Run("can't lock if apply is omitted from userConfig.AllowCommands", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			l := locking.NewApplyClient(backend, true, false)
			_, err := l.LockApply()
			ErrEquals(t, "apply is omitted from AllowCommands; Apply commands are locked globally until flag is updated", err)

			// TODO: Convert Never() expectation: backend.EXPECT().LockCommand(gomock.Any().Times(0), gomock.Any())
		})

		t.Run("succeeds", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			backend.EXPECT().LockCommand(gomock.Any(), gomock.Any()).Return(applyLock, nil)
			l := locking.NewApplyClient(backend, false, false)
			lock, _ := l.LockApply()
			Assert(t, lock.Locked, "exp lock present")
		})
	})

	t.Run("UnlockApply", func(t *testing.T) {
		t.Run("backend fails", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			backend.EXPECT().UnlockCommand(gomock.Any()).Return(errExpected)
			l := locking.NewApplyClient(backend, false, false)
			err := l.UnlockApply()
			Equals(t, errExpected, err)
		})

		t.Run("can't lock if apply is omitted from userConfig.AllowCommands", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			l := locking.NewApplyClient(backend, true, false)
			err := l.UnlockApply()
			ErrEquals(t, "apply commands are disabled until AllowCommands flag is updated", err)

			// TODO: Convert Never() expectation: backend.EXPECT().UnlockCommand(gomock.Any().Times(0))
		})

		t.Run("succeeds", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			backend.EXPECT().UnlockCommand(gomock.Any()).Return(nil)
			l := locking.NewApplyClient(backend, false, false)
			err := l.UnlockApply()
			Equals(t, nil, err)
		})

	})

	t.Run("CheckApplyLock", func(t *testing.T) {
		t.Run("fails", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			backend.EXPECT().CheckCommandLock(gomock.Any()).Return(nil, errExpected)
			l := locking.NewApplyClient(backend, false, false)
			lock, err := l.CheckApplyLock()
			Equals(t, errExpected, err)
			Equals(t, lock.Locked, false)
		})

		t.Run("when apply is not in AllowCommands always return a lock", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			l := locking.NewApplyClient(backend, true, false)
			lock, err := l.CheckApplyLock()
			Ok(t, err)
			Equals(t, lock.Locked, true)
			// TODO: Convert Never() expectation: backend.EXPECT().CheckCommandLock(gomock.Any().Times(0))
		})

		t.Run("UnlockCommand succeeds", func(t *testing.T) {
			backend := mocks.NewMockBackend(ctrl)

			backend.EXPECT().CheckCommandLock(gomock.Any()).Return(applyLock, nil)
			l := locking.NewApplyClient(backend, false, false)
			lock, err := l.CheckApplyLock()
			Equals(t, nil, err)
			Assert(t, lock.Locked, "exp lock present")
		})
	})
}
