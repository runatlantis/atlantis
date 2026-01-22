// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"encoding/json"
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
	var result models.DriftStatusResponse
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, "owner/repo", result.Repository)
	Equals(t, 1, result.TotalProjects)
	Equals(t, 1, result.ProjectsWithDrift)
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
	var result models.DriftStatusResponse
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, "owner/repo", result.Repository)
	Equals(t, 0, result.TotalProjects)
	Equals(t, 0, result.ProjectsWithDrift)
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
	var result models.RemediationResult
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, "test-id", result.ID)
	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, result.SuccessCount)
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

	Equals(t, http.StatusBadRequest, w.Code)
}
