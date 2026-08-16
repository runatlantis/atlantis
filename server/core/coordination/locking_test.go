// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package coordination_test

import (
	"errors"
	"testing"
	"time"

	"strings"

	"github.com/runatlantis/atlantis/server/core/coordination"
	"github.com/runatlantis/atlantis/server/core/coordination/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

var project = models.NewProject("owner/repo", "path", "projectName")
var workspace = "workspace"
var pull = models.PullRequest{}
var user = models.User{}
var errExpected = errors.New("err")
var timeNow = time.Now().Local()
var pl = models.ProjectLock{Project: project, Pull: pull, User: user, Workspace: workspace, Time: timeNow}

func TestTryLock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().TryLock(gomock.Any()).Return(false, models.ProjectLock{}, errExpected)
	t.Log("when the database returns an error, TryLock should return that error")
	l := coordination.NewLockingClient(database)
	_, err := l.TryLock(project, workspace, pull, user)
	Equals(t, errExpected, err)
}

func TestTryLock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	currLock := models.ProjectLock{}
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().TryLock(gomock.Any()).Return(true, currLock, nil)
	l := coordination.NewLockingClient(database)
	r, err := l.TryLock(project, workspace, pull, user)
	Ok(t, err)
	Equals(t, coordination.TryLockResponse{LockAcquired: true, CurrLock: currLock, LockKey: "owner/repo/path/workspace/projectName"}, r)
}

func TestUnlock_InvalidKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	l := coordination.NewLockingClient(database)

	_, err := l.Unlock("invalidkey")
	Assert(t, err != nil, "expected err")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected err")
}

func TestUnlock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().Unlock(project, "workspace").Return(nil, errExpected).Times(1)
	l := coordination.NewLockingClient(database)
	_, err := l.Unlock("owner/repo/path/workspace/projectName")
	Equals(t, errExpected, err)
}

func TestUnlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().Unlock(gomock.Any(), gomock.Any()).Return(&pl, nil)
	l := coordination.NewLockingClient(database)
	lock, err := l.Unlock("owner/repo/path/workspace/projectName")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestUnlockIfOwnedByPull_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().UnlockIfOwnedByPull(project, workspace, 1).Return(nil, errExpected)
	l := coordination.NewLockingClient(database)

	_, err := l.UnlockIfOwnedByPull(project, workspace, 1)
	Equals(t, errExpected, err)
}

func TestUnlockIfOwnedByPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().UnlockIfOwnedByPull(project, workspace, 1).Return(&pl, nil)
	l := coordination.NewLockingClient(database)

	lock, err := l.UnlockIfOwnedByPull(project, workspace, 1)
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestUnlockIfOwnedByPull_SlashfulRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	slashfulProject := models.NewProject("group/subgroup/repo", "terraform", "prod")
	slashfulLock := models.ProjectLock{Project: slashfulProject, Workspace: "default"}
	database.EXPECT().UnlockIfOwnedByPull(slashfulProject, "default", 1).Return(&slashfulLock, nil)
	l := coordination.NewLockingClient(database)

	lock, err := l.UnlockIfOwnedByPull(slashfulProject, "default", 1)
	Ok(t, err)
	Equals(t, &slashfulLock, lock)
}

func TestList_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().List().Return(nil, errExpected)
	l := coordination.NewLockingClient(database)
	_, err := l.List()
	Equals(t, errExpected, err)
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().List().Return([]models.ProjectLock{pl}, nil)
	l := coordination.NewLockingClient(database)
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{
		"owner/repo/path/workspace/projectName": pl,
	}, list)
}

func TestUnlockByPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().UnlockByPull("owner/repo", 1).Return(nil, errExpected)
	l := coordination.NewLockingClient(database)
	_, err := l.UnlockByPull("owner/repo", 1)
	Equals(t, errExpected, err)
}

func TestGetLock_BadKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	l := coordination.NewLockingClient(database)
	_, err := l.GetLock("invalidkey")
	Assert(t, err != nil, "err should not be nil")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected different err")
}

func TestGetLock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().GetLock(project, workspace).Return(nil, errExpected)
	l := coordination.NewLockingClient(database)
	_, err := l.GetLock("owner/repo/path/workspace/projectName")
	Equals(t, errExpected, err)
}

func TestGetLock(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockStore(ctrl)
	database.EXPECT().GetLock(project, workspace).Return(&pl, nil)
	l := coordination.NewLockingClient(database)
	lock, err := l.GetLock("owner/repo/path/workspace/projectName")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestApplyLocker(t *testing.T) {
	applyLock := &command.Lock{
		CommandName: command.Apply,
		LockMetadata: command.LockMetadata{
			UnixTime: time.Now().Unix(),
		},
	}

	t.Run("LockApply", func(t *testing.T) {
		t.Run("database errors", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			database.EXPECT().LockCommand(gomock.Any(), gomock.Any()).Return(nil, errExpected)
			l := coordination.NewApplyClient(database, false, false)
			lock, err := l.LockApply()
			Equals(t, errExpected, err)
			Assert(t, !lock.Locked, "exp false")
		})

		t.Run("can't lock if apply is omitted from userConfig.AllowCommands", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			l := coordination.NewApplyClient(database, true, false)
			_, err := l.LockApply()
			ErrEquals(t, "apply is omitted from AllowCommands; Apply commands are locked globally until flag is updated", err)

			// gomock will fail if LockCommand is called unexpectedly (no EXPECT set)
		})

		t.Run("succeeds", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			database.EXPECT().LockCommand(gomock.Any(), gomock.Any()).Return(applyLock, nil)
			l := coordination.NewApplyClient(database, false, false)
			lock, _ := l.LockApply()
			Assert(t, lock.Locked, "exp lock present")
		})
	})

	t.Run("UnlockApply", func(t *testing.T) {
		t.Run("database fails", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			database.EXPECT().UnlockCommand(gomock.Any()).Return(errExpected)
			l := coordination.NewApplyClient(database, false, false)
			err := l.UnlockApply()
			Equals(t, errExpected, err)
		})

		t.Run("can't lock if apply is omitted from userConfig.AllowCommands", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			l := coordination.NewApplyClient(database, true, false)
			err := l.UnlockApply()
			ErrEquals(t, "apply commands are disabled until AllowCommands flag is updated", err)

			// gomock will fail if UnlockCommand is called unexpectedly (no EXPECT set)
		})

		t.Run("succeeds", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			database.EXPECT().UnlockCommand(gomock.Any()).Return(nil)
			l := coordination.NewApplyClient(database, false, false)
			err := l.UnlockApply()
			Equals(t, nil, err)
		})

	})

	t.Run("CheckApplyLock", func(t *testing.T) {
		t.Run("fails", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			database.EXPECT().CheckCommandLock(gomock.Any()).Return(nil, errExpected)
			l := coordination.NewApplyClient(database, false, false)
			lock, err := l.CheckApplyLock()
			Equals(t, errExpected, err)
			Equals(t, lock.Locked, false)
		})

		t.Run("when apply is not in AllowCommands always return a lock", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			l := coordination.NewApplyClient(database, true, false)
			lock, err := l.CheckApplyLock()
			Ok(t, err)
			Equals(t, lock.Locked, true)
			// gomock will fail if CheckCommandLock is called unexpectedly (no EXPECT set)
		})

		t.Run("UnlockCommand succeeds", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockStore(ctrl)

			database.EXPECT().CheckCommandLock(gomock.Any()).Return(applyLock, nil)
			l := coordination.NewApplyClient(database, false, false)
			lock, err := l.CheckApplyLock()
			Equals(t, nil, err)
			Assert(t, lock.Locked, "exp lock present")
		})
	})
}

func TestIsCurrentLocking_ValidKey(t *testing.T) {
	t.Log("IsCurrentLocking should succeed with valid key format")
	key := "owner/repo/path/workspace/projectName"
	project, workspace, err := models.ParseLockKey(key)
	Ok(t, err)
	Equals(t, "owner/repo", project.RepoFullName)
	Equals(t, "path", project.Path)
	Equals(t, "workspace", workspace)
	Equals(t, "projectName", project.ProjectName)
}

func TestIsCurrentLocking_ValidKeyWithNestedPath(t *testing.T) {
	t.Log("IsCurrentLocking should succeed with nested path")
	key := "owner/repo/parent/child/path/workspace/projectName"
	project, workspace, err := models.ParseLockKey(key)
	Ok(t, err)
	Equals(t, "owner/repo", project.RepoFullName)
	Equals(t, "parent/child/path", project.Path)
	Equals(t, "workspace", workspace)
	Equals(t, "projectName", project.ProjectName)
}

func TestIsCurrentLocking_InvalidKeyOldFormat(t *testing.T) {
	t.Log("IsCurrentLocking should fail with old format key (3 parts)")
	key := "owner/repo/path/workspace"
	_, _, err := models.ParseLockKey(key)
	Assert(t, err != nil, "expected error for old format")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected invalid key format error")
}

func TestIsCurrentLocking_InvalidKeySinglePart(t *testing.T) {
	t.Log("IsCurrentLocking should fail with single part key")
	key := "invalidkey"
	_, _, err := models.ParseLockKey(key)
	Assert(t, err != nil, "expected error for invalid key")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected invalid key format error")
}

func TestIsCurrentLocking_EmptyKey(t *testing.T) {
	t.Log("IsCurrentLocking should fail with empty key")
	key := ""
	_, _, err := models.ParseLockKey(key)
	Assert(t, err != nil, "expected error for empty key")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected invalid key format error")
}
