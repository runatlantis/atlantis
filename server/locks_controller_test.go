package server_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/locking/mocks"
	mocks2 "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	sMocks "github.com/runatlantis/atlantis/server/mocks"
)

func AnyRepo() models.Repo {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(models.Repo{})))
	return models.Repo{}
}

func TestGetLockRoute_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc := server.LocksController{
		Logger: logging.NewNoopLogger(),
	}
	lc.GetLock(w, req)
	responseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestGetLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	lc := server.LocksController{
		Logger: logging.NewNoopLogger(),
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "%A@"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
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
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
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
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
	responseContains(t, w, http.StatusNotFound, "No lock found at id \"id\"")
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
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
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

func TestDeleteLock_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc := server.LocksController{Logger: logging.NewNoopLogger()}
	lc.DeleteLock(w, req)
	responseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestDeleteLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	lc := server.LocksController{Logger: logging.NewNoopLogger()}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "%A@"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	responseContains(t, w, http.StatusBadRequest, "Invalid lock id \"%A@\"")
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
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
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
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	responseContains(t, w, http.StatusNotFound, "No lock found at id \"id\"")
}

func TestDeleteLock_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	RegisterMockTestingT(t)

	cp := vcsmocks.NewMockClientProxy()
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	lc := server.LocksController{
		Locker:    l,
		Logger:    logging.NewNoopLogger(),
		VCSClient: cp,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	responseContains(t, w, http.StatusOK, "Deleted lock id \"id\"")
	cp.VerifyWasCalled(Never()).CreateComment(AnyRepo(), AnyInt(), AnyString())
}

func TestDeleteLock_CommentFailed(t *testing.T) {
	t.Log("If the commenting fails we return an error")
	RegisterMockTestingT(t)

	cp := vcsmocks.NewMockClientProxy()
	workingDir := mocks2.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	When(cp.CreateComment(AnyRepo(), AnyInt(), AnyString())).ThenReturn(errors.New("err"))
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}, nil)
	lc := server.LocksController{
		Locker:           l,
		Logger:           logging.NewNoopLogger(),
		VCSClient:        cp,
		WorkingDir:       workingDir,
		WorkingDirLocker: workingDirLocker,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	responseContains(t, w, http.StatusInternalServerError, "Failed commenting on pull request: err")
}

func TestDeleteLock_CommentSuccess(t *testing.T) {
	t.Log("We should comment back on the pull request if the lock is deleted")
	RegisterMockTestingT(t)

	cp := vcsmocks.NewMockClientProxy()
	l := mocks.NewMockLocker()
	workingDir := mocks2.NewMockWorkingDir()
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
	lc := server.LocksController{
		Locker:           l,
		Logger:           logging.NewNoopLogger(),
		VCSClient:        cp,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	responseContains(t, w, http.StatusOK, "Deleted lock id \"id\"")
	cp.VerifyWasCalled(Once()).CreateComment(pull.BaseRepo, pull.Num,
		"**Warning**: The plan for dir: `path` workspace: `workspace` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again.")
	workingDir.VerifyWasCalledOnce().DeleteForWorkspace(pull.BaseRepo, pull, "workspace")
}
