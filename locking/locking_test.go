package locking_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/locking/mocks"
	"github.com/hootsuite/atlantis/models"
	. "github.com/hootsuite/atlantis/testing_util"
	. "github.com/petergtz/pegomock"
	"reflect"
	"strings"
)

var project = models.NewProject("owner/repo", "path")
var env = "env"
var pull = models.PullRequest{}
var user = models.User{}
var expectedErr = errors.New("err")
var timeNow = time.Now().Local()
var pl = models.ProjectLock{project, pull, user, env, timeNow}

func TestTryLock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.TryLock(AnyProjectLock())).ThenReturn(false, models.ProjectLock{}, expectedErr)
	t.Log("when the backend returns an error, TryLock should return that error")
	l := locking.NewClient(backend)
	_, err := l.TryLock(project, env, pull, user)
	Equals(t, err, err)
}

func TestTryLock_Success(t *testing.T) {
	RegisterMockTestingT(t)
	currLock := models.ProjectLock{}
	backend := mocks.NewMockBackend()
	When(backend.TryLock(AnyProjectLock())).ThenReturn(true, currLock, nil)
	l := locking.NewClient(backend)
	r, err := l.TryLock(project, env, pull, user)
	Ok(t, err)
	Equals(t, locking.TryLockResponse{true, currLock, "owner/repo/path/env"}, r)
}

func TestUnlock_InvalidKey(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	l := locking.NewClient(backend)

	_, err := l.Unlock("invalidkey")
	Assert(t, err != nil, "expected err")
	Assert(t, strings.Contains(err.Error(), "invalid key format"), "expected err")
}

func TestUnlock_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.Unlock(AnyProject(), AnyString())).ThenReturn(nil, expectedErr)
	l := locking.NewClient(backend)
	_, err := l.Unlock("owner/repo/path/env")
	Equals(t, err, err)
	backend.VerifyWasCalledOnce().Unlock(project, "env")
}

func TestUnlock(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.Unlock(AnyProject(), AnyString())).ThenReturn(&pl, nil)
	l := locking.NewClient(backend)
	lock, err := l.Unlock("owner/repo/path/env")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func TestList_Err(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.List()).ThenReturn(nil, expectedErr)
	l := locking.NewClient(backend)
	_, err := l.List()
	Equals(t, expectedErr, err)
}

func TestList(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.List()).ThenReturn([]models.ProjectLock{pl}, nil)
	l := locking.NewClient(backend)
	list, err := l.List()
	Ok(t, err)
	Equals(t, map[string]models.ProjectLock{
		"owner/repo/path/env": pl,
	}, list)
}

func TestUnlockByPull(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.UnlockByPull("owner/repo", 1)).ThenReturn(nil, expectedErr)
	l := locking.NewClient(backend)
	_, err := l.UnlockByPull("owner/repo", 1)
	Equals(t, expectedErr, err)
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
	When(backend.GetLock(project, env)).ThenReturn(nil, expectedErr)
	l := locking.NewClient(backend)
	_, err := l.GetLock("owner/repo/path/env")
	Equals(t, expectedErr, err)
}

func TestGetLock(t *testing.T) {
	RegisterMockTestingT(t)
	backend := mocks.NewMockBackend()
	When(backend.GetLock(project, env)).ThenReturn(&pl, nil)
	l := locking.NewClient(backend)
	lock, err := l.GetLock("owner/repo/path/env")
	Ok(t, err)
	Equals(t, &pl, lock)
}

func AnyProjectLock() models.ProjectLock {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(models.ProjectLock{})))
	return models.ProjectLock{}
}

func AnyProject() models.Project {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(models.Project{})))
	return models.Project{}
}
