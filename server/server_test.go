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
//
package server_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	sMocks "github.com/runatlantis/atlantis/server/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewServer(t *testing.T) {
	t.Log("Run through NewServer constructor")
	tmpDir, err := ioutil.TempDir("", "")
	Ok(t, err)
	_, err = server.NewServer(server.UserConfig{
		DataDir: tmpDir,
	}, server.Config{})
	Ok(t, err)
}

func TestIndex_LockErr(t *testing.T) {
	t.Log("index should return a 503 if unable to list locks")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.List()).ThenReturn(nil, errors.New("err"))
	s := server.Server{
		Locker: l,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.Index(w, eventsReq)
	responseContains(t, w, 503, "Could not retrieve locks: err")
}

func TestIndex_Success(t *testing.T) {
	t.Log("Index should render the index template successfully.")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	// These are the locks that we expect to be rendered.
	now := time.Now()
	locks := map[string]models.ProjectLock{
		"id1": {
			Pull: models.PullRequest{
				Num: 9,
			},
			Project: models.Project{
				RepoFullName: "owner/repo",
			},
			Time: now,
		},
	}
	When(l.List()).ThenReturn(locks, nil)
	it := sMocks.NewMockTemplateWriter()
	r := mux.NewRouter()
	atlantisVersion := "0.3.1"
	// Need to create a lock route since the server expects this route to exist.
	r.NewRoute().Path("").Name(server.LockRouteName)
	s := server.Server{
		Locker:          l,
		IndexTemplate:   it,
		Router:          r,
		AtlantisVersion: atlantisVersion,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.Index(w, eventsReq)
	it.VerifyWasCalledOnce().Execute(w, server.IndexData{
		Locks: []server.LockIndexData{
			{
				LockURL:      "",
				RepoFullName: "owner/repo",
				PullNum:      9,
				Time:         now,
			},
		},
		AtlantisVersion: atlantisVersion,
	})
	responseContains(t, w, http.StatusOK, "")
}

func TestGetLockRoute_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s := server.Server{}
	s.GetLockRoute(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestGetLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	s := server.Server{}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.GetLock(w, eventsReq, "%A@")
	responseContains(t, w, http.StatusBadRequest, "Invalid lock id")
}

func TestGetLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(nil, errors.New("err"))
	s := server.Server{
		Locker: l,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.GetLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusInternalServerError, "err")
}

func TestGetLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we get a 404")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(nil, nil)
	s := server.Server{
		Locker: l,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.GetLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusNotFound, "No lock found at that id")
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
	s := server.Server{
		Locker:             l,
		LockDetailTemplate: tmpl,
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.GetLock(w, eventsReq, "id")
	tmpl.VerifyWasCalledOnce().Execute(w, server.LockDetailData{
		LockKeyEncoded:  "id",
		LockKey:         "id",
		RepoOwner:       "owner",
		RepoName:        "repo",
		PullRequestLink: "url",
		LockedBy:        "lkysow",
		Workspace:       "workspace",
	})
	responseContains(t, w, http.StatusOK, "")
}

func TestDeleteLockRoute_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s := server.Server{Logger: logging.NewNoopLogger()}
	s.DeleteLockRoute(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestDeleteLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	s := server.Server{Logger: logging.NewNoopLogger()}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.DeleteLock(w, eventsReq, "%A@")
	responseContains(t, w, http.StatusBadRequest, "Invalid lock id")
}

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	s := server.Server{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.DeleteLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusInternalServerError, "err")
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we get a 404")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, nil)
	s := server.Server{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.DeleteLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusNotFound, "No lock found at that id")
}

func TestDeleteLock_Success(t *testing.T) {
	t.Log("If the lock is deleted successfully we get a 200")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	s := server.Server{
		Locker: l,
		Logger: logging.NewNoopLogger(),
	}
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.DeleteLock(w, eventsReq, "id")
	responseContains(t, w, http.StatusOK, "Deleted lock id id")
}

func responseContains(t *testing.T, r *httptest.ResponseRecorder, status int, bodySubstr string) {
	Equals(t, status, r.Result().StatusCode)
	body, _ := ioutil.ReadAll(r.Result().Body)
	Assert(t, strings.Contains(string(body), bodySubstr), "exp %q to be contained in %q", bodySubstr, string(body))
}
