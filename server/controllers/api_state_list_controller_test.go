// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStateListAPIRequiresOptInAndAuthentication(t *testing.T) {
	name := command.StateList
	for _, tc := range []struct {
		token   string
		enabled bool
		status  int
	}{
		{atlantisToken, false, http.StatusForbidden},
		{"", true, http.StatusUnauthorized},
		{"incorrect", true, http.StatusUnauthorized},
	} {
		t.Run(name.String()+tc.token, func(t *testing.T) {
			ac, _, runner := setup(t)
			if tc.enabled {
				ac.AllowAPICommands = []command.Name{name}
			}
			ac.ProjectStateListCommandRunner = runner
			req := httptest.NewRequest(http.MethodPost, "/api/state/list", bytes.NewBufferString(`{}`))
			req.Header.Set(atlantisTokenHeader, tc.token)
			w := httptest.NewRecorder()
			ac.StateList(w, req)
			Equals(t, tc.status, w.Code)
			runner.VerifyWasCalled(Never()).StateList(Any[command.ProjectContext]())
		})
	}
}

func TestStateListAPIExecutesRequestedCommand(t *testing.T) {
	name := command.StateList
	t.Run(name.String(), func(t *testing.T) {
		ac, builder, runner := setup(t)
		ac.AllowAPICommands = []command.Name{name}
		ac.ProjectStateListCommandRunner = runner
		var args []string
		When(builder.BuildStateListCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(params []Param) ReturnValues {
			cc := params[1].(*events.CommentCommand)
			Equals(t, name, cc.Name)
			Equals(t, args, cc.Flags)
			return ReturnValues{[]command.ProjectContext{{CommandName: name, CommentArgs: cc.Flags, ProjectName: "app"}}, nil}
		})
		When(runner.StateList(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{StateListSuccess: &models.StateListSuccess{Output: "value"}})
		body, err := json.Marshal(controllers.APIRequest{Repository: "Repo", Ref: "main", Type: "Gitlab", PR: 1, Projects: []string{"app"}, Arguments: args})
		Ok(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/state/list", bytes.NewReader(body))
		req.Header.Set(atlantisTokenHeader, atlantisToken)
		w := httptest.NewRecorder()
		ac.StateList(w, req)
		Assert(t, w.Code == http.StatusOK, "unexpected response %d: %s", w.Code, w.Body.String())
		var result command.Result
		Ok(t, json.Unmarshal(w.Body.Bytes(), &result))
		Equals(t, "value", result.ProjectResults[0].StateListSuccess.Output)
		runner.VerifyWasCalledOnce().StateList(Any[command.ProjectContext]())
	})
}

func TestStateListAPIReportsCheckoutErrors(t *testing.T) {
	ac, builder, runner := setup(t)
	ac.AllowAPICommands = []command.Name{command.StateList}
	ac.ProjectStateListCommandRunner = runner
	When(ac.WorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn("", errors.New("cloning repository"))
	req := httptest.NewRequest(http.MethodPost, "/api/state/list", bytes.NewBufferString(`{"Repository":"Repo","Ref":"main","Type":"Gitlab","PR":1,"Projects":["app"]}`))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.StateList(w, req)
	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "cloning repository"), "expected checkout error: %s", w.Body.String())
	builder.VerifyWasCalled(Never()).BuildStateListCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	runner.VerifyWasCalled(Never()).StateList(Any[command.ProjectContext]())
}

func TestStateListAPIAuthorization(t *testing.T) {
	for _, rule := range []string{"platform:state_list", "*:state", "*:state_list", "external deny", "external allow"} {
		t.Run(rule, func(t *testing.T) {
			ac, _, runner := setup(t)
			ac.AllowAPICommands = []command.Name{command.StateList}
			ac.ProjectStateListCommandRunner = runner
			When(runner.StateList(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{StateListSuccess: &models.StateListSuccess{}})
			setupAPIProjectBuilder(t, ac)
			var external *apiTeamAuthorizationRunner
			if strings.HasPrefix(rule, "external") {
				external = &apiTeamAuthorizationRunner{output: "deny"}
				if rule == "external allow" {
					external.output = "pass"
				}
				ac.TeamAllowlistChecker = &events.ExternalTeamAllowlistChecker{Command: "authorize", ExternalTeamAllowlistRunner: external}
			} else {
				checker, err := command.NewTeamAllowlistChecker(rule)
				Ok(t, err)
				ac.TeamAllowlistChecker = checker
			}
			body, err := json.Marshal(controllers.APIRequest{Repository: "Repo", Ref: "main", Type: "Gitlab", PR: 1, Projects: []string{"app"}, Arguments: []string{}})
			Ok(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/state/list", bytes.NewReader(body))
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			ac.StateList(w, req)
			if rule == "*:state_list" || rule == "external allow" {
				Equals(t, http.StatusOK, w.Code)
				runner.VerifyWasCalledOnce().StateList(Any[command.ProjectContext]())
			} else {
				ResponseContains(t, w, http.StatusForbidden, "not authorized")
				runner.VerifyWasCalled(Never()).StateList(Any[command.ProjectContext]())
			}
			if external != nil {
				Equals(t, 1, external.calls)
			}
		})
	}
}

func TestStateListAPIIgnoredTargetReturnsEmptySuccess(t *testing.T) {
	ac, builder, runner := setup(t)
	ac.AllowAPICommands = []command.Name{command.StateList}
	ac.ProjectStateListCommandRunner = runner
	When(builder.BuildStateListCommands(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil, events.ErrIgnoredTargetedDir)
	body, err := json.Marshal(controllers.APIRequest{Repository: "Repo", Ref: "main", Type: "Gitlab", PR: 1, Paths: []controllers.APIRequestPath{{Directory: "ignored", Workspace: "default"}}, Arguments: []string{}})
	Ok(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/state/list", bytes.NewReader(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.StateList(w, req)
	Equals(t, http.StatusOK, w.Code)
	var result command.Result
	Ok(t, json.Unmarshal(w.Body.Bytes(), &result))
	Equals(t, 0, len(result.ProjectResults))
	runner.VerifyWasCalled(Never()).StateList(Any[command.ProjectContext]())
}
