// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-github/v45/github"
	. "github.com/petergtz/pegomock"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/events/mocks"
	"github.com/runatlantis/atlantis/server/events"
	emocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	httputils "github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
	. "github.com/runatlantis/atlantis/testing"
	gitlab "github.com/xanzy/go-gitlab"
)

const gitlabHeader = "X-Gitlab-Event"

var secret = []byte("secret")

func AnyRepo() models.Repo {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(models.Repo{})))
	return models.Repo{}
}

func AnyStatus() []*github.RepoStatus {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(github.RepoStatus{})))
	return []*github.RepoStatus{}
}

func TestPost_NotGitlab(t *testing.T) {
	t.Log("when the request is not for gitlab or github a 400 is returned")
	e, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request")
}

func TestPost_UnsupportedVCSGitlab(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
}

func TestPost_InvalidGitlabSecret(t *testing.T) {
	t.Log("when the gitlab payload can't be validated a 400 is returned")
	e, gl, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_UnsupportedGitlabEvent(t *testing.T) {
	t.Log("when the event type is an unsupported gitlab event we ignore it")
	e, gl, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

// Test that if the comment comes from a commit rather than a merge request,
// we give an error and ignore it.
func TestPost_GitlabCommentOnCommit(t *testing.T) {
	e, gl, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.CommitCommentEvent{}, nil)
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring comment on commit event")
}

func TestPost_GitlabCommentSuccess(t *testing.T) {
	t.Log("when the event is a gitlab comment with a valid command we call the command handler")
	e, gl, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Processing...\n")
}

func TestPost_GitlabMergeRequestInvalid(t *testing.T) {
	t.Log("when the event is a gitlab merge request with invalid data we return a 400")
	e, gl, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeEvent{}, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGitlabMergeRequestEvent(gitlab.MergeEvent{})).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Error parsing webhook: err")
}

func TestPost_GitlabMergeRequestClosedErrCleaningPull(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a closed gitlab merge request and an error occurs calling CleanUpPull we return a 500")
	e, gl, p, c, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	var event gitlab.MergeEvent
	event.ObjectAttributes.Action = "close"
	When(gl.ParseAndValidate(req, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGitlabMergeRequestEvent(event)).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(c.CleanUpPull(repo, pullRequest)).ThenReturn(errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: err")
}

func TestPost_GitlabMergeRequestSuccess(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a gitlab merge request and the cleanup works we return a 200")
	e, gl, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeEvent{}, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGitlabMergeRequestEvent(gitlab.MergeEvent{})).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Pull request cleaned successfully")
}

func TestPost_GitlabMergeRequestError(t *testing.T) {
	t.Log("if getting the gitlab merge request fails we return 400 with error")

	// setup
	e, gl, p, _, _, _, glGetter := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	repo := models.Repo{}
	pullRequest := models.PullRequest{}
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(p.ParseGitlabMergeRequestEvent(gitlab.MergeEvent{})).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(glGetter.GetMergeRequest(repo.FullName, pullRequest.Num)).ThenReturn(nil, errors.New("error"))
	w := httptest.NewRecorder()

	// act
	e.Post(w, req)

	// assert
	ResponseContains(t, w, http.StatusBadRequest, "error")
}

// Test Bitbucket server pull closed events.
func TestPost_BBServerPullClosed(t *testing.T) {
	cases := []struct {
		header string
	}{
		{
			"pr:deleted",
		},
		{
			"pr:merged",
		},
		{
			"pr:declined",
		},
	}

	for _, c := range cases {
		t.Run(c.header, func(t *testing.T) {
			RegisterMockTestingT(t)
			allowlist, err := events.NewRepoAllowlistChecker("*")
			Ok(t, err)
			ctxLogger := logging.NewNoopCtxLogger(t)
			scope, _, _ := metrics.NewLoggingScope(ctxLogger, "null")
			ec := &events_controllers.VCSEventsController{
				Parser: &events.EventParser{
					BitbucketUser:      "bb-user",
					BitbucketToken:     "bb-token",
					BitbucketServerURL: "https://bbserver.com",
				},
				CommentEventHandler:  noopCommentHandler{},
				PREventHandler:       noopPRHandler{},
				RepoAllowlistChecker: allowlist,
				SupportedVCSHosts:    []models.VCSHostType{models.BitbucketServer},
				VCSClient:            nil,
				Logger:               ctxLogger,
				Scope:                scope,
			}

			// Build HTTP request.
			requestBytes, err := os.ReadFile(filepath.Join("testfixtures", "bb-server-pull-deleted-event.json"))
			// Replace the eventKey field with our event type.
			requestJSON := strings.Replace(string(requestBytes), `"eventKey":"pr:deleted",`, fmt.Sprintf(`"eventKey":"%s",`, c.header), -1)
			Ok(t, err)
			req, err := http.NewRequest("POST", "/events", bytes.NewBuffer([]byte(requestJSON)))
			Ok(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Event-Key", c.header)
			req.Header.Set("X-Request-ID", "request-id")

			// Send the request.
			w := httptest.NewRecorder()
			ec.Post(w, req)

			// Make our assertions.
			ResponseContains(t, w, 200, "Processing...")
		})
	}
}

func TestPost_PullOpenedOrUpdated(t *testing.T) {
	cases := []struct {
		Description string
		HostType    models.VCSHostType
		Action      string
	}{
		{
			"gitlab opened",
			models.Gitlab,
			"open",
		},
		{
			"gitlab update",
			models.Gitlab,
			"update",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			e, gl, p, _, _, _, _ := setup(t)
			req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
			var pullRequest models.PullRequest
			var repo models.Repo

			switch c.HostType {
			case models.Gitlab:
				req.Header.Set(gitlabHeader, "value")
				var event gitlab.MergeEvent
				event.ObjectAttributes.Action = c.Action
				When(gl.ParseAndValidate(req, secret)).ThenReturn(event, nil)
				repo = models.Repo{}
				pullRequest = models.PullRequest{State: models.ClosedPullState}
				When(p.ParseGitlabMergeRequestEvent(event)).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)
			}

			w := httptest.NewRecorder()
			e.Post(w, req)
			ResponseContains(t, w, http.StatusOK, "Processing...")
		})
	}
}

func setup(t *testing.T) (events_controllers.VCSEventsController, *mocks.MockGitlabRequestParserValidator, *emocks.MockEventParsing, *emocks.MockPullCleaner, *vcsmocks.MockClient, *emocks.MockCommentParsing, *mocks.MockGitlabMergeRequestGetter) {
	RegisterMockTestingT(t)
	gitLabMock := mocks.NewMockGitlabMergeRequestGetter()
	azureDevopsMock := mocks.NewMockAzureDevopsPullGetter()
	gl := mocks.NewMockGitlabRequestParserValidator()
	p := emocks.NewMockEventParsing()
	cp := emocks.NewMockCommentParsing()
	c := emocks.NewMockPullCleaner()
	vcsmock := vcsmocks.NewMockClient()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	Ok(t, err)
	ctxLogger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(ctxLogger, "null")
	e := events_controllers.VCSEventsController{
		Logger:                       ctxLogger,
		Scope:                        scope,
		Parser:                       p,
		CommentEventHandler:          noopCommentHandler{},
		PREventHandler:               noopPRHandler{},
		CommentParser:                cp,
		SupportedVCSHosts:            []models.VCSHostType{models.Gitlab},
		GitlabWebhookSecret:          secret,
		GitlabRequestParserValidator: gl,
		RepoAllowlistChecker:         repoAllowlistChecker,
		VCSClient:                    vcsmock,

		GitlabMergeRequestGetter: gitLabMock,
		AzureDevopsPullGetter:    azureDevopsMock,
	}
	return e, gl, p, c, vcsmock, cp, gitLabMock
}

// This struct shouldn't be using these anyways since it should be broken down into separate packages (ie. see github)
// therefore we're just using noop implementations here for now.
// agreed this means we're not verifying any of the arguments passed in, but that's fine since we don't use any of these providers
// atm.
type noopPRHandler struct{}

func (h noopPRHandler) Handle(ctx context.Context, request *httputils.BufferedRequest, event event_types.PullRequest) error {
	return nil
}

type noopCommentHandler struct{}

func (h noopCommentHandler) Handle(ctx context.Context, request *httputils.BufferedRequest, event event_types.Comment) error {
	return nil
}
