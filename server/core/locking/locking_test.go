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
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

var project = models.NewProject("owner/repo", "path", "projectName")
var workspace = "workspace"
var pull = models.PullRequest{}
var user = models.User{}
var errExpected = errors.New("err")
var timeNow = time.Now().Local()
var pl = models.ProjectLock{Project: project, Pull: pull, User: user, Workspace: workspace, Time: timeNow}

func TestTryLock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.TryLock(Any[models.ProjectLock]())).ThenReturn(false, models.ProjectLock{}, errExpected)
	t.Log("when the database returns an error, TryLock should return that error")
	l := locking.NewClient(database)
	_, err := l.TryLock(project, workspace, pull, user)
	Equals(t, err, err)
}

func TestTryLock_Success(t *testing.T) {
	RegisterMockTestingT(t)
	currLock := models.ProjectLock{}
	database := mocks.NewMockDatabase()
	When(database.TryLock(Any[models.ProjectLock]())).ThenReturn(true, currLock, nil)
	l := locking.NewClient(database)
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace/projectName"}, r)
}

func TestUnlock_InvalidKey(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	l := locking.NewClient(database)

	_, err := l.Unlock("invalidkey")
	Assert(t, err != nil, "expected err")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected err")
}

func TestUnlock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.Unlock(Any[models.Project](), Any[string]())).ThenReturn(nil, errExpected)
	l := locking.NewClient(database)
	_, err := l.Unlock("owner/repo/path/workspace/projectName")
	Equals(t, err, err)
	database.VerifyWasCalledOnce().Unlock(project, "workspace")
}

func TestUnlock(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.Unlock(Any[models.Project](), Any[string]())).ThenReturn(&pl, nil)
	l := locking.NewClient(database)
	lock, err := l.Unlock("owner/repo/path/workspace/projectName")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestList_Err(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.List()).ThenReturn(nil, errExpected)
	l := locking.NewClient(database)
	_, err := l.List()
	Equals(t, errExpected, err)
}

func TestList(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.List()).ThenReturn([]models.ProjectLock{pl}, nil)
	l := locking.NewClient(database)
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{
		"owner/repo/path/workspace/projectName": pl,
	}, list)
}

func TestUnlockByPull(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.UnlockByPull("owner/repo", 1)).ThenReturn(nil, errExpected)
	l := locking.NewClient(database)
	_, err := l.UnlockByPull("owner/repo", 1)
	Equals(t, errExpected, err)
}

func TestGetLock_BadKey(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	l := locking.NewClient(database)
	_, err := l.GetLock("invalidkey")
	Assert(t, err != nil, "err should not be nil")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected different err")
}

func TestGetLock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.GetLock(project, workspace)).ThenReturn(nil, errExpected)
	l := locking.NewClient(database)
	_, err := l.GetLock("owner/repo/path/workspace/projectName")
	Equals(t, errExpected, err)
}

func TestGetLock(t *testing.T) {
	RegisterMockTestingT(t)
	database := mocks.NewMockDatabase()
	When(database.GetLock(project, workspace)).ThenReturn(&pl, nil)
	l := locking.NewClient(database)
	lock, err := l.GetLock("owner/repo/path/workspace/projectName")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestTryLock_NoOpLocker(t *testing.T) {
	RegisterMockTestingT(t)
	currLock := models.ProjectLock{}
	l := locking.NewNoOpLocker()
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace/projectName"}, r)
}

func TestUnlock_NoOpLocker(t *testing.T) {
	l := locking.NewNoOpLocker()
	lock, err := l.Unlock("owner/repo/path/workspace/projectName")
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
	lock, err := l.GetLock("owner/repo/path/workspace/projectName")
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
		t.Run("database errors", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			When(database.LockCommand(Any[command.Name](), Any[time.Time]())).ThenReturn(nil, errExpected)
			l := locking.NewApplyClient(database, false, false)
			lock, err := l.LockApply()
			Equals(t, errExpected, err)
			Assert(t, !lock.Locked, "exp false")
		})

		t.Run("can't lock if apply is omitted from userConfig.AllowCommands", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			l := locking.NewApplyClient(database, true, false)
			_, err := l.LockApply()
			ErrEquals(t, "apply is omitted from AllowCommands; Apply commands are locked globally until flag is updated", err)

			database.VerifyWasCalled(Never()).LockCommand(Any[command.Name](), Any[time.Time]())
		})

		t.Run("succeeds", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			When(database.LockCommand(Any[command.Name](), Any[time.Time]())).ThenReturn(applyLock, nil)
			l := locking.NewApplyClient(database, false, false)
			lock, _ := l.LockApply()
			Assert(t, lock.Locked, "exp lock present")
		})
	})

	t.Run("UnlockApply", func(t *testing.T) {
		t.Run("database fails", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			When(database.UnlockCommand(Any[command.Name]())).ThenReturn(errExpected)
			l := locking.NewApplyClient(database, false, false)
			err := l.UnlockApply()
			Equals(t, errExpected, err)
		})

		t.Run("can't lock if apply is omitted from userConfig.AllowCommands", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			l := locking.NewApplyClient(database, true, false)
			err := l.UnlockApply()
			ErrEquals(t, "apply commands are disabled until AllowCommands flag is updated", err)

			database.VerifyWasCalled(Never()).UnlockCommand(Any[command.Name]())
		})

		t.Run("succeeds", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			When(database.UnlockCommand(Any[command.Name]())).ThenReturn(nil)
			l := locking.NewApplyClient(database, false, false)
			err := l.UnlockApply()
			Equals(t, nil, err)
		})

	})

	t.Run("CheckApplyLock", func(t *testing.T) {
		t.Run("fails", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			When(database.CheckCommandLock(Any[command.Name]())).ThenReturn(nil, errExpected)
			l := locking.NewApplyClient(database, false, false)
			lock, err := l.CheckApplyLock()
			Equals(t, errExpected, err)
			Equals(t, lock.Locked, false)
		})

		t.Run("when apply is not in AllowCommands always return a lock", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			l := locking.NewApplyClient(database, true, false)
			lock, err := l.CheckApplyLock()
			Ok(t, err)
			Equals(t, lock.Locked, true)
			database.VerifyWasCalled(Never()).CheckCommandLock(Any[command.Name]())
		})

		t.Run("UnlockCommand succeeds", func(t *testing.T) {
			database := mocks.NewMockDatabase()

			When(database.CheckCommandLock(Any[command.Name]())).ThenReturn(applyLock, nil)
			l := locking.NewApplyClient(database, false, false)
			lock, err := l.CheckApplyLock()
			Equals(t, nil, err)
			Assert(t, lock.Locked, "exp lock present")
		})
	})
}

func TestIsCurrentLocking_ValidKey(t *testing.T) {
	t.Log("IsCurrentLocking should succeed with valid key format")
	key := "owner/repo/path/workspace/projectName"
	matches, err := locking.IsCurrentLocking(key)
	Ok(t, err)
	Equals(t, 5, len(matches))
	Equals(t, "owner/repo", matches[1])
	Equals(t, "path", matches[2])
	Equals(t, "workspace", matches[3])
	Equals(t, "projectName", matches[4])
}

func TestIsCurrentLocking_ValidKeyWithNestedPath(t *testing.T) {
	t.Log("IsCurrentLocking should succeed with nested path")
	key := "owner/repo/parent/child/path/workspace/projectName"
	matches, err := locking.IsCurrentLocking(key)
	Ok(t, err)
	Equals(t, 5, len(matches))
	Equals(t, "owner/repo", matches[1])
	Equals(t, "parent/child/path", matches[2])
	Equals(t, "workspace", matches[3])
	Equals(t, "projectName", matches[4])
}

func TestIsCurrentLocking_InvalidKeyOldFormat(t *testing.T) {
	t.Log("IsCurrentLocking should fail with old format key (3 parts)")
	key := "owner/repo/path/workspace"
	_, err := locking.IsCurrentLocking(key)
	Assert(t, err != nil, "expected error for old format")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected invalid key format error")
}

func TestIsCurrentLocking_InvalidKeySinglePart(t *testing.T) {
	t.Log("IsCurrentLocking should fail with single part key")
	key := "invalidkey"
	_, err := locking.IsCurrentLocking(key)
	Assert(t, err != nil, "expected error for invalid key")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected invalid key format error")
}

func TestIsCurrentLocking_EmptyKey(t *testing.T) {
	t.Log("IsCurrentLocking should fail with empty key")
	key := ""
	_, err := locking.IsCurrentLocking(key)
	Assert(t, err != nil, "expected error for empty key")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected invalid key format error")
}
