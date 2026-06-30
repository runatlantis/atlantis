// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
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
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

const atlantisTokenHeader = "X-Atlantis-Token"
const atlantisToken = "token"

type recordingDriftSender struct {
	calls   int
	results []webhooks.DriftResult
}

func (r *recordingDriftSender) Send(_ logging.SimpleLogging, result webhooks.DriftResult) error {
	r.calls++
	r.results = append(r.results, result)
	return nil
}

type fakeControllerLivePullHeadFetcher struct {
	pull models.PullRequest
	err  error
}

func (f fakeControllerLivePullHeadFetcher) GetLivePullIdentity(command.ProjectContext) (models.PullRequest, error) {
	return f.pull, f.err
}

type recordingPullStatusFetcher struct {
	statuses []*models.PullStatus
	calls    []models.PullRequest
}

func (f *recordingPullStatusFetcher) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	f.calls = append(f.calls, pull)
	if len(f.statuses) == 0 {
		return nil, nil
	}
	status := f.statuses[0]
	f.statuses = f.statuses[1:]
	return status, nil
}

func TestAPIController_Plan(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	cases := []struct {
		repository string
		ref        string
		vcsType    string
		pr         int
		projects   []string
		paths      []controllers.APIRequestPath
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
			paths: []controllers.APIRequestPath{
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
			paths: []controllers.APIRequestPath{
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

func TestAPIController_PlanSortsByExecutionOrder(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			commentCommand := args[1].(*events.CommentCommand)
			orderGroup := 2
			if commentCommand.ProjectName == "first-group" {
				orderGroup = 1
			}
			return ReturnValues{[]command.ProjectContext{{
				CommandName:               command.Plan,
				ProjectName:               commentCommand.ProjectName,
				ExecutionOrderGroup:       orderGroup,
				AbortOnExecutionOrderFail: true,
			}}, nil}
		})
	var planOrder []string
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			planOrder = append(planOrder, projectCtx.ProjectName)
			return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"second-group", "first-group"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	Equals(t, http.StatusOK, w.Code)
	Equals(t, []string{"first-group", "second-group"}, planOrder)
}

func TestAPIController_PlanProjectFailureReturnsLegacyNon2xx(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		Error: errors.New("plan failed"),
	})

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

	Equals(t, http.StatusInternalServerError, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	Assert(t, !strings.Contains(string(responseBody), "\"success\":true"), "legacy plan response must not use success envelope: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "\"ProjectResults\""), "expected legacy ProjectResults body: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "\"Error\":{}"), "expected legacy project Error object in response: %s", responseBody)
	Assert(t, !strings.Contains(string(responseBody), "plan failed"), "legacy project Error must not be stringified: %s", responseBody)
}

func TestAPIController_PlanFailsClosedOnTeamAllowlistDenial(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	var capturedCtx *command.Context
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			return ReturnValues{[]command.ProjectContext{}, nil}
		})

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

	Equals(t, http.StatusOK, w.Code)
	Assert(t, capturedCtx != nil, "expected BuildPlanCommands to be called")
	Assert(t, capturedCtx.FailOnTeamAllowlistDenied, "API plan should fail closed on team allowlist denial")
}

func TestAPIController_PlanSuccessReturnsLegacyShape(t *testing.T) {
	ac, _, _ := setup(t)

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

	Equals(t, http.StatusOK, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	Assert(t, !strings.Contains(string(responseBody), "\"success\":true"), "legacy plan response must not use success envelope: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "\"ProjectResults\""), "expected legacy ProjectResults body: %s", responseBody)
}

func TestAPIController_PlanPublishesNormalCommitStatus(t *testing.T) {
	ac, _, _ := setup(t)
	commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

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

	Equals(t, http.StatusOK, w.Code)
	commitStatusUpdater.VerifyWasCalled(Times(1)).
		UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.PendingCommitStatus), Eq(command.Plan))
}

func TestAPIController_PlanRunsPolicyChecksForAPI(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)
	postWorkflowHooksRunner := ac.PostWorkflowHooksCommandRunner.(*MockPostWorkflowHooksCommandRunner)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: "policy", Passed: true}},
		}})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	Equals(t, http.StatusOK, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	var result command.Result
	Ok(t, json.Unmarshal(responseBody, &result))
	Equals(t, 2, len(result.ProjectResults))
	Equals(t, command.Plan, result.ProjectResults[0].Command)
	Equals(t, command.PolicyCheck, result.ProjectResults[1].Command)
	projectCommandRunner.VerifyWasCalled(Once()).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Once()).PolicyCheck(Any[command.ProjectContext]())

	_, preHookCmds := preWorkflowHooksRunner.VerifyWasCalled(Times(1)).
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetAllCapturedArguments()
	Assert(t, len(preHookCmds) == 1 && preHookCmds[0] != nil, "expected one non-nil pre-workflow hook command")
	Equals(t, command.Plan, preHookCmds[0].Name)
	Equals(t, "app", preHookCmds[0].ProjectName)

	_, postHookCmds := postWorkflowHooksRunner.VerifyWasCalled(Times(1)).
		RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetAllCapturedArguments()
	Assert(t, len(postHookCmds) == 1 && postHookCmds[0] != nil, "expected one non-nil post-workflow hook command")
	Equals(t, command.Plan, postHookCmds[0].Name)
	Equals(t, "app", postHookCmds[0].ProjectName)
}

func TestAPIController_PlanPropagatesPolicyCheckFailure(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{Failure: "policy failed"})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	ResponseContains(t, w, http.StatusInternalServerError, "policy failed")
	projectCommandRunner.VerifyWasCalled(Once()).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Once()).PolicyCheck(Any[command.ProjectContext]())
}

func TestAPIController_PlanLoadsPullStatusBeforePolicyChecks(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	policyStatus := []models.PolicySetStatus{{
		PolicySetName: "policy",
		Passed:        false,
		Approvals: []models.PolicySetApproval{{
			Approver: "owner",
			Hashes:   []string{"hash"},
		}},
		Hashes: []string{"hash"},
	}}
	fetcher := &recordingPullStatusFetcher{statuses: []*models.PullStatus{{
		Pull: models.PullRequest{Num: 42},
		Projects: []models.ProjectStatus{{
			ProjectName:  "app",
			RepoRelDir:   "app",
			Workspace:    events.DefaultWorkspace,
			PolicyStatus: policyStatus,
		}},
	}}}
	ac.PullStatusFetcher = fetcher

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			Assert(t, ctx.PullStatus != nil, "expected pull status before building plan commands")
			return ReturnValues{[]command.ProjectContext{
				{
					CommandName: command.Plan,
					ProjectName: "app",
					RepoRelDir:  "app",
					Workspace:   events.DefaultWorkspace,
				},
				{
					CommandName:         command.PolicyCheck,
					ProjectName:         "app",
					RepoRelDir:          "app",
					Workspace:           events.DefaultWorkspace,
					ProjectPolicyStatus: ctx.PullStatus.Projects[0].PolicyStatus,
				},
			}, nil}
		})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: "policy", Passed: true}},
		}})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Equals(t, 1, len(fetcher.calls))
	policyCtx := projectCommandRunner.VerifyWasCalled(Once()).
		PolicyCheck(Any[command.ProjectContext]()).
		GetCapturedArguments()
	Equals(t, policyStatus, policyCtx.ProjectPolicyStatus)
}

func TestAPIController_PlanWithoutPRDoesNotLoadPullStatus(t *testing.T) {
	ac, _, _ := setup(t)
	fetcher := &recordingPullStatusFetcher{statuses: []*models.PullStatus{{
		Projects: []models.ProjectStatus{{ProjectName: "app"}},
	}}}
	ac.PullStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Equals(t, 0, len(fetcher.calls))
}

func TestAPIController_ApplyWithoutPRDoesNotLoadPullStatus(t *testing.T) {
	ac, _, _ := setup(t)
	fetcher := &recordingPullStatusFetcher{statuses: []*models.PullStatus{{
		Projects: []models.ProjectStatus{{ProjectName: "app"}},
	}}}
	ac.PullStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Equals(t, 0, len(fetcher.calls))
}

func TestAPIController_ApplyReportsPolicyCheckFailurePerProject(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName:       command.Apply,
			ProjectName:       "app",
			RepoRelDir:        "app",
			Workspace:         events.DefaultWorkspace,
			PullStatus:        &models.PullStatus{},
			ApplyRequirements: []string{"policies_passed"},
		}}, nil)
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{Failure: "policy failed"})
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{Failure: "All policies must pass for project before running apply."})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusInternalServerError, "All policies must pass")
	projectCommandRunner.VerifyWasCalled(Once()).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Once()).PolicyCheck(Any[command.ProjectContext]())
	projectCommandBuilder.VerifyWasCalled(Once()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Once()).Apply(Any[command.ProjectContext]())
}

func TestAPIController_ApplyContinuesAfterMixedPolicyCheckFailure(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			cmd := args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{
				{
					CommandName: command.Plan,
					ProjectName: cmd.ProjectName,
					RepoRelDir:  cmd.ProjectName,
					Workspace:   events.DefaultWorkspace,
				},
				{
					CommandName: command.PolicyCheck,
					ProjectName: cmd.ProjectName,
					RepoRelDir:  cmd.ProjectName,
					Workspace:   events.DefaultWorkspace,
				},
			}, nil}
		})
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			cmd := args[1].(*events.CommentCommand)
			Assert(t, ctx.PullStatus != nil, "expected seeded pull status before apply command build")
			return ReturnValues{[]command.ProjectContext{{
				CommandName:       command.Apply,
				ProjectName:       cmd.ProjectName,
				RepoRelDir:        cmd.ProjectName,
				Workspace:         events.DefaultWorkspace,
				PullStatus:        ctx.PullStatus,
				ApplyRequirements: []string{"policies_passed"},
			}}, nil}
		})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			if projectCtx.ProjectName == "app-a" {
				return ReturnValues{command.ProjectCommandOutput{Failure: "policy failed for app-a"}}
			}
			return ReturnValues{command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
				PolicySetResults: []models.PolicySetResult{{PolicySetName: "policy", Passed: true}},
			}}}
		})

	var applyCalls []string
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			applyCalls = append(applyCalls, projectCtx.ProjectName)
			if projectCtx.ProjectName == "app-a" {
				return ReturnValues{command.ProjectCommandOutput{Failure: "All policies must pass for project before running apply."}}
			}
			return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "applied app-b"}}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app-a", "app-b"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	Assert(t, strings.Contains(string(responseBody), "applied app-b"), "expected app-b apply result: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "All policies must pass"), "expected app-a policy failure: %s", responseBody)
	projectCommandBuilder.VerifyWasCalled(Times(2)).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Times(2)).Apply(Any[command.ProjectContext]())
	Equals(t, []string{"app-a", "app-b"}, applyCalls)
}

func TestAPIController_NonPRSetupErrorCleansSyntheticWorkingDir(t *testing.T) {
	ac, _, _ := setup(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn("", errors.New("clone failed"))
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(nil)

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

	Equals(t, http.StatusInternalServerError, w.Code)
	workingDir.VerifyWasCalled(Once()).Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
}

func TestAPIController_PlanSetupErrorRedactsCredentials(t *testing.T) {
	ac, _, _ := setup(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn("", errors.New("fatal: Authentication failed for 'https://user:super-secret@example.com/Repo.git'"))
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(nil)

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

	Equals(t, http.StatusInternalServerError, w.Code)
	responseBody := w.Body.String()
	Assert(t, !strings.Contains(responseBody, "super-secret"), "response leaked token: %s", responseBody)
	Assert(t, !strings.Contains(responseBody, "user:super-secret"), "response leaked credentials: %s", responseBody)
	Assert(t, strings.Contains(responseBody, "redacted"), "response should preserve redacted URL context: %s", responseBody)
}

func TestAPIController_DetectDriftSetupErrorRedactsCredentials(t *testing.T) {
	ac, _, _ := setup(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn("", errors.New("fatal: Authentication failed for 'https://user:super-secret@example.com/Repo.git'"))
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(nil)
	ac.DriftStorage = driftmocks.NewMockStorage()

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
	responseBody := w.Body.String()
	Assert(t, !strings.Contains(responseBody, "super-secret"), "response leaked token: %s", responseBody)
	Assert(t, !strings.Contains(responseBody, "user:super-secret"), "response leaked credentials: %s", responseBody)
}

func TestAPIController_LegacyPlanApplyErrorsReturnLegacyShape(t *testing.T) {
	type legacyHandler struct {
		name string
		call func(*controllers.APIController, http.ResponseWriter, *http.Request)
	}
	handlers := []legacyHandler{
		{name: "plan", call: (*controllers.APIController).Plan},
		{name: "apply", call: (*controllers.APIController).Apply},
	}
	validBody, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})

	for _, handler := range handlers {
		for _, tc := range []struct {
			name       string
			body       string
			secret     string
			apiSecret  []byte
			statusCode int
		}{
			{
				name:       "auth failure",
				body:       string(validBody),
				apiSecret:  []byte(atlantisToken),
				statusCode: http.StatusUnauthorized,
			},
			{
				name:       "API disabled",
				body:       string(validBody),
				secret:     atlantisToken,
				apiSecret:  nil,
				statusCode: http.StatusServiceUnavailable,
			},
			{
				name:       "malformed JSON",
				body:       "{",
				secret:     atlantisToken,
				apiSecret:  []byte(atlantisToken),
				statusCode: http.StatusBadRequest,
			},
			{
				name:       "missing fields",
				body:       "{}",
				secret:     atlantisToken,
				apiSecret:  []byte(atlantisToken),
				statusCode: http.StatusBadRequest,
			},
		} {
			t.Run(handler.name+" "+tc.name, func(t *testing.T) {
				ac, _, _ := setup(t)
				ac.APISecret = tc.apiSecret

				req, _ := http.NewRequest("POST", "", strings.NewReader(tc.body))
				if tc.secret != "" {
					req.Header.Set(atlantisTokenHeader, tc.secret)
				}
				w := httptest.NewRecorder()
				handler.call(ac, w, req)

				Equals(t, tc.statusCode, w.Code)
				responseBody, _ := io.ReadAll(w.Result().Body)
				var legacyErr map[string]string
				Ok(t, json.Unmarshal(responseBody, &legacyErr))
				Assert(t, legacyErr["error"] != "", "expected legacy error field: %s", responseBody)
				Assert(t, !strings.Contains(string(responseBody), "\"success\""), "legacy error must not use success envelope: %s", responseBody)
				Assert(t, !strings.Contains(string(responseBody), "\"request_id\""), "legacy error must not include request_id envelope field: %s", responseBody)
			})
		}
	}
}

func TestAPIController_Apply(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	cases := []struct {
		repository string
		ref        string
		vcsType    string
		pr         int
		projects   []string
		paths      []controllers.APIRequestPath
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
			paths: []controllers.APIRequestPath{
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
			paths: []controllers.APIRequestPath{
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

func TestAPIController_ApplySortsByExecutionOrder(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	buildCommands := func(args []Param) ReturnValues {
		commentCommand := args[1].(*events.CommentCommand)
		orderGroup := 2
		if commentCommand.ProjectName == "first-group" {
			orderGroup = 1
		}
		return ReturnValues{[]command.ProjectContext{{
			CommandName:               commentCommand.Name,
			ProjectName:               commentCommand.ProjectName,
			ExecutionOrderGroup:       orderGroup,
			AbortOnExecutionOrderFail: true,
		}}, nil}
	}
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(buildCommands)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(buildCommands)

	var planOrder, applyOrder []string
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			planOrder = append(planOrder, projectCtx.ProjectName)
			return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
		})
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			applyOrder = append(applyOrder, projectCtx.ProjectName)
			return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "success"}}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"second-group", "first-group"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	Equals(t, http.StatusOK, w.Code)
	Equals(t, []string{"first-group", "second-group"}, planOrder)
	Equals(t, []string{"first-group", "second-group"}, applyOrder)
}

func TestAPIController_ApplyProjectFailureReturnsLegacyNon2xx(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		Error: errors.New("apply failed"),
	})

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

	Equals(t, http.StatusInternalServerError, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	Assert(t, !strings.Contains(string(responseBody), "\"success\":true"), "legacy apply response must not use success envelope: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "\"ProjectResults\""), "expected legacy ProjectResults body: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "\"Error\":{}"), "expected legacy project Error object in response: %s", responseBody)
	Assert(t, !strings.Contains(string(responseBody), "apply failed"), "legacy project Error must not be stringified: %s", responseBody)
}

func TestAPIController_ApplyContinuesAfterPrePlanProjectError(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			cmd := args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: cmd.ProjectName,
				RepoRelDir:  cmd.RepoRelDir,
				Workspace:   events.DefaultWorkspace,
			}}, nil}
		})
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			cmd := args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Apply,
				ProjectName: cmd.ProjectName,
				RepoRelDir:  cmd.RepoRelDir,
				Workspace:   events.DefaultWorkspace,
			}}, nil}
		})
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			if projectCtx.ProjectName == "broken" {
				return ReturnValues{command.ProjectCommandOutput{Error: errors.New("plan failed")}}
			}
			return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."}}}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"broken", "ready"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	Equals(t, http.StatusOK, w.Code)
	projectCommandRunner.VerifyWasCalled(Times(2)).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Times(2)).Apply(Any[command.ProjectContext]())
}

func TestAPIController_NoPRRequestsUseSyntheticHardenedAPIContext(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*controllers.APIController, http.ResponseWriter, *http.Request)
	}{
		{name: "plan", call: (*controllers.APIController).Plan},
		{name: "apply", call: (*controllers.APIController).Apply},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ac, projectCommandBuilder, _ := setup(t)
			repoDir, _ := initAPIControllerGitRepo(t)
			workingDir := ac.WorkingDir.(*MockWorkingDir)
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
				ThenReturn(repoDir, nil)
			var capturedCtx *command.Context
			When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
				Then(func(args []Param) ReturnValues {
					capturedCtx = args[0].(*command.Context)
					return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
				})

			body, _ := json.Marshal(controllers.APIRequest{
				Repository: "Repo",
				Ref:        "main",
				Type:       "Gitlab",
				Projects:   []string{"default"},
			})
			req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			tc.call(ac, w, req)

			Equals(t, http.StatusOK, w.Code)
			Assert(t, capturedCtx != nil, "expected plan command builder to be called")
			Assert(t, capturedCtx.Pull.Num < 0, "expected no-PR API request to use an isolated synthetic pull number")
			Assert(t, capturedCtx.Pull.HardenedNonPRRefCheckout, "expected no-PR API request to use hardened checkout")
			Assert(t, capturedCtx.SkipPRModifiedFiles, "expected no-PR API request to skip PR modified-file lookups")
			Assert(t, capturedCtx.FailOnTeamAllowlistDenied, "expected no-PR API request to fail closed on team allowlist denial")
			Assert(t, capturedCtx.RunPolicyChecks, "expected API request to run policy checks when generated")
			Assert(t, capturedCtx.SortByExecutionOrder, "expected API request to honor execution-order sorting")
			Assert(t, capturedCtx.ExactProjectNameMatching, "expected API project selectors to use exact names")
		})
	}
}

func TestAPIController_ApplyFailsClosedOnTeamAllowlistDenial(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	var capturedCtx *command.Context
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			return ReturnValues{[]command.ProjectContext{}, nil}
		})

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

	Equals(t, http.StatusOK, w.Code)
	Assert(t, capturedCtx != nil, "expected BuildPlanCommands to be called")
	Assert(t, capturedCtx.FailOnTeamAllowlistDenied, "API apply should fail closed on team allowlist denial")
}

func TestAPIController_ApplySuccessReturnsLegacyShape(t *testing.T) {
	ac, _, _ := setup(t)

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

	Equals(t, http.StatusOK, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	Assert(t, !strings.Contains(string(responseBody), "\"success\":true"), "legacy apply response must not use success envelope: %s", responseBody)
	Assert(t, strings.Contains(string(responseBody), "\"ProjectResults\""), "expected legacy ProjectResults body: %s", responseBody)
}

func TestAPIController_ApplyPlanFailureStillRunsApplyForLegacyCompatibility(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		Error: errors.New("plan failed"),
	})

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

	Equals(t, http.StatusOK, w.Code)
	projectCommandRunner.VerifyWasCalled(Times(1)).Apply(Any[command.ProjectContext]())
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

func TestAPIController_Plan_SkipsIgnoredPathsWithoutShiftingHookCommands(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(args []Param) ReturnValues {
		commentCommand := args[1].(*events.CommentCommand)
		if commentCommand.RepoRelDir == "ignored" {
			return ReturnValues{[]command.ProjectContext{}, events.ErrIgnoredTargetedDir}
		}
		return ReturnValues{[]command.ProjectContext{{
			CommandName: command.Plan,
			RepoRelDir:  commentCommand.RepoRelDir,
			Workspace:   commentCommand.Workspace,
		}}, nil}
	})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Paths: []controllers.APIRequestPath{
			{
				Directory: "ignored",
				Workspace: "ignored-workspace",
			},
			{
				Directory: "kept",
				Workspace: "kept-workspace",
			},
		},
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	projectCommandRunner.VerifyWasCalledOnce().Plan(Any[command.ProjectContext]())
	_, capturedCmd := preWorkflowHooksRunner.VerifyWasCalledOnce().
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()

	Equals(t, "kept", capturedCmd.RepoRelDir)
	Equals(t, "kept-workspace", capturedCmd.Workspace)
}

func TestAPIController_Plan_AllIgnoredPathsNoOp(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t, func(config *apiControllerTestConfig) {
		config.allowUnlockByPull = false
	})
	commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, events.ErrIgnoredTargetedDir)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Paths: []controllers.APIRequestPath{
			{
				Directory: "ignored",
				Workspace: "default",
			},
		},
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	projectCommandRunner.VerifyWasCalled(Never()).Plan(Any[command.ProjectContext]())
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombined(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
}

func TestAPIController_Apply_AllIgnoredPathsNoOp(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t, func(config *apiControllerTestConfig) {
		config.allowUnlockByPull = false
	})
	commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, events.ErrIgnoredTargetedDir)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, events.ErrIgnoredTargetedDir)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Paths: []controllers.APIRequestPath{
			{
				Directory: "ignored",
				Workspace: "default",
			},
		},
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	projectCommandRunner.VerifyWasCalled(Never()).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombined(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
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
	ac.Locker.(*MockLocker).EXPECT().List().Return(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	Equals(t, http.StatusOK, w.Code)

	responseBody, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResult
	Ok(t, json.Unmarshal(responseBody, &result))
	Assert(t, !strings.Contains(string(responseBody), "\"success\":true"), "legacy locks response must not use success envelope: %s", responseBody)

	// Verify the lock details
	Assert(t, len(result.Locks) == 1, "expected 1 lock")
	lock := result.Locks[0]
	Equals(t, "lock-id", lock.Name)
	Equals(t, "terraform", lock.ProjectName)
	Equals(t, "owner/repo", lock.ProjectRepo)
	Equals(t, "/path", lock.ProjectRepoPath)
	Equals(t, "default", lock.Workspace)
	Equals(t, 123, lock.PullID)
	Equals(t, "url", lock.PullURL)
	Equals(t, "jdoe", lock.User)
}

func TestAPIController_PlanFetchesPullReqStatus(t *testing.T) {
	ac, _, _ := setup(t)
	fetcher := NewMockPullReqStatusFetcher()
	When(fetcher.FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())).
		ThenReturn(models.PullReqStatus{
			ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
			MergeableStatus: models.MergeableStatus{IsMergeable: true},
		}, nil)
	ac.PullReqStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	fetcher.VerifyWasCalled(Times(1)).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
}

func TestAPIController_PlanSkipsPullReqStatusWhenNoPR(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	fetcher := NewMockPullReqStatusFetcher()
	ac.PullReqStatusFetcher = fetcher

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

	fetcher.VerifyWasCalled(Never()).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
	planCtx, _ := projectCommandBuilder.VerifyWasCalledOnce().
		BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()
	Assert(t, planCtx.Pull.Num < 0, "expected no-PR API request to use an isolated synthetic pull number")
	Assert(t, planCtx.Pull.HardenedNonPRRefCheckout, "expected no-PR API request to use hardened checkout")
}

func TestAPIController_PlanContinuesWhenPullReqStatusFetchFails(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	fetcher := NewMockPullReqStatusFetcher()
	When(fetcher.FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())).
		ThenReturn(models.PullReqStatus{}, errors.New("api error"))
	ac.PullReqStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	fetcher.VerifyWasCalled(Times(1)).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
	projectCommandBuilder.VerifyWasCalled(Times(1)).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Times(1)).Plan(Any[command.ProjectContext]())
}

func TestAPISetup_RefreshesPullRequestStatusAfterResolvingPRHead(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	fetcher := NewMockPullReqStatusFetcher()
	mockCall := When(fetcher.FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]()))
	mockCall = mockCall.Then(func(args []Param) ReturnValues {
		pull := args[1].(models.PullRequest)
		Equals(t, "main", pull.HeadCommit)
		return ReturnValues{models.PullReqStatus{
			ApprovalStatus:  models.ApprovalStatus{IsApproved: false},
			MergeableStatus: models.MergeableStatus{IsMergeable: false},
		}, nil}
	})
	mockCall.Then(func(args []Param) ReturnValues {
		pull := args[1].(models.PullRequest)
		Equals(t, headCommit, pull.HeadCommit)
		return ReturnValues{models.PullReqStatus{
			ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
			MergeableStatus: models.MergeableStatus{IsMergeable: true},
		}, nil}
	})
	ac.PullReqStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	fetcher.VerifyWasCalled(Times(2)).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
	planCtx, _ := projectCommandBuilder.VerifyWasCalled(Times(1)).
		BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()
	Equals(t, headCommit, planCtx.Pull.HeadCommit)
	Assert(t, planCtx.PullRequestStatus.ApprovalStatus.IsApproved,
		"expected plan commands to use resolved-head approved status")
	Assert(t, planCtx.PullRequestStatus.MergeableStatus.IsMergeable,
		"expected plan commands to use resolved-head mergeable status")
}

func TestAPISetup_PRWithoutBaseBranchSeedsLivePRBase(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	ac.LivePullHeadFetcher = fakeControllerLivePullHeadFetcher{
		pull: models.PullRequest{HeadCommit: headCommit, BaseBranch: "main"},
	}
	var capturedCtx *command.Context
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "feature",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Assert(t, capturedCtx != nil, "expected plan command builder to be called")
	Equals(t, headCommit, capturedCtx.Pull.HeadCommit)
	Equals(t, "feature", capturedCtx.Pull.HeadBranch)
	Equals(t, "main", capturedCtx.Pull.BaseBranch)
}

func TestAPISetup_PRWithoutBaseBranchSeedsLiveBaseBeforeClone(t *testing.T) {
	ac, _, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	var clonePull models.PullRequest
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		Then(func(args []Param) ReturnValues {
			clonePull = args[2].(models.PullRequest)
			return ReturnValues{repoDir, nil}
		})
	ac.LivePullHeadFetcher = fakeControllerLivePullHeadFetcher{
		pull: models.PullRequest{HeadCommit: headCommit, BaseBranch: "main"},
	}

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "feature",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Equals(t, "main", clonePull.BaseBranch)
	Equals(t, "feature", clonePull.HeadBranch)
}

func TestAPISetup_MergeCheckoutCloneReceivesLivePRBase(t *testing.T) {
	ac, _, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	var clonePull models.PullRequest
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		Then(func(args []Param) ReturnValues {
			clonePull = args[2].(models.PullRequest)
			return ReturnValues{repoDir, nil}
		})
	ac.LivePullHeadFetcher = fakeControllerLivePullHeadFetcher{
		pull: models.PullRequest{HeadCommit: headCommit, BaseBranch: "main"},
	}

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "feature",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Equals(t, "main", clonePull.BaseBranch)
}

func TestAPIApply_PRRefDoesNotUseHeadRefAsBaseBranch(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	ac.LivePullHeadFetcher = fakeControllerLivePullHeadFetcher{
		pull: models.PullRequest{HeadCommit: headCommit, BaseBranch: "main"},
	}
	var applyCtx *command.Context
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			applyCtx = args[0].(*command.Context)
			return ReturnValues{[]command.ProjectContext{{CommandName: command.Apply}}, nil}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "feature",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	Assert(t, applyCtx != nil, "expected apply command builder to be called")
	Equals(t, "feature", applyCtx.Pull.HeadBranch)
	Equals(t, "main", applyCtx.Pull.BaseBranch)
	Assert(t, applyCtx.PullStatus != nil, "expected API apply to seed in-memory PullStatus")
	Equals(t, "main", applyCtx.PullStatus.Pull.BaseBranch)
}

func TestAPIApply_PRWithoutBaseBranchCurrentLiveBaseSucceeds(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	ac.LivePullHeadFetcher = fakeControllerLivePullHeadFetcher{
		pull: models.PullRequest{HeadCommit: headCommit, BaseBranch: "main"},
	}
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{CommandName: command.Apply}}, nil)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "feature",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	projectCommandRunner.VerifyWasCalledOnce().Apply(Any[command.ProjectContext]())
}

func TestAPIApply_NonPRMutableRefChangedDuringApplyDoesNotPublishSuccess(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	repoDir, _ := initAPIControllerGitRepoWithOrigin(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).
		Then(func([]Param) ReturnValues {
			advanceAPIControllerGitMain(t, repoDir)
			return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "success"}}
		})
	commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		BaseBranch: "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	ResponseContains(t, w, http.StatusInternalServerError, "changed")
	commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq(models.SuccessCommitStatus),
		Eq(command.Apply),
		Any[models.ProjectCounts](),
	)
}

func TestAPIApply_NonPRReleaseStableBranchChangedNoProjectsDoesNotPublishZeroZeroSuccess(t *testing.T) {
	for _, ref := range []string{"release", "stable"} {
		t.Run(ref, func(t *testing.T) {
			ac, projectCommandBuilder, projectCommandRunner := setup(t)
			repoDir, _ := initAPIControllerGitRepoWithOrigin(t)
			createAPIControllerGitBranch(t, repoDir, ref)
			workingDir := ac.WorkingDir.(*MockWorkingDir)
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
				ThenReturn(repoDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
				ThenReturn(repoDir, nil)
			When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
				Then(func([]Param) ReturnValues {
					advanceAPIControllerGitBranch(t, repoDir, ref)
					return ReturnValues{[]command.ProjectContext{}, nil}
				})
			commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

			body, _ := json.Marshal(controllers.APIRequest{
				Repository: "Repo",
				Ref:        ref,
				Type:       "Gitlab",
				Projects:   []string{"default"},
			})
			req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			ac.Apply(w, req)

			Equals(t, http.StatusInternalServerError, w.Code)
			ResponseContains(t, w, http.StatusInternalServerError, "changed")
			projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
			commitStatusUpdater.VerifyWasCalled(Never()).UpdateCombinedCount(
				Any[logging.SimpleLogging](),
				Any[models.Repo](),
				Any[models.PullRequest](),
				Eq(models.SuccessCommitStatus),
				Eq(command.Apply),
				Any[models.ProjectCounts](),
			)
		})
	}
}

func TestAPIApply_NonPRAmbiguousBranchUnchangedNoProjectsSucceeds(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	repoDir, _ := initAPIControllerGitRepoWithOrigin(t)
	createAPIControllerGitBranch(t, repoDir, "release")
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)
	commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "release",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
	commitStatusUpdater.VerifyWasCalled(Times(2)).UpdateCombinedCount(
		Any[logging.SimpleLogging](),
		Any[models.Repo](),
		Any[models.PullRequest](),
		Eq(models.SuccessCommitStatus),
		Eq(command.Apply),
		Eq(models.ProjectCounts{}),
	)
}

func TestAPIApply_NonPRImmutableSHANoProjectsSucceeds(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	repoDir, commit := initAPIControllerGitRepoWithOrigin(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        commit,
		BaseBranch: "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestAPIApply_NonPRMutableRefUnchangedAllowsApply(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	repoDir, _ := initAPIControllerGitRepoWithOrigin(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		BaseBranch: "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	projectCommandRunner.VerifyWasCalledOnce().Apply(Any[command.ProjectContext]())
}

func TestAPIApply_NonPRImmutableSHAAllowsApply(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	repoDir, commit := initAPIControllerGitRepoWithOrigin(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        commit,
		BaseBranch: "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)

	ResponseContains(t, w, http.StatusOK, "")
	projectCommandRunner.VerifyWasCalledOnce().Apply(Any[command.ProjectContext]())
}

func TestAPIApply_UsesResolvedHeadPullRequestStatusForPlanRequirements(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).
		ThenReturn(repoDir, nil)
	fetcher := NewMockPullReqStatusFetcher()
	mockCall := When(fetcher.FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]()))
	mockCall = mockCall.ThenReturn(models.PullReqStatus{}, nil)
	mockCall = mockCall.ThenReturn(models.PullReqStatus{
		ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
		MergeableStatus: models.MergeableStatus{IsMergeable: true},
	}, nil)
	mockCall.ThenReturn(models.PullReqStatus{
		ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
		MergeableStatus: models.MergeableStatus{IsMergeable: true},
	}, nil)
	ac.PullReqStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	fetcher.VerifyWasCalled(Times(3)).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
	planCtx, _ := projectCommandBuilder.VerifyWasCalled(Times(1)).
		BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()
	Equals(t, headCommit, planCtx.Pull.HeadCommit)
	Assert(t, planCtx.PullRequestStatus.ApprovalStatus.IsApproved,
		"expected pre-apply plan to use resolved-head approved status")
	Assert(t, planCtx.PullRequestStatus.MergeableStatus.IsMergeable,
		"expected pre-apply plan to use resolved-head mergeable status")
}

func TestAPIController_ApplyRefreshesPullReqStatusAfterPlan(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	fetcher := NewMockPullReqStatusFetcher()
	mockCall := When(fetcher.FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]()))
	mockCall = mockCall.ThenReturn(models.PullReqStatus{
		ApprovalStatus:  models.ApprovalStatus{IsApproved: false},
		MergeableStatus: models.MergeableStatus{IsMergeable: false},
	}, nil)
	mockCall.ThenReturn(models.PullReqStatus{
		ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
		MergeableStatus: models.MergeableStatus{IsMergeable: true},
	}, nil)
	ac.PullReqStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	fetcher.VerifyWasCalled(Times(2)).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
	applyCtx, _ := projectCommandBuilder.VerifyWasCalled(Times(1)).
		BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()
	Assert(t, applyCtx.PullRequestStatus.ApprovalStatus.IsApproved,
		"expected apply commands to use refreshed approved status")
	Assert(t, applyCtx.PullRequestStatus.MergeableStatus.IsMergeable,
		"expected apply commands to use refreshed mergeable status")
}

func TestAPIController_ApplyClearsPullReqStatusWhenPostPlanRefreshFails(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	fetcher := NewMockPullReqStatusFetcher()
	mockCall := When(fetcher.FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]()))
	mockCall = mockCall.ThenReturn(models.PullReqStatus{
		ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
		MergeableStatus: models.MergeableStatus{IsMergeable: true},
	}, nil)
	mockCall.ThenReturn(models.PullReqStatus{}, errors.New("api error"))
	ac.PullReqStatusFetcher = fetcher

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	fetcher.VerifyWasCalled(Times(2)).FetchPullStatus(Any[logging.SimpleLogging](), Any[models.PullRequest]())
	applyCtx, _ := projectCommandBuilder.VerifyWasCalled(Times(1)).
		BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetCapturedArguments()
	Assert(t, !applyCtx.PullRequestStatus.ApprovalStatus.IsApproved,
		"expected apply commands to clear stale approved status")
	Assert(t, !applyCtx.PullRequestStatus.MergeableStatus.IsMergeable,
		"expected apply commands to clear stale mergeable status")
}

func TestAPIController_ApplyUpdatesPullStatusBetweenSequentialProjects(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	pullStatus := &models.PullStatus{
		Projects: []models.ProjectStatus{
			{
				ProjectName: "network",
				RepoRelDir:  "network",
				Workspace:   "default",
				Status:      models.PlannedPlanStatus,
			},
			{
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   "default",
				Status:      models.PlannedPlanStatus,
			},
		},
	}

	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			ctx.PullStatus = pullStatus
			return ReturnValues{[]command.ProjectContext{
				{
					CommandName:   command.Apply,
					ProjectName:   "network",
					RepoRelDir:    "network",
					Workspace:     "default",
					PullStatus:    pullStatus,
					ApplyCmd:      "atlantis apply -p network",
					BaseRepo:      ctx.Pull.BaseRepo,
					HeadRepo:      ctx.HeadRepo,
					Pull:          ctx.Pull,
					PullReqStatus: ctx.PullRequestStatus,
				},
				{
					CommandName:   command.Apply,
					ProjectName:   "app",
					RepoRelDir:    "app",
					Workspace:     "default",
					PullStatus:    pullStatus,
					DependsOn:     []string{"network"},
					ApplyCmd:      "atlantis apply -p app",
					BaseRepo:      ctx.Pull.BaseRepo,
					HeadRepo:      ctx.HeadRepo,
					Pull:          ctx.Pull,
					PullReqStatus: ctx.PullRequestStatus,
				},
			}, nil}
		})

	applyCalls := 0
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).Then(func(args []Param) ReturnValues {
		applyCalls++
		projectCtx := args[0].(command.ProjectContext)
		if projectCtx.ProjectName == "app" {
			Equals(t, models.AppliedPlanStatus, pullStatus.Projects[0].Status)
		}
		return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "success"}}
	})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"all"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	Equals(t, 2, applyCalls)
	Equals(t, models.AppliedPlanStatus, pullStatus.Projects[0].Status)
	Equals(t, models.AppliedPlanStatus, pullStatus.Projects[1].Status)
}

func TestAPIController_ApplySeedsPolicyStatusFromAPIPlan(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)

	var capturedPullStatus *models.PullStatus
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			capturedPullStatus = ctx.PullStatus
			Assert(t, capturedPullStatus != nil, "expected pull status before building apply commands")
			Equals(t, 1, len(capturedPullStatus.Projects))
			Equals(t, models.PlannedNoChangesPlanStatus, capturedPullStatus.Projects[0].Status)
			Equals(t, 1, len(capturedPullStatus.Projects[0].PolicyStatus))
			Equals(t, "policy", capturedPullStatus.Projects[0].PolicyStatus[0].PolicySetName)
			Assert(t, capturedPullStatus.Projects[0].PolicyStatus[0].Passed, "expected policy status to pass")

			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Apply,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
				PullStatus:  capturedPullStatus,
			}}, nil}
		})

	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes. Infrastructure is up-to-date."},
		})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: "policy", Passed: true}},
		}})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	Assert(t, capturedPullStatus != nil, "expected apply command builder to be called")
	Equals(t, 1, len(capturedPullStatus.Projects))
	Equals(t, 1, len(capturedPullStatus.Projects[0].PolicyStatus))
	projectCommandRunner.VerifyWasCalled(Once()).PolicyCheck(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Once()).Apply(Any[command.ProjectContext]())
}

func TestAPIController_ApplyRefreshesStalePullStatusPullAfterAPIPlan(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	policyStatus := []models.PolicySetStatus{{
		PolicySetName: "policy",
		Passed:        false,
		Approvals: []models.PolicySetApproval{{
			Approver: "owner",
			Hashes:   []string{"hash"},
		}},
		Hashes: []string{"hash"},
	}}
	fetcher := &recordingPullStatusFetcher{statuses: []*models.PullStatus{{
		Pull: models.PullRequest{
			Num:        42,
			HeadBranch: "old-head",
			HeadCommit: "old-head",
			BaseBranch: "old-base",
		},
		Projects: []models.ProjectStatus{{
			ProjectName:  "app",
			RepoRelDir:   "app",
			Workspace:    events.DefaultWorkspace,
			PolicyStatus: policyStatus,
		}},
	}}}
	ac.PullStatusFetcher = fetcher

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "app",
			RepoRelDir:  "app",
			Workspace:   events.DefaultWorkspace,
		}}, nil)

	var capturedCtx *command.Context
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			capturedCtx = ctx
			Assert(t, ctx.PullStatus != nil, "expected pull status before building apply commands")
			Equals(t, ctx.Pull, ctx.PullStatus.Pull)
			Equals(t, "current-head", ctx.PullStatus.Pull.HeadCommit)
			Equals(t, "main", ctx.PullStatus.Pull.BaseBranch)
			Equals(t, policyStatus, ctx.PullStatus.Projects[0].PolicyStatus)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Apply,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
				PullStatus:  ctx.PullStatus,
			}}, nil}
		})

	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "current-head",
		BaseBranch: "main",
		Type:       "Gitlab",
		PR:         42,
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Apply(w, req)
	ResponseContains(t, w, http.StatusOK, "")

	Assert(t, capturedCtx != nil, "expected apply command builder to be called")
	Equals(t, 1, len(fetcher.calls))
	Equals(t, 42, fetcher.calls[0].Num)
}

func TestAPIController_ListLocksEmpty(t *testing.T) {
	ac, _, _ := setup(t)

	mockLocks := map[string]models.ProjectLock{}
	ac.Locker.(*MockLocker).EXPECT().List().Return(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	Equals(t, http.StatusOK, w.Code)

	responseBody, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResult
	Ok(t, json.Unmarshal(responseBody, &result))
	Assert(t, !strings.Contains(string(responseBody), "\"success\":true"), "legacy locks response must not use success envelope: %s", responseBody)

	// Verify empty result
	Equals(t, 0, len(result.Locks))
}

func TestAPIController_ListLocksErrorReturnsLegacyShape(t *testing.T) {
	ac, _, _ := setup(t)
	ac.Locker.(*MockLocker).EXPECT().List().Return(nil, errors.New("lock list failed"))

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	responseBody, _ := io.ReadAll(w.Result().Body)
	var legacyErr map[string]string
	Ok(t, json.Unmarshal(responseBody, &legacyErr))
	Equals(t, "lock list failed", legacyErr["error"])
	Assert(t, !strings.Contains(string(responseBody), "\"success\""), "legacy locks error must not use success envelope: %s", responseBody)
}

type apiControllerTestConfig struct {
	allowUnlockByPull bool
}

func setup(t *testing.T, options ...func(*apiControllerTestConfig)) (*controllers.APIController, *MockProjectCommandBuilder, *MockProjectCommandRunner) {
	RegisterMockTestingT(t)
	config := &apiControllerTestConfig{
		allowUnlockByPull: true,
	}
	for _, option := range options {
		option(config)
	}

	gmockCtrl := gomock.NewController(t)
	locker := NewMockLocker(gmockCtrl)
	if config.allowUnlockByPull {
		// Allow incidental calls to UnlockByPull (called internally during plan/apply operations)
		locker.EXPECT().UnlockByPull(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	}
	logger := logging.NewNoopLogger(t)
	parser := NewMockEventParsing()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	scope := metricstest.NewLoggingScope(t, logger, "null")
	vcsClient := NewMockClient()
	workingDir := NewMockWorkingDir()
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(nil)
	Ok(t, err)
	When(vcsClient.GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())).
		Then(func(args []Param) ReturnValues {
			repository := args[2].(string)
			return ReturnValues{fmt.Sprintf("https://example.com/%s.git", repository), nil}
		})
	When(parser.ParseAPIPlanRequest(Any[models.VCSHostType](), Any[string](), Any[string]())).
		Then(func(args []Param) ReturnValues {
			vcsType := args[0].(models.VCSHostType)
			repository := args[1].(string)
			return ReturnValues{models.Repo{
				FullName: repository,
				VCSHost: models.VCSHost{
					Hostname: testVCSHostname(vcsType),
					Type:     vcsType,
				},
			}, nil}
		})

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
		APISecret:                       []byte(atlantisToken),
		Locker:                          locker,
		Logger:                          logger,
		Scope:                           scope,
		Parser:                          parser,
		ProjectCommandBuilder:           projectCommandBuilder,
		ProjectPlanCommandRunner:        projectCommandRunner,
		ProjectPolicyCheckCommandRunner: projectCommandRunner,
		ProjectApplyCommandRunner:       projectCommandRunner,
		PreWorkflowHooksCommandRunner:   preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner:  postWorkflowHooksCommandRunner,
		VCSClient:                       vcsClient,
		RepoAllowlistChecker:            repoAllowlistChecker,
		WorkingDir:                      workingDir,
		WorkingDirLocker:                workingDirLocker,
		CommitStatusUpdater:             commitStatusUpdater,
	}
	return ac, projectCommandBuilder, projectCommandRunner
}

func testVCSHostname(vcsType models.VCSHostType) string {
	switch vcsType {
	case models.Github:
		return "github.com"
	case models.Gitlab:
		return "gitlab.com"
	case models.BitbucketCloud:
		return "bitbucket.org"
	case models.BitbucketServer:
		return "bitbucket.example.com"
	case models.AzureDevops:
		return "dev.azure.com"
	case models.Gitea:
		return "gitea.example.com"
	default:
		return "example.com"
	}
}

func initAPIControllerGitRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir, err := os.MkdirTemp("", "api-controller-git-*")
	Ok(t, err)
	t.Cleanup(func() {
		for range 5 {
			if err := os.RemoveAll(repoDir); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
	runAPIControllerGit(t, repoDir, "init", "--initial-branch=main")
	runAPIControllerGit(t, repoDir, "config", "--local", "gc.auto", "0")
	runAPIControllerGit(t, repoDir, "config", "--local", "maintenance.auto", "false")
	runAPIControllerGit(t, repoDir, "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runAPIControllerGit(t, repoDir, "config", "--local", "user.name", "atlantisbot")
	runAPIControllerGit(t, repoDir, "config", "--local", "commit.gpgsign", "false")
	runAPIControllerGit(t, repoDir, "commit", "--allow-empty", "-m", "initial commit")
	return repoDir, strings.TrimSpace(runAPIControllerGit(t, repoDir, "rev-parse", "HEAD"))
}

func initAPIControllerGitRepoWithOrigin(t *testing.T) (string, string) {
	t.Helper()
	repoDir, commit := initAPIControllerGitRepo(t)
	originDir := filepath.Join(t.TempDir(), "origin.git")
	runAPIControllerGit(t, repoDir, "init", "--bare", originDir)
	runAPIControllerGit(t, originDir, "config", "gc.auto", "0")
	runAPIControllerGit(t, originDir, "config", "maintenance.auto", "false")
	runAPIControllerGit(t, repoDir, "remote", "add", "origin", "file://"+originDir)
	runAPIControllerGit(t, repoDir, "push", "-q", "-u", "origin", "main")
	return repoDir, commit
}

func createAPIControllerGitBranch(t *testing.T, repoDir, branch string) string {
	t.Helper()
	runAPIControllerGit(t, repoDir, "checkout", "-q", "-b", branch)
	runAPIControllerGit(t, repoDir, "push", "-q", "-u", "origin", branch)
	return strings.TrimSpace(runAPIControllerGit(t, repoDir, "rev-parse", "HEAD"))
}

func advanceAPIControllerGitBranch(t *testing.T, repoDir, branch string) string {
	t.Helper()
	runAPIControllerGit(t, repoDir, "checkout", "-q", branch)
	changedTF := []byte("resource \"null_resource\" \"" + strings.ReplaceAll(branch, "-", "_") + "_changed\" {}\n")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), changedTF, 0600))
	runAPIControllerGit(t, repoDir, "add", "main.tf")
	runAPIControllerGit(t, repoDir, "commit", "-q", "-m", "advance "+branch)
	runAPIControllerGit(t, repoDir, "push", "-q", "origin", "HEAD:"+branch)
	return strings.TrimSpace(runAPIControllerGit(t, repoDir, "rev-parse", "HEAD"))
}

func advanceAPIControllerGitMain(t *testing.T, repoDir string) string {
	t.Helper()
	runAPIControllerGit(t, repoDir, "checkout", "-q", "main")
	mainTF := []byte("resource \"null_resource\" \"changed\" {}\n")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), mainTF, 0600))
	runAPIControllerGit(t, repoDir, "add", "main.tf")
	runAPIControllerGit(t, repoDir, "commit", "-q", "-m", "advance main")
	runAPIControllerGit(t, repoDir, "push", "-q", "origin", "HEAD:main")
	return strings.TrimSpace(runAPIControllerGit(t, repoDir, "rev-parse", "HEAD"))
}

func runAPIControllerGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...) // nolint: gosec
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running git %s: %s: %v", strings.Join(args, " "), output, err)
	}
	return string(output)
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

func TestAPIResponseEnvelopeIncludesNullInactiveFields(t *testing.T) {
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
	responseBody := w.Body.String()
	Assert(t, strings.Contains(responseBody, "\"error\":null"), "success envelope must include error:null: %s", responseBody)

	errReq, _ := http.NewRequest("GET", "/api/drift/status", nil)
	errReq.Header.Set(atlantisTokenHeader, atlantisToken)
	errW := httptest.NewRecorder()
	ac.DriftStatus(errW, errReq)

	Equals(t, http.StatusBadRequest, errW.Code)
	errorBody := errW.Body.String()
	Assert(t, strings.Contains(errorBody, "\"data\":null"), "error envelope must include data:null: %s", errorBody)
}

func TestAPIController_DriftStatus(t *testing.T) {
	ac, _, _ := setup(t)
	driftStorage := driftmocks.NewMockStorage()
	checkTime := time.Date(2025, 1, 21, 10, 30, 0, 0, time.UTC)
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
	When(driftStorage.Get(Eq("github.com/owner/repo"), Any[drift.GetOptions]())).ThenReturn(mockDrifts, nil)

	ac.DriftStorage = driftStorage

	req, _ := http.NewRequest("GET", "?repository=owner/repo&type=Github", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusOK, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftStatusAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "owner/repo", result.Repository)
	Equals(t, 1, result.Summary.TotalProjects)
	Equals(t, 1, result.Summary.ProjectsWithDrift)
	Equals(t, checkTime, result.CheckedAt)
}

func TestAPIController_DriftStatus_Unauthorized(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
	driftStorage := driftmocks.NewMockStorage()

	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo", nil)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusUnauthorized, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeUnauthorized, apiErr.Code)
	driftStorage.VerifyWasCalled(Never()).Get(Any[string](), Any[drift.GetOptions]())
}

func TestAPIController_DriftStatus_NoStorage(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

	// Controller without drift storage
	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: nil,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
}

func TestAPIController_DriftStatus_MissingRepository(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
	driftStorage := driftmocks.NewMockStorage()

	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
}

func TestAPIController_DriftStatus_MissingType(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
	driftStorage := driftmocks.NewMockStorage()

	ac := controllers.APIController{
		APISecret:    []byte(atlantisToken),
		Logger:       logger,
		Locker:       locker,
		DriftStorage: driftStorage,
	}

	req, _ := http.NewRequest("GET", "?repository=owner/repo", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
	driftStorage.VerifyWasCalled(Never()).Get(Any[string](), Any[drift.GetOptions]())
}

func TestAPIController_DriftStatus_UnsupportedVCSType(t *testing.T) {
	ac, _, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()
	vcsClient := ac.VCSClient.(*MockClient)

	req, _ := http.NewRequest("GET", "?repository=owner/repo&type=BitbucketCloud", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
	Assert(t, strings.Contains(apiErr.Message, "unsupported VCS type"), "expected unsupported type error, got %q", apiErr.Message)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
}

func TestAPIController_DriftStatus_InvalidRepository(t *testing.T) {
	ac, _, _ := setup(t)
	driftStorage := driftmocks.NewMockStorage()
	ac.DriftStorage = driftStorage
	vcsClient := ac.VCSClient.(*MockClient)

	req, _ := http.NewRequest("GET", "?repository=owner&type=Github", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	driftStorage.VerifyWasCalled(Never()).Get(Any[string](), Any[drift.GetOptions]())
}

func TestAPIController_DriftStatus_WithFilters(t *testing.T) {
	ac, _, _ := setup(t)
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
	When(driftStorage.Get(Eq("github.com/owner/repo"), Any[drift.GetOptions]())).ThenReturn(mockDrifts, nil)

	ac.DriftStorage = driftStorage

	req, _ := http.NewRequest("GET", "?repository=owner/repo&type=Github&project=project1&path=./modules/vpc/&workspace=staging&ref=refs/heads/main&base_branch=refs/heads/main", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DriftStatus(w, req)

	Equals(t, http.StatusOK, w.Code)

	// Verify the storage was called with correct options
	_, opts := driftStorage.VerifyWasCalledOnce().Get(Eq("github.com/owner/repo"), Any[drift.GetOptions]()).GetCapturedArguments()
	Equals(t, "project1", opts.ProjectName)
	Equals(t, "modules/vpc", opts.Path)
	Equals(t, "staging", opts.Workspace)
	Equals(t, "main", opts.Ref)
	Equals(t, "main", opts.BaseBranch)
}

func TestAPIController_DriftStatus_InvalidPathFilter(t *testing.T) {
	ac, _, _ := setup(t)
	driftStorage := driftmocks.NewMockStorage()
	ac.DriftStorage = driftStorage

	for _, path := range []string{"../env", "envs/*"} {
		req, _ := http.NewRequest("GET", "?repository=owner/repo&type=Github&path="+url.QueryEscape(path), nil)
		req.Header.Set(atlantisTokenHeader, atlantisToken)
		w := httptest.NewRecorder()
		ac.DriftStatus(w, req)

		Equals(t, http.StatusBadRequest, w.Code)
	}
	driftStorage.VerifyWasCalled(Never()).Get(Any[string](), Any[drift.GetOptions]())
}

func TestAPIController_DriftStatus_Empty(t *testing.T) {
	ac, _, _ := setup(t)
	driftStorage := driftmocks.NewMockStorage()

	When(driftStorage.Get(Eq("github.com/owner/repo"), Any[drift.GetOptions]())).ThenReturn([]models.ProjectDrift{}, nil)

	ac.DriftStorage = driftStorage

	req, _ := http.NewRequest("GET", "?repository=owner/repo&type=Github", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
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

func TestAPIController_DriftReadEndpointsRejectNonAllowlistedRepos(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*controllers.APIController, *httptest.ResponseRecorder, *http.Request)
	}{
		{
			name: "status",
			call: func(ac *controllers.APIController, w *httptest.ResponseRecorder, req *http.Request) {
				ac.DriftStatus(w, req)
			},
		},
		{
			name: "remediation history",
			call: func(ac *controllers.APIController, w *httptest.ResponseRecorder, req *http.Request) {
				ac.ListRemediationResults(w, req)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ac, _, _ := setup(t)
			repoAllowlistChecker, err := events.NewRepoAllowlistChecker("github.com/nevermatch")
			Ok(t, err)
			ac.RepoAllowlistChecker = repoAllowlistChecker
			driftStorage := driftmocks.NewMockStorage()
			remediationService := driftmocks.NewMockRemediationService()
			ac.DriftStorage = driftStorage
			ac.RemediationService = remediationService

			req, _ := http.NewRequest("GET", "?repository=owner/repo&type=Github", nil)
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			tc.call(ac, w, req)

			Equals(t, http.StatusForbidden, w.Code)
			driftStorage.VerifyWasCalled(Never()).Get(Any[string](), Any[drift.GetOptions]())
			remediationService.VerifyWasCalled(Never()).ListResults(Any[string](), Any[int]())
		})
	}
}

func TestAPIController_RemediateReadEndpointsRejectInvalidRepositoryBeforeRead(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*controllers.APIController, *httptest.ResponseRecorder, *http.Request)
	}{
		{
			name: "history",
			call: func(ac *controllers.APIController, w *httptest.ResponseRecorder, req *http.Request) {
				ac.ListRemediationResults(w, req)
			},
		},
		{
			name: "result",
			call: func(ac *controllers.APIController, w *httptest.ResponseRecorder, req *http.Request) {
				ac.GetRemediationResult(w, req)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ac, _, _ := setup(t)
			remediationService := driftmocks.NewMockRemediationService()
			ac.RemediationService = remediationService
			vcsClient := ac.VCSClient.(*MockClient)

			req, _ := http.NewRequest("GET", "?id=result-id&repository=owner&type=Github", nil)
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			tc.call(ac, w, req)

			Equals(t, http.StatusBadRequest, w.Code)
			vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
			remediationService.VerifyWasCalled(Never()).GetResult(Any[string]())
			remediationService.VerifyWasCalled(Never()).ListResults(Any[string](), Any[int]())
		})
	}
}

func TestAPIController_DetectDriftRejectsNonAllowlistedRepoBeforeMutation(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("github.com/allowed/repo")
	Ok(t, err)
	ac.RepoAllowlistChecker = repoAllowlistChecker
	driftStorage := driftmocks.NewMockStorage()
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusForbidden, w.Code)
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	driftStorage.VerifyWasCalled(Never()).Store(Any[string](), Any[models.ProjectDrift]())
}

func TestAPIController_DetectDrift_InvalidRepositoryReturnsBadRequest(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()
	vcsClient := ac.VCSClient.(*MockClient)

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "owner",
		Ref:        "main",
		Type:       "Github",
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
}

func TestAPIController_RemediateRejectsNonAllowlistedRepoBeforeMutation(t *testing.T) {
	ac, _, projectCommandRunner := setup(t)
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("github.com/allowed/repo")
	Ok(t, err)
	ac.RepoAllowlistChecker = repoAllowlistChecker
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/remediate", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusForbidden, w.Code)
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestAPIController_Remediate_InvalidRepositoryReturnsBadRequest(t *testing.T) {
	ac, _, _ := setup(t)
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService
	vcsClient := ac.VCSClient.(*MockClient)

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
	})
	req, _ := http.NewRequest("POST", "/api/drift/remediate", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
}

func TestAPIController_Remediate(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
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
	var capturedRequest models.RemediationRequest
	When(remediationService.Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())).
		Then(func(args []Param) ReturnValues {
			capturedRequest = args[0].(models.RemediationRequest)
			return ReturnValues{mockResult, nil}
		})

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
		Ref:        "refs/heads/main",
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
	Equals(t, "owner/repo", capturedRequest.Repository)
	Equals(t, "main", capturedRequest.Ref)
	Equals(t, "refs/heads/main", capturedRequest.ExecutionRef)
	Equals(t, "main", capturedRequest.BaseBranch)
	Equals(t, "github.com/owner/repo", capturedRequest.StorageRepository)
}

func TestAPIController_Remediate_ProjectFailuresReturnNon2xx(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
	parser := NewMockEventParsing()
	vcsClient := NewMockClient()
	repoAllowlistChecker, _ := events.NewRepoAllowlistChecker("*")

	remediationService := driftmocks.NewMockRemediationService()
	mockResult := &models.RemediationResult{
		ID:            "test-id",
		Repository:    "owner/repo",
		Ref:           "main",
		Action:        models.RemediationAutoApply,
		Status:        models.RemediationStatusFailed,
		TotalProjects: 1,
		FailureCount:  1,
		Projects: []models.ProjectRemediationResult{
			{
				ProjectName: "project1",
				Status:      models.RemediationStatusFailed,
				Error:       "plan had errors",
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
		APISecret:              []byte(atlantisToken),
		Logger:                 logger,
		Locker:                 locker,
		Parser:                 parser,
		VCSClient:              vcsClient,
		RepoAllowlistChecker:   repoAllowlistChecker,
		RemediationService:     remediationService,
		EnableDriftRemediation: true,
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusConflict, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeConflict, apiErr.Code)
	Assert(t, apiErr.Details != nil, "expected failed remediation details")
}

func TestAPIController_Remediate_ServiceFailureWithoutProjectsReturnsInternalError(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
	parser := NewMockEventParsing()
	vcsClient := NewMockClient()
	repoAllowlistChecker, _ := events.NewRepoAllowlistChecker("*")

	remediationService := driftmocks.NewMockRemediationService()
	mockResult := &models.RemediationResult{
		ID:         "test-id",
		Repository: "owner/repo",
		Ref:        "main",
		Action:     models.RemediationPlanOnly,
		Status:     models.RemediationStatusFailed,
		Error:      "failed to load cached drift: storage unavailable",
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

	Equals(t, http.StatusInternalServerError, w.Code)
	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeInternal, apiErr.Code)
}

func TestAPIController_RemediateProjectSelectorsUseExactMatching(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	ac.RemediationService = drift.NewInMemoryRemediationService(drift.NewInMemoryStorage())
	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

	var capturedCtx *command.Context
	var capturedCmd *events.CommentCommand
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			capturedCmd = args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: "app.prod",
				RepoRelDir:  "env",
				Workspace:   events.DefaultWorkspace,
			}}, nil}
		})
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes. Your infrastructure matches the configuration."},
	})

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
		Projects:   []string{"app.prod"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/remediate", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusOK, w.Code)
	Assert(t, capturedCtx != nil, "expected project command builder context")
	Assert(t, capturedCtx.ExactProjectNameMatching, "remediation API project selectors must use exact matching")
	Assert(t, capturedCmd != nil, "expected remediation project selector")
	Equals(t, "app.prod", capturedCmd.ProjectName)
}

func TestAPIController_Remediate_PartialFailuresReturnMultiStatus(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
	parser := NewMockEventParsing()
	vcsClient := NewMockClient()
	repoAllowlistChecker, _ := events.NewRepoAllowlistChecker("*")

	remediationService := driftmocks.NewMockRemediationService()
	mockResult := &models.RemediationResult{
		ID:            "test-id",
		Repository:    "owner/repo",
		Ref:           "main",
		Action:        models.RemediationAutoApply,
		Status:        models.RemediationStatusPartial,
		TotalProjects: 2,
		SuccessCount:  1,
		FailureCount:  1,
		Projects: []models.ProjectRemediationResult{
			{ProjectName: "project1", Status: models.RemediationStatusSuccess},
			{ProjectName: "project2", Status: models.RemediationStatusFailed, Error: "apply had errors"},
		},
	}
	When(remediationService.Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())).ThenReturn(mockResult, nil)

	When(vcsClient.GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Eq("owner/repo"))).ThenReturn("https://github.com/owner/repo.git", nil)
	When(parser.ParseAPIPlanRequest(Any[models.VCSHostType](), Eq("owner/repo"), Any[string]())).ThenReturn(models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Hostname: "github.com"},
	}, nil)

	ac := controllers.APIController{
		APISecret:              []byte(atlantisToken),
		Logger:                 logger,
		Locker:                 locker,
		Parser:                 parser,
		VCSClient:              vcsClient,
		RepoAllowlistChecker:   repoAllowlistChecker,
		RemediationService:     remediationService,
		EnableDriftRemediation: true,
	}

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusMultiStatus, w.Code)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.RemediationResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, "partial", result.Status)
	Equals(t, 1, result.Summary.SuccessCount)
	Equals(t, 1, result.Summary.FailureCount)
}

func TestAPIController_Remediate_AutoApplyRequiresOptIn(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
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
		Action:     models.RemediationAutoApply,
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusServiceUnavailable, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Equals(t, controllers.ErrCodeServiceUnavailable, apiErr.Code)
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
}

func TestAPIController_Remediate_InvalidBaseBranchReturnsBadRequest(t *testing.T) {
	ac, _, _ := setup(t)
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		BaseBranch: "refs/pull/1/head",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
	})
	req, _ := http.NewRequest("POST", "/api/drift/remediate", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "base_branch"), "expected base_branch validation error, got %#v", apiErr.Details)
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
}

func TestAPIController_Remediate_InvalidRefReturnsBadRequest(t *testing.T) {
	ac, _, _ := setup(t)
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService
	vcsClient := ac.VCSClient.(*MockClient)

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "refs/pull/123/head",
		BaseBranch: "main",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
	})
	req, _ := http.NewRequest("POST", "/api/drift/remediate", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "ref"), "expected ref validation error, got %#v", apiErr.Details)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
}

func TestAPIController_Remediate_InvalidWorkspaceReturnsBadRequest(t *testing.T) {
	ac, _, _ := setup(t)
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService
	vcsClient := ac.VCSClient.(*MockClient)

	body, _ := json.Marshal(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
		Paths:      []models.DriftDetectionPath{{Directory: "env", Workspace: "../../tmp/plan"}},
	})
	req, _ := http.NewRequest("POST", "/api/drift/remediate", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "paths"), "expected paths validation error, got %#v", apiErr.Details)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
}

func TestAPIController_Remediate_NoService(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

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
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
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
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
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

func TestAPIController_Remediate_EmptyProjectName(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

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
		Projects:   []string{""},
	})

	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Remediate(w, req)

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
	remediationService.VerifyWasCalled(Never()).Remediate(Any[models.RemediationRequest](), Any[drift.RemediationExecutor]())
}

func TestAPIController_Remediate_APIDisabled(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

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
	ac, _, _ := setup(t)

	remediationService := driftmocks.NewMockRemediationService()
	mockResult := &models.RemediationResult{
		ID:                "test-id-123",
		Repository:        "owner/repo",
		StorageRepository: "github.com/owner/repo",
		Ref:               "main",
		Action:            models.RemediationPlanOnly,
		Status:            models.RemediationStatusSuccess,
		TotalProjects:     1,
		SuccessCount:      1,
		Projects: []models.ProjectRemediationResult{
			{
				ProjectName: "project1",
				Status:      models.RemediationStatusSuccess,
				PlanOutput:  "Plan: 0 to add, 0 to change, 0 to destroy",
			},
		},
	}
	When(remediationService.GetResult(Eq("test-id-123"))).ThenReturn(mockResult, nil)
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id-123?repository=owner/repo&type=Github", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id-123"})
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
	ac, _, _ := setup(t)

	remediationService := driftmocks.NewMockRemediationService()
	When(remediationService.GetResult(Eq("nonexistent-id"))).ThenReturn(nil, fmt.Errorf("remediation result not found: nonexistent-id"))
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate/nonexistent-id?repository=owner/repo&type=Github", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent-id"})
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusNotFound, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeNotFound, apiErr.Code)
}

func TestAPIController_GetRemediationResult_RequiresRepositoryScope(t *testing.T) {
	ac, _, _ := setup(t)
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id?type=Github", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	remediationService.VerifyWasCalled(Never()).GetResult(Any[string]())
}

func TestAPIController_GetRemediationResult_RejectsNonAllowlistedRepoBeforeLookup(t *testing.T) {
	ac, _, _ := setup(t)
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("github.com/allowed/repo")
	Ok(t, err)
	ac.RepoAllowlistChecker = repoAllowlistChecker
	remediationService := driftmocks.NewMockRemediationService()
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id?repository=owner/repo&type=Github", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusForbidden, w.Code)
	remediationService.VerifyWasCalled(Never()).GetResult(Any[string]())
}

func TestAPIController_GetRemediationResult_RejectsMismatchedRepo(t *testing.T) {
	ac, _, _ := setup(t)
	remediationService := driftmocks.NewMockRemediationService()
	When(remediationService.GetResult(Eq("test-id"))).ThenReturn(&models.RemediationResult{
		ID:                "test-id",
		Repository:        "other/repo",
		StorageRepository: "github.com/other/repo",
		Ref:               "main",
		Status:            models.RemediationStatusSuccess,
	}, nil)
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id?repository=owner/repo&type=Github", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.GetRemediationResult(w, req)

	Equals(t, http.StatusNotFound, w.Code)
}

func TestAPIController_GetRemediationResult_MissingID(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

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
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

	ac := controllers.APIController{
		APISecret:          []byte(atlantisToken),
		Logger:             logger,
		Locker:             locker,
		RemediationService: nil, // No service
	}

	req, _ := http.NewRequest("GET", "/api/drift/remediate/test-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
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
	ac, _, _ := setup(t)

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
	When(remediationService.ListResults(Eq("github.com/owner/repo"), Eq(10))).ThenReturn(mockResults, nil)
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo&type=Github", nil)
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
	ac, _, _ := setup(t)

	remediationService := driftmocks.NewMockRemediationService()
	mockResults := []*models.RemediationResult{
		{
			ID:         "result-1",
			Repository: "owner/repo",
			Status:     models.RemediationStatusSuccess,
		},
	}
	When(remediationService.ListResults(Eq("github.com/owner/repo"), Eq(5))).ThenReturn(mockResults, nil)
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo&type=Github&limit=5", nil)
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.ListRemediationResults(w, req)

	Equals(t, http.StatusOK, w.Code)
}

func TestAPIController_ListRemediationResults_InvalidLimit(t *testing.T) {
	cases := []string{"5abc", "5+6", "abc", "", "0", "-1"}
	for _, limit := range cases {
		t.Run(limit, func(t *testing.T) {
			ac, _, _ := setup(t)
			remediationService := driftmocks.NewMockRemediationService()
			ac.RemediationService = remediationService

			req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo&type=Github&limit="+limit, nil)
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			ac.ListRemediationResults(w, req)

			Equals(t, http.StatusBadRequest, w.Code)
			response, _ := io.ReadAll(w.Result().Body)
			apiErr := parseAPIError(t, response)
			Equals(t, controllers.ErrCodeValidation, apiErr.Code)
			remediationService.VerifyWasCalled(Never()).ListResults(Any[string](), Any[int]())
		})
	}
}

func TestAPIController_ListRemediationResults_MissingType(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

	remediationService := driftmocks.NewMockRemediationService()

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

	Equals(t, http.StatusBadRequest, w.Code)

	response, _ := io.ReadAll(w.Result().Body)
	apiErr := parseAPIError(t, response)
	Equals(t, controllers.ErrCodeValidation, apiErr.Code)
	remediationService.VerifyWasCalled(Never()).ListResults(Any[string](), Any[int]())
}

func TestAPIController_ListRemediationResults_MissingRepository(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

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
	ac, _, _ := setup(t)

	remediationService := driftmocks.NewMockRemediationService()
	When(remediationService.ListResults(Eq("github.com/owner/repo"), Eq(10))).ThenReturn([]*models.RemediationResult{}, nil)
	ac.RemediationService = remediationService

	req, _ := http.NewRequest("GET", "/api/drift/remediate?repository=owner/repo&type=Github", nil)
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
	ac, projectCommandBuilder, _ := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			Assert(t, !ctx.SkipPRRequirements, "drift detection should fail closed for PR-state plan requirements")
			Assert(t, ctx.FailOnTeamAllowlistDenied, "drift detection should fail closed on team allowlist denial")
			return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
		})

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

func TestAPIController_DetectDrift_TeamAllowlistDeniedReturnsForbidden(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	driftStorage := driftmocks.NewMockStorage()
	ac.DriftStorage = driftStorage
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn(nil, fmt.Errorf("failed to build command: %w", events.ErrTeamAllowlistDenied))

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

	Equals(t, http.StatusForbidden, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Equals(t, controllers.ErrCodeForbidden, apiErr.Code)
	Assert(t, strings.Contains(apiErr.Message, "team allowlist denied"), "expected team allowlist error, got %q", apiErr.Message)
}

func TestAPIController_DetectDriftSuppressesNormalCommitStatus(t *testing.T) {
	ac, _, _ := setup(t)
	commitStatusUpdater := ac.CommitStatusUpdater.(*MockCommitStatusUpdater)

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
	commitStatusUpdater.VerifyWasCalled(Never()).
		UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())
	commitStatusUpdater.VerifyWasCalled(Never()).
		UpdateCombinedCount(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name](), Any[models.ProjectCounts]())
}

func TestAPIController_DetectDriftSendsWebhookWhenDriftDetected(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "app",
			RepoRelDir:  "app",
			Workspace:   events.DefaultWorkspace,
		}}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
	})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage
	sender := &recordingDriftSender{}
	ac.DriftWebhookSender = &webhooks.DriftWebhookSender{Webhooks: []webhooks.DriftSender{sender}}

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	Equals(t, 1, sender.calls)
	Equals(t, 1, len(sender.results))
	Equals(t, "Repo", sender.results[0].Repository)
	Equals(t, "main", sender.results[0].Ref)
	Equals(t, 1, sender.results[0].ProjectsWithDrift)
	Equals(t, 1, len(sender.results[0].Projects))
	Equals(t, "app", sender.results[0].Projects[0].ProjectName)
	Equals(t, "app", sender.results[0].Projects[0].Path)
	Equals(t, events.DefaultWorkspace, sender.results[0].Projects[0].Workspace)
}

func TestAPIController_DetectDriftSendsWebhookWhenNoDrift(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "app",
			RepoRelDir:  "app",
			Workspace:   events.DefaultWorkspace,
		}}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes. Your infrastructure matches the configuration."},
	})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage
	sender := &recordingDriftSender{}
	ac.DriftWebhookSender = &webhooks.DriftWebhookSender{Webhooks: []webhooks.DriftSender{sender}}

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	Equals(t, 1, sender.calls)
	Equals(t, 1, len(sender.results))
	Equals(t, 0, sender.results[0].ProjectsWithDrift)
	Equals(t, 1, len(sender.results[0].Projects))
	Assert(t, !sender.results[0].Projects[0].HasDrift, "expected no-drift webhook project")
}

func TestAPIController_DetectDriftEmptySelectorsDiscoverAllProjects(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	var capturedCtx *command.Context
	var capturedCmd *events.CommentCommand
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			capturedCmd = args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			}}, nil}
		})
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
	})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

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
	Assert(t, capturedCmd != nil, "expected project command builder to be called")
	Assert(t, capturedCmd.DiscoverAllProjects, "expected empty drift selectors to request full project discovery")
	Assert(t, capturedCtx != nil, "expected command context")
	Assert(t, capturedCtx.ExactProjectNameMatching, "expected drift API project selectors to use exact matching")
	Assert(t, capturedCtx.SortByExecutionOrder, "expected drift API contexts to opt into execution-order sorting")
}

func TestAPIController_DetectDriftNormalizesBranchRefsForSelectionAndStorage(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	var capturedCtx *command.Context
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			}}, nil}
		})
	driftStorage := driftmocks.NewMockStorage()
	var stored models.ProjectDrift
	When(driftStorage.Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]())).
		Then(func(args []Param) ReturnValues {
			stored = args[1].(models.ProjectDrift)
			return ReturnValues{nil}
		})
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "refs/heads/main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	Assert(t, capturedCtx != nil, "expected project builder to be called")
	Equals(t, "main", capturedCtx.Pull.BaseBranch)
	Equals(t, "main", capturedCtx.Pull.HeadBranch)
	Equals(t, "main", stored.Ref)
	Equals(t, "main", stored.BaseBranch)
}

func TestAPIController_DetectDrift_PathScopedStoresDirectoryAndWorkspace(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)

	var capturedCmd *events.CommentCommand
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCmd = args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				RepoRelDir:  capturedCmd.RepoRelDir,
				Workspace:   capturedCmd.Workspace,
			}}, nil}
		})

	driftStorage := driftmocks.NewMockStorage()
	var stored models.ProjectDrift
	When(driftStorage.Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]())).
		Then(func(args []Param) ReturnValues {
			stored = args[1].(models.ProjectDrift)
			return ReturnValues{nil}
		})
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Paths: []models.DriftDetectionPath{{
			Directory: "./modules/app/",
			Workspace: "staging",
		}},
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	Assert(t, capturedCmd != nil, "expected plan command to be built")
	Equals(t, "modules/app", capturedCmd.RepoRelDir)
	Equals(t, "staging", capturedCmd.Workspace)
	Equals(t, "modules/app", stored.Path)
	Equals(t, "staging", stored.Workspace)
	Equals(t, "main", stored.Ref)
}

func TestAPIController_DetectDrift_RejectsMixedProjectAndPathSelectors(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
			Workspace: "prod",
		}},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
}

func TestAPIController_DetectDrift_SHARefRequiresBaseBranch(t *testing.T) {
	ac, _, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "0123456789abcdef0123456789abcdef01234567",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "base_branch"), "expected base_branch validation error, got %#v", apiErr.Details)
}

func TestAPIController_DetectDrift_InvalidBaseBranchReturnsBadRequest(t *testing.T) {
	ac, _, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		BaseBranch: "refs/merge-requests/1/head",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "base_branch"), "expected base_branch validation error, got %#v", apiErr.Details)
}

func TestAPIController_DetectDrift_InvalidRefReturnsBadRequest(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()
	vcsClient := ac.VCSClient.(*MockClient)

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "refs/pull/123/head",
		BaseBranch: "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "ref"), "expected ref validation error, got %#v", apiErr.Details)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
}

func TestAPIController_DetectDrift_InvalidWorkspaceReturnsBadRequest(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	ac.DriftStorage = driftmocks.NewMockStorage()
	vcsClient := ac.VCSClient.(*MockClient)

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Paths:      []models.DriftDetectionPath{{Directory: "env", Workspace: "../../tmp/plan"}},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
	apiErr := parseAPIError(t, w.Body.Bytes())
	Assert(t, strings.Contains(fmt.Sprintf("%v", apiErr.Details), "paths"), "expected paths validation error, got %#v", apiErr.Details)
	vcsClient.VerifyWasCalled(Never()).GetCloneURL(Any[logging.SimpleLogging](), Any[models.VCSHostType](), Any[string]())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
}

func TestAPIController_DetectDrift_SHARefUsesBaseBranchAndStoresResolvedCommit(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	ref := "0123456789abcdef0123456789abcdef01234567"
	var capturedCtx *command.Context
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCtx = args[0].(*command.Context)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			}}, nil}
		})

	driftStorage := driftmocks.NewMockStorage()
	var stored models.ProjectDrift
	When(driftStorage.Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]())).
		Then(func(args []Param) ReturnValues {
			stored = args[1].(models.ProjectDrift)
			return ReturnValues{nil}
		})
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        ref,
		BaseBranch: "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	Assert(t, capturedCtx != nil, "expected plan command context")
	Equals(t, "main", capturedCtx.Pull.BaseBranch)
	Equals(t, ref, capturedCtx.Pull.HeadBranch)
	Equals(t, "main", stored.BaseBranch)
	Equals(t, ref, stored.Ref)
	Equals(t, ref, stored.ResolvedCommit)
}

func TestAPIController_DetectDrift_PartialProjectFailure(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "ok",
				RepoRelDir:  "ok",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.Plan,
				ProjectName: "bad",
				RepoRelDir:  "bad",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		Then(func(args []Param) ReturnValues {
			projectCtx := args[0].(command.ProjectContext)
			if projectCtx.ProjectName == "bad" {
				return ReturnValues{command.ProjectCommandOutput{Error: errors.New("terraform plan failed")}}
			}
			return ReturnValues{command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
			}}
		})

	driftStorage := driftmocks.NewMockStorage()
	stored := map[string]models.ProjectDrift{}
	When(driftStorage.Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]())).
		Then(func(args []Param) ReturnValues {
			projectDrift := args[1].(models.ProjectDrift)
			stored[projectDrift.ProjectName] = projectDrift
			return ReturnValues{nil}
		})
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusMultiStatus, w.Code)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftDetectionResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, 2, result.Summary.TotalProjects)
	Equals(t, 1, result.Summary.ProjectsWithErrors)
	Equals(t, "terraform plan failed", stored["bad"].Error)
	Equals(t, false, stored["bad"].Drift.HasDrift)
	Equals(t, true, stored["ok"].Drift.HasDrift)
}

func TestAPIController_DetectDrift_StorageFailureIsProjectError(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "stored",
				RepoRelDir:  "stored",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.Plan,
				ProjectName: "unstored",
				RepoRelDir:  "unstored",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
		})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]())).
		Then(func(args []Param) ReturnValues {
			projectDrift := args[1].(models.ProjectDrift)
			if projectDrift.ProjectName == "unstored" {
				return ReturnValues{errors.New("store unavailable")}
			}
			return ReturnValues{nil}
		})
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"all"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusMultiStatus, w.Code)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftDetectionResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, 2, result.Summary.TotalProjects)
	Equals(t, 1, result.Summary.ProjectsWithErrors)
	var unstored controllers.DriftProjectAPI
	for _, project := range result.Projects {
		if project.ProjectName == "unstored" {
			unstored = project
			break
		}
	}
	Assert(t, strings.Contains(unstored.Error, "storing drift result: store unavailable"), "expected storage error, got %q", unstored.Error)
	Equals(t, true, unstored.HasDrift)
	driftStorage.VerifyWasCalled(Never()).DeleteMatching(Any[string](), Any[drift.GetOptions]())
}

func TestAPIController_DetectDrift_PolicyCheckContextsUsePolicyRunner(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	postWorkflowHooksCommandRunner := ac.PostWorkflowHooksCommandRunner.(*MockPostWorkflowHooksCommandRunner)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."}})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: "policy", Passed: true}},
		}})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	projectCommandRunner.VerifyWasCalled(Times(1)).PolicyCheck(Any[command.ProjectContext]())
	postWorkflowHooksCommandRunner.VerifyWasCalled(Times(1)).RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())
	driftStorage.VerifyWasCalled(Times(1)).Store(Any[string](), Any[models.ProjectDrift]())
}

func TestAPIController_DetectDrift_SkipsPolicyChecksWhenPlanFails(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{Error: errors.New("terraform plan failed")})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusMultiStatus, w.Code)
	projectCommandRunner.VerifyWasCalled(Never()).PolicyCheck(Any[command.ProjectContext]())
	_, stored := driftStorage.VerifyWasCalledOnce().
		Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]()).
		GetCapturedArguments()
	Equals(t, "terraform plan failed", stored.Error)
}

func TestAPIController_DetectDrift_PolicyCheckFailureIsVisibleAndSuppressesWebhook(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
		})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{Failure: "policy denied"})

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage
	sender := &recordingDriftSender{}
	ac.DriftWebhookSender = &webhooks.DriftWebhookSender{Webhooks: []webhooks.DriftSender{sender}}

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusMultiStatus, w.Code)
	_, stored := driftStorage.VerifyWasCalledOnce().
		Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]()).
		GetCapturedArguments()
	Assert(t, strings.Contains(stored.Error, "policy_check failed: policy denied"), "expected policy error in stored drift, got %q", stored.Error)
	Equals(t, 0, sender.calls)
}

func TestAPIController_DetectDrift_FullDetectionRemovesStaleSameRefRecords(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "stale",
		Path:        "stale",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		LastChecked: time.Now(),
	}))
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "other-ref",
		Path:        "other-ref",
		Workspace:   events.DefaultWorkspace,
		Ref:         "dev",
		BaseBranch:  "dev",
		LastChecked: time.Now(),
	}))

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "current",
			RepoRelDir:  "current",
			Workspace:   events.DefaultWorkspace,
		}}, nil)

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
	mainRecords, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main"})
	Ok(t, err)
	Equals(t, 1, len(mainRecords))
	Equals(t, "current", mainRecords[0].ProjectName)
	devRecords, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "dev"})
	Ok(t, err)
	Equals(t, 1, len(devRecords))
	Equals(t, "other-ref", devRecords[0].ProjectName)
}

func TestAPIController_DetectDrift_FullDetectionDeletesStaleUnnamedWithoutDeletingNamedSamePath(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "",
		Path:        "app",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		LastChecked: time.Now(),
	}))

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "app",
			RepoRelDir:  "app",
			Workspace:   events.DefaultWorkspace,
		}}, nil)

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
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "app", records[0].ProjectName)
	Equals(t, "app", records[0].Path)
}

func TestAPIController_DetectDrift_ScopedDetectionKeepsUnrelatedRecords(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "stale",
		Path:        "stale",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		LastChecked: time.Now(),
	}))

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "current",
			RepoRelDir:  "current",
			Workspace:   events.DefaultWorkspace,
		}}, nil)

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"current"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	mainRecords, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main"})
	Ok(t, err)
	Equals(t, 2, len(mainRecords))
}

func TestAPIController_DetectDrift_NonFatalPreHookFailureSkipsStaleReconciliation(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "cached",
		Path:        "cached",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)
	preWorkflowHooksRunner := ac.PreWorkflowHooksCommandRunner.(*MockPreWorkflowHooksCommandRunner)
	When(preWorkflowHooksRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn(errors.New("generate config failed"))

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
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "cached", records[0].ProjectName)
}

func TestAPIController_DetectDrift_CommandBuildFailureSkipsStaleReconciliation(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "cached",
		Path:        "cached",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn(nil, errors.New("project discovery failed"))

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "cached", records[0].ProjectName)
}

func TestAPIController_DetectDrift_ZeroProjectFullDetectionReconcilesStaleRecords(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	sender := &recordingDriftSender{}
	ac.DriftWebhookSender = &webhooks.DriftWebhookSender{Webhooks: []webhooks.DriftSender{sender}}
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "cached",
		Path:        "cached",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)

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
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftDetectionResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, 0, result.Summary.TotalProjects)
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 0, len(records))
	Equals(t, 1, sender.calls)
	Equals(t, 1, len(sender.results))
	Equals(t, "main", sender.results[0].Ref)
	Equals(t, 0, sender.results[0].TotalProjects)
}

func TestAPIController_DetectDrift_ReconciliationKeepsNewerRecords(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "newer",
		Path:        "newer",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now().Add(time.Hour),
	}))
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)

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
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "newer", records[0].ProjectName)
}

func TestAPIController_DetectDrift_ReconciliationUsesDetectionStartTime(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			time.Sleep(10 * time.Millisecond)
			Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
				ProjectName: "newer",
				Path:        "newer",
				Workspace:   events.DefaultWorkspace,
				Ref:         "main",
				BaseBranch:  "main",
				Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
				LastChecked: time.Now(),
			}))
			return ReturnValues{[]command.ProjectContext{}, nil}
		})

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
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "newer", records[0].ProjectName)
}

func TestAPIController_DetectDrift_FullDetectionSerializesByRepoRefBase(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"

	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondEntered := make(chan struct{})
	var mu sync.Mutex
	buildCalls := 0
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			mu.Lock()
			buildCalls++
			call := buildCalls
			mu.Unlock()

			switch call {
			case 1:
				close(firstEntered)
				<-releaseFirst
				return ReturnValues{[]command.ProjectContext{{
					CommandName: command.Plan,
					ProjectName: "stale",
					RepoRelDir:  "stale",
					Workspace:   events.DefaultWorkspace,
				}}, nil}
			case 2:
				close(secondEntered)
				return ReturnValues{[]command.ProjectContext{{
					CommandName: command.Plan,
					ProjectName: "fresh",
					RepoRelDir:  "fresh",
					Workspace:   events.DefaultWorkspace,
				}}, nil}
			default:
				return ReturnValues{[]command.ProjectContext{}, nil}
			}
		})

	runDetect := func(done chan<- *httptest.ResponseRecorder) {
		body, _ := json.Marshal(models.DriftDetectionRequest{
			Repository: "Repo",
			Ref:        "main",
			Type:       "Gitlab",
		})
		req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
		req.Header.Set(atlantisTokenHeader, atlantisToken)
		w := httptest.NewRecorder()
		ac.DetectDrift(w, req)
		done <- w
	}
	waitResponse := func(done <-chan *httptest.ResponseRecorder) *httptest.ResponseRecorder {
		select {
		case w := <-done:
			return w
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for drift detection")
			return nil
		}
	}

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	secondDone := make(chan *httptest.ResponseRecorder, 1)
	go runDetect(firstDone)
	select {
	case <-firstEntered:
	case <-time.After(time.Second):
		t.Fatal("first full detection did not start")
	}

	go runDetect(secondDone)
	select {
	case <-secondEntered:
		t.Fatal("second full detection entered project discovery before first detection finished")
	case <-time.After(100 * time.Millisecond):
	}

	close(releaseFirst)
	firstResponse := waitResponse(firstDone)
	secondResponse := waitResponse(secondDone)
	Equals(t, http.StatusOK, firstResponse.Code)
	Equals(t, http.StatusOK, secondResponse.Code)

	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "fresh", records[0].ProjectName)
}

func TestAPIController_DetectDrift_ZeroProjectScopedDetectionKeepsStaleRecords(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	storage := drift.NewInMemoryStorage()
	ac.DriftStorage = storage
	repositoryKey := "gitlab.com/Repo"
	Ok(t, storage.Store(repositoryKey, models.ProjectDrift{
		ProjectName: "cached",
		Path:        "cached",
		Workspace:   events.DefaultWorkspace,
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{}, nil)

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"missing"},
	})
	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)
	records, err := storage.Get(repositoryKey, drift.GetOptions{Ref: "main", BaseBranch: "main"})
	Ok(t, err)
	Equals(t, 1, len(records))
	Equals(t, "cached", records[0].ProjectName)
}

func TestAPIController_DetectDrift_ResolvesSyntheticRefBeforePlanReuse(t *testing.T) {
	ac, projectCommandBuilder, _ := setup(t)
	repoDir, headCommit := initAPIControllerGitRepo(t)

	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(repoDir, nil)

	var capturedHeadCommit string
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			capturedHeadCommit = ctx.Pull.HeadCommit
			return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
		})

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
	Equals(t, headCommit, capturedHeadCommit)
}

func TestAPIController_DetectDrift_SkipsPolicyCheckResults(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "modules/app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
			},
		})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).
		ThenReturn(command.ProjectCommandOutput{
			PolicyCheckResults: &models.PolicyCheckResults{},
		})

	body, _ := json.Marshal(models.DriftDetectionRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"app"},
	})

	req, _ := http.NewRequest("POST", "/api/drift/detect", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.DetectDrift(w, req)

	Equals(t, http.StatusOK, w.Code)

	_, storedDrift := driftStorage.VerifyWasCalledOnce().
		Store(Eq("gitlab.com/Repo"), Any[models.ProjectDrift]()).
		GetCapturedArguments()
	Equals(t, "app", storedDrift.ProjectName)
	Equals(t, "modules/app", storedDrift.Path)
	Equals(t, events.DefaultWorkspace, storedDrift.Workspace)
	Equals(t, true, storedDrift.Drift.HasDrift)
	Equals(t, 1, storedDrift.Drift.ToAdd)

	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.DriftDetectionResultAPI
	parseAPIResponse(t, response, &result)
	Equals(t, 1, len(result.Projects))
	Equals(t, "app", result.Projects[0].ProjectName)
	Equals(t, true, result.Projects[0].HasDrift)
	Equals(t, 1, result.Projects[0].Drift.ToAdd)
}

func TestAPIController_DetectDrift_CleansSyntheticWorkdir(t *testing.T) {
	ac, _, _ := setup(t)

	driftStorage := driftmocks.NewMockStorage()
	When(driftStorage.Store(Any[string](), Any[models.ProjectDrift]())).ThenReturn(nil)
	ac.DriftStorage = driftStorage

	workingDir := ac.WorkingDir.(*MockWorkingDir)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

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

	_, _, pull := workingDir.VerifyWasCalledOnce().
		Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]()).
		GetCapturedArguments()
	Assert(t, pull.Num < 0, "expected synthetic non-PR pull number, got %d", pull.Num)
}

func TestAPIController_DetectDrift_NoStorage(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)

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
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
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
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	locker := NewMockLocker(gmockCtrl)
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

	// Pre-workflow hooks are called once before project discovery so generated
	// config is used consistently by discovery and execution.
	_, capturedCmds := preWorkflowHooksRunner.VerifyWasCalled(Times(1)).
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetAllCapturedArguments()

	Assert(t, len(capturedCmds) == 1,
		"expected 1 pre-workflow hook call, got %d", len(capturedCmds))

	Assert(t, capturedCmds[0].Name == command.Plan,
		"expected first pre-hook command to be Plan, got %s", capturedCmds[0].Name.String())
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
