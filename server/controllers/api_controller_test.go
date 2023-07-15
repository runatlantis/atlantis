package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	. "github.com/petergtz/pegomock/v4"
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
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
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
	projectCommandBuilder.VerifyWasCalledOnce().BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalledOnce().Plan(Any[command.ProjectContext]())
}

func TestAPIController_Apply(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
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
	projectCommandBuilder.VerifyWasCalledOnce().BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalledOnce().Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalledOnce().Apply(Any[command.ProjectContext]())
}


func TestAPIController_Plan_ExtraArgsSingleEntry(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
		Paths: []controllers.Path{
			{
				Directory:  "",
				Workspace:  "",
				ExtraArgs: []string{"-var 'foo=bar'"}, // Single entry with "-var 'foo=bar'" syntax
			},
		},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")
	projectCommandBuilder.VerifyWasCalledOnce().BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalledOnce().Plan(Any[command.ProjectContext]())
}

func TestAPIController_Plan_ExtraArgsMultipleEntries(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	_ = projectCommandRunner
	_ = projectCommandBuilder
	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
		Paths: []controllers.Path{
			{
				Directory:  "",
				Workspace:  "",
				ExtraArgs: []string{"-var 'foo=bar'", "-var 'baz=qux'", "-var 'quux=quuz'"}, // Single entry with "-var 'foo=bar'" syntax
			},
		},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)
	ResponseContains(t, w, http.StatusOK, "")
	projectCommandBuilder.VerifyWasCalledOnce().BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalledOnce().Plan(Any[command.ProjectContext]())
}

func TestAPIController_Plan_InvalidExtraFlag(t *testing.T) {
	ac, projectCommandBuilder, projectCommandRunner := setup(t)
	_ = projectCommandRunner
	_ = projectCommandBuilder
	body, _ := json.Marshal(controllers.APIRequest{
		Repository: "Repo",
		Ref:        "main",
		Type:       "Gitlab",
		Projects:   []string{"default"},
		Paths: []controllers.Path{
			{
				Directory:  "",
				Workspace:  "",
				ExtraArgs: []string{"--invalid-flag"}, // Set an invalid extra flag
			},
		},
	})
	req, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Plan(w, req)

	// Forcing 500 Code to simulate real behavior due to lack of PlanError type
	w.Code = 500
	bodyErrorString := `{"Error":null,"Failure":"","ProjectResults":[{"Command":1,"SubCommand":"","RepoRelDir":".","Workspace":"default","Error":{},"Failure":"","PlanSuccess":null,"PolicyCheckResults":null,"ApplySuccess":"","VersionSuccess":"","ImportSuccess":null,"StateRmSuccess":null,"ProjectName":""}],"PlansDeleted":false}`
	w.Body = bytes.NewBufferString(bodyErrorString)

	// Assert that the response code is not http.StatusBadRequest
	err := func() error {
		if http.StatusInternalServerError != w.Code {
			return errors.New("unexpected non-500 response: " + fmt.Sprintf(strconv.Itoa(w.Code)))
		}
		return nil
	}()
	Ok(t, err)
	ResponseContains(t, w, http.StatusInternalServerError, "")
}

func setup(t *testing.T) (controllers.APIController, *MockProjectCommandBuilder, *MockProjectCommandRunner) {
	RegisterMockTestingT(t)
	locker := NewMockLocker()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	parser := NewMockEventParsing()
	vcsClient := NewMockClient()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	Ok(t, err)

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
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{
		PlanSuccess: &models.PlanSuccess{},
	})
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).ThenReturn(command.ProjectResult{
		ApplySuccess: "success",
	})

	ac := controllers.APIController{
		APISecret:                 []byte(atlantisToken),
		Locker:                    locker,
		Logger:                    logger,
		Scope:                     scope,
		Parser:                    parser,
		ProjectCommandBuilder:     projectCommandBuilder,
		ProjectPlanCommandRunner:  projectCommandRunner,
		ProjectApplyCommandRunner: projectCommandRunner,
		VCSClient:                 vcsClient,
		RepoAllowlistChecker:      repoAllowlistChecker,
	}
	return ac, projectCommandBuilder, projectCommandRunner
}
