// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	tMocks "github.com/runatlantis/atlantis/server/controllers/web_templates/mocks"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"

	"github.com/gorilla/mux"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"

	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	mocks2 "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestCreateApplyLock(t *testing.T) {
	t.Run("Creates apply lock", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
		w := httptest.NewRecorder()

		layout := "2006-01-02T15:04:05.000Z"
		strLockTime := "2020-09-01T00:45:26.371Z"
		expLockTime := "2020-09-01 00:45:26"
		lockTime, _ := time.Parse(layout, strLockTime)

		ctrl := gomock.NewController(t)
		l := mocks.NewMockApplyLocker(ctrl)
		l.EXPECT().LockApply().Return(locking.ApplyCommandLock{
			Locked: true,
			Time:   lockTime,
		}, nil)

		lc := controllers.LocksController{
			Logger:      logging.NewNoopLogger(t),
			ApplyLocker: l,
		}
		lc.LockApply(w, req)

		ResponseContains(t, w, http.StatusOK, fmt.Sprintf("Apply Lock is acquired on %s", expLockTime))
	})

	t.Run("Apply lock creation fails", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)
		l := mocks.NewMockApplyLocker(ctrl)
		l.EXPECT().LockApply().Return(locking.ApplyCommandLock{
			Locked: false,
		}, errors.New("failed to acquire lock"))

		lc := controllers.LocksController{
			Logger:      logging.NewNoopLogger(t),
			ApplyLocker: l,
		}
		lc.LockApply(w, req)

		ResponseContains(t, w, http.StatusInternalServerError, "creating apply lock failed with: failed to acquire lock")
	})
}

func TestUnlockApply(t *testing.T) {
	t.Run("Apply lock deleted successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)
		l := mocks.NewMockApplyLocker(ctrl)
		l.EXPECT().UnlockApply().Return(nil)

		lc := controllers.LocksController{
			Logger:      logging.NewNoopLogger(t),
			ApplyLocker: l,
		}
		lc.UnlockApply(w, req)

		ResponseContains(t, w, http.StatusOK, "Deleted apply lock")
	})

	t.Run("Apply lock deletion failed", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)
		l := mocks.NewMockApplyLocker(ctrl)
		l.EXPECT().UnlockApply().Return(errors.New("failed to delete lock"))

		lc := controllers.LocksController{
			Logger:      logging.NewNoopLogger(t),
			ApplyLocker: l,
		}
		lc.UnlockApply(w, req)

		ResponseContains(t, w, http.StatusInternalServerError, "deleting apply lock failed with: failed to delete lock")
	})
}

func TestGetLockRoute_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc := controllers.LocksController{
		Logger: logging.NewNoopLogger(t),
	}
	lc.GetLock(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestGetLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	lc := controllers.LocksController{
		Logger: logging.NewNoopLogger(t),
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "%A@"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Invalid lock id")
}

func TestGetLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	ctrl := gomock.NewController(t)
	l := mocks.NewMockLocker(ctrl)
	l.EXPECT().GetLock("id").Return(nil, errors.New("err"))
	lc := controllers.LocksController{
		Logger: logging.NewNoopLogger(t),
		Locker: l,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
	ResponseContains(t, w, http.StatusInternalServerError, "err")
}

func TestGetLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we get a 404")
	ctrl := gomock.NewController(t)
	l := mocks.NewMockLocker(ctrl)
	l.EXPECT().GetLock("id").Return(nil, nil)
	lc := controllers.LocksController{
		Logger: logging.NewNoopLogger(t),
		Locker: l,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
	ResponseContains(t, w, http.StatusNotFound, "No lock found at id 'id'")
}

func TestGetLock_Success(t *testing.T) {
	t.Log("Should be able to render a lock successfully")
	RegisterMockTestingT(t) // needed for pegomock TemplateWriter mock
	ctrl := gomock.NewController(t)
	l := mocks.NewMockLocker(ctrl)
	l.EXPECT().GetLock("id").Return(&models.ProjectLock{
		Project:   models.Project{RepoFullName: "owner/repo", Path: "path"},
		Pull:      models.PullRequest{URL: "url", Author: "lkysow"},
		Workspace: "workspace",
	}, nil)
	tmpl := tMocks.NewMockTemplateWriter()
	atlantisURL, err := url.Parse("https://example.com/basepath")
	Ok(t, err)
	lc := controllers.LocksController{
		Logger:             logging.NewNoopLogger(t),
		Locker:             l,
		LockDetailTemplate: tmpl,
		AtlantisVersion:    "1300135",
		AtlantisURL:        atlantisURL,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
	tmpl.VerifyWasCalledOnce().Execute(w, web_templates.LockDetailData{
		LockKeyEncoded:  "id",
		LockKey:         "id",
		RepoOwner:       "owner",
		RepoName:        "repo",
		PullRequestLink: "url",
		LockedBy:        "lkysow",
		Workspace:       "workspace",
		AtlantisVersion: "1300135",
		CleanedBasePath: "/basepath",
	})
	ResponseContains(t, w, http.StatusOK, "")
}

func TestDeleteLock_NoLockID(t *testing.T) {
	t.Log("If there is no lock ID in the request then we should get a 400")
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	lc := controllers.LocksController{Logger: logging.NewNoopLogger(t)}
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "No lock id in request")
}

func TestDeleteLock_InvalidLockID(t *testing.T) {
	t.Log("If the lock ID is invalid then we should get a 400")
	lc := controllers.LocksController{Logger: logging.NewNoopLogger(t)}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "%A@"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Invalid lock id '%A@'")
}

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	RegisterMockTestingT(t)
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(nil, false, errors.New("err"))
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusInternalServerError, "err")
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we get a 404")
	RegisterMockTestingT(t)
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(nil, false, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusNotFound, "No lock found at id 'id'")
}

func TestDeleteLock_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	RegisterMockTestingT(t)
	cp := vcsmocks.NewMockClient()
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(&models.ProjectLock{}, false, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id 'id'")
	cp.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestDeleteLock_UpdateProjectStatus(t *testing.T) {
	RegisterMockTestingT(t)
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	defer closeTestDatabase(t, database)
	pull := models.PullRequest{Num: 1, BaseRepo: models.Repo{FullName: "owner/repo"}}
	project := models.NewProject(pull.BaseRepo.FullName, "path", "")
	locker := locking.NewClient(database)
	held, err := locker.TryLock(project, "workspace", pull, models.User{})
	Ok(t, err)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, RepoRelDir: "path", Workspace: "workspace", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	Ok(t, err)
	workingDir := mocks2.NewMockWorkingDir()
	lc := controllers.LocksController{
		DeleteLockCommand: &events.DefaultDeleteLockCommand{Locker: locker, Database: database, WorkingDir: workingDir},
		Database:          database, Logger: logging.NewNoopLogger(t), VCSClient: vcsmocks.NewMockClient(),
	}
	req := httptest.NewRequest("DELETE", "/locks", nil)
	req = mux.SetURLVars(req, map[string]string{"id": held.LockKey})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	Equals(t, http.StatusOK, w.Code)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)
}

func TestDeleteLock_CommentFailed(t *testing.T) {
	t.Log("If the commenting fails we still return success")
	RegisterMockTestingT(t)
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(&models.ProjectLock{
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}, true, nil)
	cp := vcsmocks.NewMockClient()
	workingDir := mocks2.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	var database db.Database
	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	Ok(t, err)
	defer closeTestDatabase(t, database)
	When(cp.CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())).ThenReturn(errors.New("err"))
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
		WorkingDir:        workingDir,
		WorkingDirLocker:  workingDirLocker,
		Database:          database,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id 'id'")
}

func TestDeleteLock_CommentSuccess(t *testing.T) {
	t.Log("We should comment back on the pull request if the lock is deleted")
	RegisterMockTestingT(t)
	cp := vcsmocks.NewMockClient()
	dlc := mocks2.NewMockDeleteLockCommand()
	workingDir := mocks2.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	var database db.Database
	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	Ok(t, err)
	defer closeTestDatabase(t, database)

	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(&models.ProjectLock{
		Pull:      pull,
		Workspace: "workspace",
		Project: models.Project{
			Path:         "path",
			RepoFullName: "owner/repo",
		},
	}, true, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
		Database:          database,
		WorkingDir:        workingDir,
		WorkingDirLocker:  workingDirLocker,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id 'id'")
	cp.VerifyWasCalled(Once()).CreateComment(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num),
		Eq("**Warning**: The plan for dir: `path` workspace: `workspace` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again."), Eq(""))
}

func closeTestDatabase(t *testing.T, database db.Database) {
	t.Helper()
	Ok(t, database.Close())
}

func TestDeleteLock_ConflictDoesNotCommentDiscarded(t *testing.T) {
	RegisterMockTestingT(t)
	deleter := mocks2.NewMockDeleteLockCommand()
	client := vcsmocks.NewMockClient()
	When(deleter.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(nil, false, db.ErrPlanStatusNotFound)
	controller := controllers.LocksController{DeleteLockCommand: deleter, VCSClient: client, Logger: logging.NewNoopLogger(t)}
	request := mux.SetURLVars(httptest.NewRequest("DELETE", "/locks", nil), map[string]string{"id": "id"})
	response := httptest.NewRecorder()
	controller.DeleteLock(response, request)
	Equals(t, http.StatusConflict, response.Code)
	client.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestDeleteLock_AppliedUnlockDoesNotCommentDiscarded(t *testing.T) {
	RegisterMockTestingT(t)
	deleter := mocks2.NewMockDeleteLockCommand()
	client := vcsmocks.NewMockClient()
	When(deleter.DeleteLock(Any[logging.SimpleLogging](), Eq("id"), Eq[command.PublicationWriteMode](command.NoClaim{}))).ThenReturn(&models.ProjectLock{Pull: models.PullRequest{Num: 1, BaseRepo: models.Repo{FullName: "owner/repo"}}}, false, nil)
	controller := controllers.LocksController{DeleteLockCommand: deleter, VCSClient: client, Logger: logging.NewNoopLogger(t)}
	request := mux.SetURLVars(httptest.NewRequest("DELETE", "/locks", nil), map[string]string{"id": "id"})
	response := httptest.NewRecorder()
	controller.DeleteLock(response, request)
	Equals(t, http.StatusOK, response.Code)
	client.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestDeleteLock_PublicationCancellationDoesNotDiscard(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		name := "request canceled"
		if shutdown {
			name = "server shutdown"
		}
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			storage, err := boltdb.New(t.TempDir())
			Ok(t, err)
			t.Cleanup(func() { Ok(t, storage.Close()) })
			pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "owner/repo"}}
			locker := mocks.NewMockLocker(gomock.NewController(t))
			locker.EXPECT().GetLock("id").Return(&models.ProjectLock{Pull: pull}, nil)
			deleter := mocks2.NewMockDeleteLockCommand()
			client := vcsmocks.NewMockClient()
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			request := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/locks", nil), map[string]string{"id": "id"})
			coordinator := events.NewPublicationCoordinator(storage, context.Background())
			expected := http.StatusRequestTimeout
			if shutdown {
				coordinator.Shutdown = canceled
				expected = http.StatusServiceUnavailable
			} else {
				request = mux.SetURLVars(request.WithContext(canceled), map[string]string{"id": "id"})
			}
			controller := controllers.LocksController{Publication: coordinator, Locker: locker, DeleteLockCommand: deleter, VCSClient: client, Logger: logging.NewNoopLogger(t)}
			response := httptest.NewRecorder()
			controller.DeleteLock(response, request)
			Equals(t, expected, response.Code)
			deleter.VerifyWasCalled(Never()).DeleteLock(Any[logging.SimpleLogging](), Any[string](), Any[command.PublicationWriteMode]())
			client.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
			lease, err := storage.GetPublicationLease(context.Background(), pull)
			Ok(t, err)
			Assert(t, lease == nil, "canceled discard must not acquire ownership")
		})
	}
}
