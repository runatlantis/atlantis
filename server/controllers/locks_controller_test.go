package controllers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	tMocks "github.com/runatlantis/atlantis/server/controllers/templates/mocks"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"

	"github.com/gorilla/mux"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"

	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	mocks2 "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func AnyRepo() models.Repo {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(models.Repo{})))
	return models.Repo{}
}

func TestCreateApplyLock(t *testing.T) {
	t.Run("Creates apply lock", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
		w := httptest.NewRecorder()

		layout := "2006-01-02T15:04:05.000Z"
		strLockTime := "2020-09-01T00:45:26.371Z"
		expLockTime := "2020-09-01 00:45:26"
		lockTime, _ := time.Parse(layout, strLockTime)

		l := mocks.NewMockApplyLocker()
		When(l.LockApply()).ThenReturn(locking.ApplyCommandLock{
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

		l := mocks.NewMockApplyLocker()
		When(l.LockApply()).ThenReturn(locking.ApplyCommandLock{
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

		l := mocks.NewMockApplyLocker()
		When(l.UnlockApply()).ThenReturn(nil)

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

		l := mocks.NewMockApplyLocker()
		When(l.UnlockApply()).ThenReturn(errors.New("failed to delete lock"))

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
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(nil, errors.New("err"))
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
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.GetLock("id")).ThenReturn(nil, nil)
	lc := controllers.LocksController{
		Logger: logging.NewNoopLogger(t),
		Locker: l,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.GetLock(w, req)
	ResponseContains(t, w, http.StatusNotFound, "No lock found at id \"id\"")
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
	tmpl.VerifyWasCalledOnce().Execute(w, templates.LockDetailData{
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
	ResponseContains(t, w, http.StatusBadRequest, "Invalid lock id \"%A@\"")
}

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, a 500 is returned")
	RegisterMockTestingT(t)
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock("id")).ThenReturn(nil, errors.New("err"))
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
	When(dlc.DeleteLock("id")).ThenReturn(nil, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusNotFound, "No lock found at id \"id\"")
}

func TestDeleteLock_OldFormat(t *testing.T) {
	t.Log("If the lock doesn't have BaseRepo set it is deleted successfully")
	RegisterMockTestingT(t)
	cp := vcsmocks.NewMockClient()
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock("id")).ThenReturn(&models.ProjectLock{}, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id \"id\"")
	cp.VerifyWasCalled(Never()).CreateComment(AnyRepo(), AnyInt(), AnyString(), AnyString())
}

func TestDeleteLock_UpdateProjectStatus(t *testing.T) {
	t.Log("When deleting a lock, pull status has to be updated to reflect discarded plan")
	RegisterMockTestingT(t)

	repoName := "owner/repo"
	projectPath := "path"
	workspaceName := "workspace"

	cp := vcsmocks.NewMockClient()
	l := mocks2.NewMockDeleteLockCommand()
	workingDir := mocks2.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: repoName},
	}
	When(l.DeleteLock("id")).ThenReturn(&models.ProjectLock{
		Pull:      pull,
		Workspace: workspaceName,
		Project: models.Project{
			Path:         projectPath,
			RepoFullName: repoName,
		},
	}, nil)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	db, err := db.New(tmp)
	Ok(t, err)
	// Seed the DB with a successful plan for that project (that is later discarded).
	_, err = db.UpdatePullWithResults(pull, []models.ProjectResult{
		{
			Command:    models.PlanCommand,
			RepoRelDir: projectPath,
			Workspace:  workspaceName,
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "tf-output",
				LockURL:         "lock-url",
			},
		},
	})
	Ok(t, err)
	lc := controllers.LocksController{
		DeleteLockCommand: l,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
		WorkingDirLocker:  workingDirLocker,
		WorkingDir:        workingDir,
		DB:                db,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id \"id\"")
	status, err := db.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status.Projects != nil, "status projects was nil")
	Equals(t, []models.ProjectStatus{
		{
			Workspace:  workspaceName,
			RepoRelDir: projectPath,
			Status:     models.DiscardedPlanStatus,
		},
	}, status.Projects)
}

func TestDeleteLock_CommentFailed(t *testing.T) {
	t.Log("If the commenting fails we still return success")
	RegisterMockTestingT(t)
	dlc := mocks2.NewMockDeleteLockCommand()
	When(dlc.DeleteLock("id")).ThenReturn(&models.ProjectLock{
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}, nil)
	cp := vcsmocks.NewMockClient()
	workingDir := mocks2.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	When(cp.CreateComment(AnyRepo(), AnyInt(), AnyString(), AnyString())).ThenReturn(errors.New("err"))
	tmp, cleanup := TempDir(t)
	defer cleanup()
	db, err := db.New(tmp)
	Ok(t, err)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
		WorkingDir:        workingDir,
		WorkingDirLocker:  workingDirLocker,
		DB:                db,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id \"id\"")
}

func TestDeleteLock_CommentSuccess(t *testing.T) {
	t.Log("We should comment back on the pull request if the lock is deleted")
	RegisterMockTestingT(t)
	cp := vcsmocks.NewMockClient()
	dlc := mocks2.NewMockDeleteLockCommand()
	workingDir := mocks2.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	When(dlc.DeleteLock("id")).ThenReturn(&models.ProjectLock{
		Pull:      pull,
		Workspace: "workspace",
		Project: models.Project{
			Path:         "path",
			RepoFullName: "owner/repo",
		},
	}, nil)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	db, err := db.New(tmp)
	Ok(t, err)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
		DB:                db,
		WorkingDir:        workingDir,
		WorkingDirLocker:  workingDirLocker,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id \"id\"")
	cp.VerifyWasCalled(Once()).CreateComment(pull.BaseRepo, pull.Num,
		"**Warning**: The plan for dir: `path` workspace: `workspace` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again.", "")
}
