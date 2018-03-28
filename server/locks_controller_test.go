package server_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	sMocks "github.com/runatlantis/atlantis/server/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func AnyRepo() models.Repo {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(models.Repo{})))
	return models.Repo{}
}

func TestGetLockRoute_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc := server.LocksController{
		Logger: logging.NewNoopLogger(),
	}
	lc.GetLockRoute(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestGetLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	lc := server.LocksController{
		Logger: logging.NewNoopLogger(),
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.GetLock(w, eventsReq, "%A@")
	responseContains(t, w, http.StatusBadRequest, "Invalid lock id")
}

func TestGetLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(nil, errors.New("err"))
	lc := server.LocksController{
		Logger: logging.NewNoopLogger(),
		Locker: l,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.GetLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusInternalServerError, "err")
}

func TestGetLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we get a 404")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(nil, nil)
	lc := server.LocksController{
		Logger: logging.NewNoopLogger(),
		Locker: l,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.GetLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusNotFound, "no corresponding lock for given id")
}

func TestGetLock_Success(t *testing.T) {
	t.Log("Should be able to render a lock successfully")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(&models.ProjectLock{
		Project:   models.Project{RepoFullName: "owner/repo", Path: "path"},
		Pull:      models.PullRequest{URL: "url", Author: "lkysow"},
		Workspace: "workspace",
	}, nil)
	tmpl := sMocks.NewMockTemplateWriter()
	lc := server.LocksController{
		Logger:             logging.NewNoopLogger(),
		Locker:             l,
		LockDetailTemplate: tmpl,
		AtlantisVersion:    "1300135",
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.GetLock(w, eventsReq, "id")
	t.Log(w.Code)
	tmpl.VerifyWasCalledOnce().Execute(w, server.LockDetailData{
		LockKeyEncoded:  "id",
		LockKey:         "id",
		RepoOwner:       "owner",
		RepoName:        "repo",
		PullRequestLink: "url",
		LockedBy:        "lkysow",
		Workspace:       "workspace",
		AtlantisVersion: "1300135",
	})
	responseContains(t, w, http.StatusOK, "")
}

func TestDeleteLockRoute_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc := server.LocksController{Logger: logging.NewNoopLogger()}
	lc.DeleteLockRoute(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestDeleteLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	lc := server.LocksController{Logger: logging.NewNoopLogger()}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.DeleteLock(w, eventsReq, "%A@")
	responseContains(t, w, http.StatusBadRequest, "Invalid lock id")
}

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	lc := server.LocksController{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.DeleteLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusInternalServerError, "err")
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we get a 404")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, nil)
	lc := server.LocksController{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc.DeleteLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusNotFound, "no corresponding lock for given id")
}

func TestDeleteLock_Success(t *testing.T) {
	t.Log("If the lock is deleted successfully we get a 200")
	cp := vcsmocks.NewMockClientProxy()
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	lc := server.LocksController{
		Locker:    l,
		Logger:    logging.NewNoopLogger(),
		VCSClient: cp,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lock := lc.DeleteLock(w, eventsReq, "id")
	Equals(t, &models.ProjectLock{}, lock)
}

func TestDeleteLock_CommentFailed(t *testing.T) {
	t.Log("If the lock is deleted successfully we get a 200")
	cp := vcsmocks.NewMockClientProxy()
	RegisterMockTestingT(t)
	When(cp.CreateComment(AnyRepo(), AnyInt(), AnyString())).ThenReturn(nil)
	lc := server.LocksController{
		Logger:    logging.NewNoopLogger(),
		VCSClient: cp,
	}
	lock := &models.ProjectLock{}
	err := lc.CommentOnPullRequest(lock)
	Equals(t, err, nil)
}
