package locking_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/locking/mocks"
	"github.com/hootsuite/atlantis/models"
	. "github.com/hootsuite/atlantis/testing_util"
)

var project = models.NewProject("owner/repo", "path")
var env = "env"
var pull = models.PullRequest{}
var user = models.User{}
var expectedErr = errors.New("err")
var timeNow = time.Now().Local()
var pl = models.ProjectLock{project, pull, user, env, timeNow}

func TestTryLock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().TryLock(gomock.Any()).Return(false, models.ProjectLock{}, expectedErr)

	t.Log("when the backend returns an error, TryLock should return that error")
	l := locking.NewClient(mock)
	_, err := l.TryLock(project, env, pull, user)
	Equals(t, err, err)
}

func TestTryLock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currLock := models.ProjectLock{}
	mock := mocks.NewMockBackend(ctrl)
	// Couldn't figure out how to match on the arg to TryLock.
	// Need to deal with the Time field
	mock.EXPECT().TryLock(gomock.Any()).Return(true, currLock, nil)

	l := locking.NewClient(mock)
	r, err := l.TryLock(project, env, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{true, currLock, "owner/repo/path/env"}, r)
}

func TestUnlock_InvalidKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	l := locking.NewClient(mock)

	_, err := l.Unlock("invalidkey")
	Assert(t, err != nil, "expected err")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected err")
}

func TestUnlock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().Unlock(project, env).Return(nil, expectedErr)

	l := locking.NewClient(mock)
	_, err := l.Unlock("owner/repo/path/env")
	Equals(t, err, err)
}

func TestUnlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().Unlock(project, env).Return(&pl, nil)

	l := locking.NewClient(mock)
	lock, err := l.Unlock("owner/repo/path/env")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestList_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().List().Return(nil, expectedErr)

	l := locking.NewClient(mock)
	_, err := l.List()
	Equals(t, expectedErr, err)
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().List().Return([]models.ProjectLock{pl}, nil)

	l := locking.NewClient(mock)
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{
		"owner/repo/path/env": pl,
	}, list)
}

func TestUnlockByPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().UnlockByPull("owner/repo", 1).Return(nil, expectedErr)

	l := locking.NewClient(mock)
	_, err := l.UnlockByPull("owner/repo", 1)
	Equals(t, expectedErr, err)
}

func TestGetLock_BadKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	l := locking.NewClient(mock)

	_, err := l.GetLock("invalidkey")
	Assert(t, err != nil, "err should not be nil")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected different err")
}

func TestGetLock_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().GetLock(project, env).Return(nil, expectedErr)
	l := locking.NewClient(mock)

	_, err := l.GetLock("owner/repo/path/env")
	Equals(t, expectedErr, err)
}

func TestGetLock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockBackend(ctrl)
	mock.EXPECT().GetLock(project, env).Return(&pl, nil)
	l := locking.NewClient(mock)

	lock, err := l.GetLock("owner/repo/path/env")
	Ok(t, err)
	Equals(t, &pl, lock)
}
