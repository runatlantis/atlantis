// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
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
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(nil, errors.New("err"))
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Locker:            mockLockLookup(t, "id", nil, errors.New("err"), 1),
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
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(nil, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Locker:            mockLockLookup(t, "id", nil, nil, 1),
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
	legacyLock := &models.ProjectLock{}
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(legacyLock, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Locker:            mockLockLookup(t, "id", legacyLock, nil, 1),
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
	projectLock := &models.ProjectLock{
		Pull:      pull,
		Workspace: workspaceName,
		Project: models.Project{
			Path:         projectPath,
			RepoFullName: repoName,
		},
	}
	When(l.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(projectLock, nil)
	var database db.Database
	tmp := t.TempDir()
	database, err := boltdb.New(tmp)
	Ok(t, err)
	defer closeTestDatabase(t, database)
	// Seed the DB with a successful plan for that project (that is later discarded).
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: projectPath,
			Workspace:  workspaceName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "tf-output",
					LockURL:         "lock-url",
				},
			},
		},
	})
	Ok(t, err)
	lc := controllers.LocksController{
		DeleteLockCommand: l,
		Locker:            mockLockLookup(t, "id", projectLock, nil, 2),
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         cp,
		WorkingDirLocker:  workingDirLocker,
		WorkingDir:        workingDir,
		Database:          database,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	w := httptest.NewRecorder()
	lc.DeleteLock(w, req)
	ResponseContains(t, w, http.StatusOK, "Deleted lock id 'id'")
	status, err := database.GetPullStatus(pull)
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

func TestDeleteLock_ActivePlanGenerationRejectsBeforeUnlock(t *testing.T) {
	RegisterMockTestingT(t)
	const (
		repoName     = "owner/repo"
		projectPath  = "path"
		workspace    = "workspace"
		projectName  = "active-plan"
		generationID = "generation-1"
	)
	pull := models.PullRequest{BaseRepo: models.Repo{FullName: repoName}}
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	defer closeTestDatabase(t, database)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command:     command.Plan,
		RepoRelDir:  projectPath,
		Workspace:   workspace,
		ProjectName: projectName,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "tf-output"},
		},
	}})
	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace:   workspace,
		RepoRelDir:  projectPath,
		ProjectName: projectName,
	}}, generationID)
	Ok(t, err)

	deleteLock := mocks2.NewMockDeleteLockCommand()
	projectLock := &models.ProjectLock{
		Pull:      pull,
		Workspace: workspace,
		Project: models.Project{
			Path:         projectPath,
			ProjectName:  projectName,
			RepoFullName: repoName,
		},
	}
	When(deleteLock.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(projectLock, nil)
	vcsClient := vcsmocks.NewMockClient()
	controller := controllers.LocksController{
		DeleteLockCommand: deleteLock,
		Locker:            mockLockLookup(t, "id", projectLock, nil, 2),
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         vcsClient,
		Database:          database,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	response := httptest.NewRecorder()

	controller.DeleteLock(response, req)

	ResponseContains(t, response, http.StatusConflict, "Lock cannot be deleted while plan generation 'generation-1' is active")
	deleteLock.VerifyWasCalled(Never()).DeleteLock(Any[logging.SimpleLogging](), Any[string]())
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "expected active generation to remain")
	Equals(t, models.ErroredPlanStatus, status.Projects[0].Status)
	Equals(t, generationID, status.Projects[0].PlanGeneration)
}

func TestDeleteLock_RetryAfterUnlockFailureUsesDiscardedTombstone(t *testing.T) {
	RegisterMockTestingT(t)
	pull := models.PullRequest{
		Num:      7,
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	projectLock := &models.ProjectLock{
		Pull:      pull,
		Workspace: events.DefaultWorkspace,
		Project: models.Project{
			Path:         "path/to/project",
			ProjectName:  "named-project",
			RepoFullName: pull.BaseRepo.FullName,
		},
	}
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	defer closeTestDatabase(t, database)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Plan, Workspace: projectLock.Workspace, RepoRelDir: projectLock.Project.Path, ProjectName: projectLock.Project.ProjectName,
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}})
	Ok(t, err)
	deleteLock := &retryDeleteLockCommand{lock: projectLock, firstErr: errors.New("lock backend unavailable")}
	vcsClient := vcsmocks.NewMockClient()
	controller := controllers.LocksController{
		DeleteLockCommand: deleteLock,
		Locker:            mockLockLookup(t, "id", projectLock, nil, 4),
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         vcsClient,
		Database:          database,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})

	first := httptest.NewRecorder()
	controller.DeleteLock(first, req)
	ResponseContains(t, first, http.StatusInternalServerError, "lock backend unavailable")
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "first attempt must retain the durable tombstone")
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)

	second := httptest.NewRecorder()
	controller.DeleteLock(second, req)
	ResponseContains(t, second, http.StatusOK, "Deleted lock id 'id'")
	Equals(t, 2, deleteLock.calls)
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq(""))
}

func TestDeleteLock_AppliedProjectUnlocksIdempotently(t *testing.T) {
	RegisterMockTestingT(t)
	pull := models.PullRequest{Num: 8, BaseRepo: models.Repo{FullName: "owner/repo"}}
	projectLock := &models.ProjectLock{
		Pull: pull, Workspace: events.DefaultWorkspace,
		Project: models.Project{Path: "applied", ProjectName: "applied-project", RepoFullName: pull.BaseRepo.FullName},
	}
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	defer closeTestDatabase(t, database)
	identity := command.ProjectResult{Workspace: projectLock.Workspace, RepoRelDir: projectLock.Project.Path, ProjectName: projectLock.Project.ProjectName}
	plan := identity
	plan.Command = command.Plan
	plan.PlanSuccess = &models.PlanSuccess{}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{plan})
	Ok(t, err)
	apply := identity
	apply.Command = command.Apply
	apply.ApplySuccess = "applied"
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{apply})
	Ok(t, err)
	deleteLock := mocks2.NewMockDeleteLockCommand()
	When(deleteLock.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(projectLock, nil)
	controller := controllers.LocksController{
		DeleteLockCommand: deleteLock,
		Locker:            mockLockLookup(t, "id", projectLock, nil, 2),
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         vcsmocks.NewMockClient(),
		Database:          database,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	response := httptest.NewRecorder()

	controller.DeleteLock(response, req)

	ResponseContains(t, response, http.StatusOK, "Deleted lock id 'id'")
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)
}

func TestDeleteLock_LegacyUnnamedLockRejectsAmbiguousNamedProjects(t *testing.T) {
	RegisterMockTestingT(t)
	pull := models.PullRequest{Num: 9, BaseRepo: models.Repo{FullName: "owner/repo"}}
	legacyLock := &models.ProjectLock{
		Pull: pull, Workspace: events.DefaultWorkspace,
		Project: models.Project{Path: "shared", RepoFullName: pull.BaseRepo.FullName},
	}
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	defer closeTestDatabase(t, database)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		{Command: command.Plan, Workspace: legacyLock.Workspace, RepoRelDir: legacyLock.Project.Path, ProjectName: "project-a", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}},
		{Command: command.Plan, Workspace: legacyLock.Workspace, RepoRelDir: legacyLock.Project.Path, ProjectName: "project-b", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}},
	})
	Ok(t, err)
	deleteLock := mocks2.NewMockDeleteLockCommand()
	controller := controllers.LocksController{
		DeleteLockCommand: deleteLock,
		Locker:            mockLockLookup(t, "id", legacyLock, nil, 2),
		Logger:            logging.NewNoopLogger(t),
		VCSClient:         vcsmocks.NewMockClient(),
		Database:          database,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req = mux.SetURLVars(req, map[string]string{"id": "id"})
	response := httptest.NewRecorder()

	controller.DeleteLock(response, req)

	ResponseContains(t, response, http.StatusConflict, "matches multiple named projects")
	deleteLock.VerifyWasCalled(Never()).DeleteLock(Any[logging.SimpleLogging](), Any[string]())
}

func TestDeleteLock_CommentFailed(t *testing.T) {
	t.Log("If the commenting fails we still return success")
	RegisterMockTestingT(t)
	dlc := mocks2.NewMockDeleteLockCommand()
	projectLock := &models.ProjectLock{
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(projectLock, nil)
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
		Locker:            mockLockLookup(t, "id", projectLock, nil, 2),
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
	projectLock := &models.ProjectLock{
		Pull:      pull,
		Workspace: "workspace",
		Project: models.Project{
			Path:         "path",
			RepoFullName: "owner/repo",
		},
	}
	When(dlc.DeleteLock(Any[logging.SimpleLogging](), Eq("id"))).ThenReturn(projectLock, nil)
	lc := controllers.LocksController{
		DeleteLockCommand: dlc,
		Locker:            mockLockLookup(t, "id", projectLock, nil, 2),
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

type retryDeleteLockCommand struct {
	lock     *models.ProjectLock
	firstErr error
	calls    int
}

func (r *retryDeleteLockCommand) DeleteLock(logging.SimpleLogging, string) (*models.ProjectLock, error) {
	r.calls++
	if r.calls == 1 {
		return nil, r.firstErr
	}
	return r.lock, nil
}

func (r *retryDeleteLockCommand) DeleteLocksByPull(logging.SimpleLogging, string, int) (int, error) {
	return 0, nil
}

func closeTestDatabase(t *testing.T, database db.Database) {
	t.Helper()
	Ok(t, database.Close())
}

func mockLockLookup(t *testing.T, id string, lock *models.ProjectLock, err error, times int) *mocks.MockLocker {
	t.Helper()
	controller := gomock.NewController(t)
	locker := mocks.NewMockLocker(controller)
	locker.EXPECT().GetLock(id).Return(lock, err).Times(times)
	return locker
}
