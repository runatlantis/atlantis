// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/drift"
	driftmocks "github.com/runatlantis/atlantis/server/core/drift/mocks"
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
	lockTime := time.Now()
	mockLock := models.ProjectLock{
		Project:   models.Project{ProjectName: "terraform", RepoFullName: "owner/repo", Path: "/path"},
		Pull:      models.PullRequest{Num: 123, URL: "url", Author: "lkysow"},
		User:      models.User{Username: "jdoe"},
		Workspace: "default",
		Time:      lockTime,
	}
	mockLocks := map[string]models.ProjectLock{
		"lock-id": mockLock,
	}
	When(ac.Locker.List()).ThenReturn(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	Equals(t, http.StatusOK, w.Code)

	responseBody, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResultAPI
	parseAPIResponse(t, responseBody, &result)

	// Verify the lock details
	Assert(t, len(result.Locks) == 1, "expected 1 lock")
	Equals(t, 1, result.TotalCount)
	lock := result.Locks[0]
	Equals(t, "lock-id", lock.ID)
	Equals(t, "terraform", lock.ProjectName)
	Equals(t, "owner/repo", lock.Repository)
	Equals(t, "/path", lock.Path)
	Equals(t, "default", lock.Workspace)
	Equals(t, 123, lock.PullRequestID)
	Equals(t, "url", lock.PullRequestURL)
	Equals(t, "jdoe", lock.LockedBy)
}

func TestAPIController_ListLocksEmpty(t *testing.T) {
	ac, _, _ := setup(t)

	mockLocks := map[string]models.ProjectLock{}
	When(ac.Locker.List()).ThenReturn(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	Equals(t, http.StatusOK, w.Code)

	responseBody, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResultAPI
	parseAPIResponse(t, responseBody, &result)

	// Verify empty result
	Equals(t, 0, len(result.Locks))
	Equals(t, 0, result.TotalCount)
}

func setup(t *testing.T) (*controllers.APIController, *MockProjectCommandBuilder, *MockProjectCommandRunner) {
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

	ac := &controllers.APIController{
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

// parseAPIResponse is a helper to extract data from the API envelope response.
func parseAPIResponse(t *testing.T, body []byte, target any) {
	t.Helper()
	var envelope controllers.APIResponse
	err := json.Unmarshal(body, &envelope)
	Ok(t, err)
	Assert(t, envelope.Success, "expected success response")

	// Re-marshal data to unmarshal into target type
	dataBytes, err := json.Marshal(envelope.Data)
	Ok(t, err)
	err = json.Unmarshal(dataBytes, target)
	Ok(t, err)
}

// parseAPIError is a helper to extract error from the API envelope response.
func parseAPIError(t *testing.T, body []byte) *controllers.APIError {
	t.Helper()
	var envelope controllers.APIResponse
	err := json.Unmarshal(body, &envelope)
	Ok(t, err)
	Assert(t, !envelope.Success, "expected error response")
	return envelope.Error
}

func TestAPIController_DriftStatus(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	driftStorage := driftmocks.NewMockStorage()
	checkTime := time.Now()
	mockDrifts := []models.ProjectDrift{
		{
			ProjectName: "project1",
			Path:        "modules/vpc",
			Workspace:   "default",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: true,
				ToAdd:    2,
			},
			LastChecked: checkTime,
		},
	}
	When(driftStorage.Get(Eq("owner/repo"), Any[drift.GetOptions]())).ThenReturn(mockDrifts, nil)

	ac := controllers.APIController{
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo", nil)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftStatusAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "owner/repo", result.Repository)
	Equals(t, 1, result.Summary.TotalProjects)
	Equals(t, 1, result.Summary.ProjectsWithDrift)
}

func TestAPIController_DriftStatus_NoStorage(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	// Controller without drift storage
	ac := controllers.APIController{
		Logger:       logger,
		Locker:       locker,
		DriftStorage: nil,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo", nil)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
}

func TestAPIController_DriftStatus_MissingRepository(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	driftStorage := driftmocks.NewMockStorage()

	ac := controllers.APIController{
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
}

func TestAPIController_DriftStatus_WithFilters(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	driftStorage := driftmocks.NewMockStorage()

	checkTime := time.Now()
	mockDrifts := []models.ProjectDrift{
		{
			ProjectName: "project1",
			Path:        "modules/vpc",
			Workspace:   "staging",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: true,
				ToChange: 1,
			},
			LastChecked: checkTime,
		},
	}
	When(driftStorage.Get(Eq("owner/repo"), Any[drift.GetOptions]())).ThenReturn(mockDrifts, nil)

	ac := controllers.APIController{
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo&project=project1&workspace=staging", nil)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusOK, w.Code)

	// Verify the storage was called with correct options
	driftStorage.VerifyWasCalledOnce().Get(Eq("owner/repo"), Any[drift.GetOptions]())
}

func TestAPIController_DriftStatus_Empty(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	driftStorage := driftmocks.NewMockStorage()

	When(driftStorage.Get(Eq("owner/repo"), Any[drift.GetOptions]())).ThenReturn([]models.ProjectDrift{}, nil)

	ac := controllers.APIController{
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo", nil)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftStatusAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "owner/repo", result.Repository)
	Equals(t, 0, result.Summary.TotalProjects)
	Equals(t, 0, result.Summary.ProjectsWithDrift)
}

func TestAPIController_Remediate(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	parser := NewMockEventParsing()
	vcsClient := NewMockClient()
	repoAllowlistChecker, _ := events.NewRepoAllowlistChecker("*")

	remediationService := driftmocks.NewMockRemediationService()
	mockResult := &models.RemediationResult{
		ID:            "test-id",
		Repository:    "owner/repo",
		Ref:           "main",
		Action:        models.RemediationPlanOnly,
		Status:        models.RemediationStatusSuccess,
		TotalProjects: 1,
		SuccessCount:  1,
		Projects: []models.ProjectRemediationResult{
			{
				ProjectName: "project1",
				Status:      models.RemediationStatusSuccess,
				PlanOutput:  "Plan: 0 to add, 0 to change, 0 to destroy",
			},
		},
	}
	When(remediationService.Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())).ThenReturn(mockResult, nil)

	When(vcsClient.GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Eq("owner/repo"))).ThenReturn("https://github.com/owner/repo.git", nil)
	When(parser.ParseAPIPlanRequest(Any[models.VCSHostType](), Eq("owner/repo"), Any[string]())).ThenReturn(models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Hostname: "github.com"},
	}, nil)

	ac := controllers.APIController{
		APISecret:            []byte(atlantisToken),
		Logger:               logger,
		Locker:               locker,
		Parser:               parser,
		VCSClient:            vcsClient,
		RepoAllowlistChecker: repoAllowlistChecker,
		RemediationService:   remediationService,
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.RemediationResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "test-id", result.ID)
	Equals(t, "success", result.Status)
	Equals(t, 1, result.Summary.SuccessCount)
}

func TestAPIController_Remediate_NoService(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: nil, // No remediation service
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
}

func TestAPIController_Remediate_Unauthorized(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	remediationService := driftmocks.NewMockRemediationService()

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, "wrong-token")
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusUnauthorized, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeUnauthorized, apiErr.Code)
}

func TestAPIController_Remediate_MissingRepository(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	remediationService := driftmocks.NewMockRemediationService()

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Ref:  "main",
		Type: "Github",
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
}

func TestAPIController_Remediate_APIDisabled(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	ac := controllers.APIController{
		APISecret: nil, // API disabled
		Logger:    logger,
		Locker:    locker,
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
}

// Phase 5: GetRemediationResult tests

func TestAPIController_GetRemediationResult(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()
	mockResult := &models.RemediationResult{
		ID:            "test-id-123",
		Repository:    "owner/repo",
		Ref:           "main",
		Action:        models.RemediationPlanOnly,
		Status:        models.RemediationStatusSuccess,
		TotalProjects: 1,
		SuccessCount:  1,
		Projects: []models.ProjectRemediationResult{
			{
				ProjectName: "project1",
				Status:      models.RemediationStatusSuccess,
				PlanOutput:  "Plan: 0 to add, 0 to change, 0 to destroy",
			},
		},
	}
	When(remediationService.GetResult(Eq("test-id-123"))).ThenReturn(mockResult, nil)

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id-123", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.RemediationResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "test-id-123", result.ID)
	Equals(t, "success", result.Status)
}

func TestAPIController_GetRemediationResult_NotFound(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()
	When(remediationService.GetResult(Eq("nonexistent-id"))).ThenReturn(nil, fmt.Errorf("remediation result not found: nonexistent-id"))

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate/nonexistent-id", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusNotFound, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeNotFound, apiErr.Code)
}

func TestAPIController_GetRemediationResult_MissingID(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate/", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
}

func TestAPIController_GetRemediationResult_NoService(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: nil, // No service
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
}

// Phase 5: ListRemediationResults tests

func TestAPIController_ListRemediationResults(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()
	mockResults := []*models.RemediationResult{
		{
			ID:            "result-1",
			Repository:    "owner/repo",
			Ref:           "main",
			Action:        models.RemediationPlanOnly,
			Status:        models.RemediationStatusSuccess,
			TotalProjects: 1,
			SuccessCount:  1,
		},
		{
			ID:            "result-2",
			Repository:    "owner/repo",
			Ref:           "develop",
			Action:        models.RemediationAutoApply,
			Status:        models.RemediationStatusPartial,
			TotalProjects: 2,
			SuccessCount:  1,
			FailureCount:  1,
		},
	}
	When(remediationService.ListResults(Eq("owner/repo"), Eq(10))).ThenReturn(mockResults, nil)

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.ListRemediationResults(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var listResponse controllers.RemediationListAPI
	parseAPIResponse(t, response, &listResponse)
	Equals(t, 2, listResponse.Count)
	Equals(t, "result-1", listResponse.Results[0].ID)
}

func TestAPIController_ListRemediationResults_WithLimit(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()
	mockResults := []*models.RemediationResult{
		{
			ID:         "result-1",
			Repository: "owner/repo",
			Status:     models.RemediationStatusSuccess,
		},
	}
	When(remediationService.ListResults(Eq("owner/repo"), Eq(5))).ThenReturn(mockResults, nil)

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo&limit=5", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.ListRemediationResults(w, req)

	Equals(t, http.StatusOK, w.Code)
}

func TestAPIController_ListRemediationResults_MissingRepository(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.ListRemediationResults(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
}

func TestAPIController_ListRemediationResults_Empty(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	remediationService := driftmocks.NewMockRemediationService()
	When(remediationService.ListResults(Eq("owner/repo"), Eq(10))).ThenReturn([]*models.RemediationResult{}, nil)

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: remediationService,
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.ListRemediationResults(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var listResponse controllers.RemediationListAPI
	parseAPIResponse(t, response, &listResponse)
	Equals(t, 0, listResponse.Count)
}

// Phase 5: DetectDrift tests

func TestAPIController_DetectDrift(t *testing.T) {
	// Use the setup function that properly configures all mocks
	ac, _, _ := setup(t)

	// Add drift storage to the controller
	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftDetectionResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "Repo", result.Repository)
}

func TestAPIController_DetectDrift_NoStorage(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()

	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: nil, // No drift storage
	}

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
}

func TestAPIController_DetectDrift_MissingRepository(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	driftStorage := driftmocks.NewMockStorage()

	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	body, _ := json.Marshal(models.DriftDetectionRequest{
		// Missing repository
		Ref:  "main",
		Type: "Github",
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
}

func TestAPIController_DetectDrift_Unauthorized(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker()
	driftStorage := driftmocks.NewMockStorage()

	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, "wrong-token")
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusUnauthorized, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeUnauthorized, apiErr.Code)
}

// TestAPIController_DetectDrift_ReturnsDetectionID verifies that DetectDrift
// returns a non-empty detection ID and that each project's DetectionID matches.
func TestAPIController_DetectDrift_ReturnsDetectionID(t *testing.T) {
	ac, _, _ := setup(t)

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftDetectionResultAPI
	parseAPIResponse(t, response, &result)

	// Detection result must have a non-empty UUID
	Assert(t, result.ID != "", "expected non-empty detection ID")

	// Each project's DetectionID must match the result ID
	for _, p := range result.Projects {
		Equals(t, result.ID, p.DetectionID)
	}
}

// TestAPIController_DetectDrift_AutoDiscoversProjects verifies that when no
// projects or paths are specified in the request, drift detection still discovers
// projects via BuildPlanCommands (auto-discovery).
func TestAPIController_DetectDrift_AutoDiscoversProjects(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	// No projects or paths specified — relies on auto-discovery
	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)

	// BuildPlanCommands should still be called once (auto-discovery with empty CommentCommand)
	projectCommandBuilder.VerifyWasCalled(Times(1)).
		BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
}

// TestAPIController_DetectDrift_PreWorkflowHooksRun verifies that pre-workflow
// hooks are executed during drift detection. The first call runs before project
// discovery (so hooks can generate atlantis.yaml), and the second call runs
// per-project inside apiPlan.
func TestAPIController_DetectDrift_PreWorkflowHooksRun(t *testing.T) {
	ac, _, _ := setup(t)

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)

	// Pre-workflow hooks are called twice:
	// 1) Before project discovery (DetectDrift handler) with Plan command
	// 2) Per-project inside apiPlan with the project-specific command
	_, capturedCmds := preWorkflowHooksRunner.VerifyWasCalled(Times(2)).
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetAllCapturedArguments()

	Assert(t, len(capturedCmds) == 2,
		"expected 2 pre-workflow hook calls, got %d", len(capturedCmds))

	// Both calls should use Plan command
	Assert(t, capturedCmds[0].Name == command.Plan,
		"expected first pre-hook command to be Plan, got %s", capturedCmds[0].Name.String())
	Assert(t, capturedCmds[1].Name == command.Plan,
		"expected second pre-hook command to be Plan, got %s", capturedCmds[1].Name.String())
}

// TestAPIController_DetectDrift_PreWorkflowHooksFailure verifies that when
// FailOnPreWorkflowHookError is true, a pre-workflow hook failure stops drift detection.
func TestAPIController_DetectDrift_PreWorkflowHooksFailure(t *testing.T) {
	ac, _, _ := setup(t)

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage
	ac.FailOnPreWorkflowHookError = true

	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)
	When(preWorkflowHooksRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn(fmt.Errorf("hook failed: generate atlantis.yaml error"))

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeInternal, apiErr.Code)
}
