// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/locking"
	. "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
)

const atlantisTokenHeader = "X-Atlantis-Token"
const atlantisToken = "token"

func TestAPIController_Plan(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	cases := []struct {
		repository string
		ref        string
		vcsType    string
		pr         int
		projects   []string
		paths      []struct {
			Directory string
			Workspace string
		}
	}{
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			projects:   []string{"default"},
		},
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			pr:         1,
		},
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			paths: []struct {
				Directory string
				Workspace string
			}{
				{
					Directory: ".",
					Workspace: "myworkspace",
				},
				{
					Directory: "./myworkspace2",
					Workspace: "myworkspace2",
				},
			},
		},
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			pr:         1,
			projects:   []string{"test"},
			paths: []struct {
				Directory string
				Workspace string
			}{
				{
					Directory: ".",
					Workspace: "myworkspace",
				},
			},
		},
	}

	expectedCalls := 0
	for _, c := range cases {
		body, _ := json.Marshal(controllers.APIRequest{
			Repository: c.repository,
			Ref:        c.ref,
			Type:       c.vcsType,
			PR:         c.pr,
			Projects:   c.projects,
			Paths:      c.paths,
		})

		req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
		req.Header.Set(atlantisTokenHeader, atlantisToken)
		w := httptest.NewRecorder()
		ac.Plan(w, req)
		ResponseContains(t, w, http.StatusOK, "")

		expectedCalls += len(c.projects)
		expectedCalls += len(c.paths)
	}

	projectCommandBuilder.VerifyWasCalled(Times(expectedCalls)).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Times(expectedCalls)).Plan(Any[command.ProjectContext]())
}

func TestAPIController_Apply(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	cases := []struct {
		repository string
		ref        string
		vcsType    string
		pr         int
		projects   []string
		paths      []struct {
			Directory string
			Workspace string
		}
	}{
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			projects:   []string{"default"},
		},
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			pr:         1,
		},
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			paths: []struct {
				Directory string
				Workspace string
			}{
				{
					Directory: ".",
					Workspace: "myworkspace",
				},
				{
					Directory: "./myworkspace2",
					Workspace: "myworkspace2",
				},
			},
		},
		{
			repository: "Repo",
			ref:        "main",
			vcsType:    "Gitlab",
			pr:         1,
			projects:   []string{"test"},
			paths: []struct {
				Directory string
				Workspace string
			}{
				{
					Directory: ".",
					Workspace: "myworkspace",
				},
			},
		},
	}

	expectedCalls := 0
	for _, c := range cases {
		body, _ := json.Marshal(controllers.APIRequest{
			Repository: c.repository,
			Ref:        c.ref,
			Type:       c.vcsType,
			PR:         c.pr,
			Projects:   c.projects,
			Paths:      c.paths,
		})

		req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
		req.Header.Set(atlantisTokenHeader, atlantisToken)
		w := httptest.NewRecorder()
		ac.Apply(w, req)
		ResponseContains(t, w, http.StatusOK, "")

		expectedCalls += len(c.projects)
		expectedCalls += len(c.paths)
	}

	projectCommandBuilder.VerifyWasCalled(Times(expectedCalls)).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Times(expectedCalls)).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Times(expectedCalls)).Apply(Any[command.ProjectContext]())
}

// TestAPIController_Plan_PreWorkflowHooksReceiveCorrectCommand verifies that when
// calling the Plan API endpoint, the pre-workflow hooks receive a CommentCommand
// with Name set to command.Plan (not the zero value which would be command.Apply).
func TestAPIController_Plan_PreWorkflowHooksReceiveCorrectCommand(t *testing.T) {
	ac, _, _ := setup(t)

	// Get access to the pre-workflow hooks mock for verification
	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	// Capture the CommentCommand passed to RunPreHooks and verify Name is Plan
	_, capturedCmd := preWorkflowHooksRunner.VerifyWasCalled(Times(1)).
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()

	Assert(t, capturedCmd.Name == command.Plan,
		"expected CommentCommand.Name to be Plan (%d), got %s (%d)",
		command.Plan, capturedCmd.Name.String(), capturedCmd.Name)
}

// TestAPIController_Apply_PreWorkflowHooksReceiveCorrectCommand verifies that when
// calling the Apply API endpoint, the pre-workflow hooks receive a CommentCommand
// with Name set to command.Apply for the apply phase (and command.Plan for the
// plan phase that runs first).
func TestAPIController_Apply_PreWorkflowHooksReceiveCorrectCommand(t *testing.T) {
	ac, _, _ := setup(t)

	// Get access to the pre-workflow hooks mock for verification
	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	// Apply calls apiPlan first (which runs pre-hooks with Plan), then apiApply (which runs pre-hooks with Apply)
	// So we expect 2 calls: first with Plan, second with Apply
	_, capturedCmds := preWorkflowHooksRunner.VerifyWasCalled(Times(2)).
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetAllCapturedArguments()

	Assert(t, len(capturedCmds) == 2,
		"expected 2 pre-workflow hook calls, got %d", len(capturedCmds))

	Assert(t, capturedCmds[0].Name == command.Plan,
		"expected first CommentCommand.Name to be Plan (%d), got %s (%d)",
		command.Plan, capturedCmds[0].Name.String(), capturedCmds[0].Name)

	Assert(t, capturedCmds[1].Name == command.Apply,
		"expected second CommentCommand.Name to be Apply (%d), got %s (%d)",
		command.Apply, capturedCmds[1].Name.String(), capturedCmds[1].Name)
}

func TestAPIController_ListLocks(t *testing.T) {
	ac, _, _ := setup(t)
	time := time.Now()
	expected := controllers.ListLocksResult{[]controllers.LockDetail{
		{
			Name:            "lock-id",
			ProjectName:     "terraform",
			ProjectRepo:     "owner/repo",
			ProjectRepoPath: "/path",
			PullID:          123,
			PullURL:         "url",
			User:            "jdoe",
			Workspace:       "default",
			Time:            time,
		},
	},
	}
	mockLock := models.ProjectLock{
		Project:   models.Project{ProjectName: "terraform", RepoFullName: "owner/repo", Path: "/path"},
		Pull:      models.PullRequest{Num: 123, URL: "url", Author: "lkysow"},
		User:      models.User{Username: "jdoe"},
		Workspace: "default",
		Time:      time,
	}
	mockLocks := map[string]models.ProjectLock{
		"lock-id": mockLock,
	}
	When(ac.Locker.List()).ThenReturn(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResult
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, expected, result)
}

func TestAPIController_ListLocksEmpty(t *testing.T) {
	ac, _, _ := setup(t)

	expected := controllers.ListLocksResult{}
	mockLocks := map[string]models.ProjectLock{}
	When(ac.Locker.List()).ThenReturn(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResult
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, expected, result)
}

func TestAPIController_ListLocksManualLock(t *testing.T) {
	ac, _, _ := setup(t)
	now := time.Now()
	mockLock := models.ProjectLock{
		Project:      models.Project{ProjectName: "terraform", RepoFullName: "owner/repo", Path: "/path"},
		User:         models.User{Username: "joseph"},
		Workspace:    "default",
		Time:         now,
		Note:         "Blocked for incident",
		IsManualLock: true,
	}
	When(ac.Locker.List()).ThenReturn(map[string]models.ProjectLock{"lock-id": mockLock}, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResult
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, 1, len(result.Locks))
	Equals(t, true, result.Locks[0].IsManualLock)
	Equals(t, "Blocked for incident", result.Locks[0].Note)
}

func TestAPIController_CreateManualLock_APIDisabled(t *testing.T) {
	ac, _, _ := setup(t)
	ac.APISecret = nil

	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusBadRequest, w.Code)
}

func TestAPIController_CreateManualLock_InvalidToken(t *testing.T) {
	ac, _, _ := setup(t)

	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(`{}`))
	req.Header.Set(atlantisTokenHeader, "wrong-token")
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusUnauthorized, w.Code)
}

func TestAPIController_CreateManualLock_MissingFields(t *testing.T) {
	ac, _, _ := setup(t)

	body := `{"repo_full_name":"owner/repo"}`
	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusBadRequest, w.Code)
}

func TestAPIController_CreateManualLock_Success(t *testing.T) {
	ac, _, _ := setup(t)

	When(ac.Locker.ManualLock(
		Any[models.Project](), Any[string](), Any[string](), Any[models.User](),
	)).ThenReturn(locking.TryLockResponse{
		LockAcquired: true,
		LockKey:      "owner/repo/./default/",
	}, nil)

	body := `{"repo_full_name":"owner/repo","workspace":"default","note":"test lock"}`
	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusOK, w.Code)

	var result controllers.ManualLockAPIResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	Ok(t, err)
	Equals(t, "owner/repo/./default/", result.LockKey)
}

func TestAPIController_CreateManualLock_AlreadyLocked(t *testing.T) {
	ac, _, _ := setup(t)

	When(ac.Locker.ManualLock(
		Any[models.Project](), Any[string](), Any[string](), Any[models.User](),
	)).ThenReturn(locking.TryLockResponse{
		LockAcquired: false,
		LockKey:      "owner/repo/./default/",
	}, nil)

	body := `{"repo_full_name":"owner/repo","workspace":"default","note":"test lock"}`
	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusConflict, w.Code)
}

func TestAPIController_CreateManualLock_LockerError(t *testing.T) {
	ac, _, _ := setup(t)

	When(ac.Locker.ManualLock(
		Any[models.Project](), Any[string](), Any[string](), Any[models.User](),
	)).ThenReturn(locking.TryLockResponse{}, fmt.Errorf("db error"))

	body := `{"repo_full_name":"owner/repo","workspace":"default","note":"test lock"}`
	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusInternalServerError, w.Code)
}

func TestAPIController_CreateManualLock_DefaultPath(t *testing.T) {
	ac, _, _ := setup(t)

	When(ac.Locker.ManualLock(
		Any[models.Project](), Any[string](), Any[string](), Any[models.User](),
	)).ThenReturn(locking.TryLockResponse{
		LockAcquired: true,
		LockKey:      "owner/repo/./default/",
	}, nil)

	// No "path" field — should default to "."
	body := `{"repo_full_name":"owner/repo","workspace":"default","note":"test"}`
	req, _ := http.NewRequest("POST", "", bytes.NewBufferString(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.CreateManualLock(w, req)
	Equals(t, http.StatusOK, w.Code)

	// Verify ManualLock was called with path "."
	project, _, _, _ := ac.Locker.(*MockLocker).VerifyWasCalledOnce().ManualLock(
		Any[models.Project](), Any[string](), Any[string](), Any[models.User](),
	).GetCapturedArguments()
	Equals(t, ".", project.Path)
}

func setup(t *testing.T) (controllers.APIController, *MockProjectCommandBuilder, *MockProjectCommandRunner) {
	RegisterMockTestingT(t)
	locker := NewMockLocker()
	logger := logging.NewNoopLogger(t)
	parser := NewMockEventParsing()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	scope := metricstest.NewLoggingScope(t, logger, "null")
	vcsClient := NewMockClient()
	workingDir := NewMockWorkingDir()
	Ok(t, err)

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Eq(events.DefaultWorkspace), Eq(events.DefaultRepoRelDir), Eq(""), Any[command.Name]())).
		ThenReturn(func() {}, nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
		}}, nil)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Apply,
		}}, nil)

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{},
	})
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		ApplySuccess: "success",
	})

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()

	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()

	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	commitStatusUpdater := NewMockCommitStatusUpdater()

	When(commitStatusUpdater.UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())).ThenReturn(nil)

	ac := controllers.APIController{
		APISecret:                      []byte(atlantisToken),
		Locker:                         locker,
		Logger:                         logger,
		Scope:                          scope,
		Parser:                         parser,
		ProjectCommandBuilder:          projectCommandBuilder,
		ProjectPlanCommandRunner:       projectCommandRunner,
		ProjectApplyCommandRunner:      projectCommandRunner,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		VCSClient:                      vcsClient,
		RepoAllowlistChecker:           repoAllowlistChecker,
		WorkingDir:                     workingDir,
		WorkingDirLocker:               workingDirLocker,
		CommitStatusUpdater:            commitStatusUpdater,
	}
	return ac, projectCommandBuilder, projectCommandRunner
}
