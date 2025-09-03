package controllers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/runatlantis/atlantis/server/controllers"
	. "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
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

	// TODO: Convert verification to gomock expectations
	// projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).Times(expectedCalls)
	// projectCommandRunner.EXPECT().Plan(gomock.Any()).Times(expectedCalls)
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

	// TODO: Convert verification to gomock expectations
	// projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).Times(expectedCalls)
	// projectCommandRunner.EXPECT().Plan(gomock.Any()).Times(expectedCalls)
	// projectCommandRunner.EXPECT().Apply(gomock.Any()).Times(expectedCalls)
}

func TestAPIController_ListLocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	locker := NewMockLocker(ctrl)
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
	
	locker.EXPECT().List().Return(mockLocks, nil)
	
	ac := controllers.APIController{
		Locker: locker,
	}

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	locker := NewMockLocker(ctrl)
	expected := controllers.ListLocksResult{}
	mockLocks := map[string]models.ProjectLock{}
	
	locker.EXPECT().List().Return(mockLocks, nil)
	
	ac := controllers.APIController{
		Locker: locker,
	}

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	locker := NewMockLocker(ctrl)
	logger := logging.NewNoopLogger(t)
	parser := NewMockEventParsing(ctrl)
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	vcsClient := NewMockClient(ctrl)
	workingDir := NewMockWorkingDir(ctrl)
	Ok(t, err)

	workingDirLocker := NewMockWorkingDirLocker(ctrl)
	workingDirLocker.EXPECT().TryLock(gomock.Any(), gomock.Any(), gomock.Eq(events.DefaultWorkspace), gomock.Eq(events.DefaultRepoRelDir)).
		Return(func() {}, nil).AnyTimes()

	projectCommandBuilder := NewMockProjectCommandBuilder(ctrl)
	projectCommandBuilder.EXPECT().BuildPlanCommands(gomock.Any(), gomock.Any()).
		Return([]command.ProjectContext{{
			CommandName: command.Plan,
		}}, nil).AnyTimes()
	projectCommandBuilder.EXPECT().BuildApplyCommands(gomock.Any(), gomock.Any()).
		Return([]command.ProjectContext{{
			CommandName: command.Apply,
		}}, nil).AnyTimes()

	projectCommandRunner := NewMockProjectCommandRunner(ctrl)
	projectCommandRunner.EXPECT().Plan(gomock.Any()).Return(command.ProjectResult{
		PlanSuccess: &models.PlanSuccess{},
	}).AnyTimes()
	projectCommandRunner.EXPECT().Apply(gomock.Any()).Return(command.ProjectResult{
		ApplySuccess: "success",
	}).AnyTimes()

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner(ctrl)

	preWorkflowHooksCommandRunner.EXPECT().RunPreHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner(ctrl)

	postWorkflowHooksCommandRunner.EXPECT().RunPostHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	commitStatusUpdater := NewMockCommitStatusUpdater(ctrl)

	commitStatusUpdater.EXPECT().UpdateCombined(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Setup locker expectations for ListLocks tests
	locker.EXPECT().List().Return(map[string]models.ProjectLock{}, nil).AnyTimes()

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
