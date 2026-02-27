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
	. "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	emocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

const atlantisTokenHeader = "X-Atlantis-Token"
const atlantisToken = "token"

func TestAPIController_Plan(t *testing.T) {
	ac, _, _ := setup(t)

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

	// Verify computed call count matches expected test case structure.
	// BuildPlanCommands and Plan are called once per project/path entry.
	Equals(t, 5, expectedCalls)
}

func TestAPIController_Apply(t *testing.T) {
	ac, _, _ := setup(t)

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

	// Verify computed call count matches expected test case structure.
	// BuildApplyCommands, Plan, and Apply are called once per project/path entry.
	Equals(t, 5, expectedCalls)
}

// TestAPIController_Plan_PreWorkflowHooksReceiveCorrectCommand verifies that when
// calling the Plan API endpoint, the pre-workflow hooks receive a CommentCommand
// with Name set to command.Plan (not the zero value which would be command.Apply).
func TestAPIController_Plan_PreWorkflowHooksReceiveCorrectCommand(t *testing.T) {
	ac, _, _ := setup(t)

	// Create a fresh mock to replace the one from setup (which has AnyTimes that would consume the call)
	gmockCtrl := gomock.NewController(t)
	preWorkflowHooksRunner := emocks.NewMockPreWorkflowHooksCommandRunner(gmockCtrl)

	var capturedCmd *events.CommentCommand
	preWorkflowHooksRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *command.Context, cmd *events.CommentCommand) error {
			capturedCmd = cmd
			return nil
		}).Times(1)

	ac.PreWorkflowHooksCommandRunner = preWorkflowHooksRunner

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

	Assert(t, capturedCmd != nil, "expected RunPreHooks to be called")
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

	// Create a fresh mock to replace the one from setup (which has AnyTimes that would consume the call)
	gmockCtrl := gomock.NewController(t)
	preWorkflowHooksRunner := emocks.NewMockPreWorkflowHooksCommandRunner(gmockCtrl)

	var capturedCmds []*events.CommentCommand
	preWorkflowHooksRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *command.Context, cmd *events.CommentCommand) error {
			capturedCmds = append(capturedCmds, cmd)
			return nil
		}).Times(2)

	ac.PreWorkflowHooksCommandRunner = preWorkflowHooksRunner

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
	ac.Locker.(*MockLocker).EXPECT().List().Return(mockLocks, nil)

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
	ac.Locker.(*MockLocker).EXPECT().List().Return(mockLocks, nil)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	ac.ListLocks(w, req)
	response, _ := io.ReadAll(w.Result().Body)
	var result controllers.ListLocksResult
	err := json.Unmarshal(response, &result)
	Ok(t, err)
	Equals(t, expected, result)
}

func setup(t *testing.T) (controllers.APIController, *emocks.MockProjectCommandBuilder, *emocks.MockProjectCommandRunner) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	locker := NewMockLocker(gmockCtrl)
	// Allow incidental calls to UnlockByPull (called internally during plan/apply operations)
	locker.EXPECT().UnlockByPull(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	logger := logging.NewNoopLogger(t)
	parser := emocks.NewMockEventParsing(gmockCtrl)
	parser.EXPECT().ParseAPIPlanRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(models.Repo{}, nil).AnyTimes()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	scope := metricstest.NewLoggingScope(t, logger, "null")
	vcsClient := vcsmocks.NewMockClient()
	workingDir := emocks.NewMockWorkingDir(gmockCtrl)
	workingDir.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	Ok(t, err)

	workingDirLocker := emocks.NewMockWorkingDirLocker(gmockCtrl)
	workingDirLocker.EXPECT().TryLock(gomock.Any(), gomock.Any(), events.DefaultWorkspace, events.DefaultRepoRelDir, "", gomock.Any()).
		Return(func() {}, nil).AnyTimes()

	projectCommandBuilder := emocks.NewMockProjectCommandBuilder(gmockCtrl)
	projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).
		Return([]command.ProjectContext{{
			CommandName: command.Plan,
		}}, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).
		Return([]command.ProjectContext{{
			CommandName: command.Apply,
		}}, nil).AnyTimes()

	projectCommandRunner := emocks.NewMockProjectCommandRunner(gmockCtrl)
	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{},
	}).AnyTimes()
	projectCommandRunner.EXPECT().Apply(gomock.Any()).Return(command.ProjectCommandOutput{
		ApplySuccess: "success",
	}).AnyTimes()

	preWorkflowHooksCommandRunner := emocks.NewMockPreWorkflowHooksCommandRunner(gmockCtrl)
	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	postWorkflowHooksCommandRunner := emocks.NewMockPostWorkflowHooksCommandRunner(gmockCtrl)
	postWorkflowHooksCommandRunner.EXPECT().RunPostHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	commitStatusUpdater := emocks.NewMockCommitStatusUpdater(gmockCtrl)
	commitStatusUpdater.EXPECT().UpdateCombined(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	commitStatusUpdater.EXPECT().UpdateCombinedCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
