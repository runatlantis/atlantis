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

func TestOutputAPIRequiresOptInAndAuthentication(t *testing.T) {
	name := command.Output
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
			ac.ProjectOutputCommandRunner = runner
			req := httptest.NewRequest(http.MethodPost, "/api/output", bytes.NewBufferString(`{}`))
			req.Header.Set(atlantisTokenHeader, tc.token)
			w := httptest.NewRecorder()
			ac.Output(w, req)
			Equals(t, tc.status, w.Code)
			runner.VerifyWasCalled(Never()).Output(Any[command.ProjectContext]())
		})
	}
}

func TestOutputAPIExecutesRequestedCommand(t *testing.T) {
	name := command.Output
	t.Run(name.String(), func(t *testing.T) {
		ac, builder, runner := setup(t)
		ac.AllowAPICommands = []command.Name{name}
		ac.ProjectOutputCommandRunner = runner
		args := []string{"-raw", "endpoint"}
		When(builder.BuildOutputCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(func(params []Param) ReturnValues {
			cc := params[1].(*events.CommentCommand)
			Equals(t, name, cc.Name)
			Equals(t, args, cc.Flags)
			return ReturnValues{[]command.ProjectContext{{CommandName: name, CommentArgs: cc.Flags, ProjectName: "app"}}, nil}
		})
		When(runner.Output(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{OutputSuccess: &models.OutputSuccess{Output: "value"}})
		body, err := json.Marshal(controllers.APIRequest{Repository: "Repo", Ref: "main", Type: "Gitlab", PR: 1, Projects: []string{"app"}, Arguments: args})
		Ok(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/output", bytes.NewReader(body))
		req.Header.Set(atlantisTokenHeader, atlantisToken)
		w := httptest.NewRecorder()
		ac.Output(w, req)
		Assert(t, w.Code == http.StatusOK, "unexpected response %d: %s", w.Code, w.Body.String())
		var result command.Result
		Ok(t, json.Unmarshal(w.Body.Bytes(), &result))
		Equals(t, "value", result.ProjectResults[0].OutputSuccess.Output)
		runner.VerifyWasCalledOnce().Output(Any[command.ProjectContext]())
	})
}

func TestOutputAPIReportsCheckoutErrors(t *testing.T) {
	ac, builder, runner := setup(t)
	ac.AllowAPICommands = []command.Name{command.Output}
	ac.ProjectOutputCommandRunner = runner
	When(ac.WorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn("", errors.New("cloning repository"))
	req := httptest.NewRequest(http.MethodPost, "/api/output", bytes.NewBufferString(`{"Repository":"Repo","Ref":"main","Type":"Gitlab","PR":1,"Projects":["app"]}`))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Output(w, req)
	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "cloning repository"), "expected checkout error: %s", w.Body.String())
	builder.VerifyWasCalled(Never()).BuildOutputCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	runner.VerifyWasCalled(Never()).Output(Any[command.ProjectContext]())
}

func TestOutputAPIRejectsArgumentsAndMissingTarget(t *testing.T) {
	for _, args := range [][]string{{"-state=/tmp/other"}, {"a", "b"}, nil} {
		t.Run(joinAPIArgs(args), func(t *testing.T) {
			ac, _, runner := setup(t)
			ac.AllowAPICommands = []command.Name{command.Output}
			ac.ProjectOutputCommandRunner = runner
			body, err := json.Marshal(controllers.APIRequest{Repository: "Repo", Ref: "main", Type: "Gitlab", PR: 1, Arguments: args})
			Ok(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/output", bytes.NewReader(body))
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			ac.Output(w, req)
			Equals(t, http.StatusBadRequest, w.Code)
			runner.VerifyWasCalled(Never()).Output(Any[command.ProjectContext]())
		})
	}
}

func joinAPIArgs(args []string) string { return string(mustJSON(args)) }
func mustJSON(v any) []byte            { b, _ := json.Marshal(v); return b }

func TestOutputAPIAuthorization(t *testing.T) {
	for _, rule := range []string{"platform:output", "*:state", "*:output", "external deny", "external allow"} {
		t.Run(rule, func(t *testing.T) {
			ac, _, runner := setup(t)
			ac.AllowAPICommands = []command.Name{command.Output}
			ac.ProjectOutputCommandRunner = runner
			When(runner.Output(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{OutputSuccess: &models.OutputSuccess{}})
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
			req := httptest.NewRequest(http.MethodPost, "/api/output", bytes.NewReader(body))
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			ac.Output(w, req)
			if rule == "*:output" || rule == "external allow" {
				Equals(t, http.StatusOK, w.Code)
				runner.VerifyWasCalledOnce().Output(Any[command.ProjectContext]())
			} else {
				ResponseContains(t, w, http.StatusForbidden, "not authorized")
				runner.VerifyWasCalled(Never()).Output(Any[command.ProjectContext]())
			}
			if external != nil {
				Equals(t, 1, external.calls)
			}
		})
	}
}

func TestOutputAPIIgnoredTargetReturnsEmptySuccess(t *testing.T) {
	ac, builder, runner := setup(t)
	ac.AllowAPICommands = []command.Name{command.Output}
	ac.ProjectOutputCommandRunner = runner
	When(builder.BuildOutputCommands(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil, events.ErrIgnoredTargetedDir)
	body, err := json.Marshal(controllers.APIRequest{Repository: "Repo", Ref: "main", Type: "Gitlab", PR: 1, Paths: []controllers.APIRequestPath{{Directory: "ignored", Workspace: "default"}}, Arguments: []string{}})
	Ok(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/output", bytes.NewReader(body))
	req.Header.Set(atlantisTokenHeader, atlantisToken)
	w := httptest.NewRecorder()
	ac.Output(w, req)
	Equals(t, http.StatusOK, w.Code)
	var result command.Result
	Ok(t, json.Unmarshal(w.Body.Bytes(), &result))
	Equals(t, 0, len(result.ProjectResults))
	runner.VerifyWasCalled(Never()).Output(Any[command.ProjectContext]())
}
