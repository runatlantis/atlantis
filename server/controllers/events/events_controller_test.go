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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	. "github.com/petergtz/pegomock/v4"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/events/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	emocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
	gitlab "github.com/xanzy/go-gitlab"
)

const githubHeader = "X-Github-Event"
const giteaHeader = "X-Gitea-Event"
const gitlabHeader = "X-Gitlab-Event"
const azuredevopsHeader = "Request-Id"

var user = []byte("user")
var secret = []byte("secret")

func TestPost_NotGithubOrGitlab(t *testing.T) {
	t.Log("when the request is not for gitlab or github a 400 is returned")
	e, _, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request")
}

func TestPost_UnsupportedVCSGithub(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitHub")
}

func TestPost_UnsupportedVCSGitea(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(giteaHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support Gitea")
}

func TestPost_UnsupportedVCSGitlab(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
}

func TestPost_InvalidGithubSecret(t *testing.T) {
	t.Log("when the github payload can't be validated a 400 is returned")
	e, v, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "value")
	When(v.Validate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_InvalidGiteaSecret(t *testing.T) {
	t.Log("when the gitea payload can't be validated a 400 is returned")
	e, v, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(giteaHeader, "value")
	When(v.Validate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "request did not pass validation")
}

func TestPost_InvalidGitlabSecret(t *testing.T) {
	t.Log("when the gitlab payload can't be validated a 400 is returned")
	e, _, gl, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_UnsupportedGithubEvent(t *testing.T) {
	t.Log("when the event type is an unsupported github event we ignore it")
	e, v, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "value")
	When(v.Validate(req, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_UnsupportedGiteaEvent(t *testing.T) {
	t.Log("when the event type is an unsupported gitea event we ignore it")
	e, v, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(giteaHeader, "value")
	e.GiteaWebhookSecret = nil
	When(v.Validate(req, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring unsupported Gitea event")
}

func TestPost_UnsupportedGitlabEvent(t *testing.T) {
	t.Log("when the event type is an unsupported gitlab event we ignore it")
	e, _, gl, _, _, _, _, _, _ := setup(t)
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
	e, _, gl, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.CommitCommentEvent{}, nil)
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring comment on commit event")
}

func TestPost_GithubCommentNotCreated(t *testing.T) {
	t.Log("when the event is a github comment but it's not a created event we ignore it")
	e, v, _, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	// comment action is deleted, not created
	event := `{"action": "deleted"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring comment event since action was not created")
}

func TestPost_GithubInvalidComment(t *testing.T) {
	t.Log("when the event is a github comment without all expected data we return a 400")
	e, v, _, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(Any[logging.SimpleLogging](), Any[*github.IssueCommentEvent]())).ThenReturn(models.Repo{}, models.User{}, 1, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Failed parsing event")
}

func TestPost_GitlabCommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a gitlab comment with an invalid command we ignore it")
	e, _, gl, _, _, _, _, _, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(cp.Parse("", models.Gitlab)).ThenReturn(events.CommentParseResult{Ignore: true})
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring non-command comment: \"\"")
}

func TestPost_GithubCommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a github comment with an invalid command we ignore it")
	e, v, _, _, p, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(Any[logging.SimpleLogging](), Any[*github.IssueCommentEvent]())).ThenReturn(models.Repo{}, models.User{}, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{Ignore: true})
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring non-command comment: \"\"")
	vcsClient.VerifyWasCalled(Never()).ReactToComment(Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(1), Eq(int64(1)), Eq("eyes"))
}

func TestPost_GitlabCommentNotAllowlisted(t *testing.T) {
	t.Log("when the event is a gitlab comment from a repo that isn't allowlisted we comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := events_controllers.VCSEventsController{
		Logger:                       logger,
		Scope:                        scope,
		CommentParser:                &events.CommentParser{ExecutableName: "atlantis"},
		GitlabRequestParserValidator: &events_controllers.DefaultGitlabRequestParserValidator{},
		Parser:                       &events.EventParser{},
		SupportedVCSHosts:            []models.VCSHostType{models.Gitlab},
		RepoAllowlistChecker:         &events.RepoAllowlistChecker{},
		VCSClient:                    vcsClient,
	}
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "gitlabMergeCommentEvent_notAllowlisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set(gitlabHeader, "Note Hook")
	w := httptest.NewRecorder()
	e.Post(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	Equals(t, http.StatusForbidden, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	exp := "Repo not allowlisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
	expRepo, _ := models.NewRepo(models.Gitlab, "gitlabhq/gitlab-test", "https://example.com/gitlabhq/gitlab-test.git", "", "")
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(expRepo), Eq(1), Eq("```\nError: This repo is not allowlisted for Atlantis.\n```"), Eq(""))
}

func TestPost_GitlabCommentNotAllowlistedWithSilenceErrors(t *testing.T) {
	t.Log("when the event is a gitlab comment from a repo that isn't allowlisted and we are silencing errors, do not comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := events_controllers.VCSEventsController{
		Logger:                       logger,
		Scope:                        scope,
		CommentParser:                &events.CommentParser{ExecutableName: "atlantis"},
		GitlabRequestParserValidator: &events_controllers.DefaultGitlabRequestParserValidator{},
		Parser:                       &events.EventParser{},
		SupportedVCSHosts:            []models.VCSHostType{models.Gitlab},
		RepoAllowlistChecker:         &events.RepoAllowlistChecker{},
		VCSClient:                    vcsClient,
		SilenceAllowlistErrors:       true,
	}
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "gitlabMergeCommentEvent_notAllowlisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set(gitlabHeader, "Note Hook")
	w := httptest.NewRecorder()
	e.Post(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	Equals(t, http.StatusForbidden, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	exp := "Repo not allowlisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())

}

func TestPost_GithubCommentNotAllowlisted(t *testing.T) {
	t.Log("when the event is a github comment from a repo that isn't allowlisted we comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := events_controllers.VCSEventsController{
		Logger:                 logger,
		Scope:                  scope,
		GithubRequestValidator: &events_controllers.DefaultGithubRequestValidator{},
		CommentParser:          &events.CommentParser{ExecutableName: "atlantis"},
		Parser:                 &events.EventParser{},
		SupportedVCSHosts:      []models.VCSHostType{models.Github},
		RepoAllowlistChecker:   &events.RepoAllowlistChecker{},
		VCSClient:              vcsClient,
	}
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "githubIssueCommentEvent_notAllowlisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "issue_comment")
	w := httptest.NewRecorder()
	e.Post(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	Equals(t, http.StatusForbidden, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	exp := "Repo not allowlisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
	expRepo, _ := models.NewRepo(models.Github, "baxterthehacker/public-repo", "https://github.com/baxterthehacker/public-repo.git", "", "")
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(expRepo), Eq(2), Eq("```\nError: This repo is not allowlisted for Atlantis.\n```"), Eq(""))
}

func TestPost_GithubCommentNotAllowlistedWithSilenceErrors(t *testing.T) {
	t.Log("when the event is a github comment from a repo that isn't allowlisted and we are silencing errors, do not comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := events_controllers.VCSEventsController{
		Logger:                 logger,
		Scope:                  scope,
		GithubRequestValidator: &events_controllers.DefaultGithubRequestValidator{},
		CommentParser:          &events.CommentParser{ExecutableName: "atlantis"},
		Parser:                 &events.EventParser{},
		SupportedVCSHosts:      []models.VCSHostType{models.Github},
		RepoAllowlistChecker:   &events.RepoAllowlistChecker{},
		VCSClient:              vcsClient,
		SilenceAllowlistErrors: true,
	}
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "githubIssueCommentEvent_notAllowlisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "issue_comment")
	w := httptest.NewRecorder()
	e.Post(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	Equals(t, http.StatusForbidden, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	exp := "Repo not allowlisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
	vcsClient.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestPost_GitlabCommentResponse(t *testing.T) {
	// When the event is a gitlab comment that warrants a comment response we comment back.
	e, _, gl, _, _, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(cp.Parse("", models.Gitlab)).ThenReturn(events.CommentParseResult{CommentResponse: "a comment"})
	w := httptest.NewRecorder()
	e.Post(w, req)
	vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(0), Eq("a comment"), Eq(""))
	ResponseContains(t, w, http.StatusOK, "Commenting back on pull request")
}

func TestPost_GithubCommentResponse(t *testing.T) {
	t.Log("when the event is a github comment that warrants a comment response we comment back")
	e, v, _, _, p, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	When(p.ParseGithubIssueCommentEvent(Any[logging.SimpleLogging](), Any[*github.IssueCommentEvent]())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{CommentResponse: "a comment"})
	w := httptest.NewRecorder()

	e.Post(w, req)
	vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(baseRepo), Eq(1), Eq("a comment"), Eq(""))
	ResponseContains(t, w, http.StatusOK, "Commenting back on pull request")
}

func TestPost_GitlabCommentSuccess(t *testing.T) {
	t.Log("when the event is a gitlab comment with a valid command we call the command handler")
	e, _, gl, _, _, cr, _, _, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	cmd := events.CommentCommand{}
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(cp.Parse(Any[string](), Eq(models.Gitlab))).ThenReturn(events.CommentParseResult{Command: &cmd})
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Processing...")

	cr.VerifyWasCalledOnce().RunCommentCommand(models.Repo{}, &models.Repo{}, nil, models.User{}, 0, &cmd)
}

func TestPost_GithubCommentSuccess(t *testing.T) {
	t.Log("when the event is a github comment with a valid command we call the command handler")
	e, v, _, _, p, cr, _, _, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	cmd := events.CommentCommand{}
	When(p.ParseGithubIssueCommentEvent(Any[logging.SimpleLogging](), Any[*github.IssueCommentEvent]())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{Command: &cmd})
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Processing...")

	cr.VerifyWasCalledOnce().RunCommentCommand(baseRepo, nil, nil, user, 1, &cmd)
}

func TestPost_GithubCommentReaction(t *testing.T) {
	t.Log("when the event is a github comment with a valid command we call the ReactToComment handler")
	e, v, _, _, p, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	testComment := "atlantis plan"
	event := fmt.Sprintf(`{"action": "created", "comment": {"body": "%v", "id": 1}}`, testComment)
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	cmd := events.CommentCommand{Name: command.Plan}
	When(p.ParseGithubIssueCommentEvent(Any[logging.SimpleLogging](), Any[*github.IssueCommentEvent]())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse(testComment, models.Github)).ThenReturn(events.CommentParseResult{Command: &cmd})
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Processing...")

	vcsClient.VerifyWasCalledOnce().ReactToComment(Any[logging.SimpleLogging](), Eq(baseRepo), Eq(1), Eq(int64(1)), Eq("eyes"))
}

func TestPost_GilabCommentReaction(t *testing.T) {
	t.Log("when the event is a gitlab comment with a valid command we call the ReactToComment handler")
	e, _, gl, _, _, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	cmd := events.CommentCommand{}
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(cp.Parse(Any[string](), Eq(models.Gitlab))).ThenReturn(events.CommentParseResult{Command: &cmd})
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Processing...")
	vcsClient.VerifyWasCalledOnce().ReactToComment(Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(0), Eq(int64(0)), Eq("eyes"))
}

func TestPost_GithubPullRequestInvalid(t *testing.T) {
	t.Log("when the event is a github pull request with invalid data we return a 400")
	e, v, _, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPullEvent(Any[logging.SimpleLogging](), Any[*github.PullRequestEvent]())).ThenReturn(models.PullRequest{}, models.OpenedPullEvent, models.Repo{}, models.Repo{}, models.User{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Error parsing pull data: err")
}

func TestPost_GitlabMergeRequestInvalid(t *testing.T) {
	t.Log("when the event is a gitlab merge request with invalid data we return a 400")
	e, _, gl, _, p, _, _, _, _ := setup(t)
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

func TestPost_GithubPullRequestNotAllowlisted(t *testing.T) {
	t.Log("when the event is a github pull request to a non-allowlisted repo we return a 400")
	e, v, _, _, _, _, _, _, _ := setup(t)
	var err error
	e.RepoAllowlistChecker, err = events.NewRepoAllowlistChecker("github.com/nevermatch")
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusForbidden, "Pull request event from non-allowlisted repo")
}

func TestPost_GitlabMergeRequestNotAllowlisted(t *testing.T) {
	t.Log("when the event is a gitlab merge request to a non-allowlisted repo we return a 400")
	e, _, gl, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")

	var err error
	e.RepoAllowlistChecker, err = events.NewRepoAllowlistChecker("github.com/nevermatch")
	Ok(t, err)
	When(gl.ParseAndValidate(req, secret)).ThenReturn(gitlab.MergeEvent{}, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGitlabMergeRequestEvent(gitlab.MergeEvent{})).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)

	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusForbidden, "Pull request event from non-allowlisted repo")
}

func TestPost_GithubPullRequestUnsupportedAction(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	e, v, _, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "unsupported"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Parser = &events.EventParser{}
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring non-actionable pull request event")
}

func TestPost_GitlabMergeRequestUnsupportedAction(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a gitlab merge request to a non-allowlisted repo we return a 400")
	e, _, gl, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	var event gitlab.MergeEvent
	event.ObjectAttributes.Action = "unsupported"
	When(gl.ParseAndValidate(req, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGitlabMergeRequestEvent(event)).ThenReturn(pullRequest, repo, repo, models.User{}, nil)

	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring non-actionable pull request event")
}

func TestPost_AzureDevopsPullRequestIgnoreEvent(t *testing.T) {
	t.Log("when the event is an azure devops pull request update that should not trigger workflow we ignore it")
	e, _, _, ado, _, _, _, _, _ := setup(t)

	event := `{
		"subscriptionId": "11111111-1111-1111-1111-111111111111",
		"notificationId": 1,
		"id": "22222222-2222-2222-2222-222222222222",
		"eventType": "git.pullrequest.updated",
		"publisherId": "tfs",
		"message": {
			"text": "Dev %s pull request 1 (Name in repo)"
		},
		"resource": {}}`

	cases := []struct {
		message string
	}{
		{
			"has changed the reviewer list on",
		},
		{
			"has approved",
		},
		{
			"has approved and left suggestions on",
		},
		{
			"is waiting for the author on",
		},
		{
			"rejected",
		},
		{
			"voted on",
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			payload := fmt.Sprintf(event, c.message)
			req, _ := http.NewRequest("GET", "", strings.NewReader(payload))
			req.Header.Set(azuredevopsHeader, "reqID")
			When(ado.Validate(req, user, secret)).ThenReturn([]byte(payload), nil)
			w := httptest.NewRecorder()
			e.Parser = &events.EventParser{}
			e.Post(w, req)
			ResponseContains(t, w, http.StatusOK, "pull request updated event is not a supported type")
		})
	}
}

func TestPost_AzureDevopsPullRequestDeletedCommentIgnoreEvent(t *testing.T) {
	t.Log("when the event is an azure devops pull request deleted comment event we ignore it")
	e, _, _, ado, _, _, _, _, _ := setup(t)

	payload := `{
		"subscriptionId": "11111111-1111-1111-1111-111111111111",
		"notificationId": 1,
		"id": "22222222-2222-2222-2222-222222222222",
		"eventType": "ms.vss-code.git-pullrequest-comment-event",
		"publisherId": "tfs",
		"message": {
			"text": "Dev has deleted a pull request comment"
		},
		"resource": {
			"comment": {
				"id": 1,
				"isDeleted": true,
				"commentType": "text"
			}
		}
	}`

	t.Run("Dev has deleted a pull request comment", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", strings.NewReader(payload))
		req.Header.Set(azuredevopsHeader, "reqID")
		When(ado.Validate(req, user, secret)).ThenReturn([]byte(payload), nil)
		w := httptest.NewRecorder()
		e.Parser = &events.EventParser{}
		e.Post(w, req)
		ResponseContains(t, w, http.StatusOK, "Ignoring comment event since it is linked to deleting a pull request comment")
	})
}

func TestPost_AzureDevopsPullRequestCommentWebhookTestIgnoreEvent(t *testing.T) {
	t.Log("when the event is an azure devops webhook test we ignore it")
	e, _, _, ado, _, _, _, _, _ := setup(t)

	event := `{
		"subscriptionId": "11111111-1111-1111-1111-111111111111",
		"notificationId": 1,
		"id": "22222222-2222-2222-2222-222222222222",
		"eventType": "%s",
		"publisherId": "tfs",
		"message": {
			"text": "%s"
		},
		"resource": {
			"pullRequest": {
				"repository":{
					"url": "https://fabrikam.visualstudio.com/DefaultCollection/_apis/git/repositories/4bc14d40-c903-45e2-872e-0462c7748079"
				}
			},
			"comment": {
				"content": "This is my comment."
			}
		}}`

	cases := []struct {
		eventType string
		message   string
	}{
		{
			"ms.vss-code.git-pullrequest-comment-event",
			"Jamal Hartnett has edited a pull request comment",
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			payload := fmt.Sprintf(event, c.eventType, c.message)
			req, _ := http.NewRequest("GET", "", strings.NewReader(payload))
			req.Header.Set(azuredevopsHeader, "reqID")
			When(ado.Validate(req, user, secret)).ThenReturn([]byte(payload), nil)
			w := httptest.NewRecorder()
			e.Parser = &events.EventParser{}
			e.Post(w, req)
			ResponseContains(t, w, http.StatusOK, "Ignoring Azure DevOps Test Event with Repo URL")
		})
	}
}

func TestPost_AzureDevopsPullRequestWebhookTestIgnoreEvent(t *testing.T) {
	t.Log("when the event is an azure devops webhook tests we ignore it")
	e, _, _, ado, _, _, _, _, _ := setup(t)

	event := `{
		"subscriptionId": "11111111-1111-1111-1111-111111111111",
		"notificationId": 1,
		"id": "22222222-2222-2222-2222-222222222222",
		"eventType": "%s",
		"publisherId": "tfs",
		"message": {
			"text": "%s"
		},
		"resource": {
			"repository":{
				"url": "https://fabrikam.visualstudio.com/DefaultCollection/_apis/git/repositories/4bc14d40-c903-45e2-872e-0462c7748079"
			}
		}}`

	cases := []struct {
		eventType string
		message   string
	}{
		{
			"git.pullrequest.created",
			"Jamal Hartnett created a new pull request",
		},
		{
			"git.pullrequest.updated",
			"Jamal Hartnett marked the pull request as completed",
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			payload := fmt.Sprintf(event, c.eventType, c.message)
			req, _ := http.NewRequest("GET", "", strings.NewReader(payload))
			req.Header.Set(azuredevopsHeader, "reqID")
			When(ado.Validate(req, user, secret)).ThenReturn([]byte(payload), nil)
			w := httptest.NewRecorder()
			e.Parser = &events.EventParser{}
			e.Post(w, req)
			ResponseContains(t, w, http.StatusOK, "Ignoring Azure DevOps Test Event with Repo URL")
		})
	}
}

func TestPost_AzureDevopsPullRequestCommentPassingIgnores(t *testing.T) {
	t.Log("when the event should not be ignored it should pass through all ignore statements without error")
	e, _, _, ado, _, _, _, _, cp := setup(t)

	testComment := "atlantis plan"
	repo := models.Repo{}
	cmd := events.CommentCommand{Name: command.Plan}
	When(e.Parser.ParseAzureDevopsRepo(Any[*azuredevops.GitRepository]())).ThenReturn(repo, nil)
	When(cp.Parse(testComment, models.AzureDevops)).ThenReturn(events.CommentParseResult{Command: &cmd})
	payload := fmt.Sprintf(`{
		"subscriptionId": "11111111-1111-1111-1111-111111111111",
		"notificationId": 1,
		"id": "22222222-2222-2222-2222-222222222222",
		"eventType": "ms.vss-code.git-pullrequest-comment-event",
		"publisherId": "tfs",
		"message": {
			"text": "Testing to see if comment passes ignore conditions"
		},
		"resource": {
			"comment": {
				"id": 1,
				"commentType": "text",
				"content": "%v"
			},
			"pullRequest": {
				"pullRequestId": 1,
				"repository": {}
			}
		}
	}`, testComment)

	t.Run("Testing to see if comment passes ignore conditions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", strings.NewReader(payload))
		req.Header.Set(azuredevopsHeader, "reqID")
		When(ado.Validate(req, user, secret)).ThenReturn([]byte(payload), nil)
		w := httptest.NewRecorder()
		e.Post(w, req)
		ResponseContains(t, w, http.StatusOK, "Processing...")
	})
}

func TestPost_GithubPullRequestClosedErrCleaningPull(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a closed pull request and we have an error calling CleanUpPull we return a 503")
	RegisterMockTestingT(t)
	e, v, _, _, p, _, c, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGithubPullEvent(Any[logging.SimpleLogging](), Any[*github.PullRequestEvent]())).ThenReturn(pull, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(c.CleanUpPull(Any[logging.SimpleLogging](), repo, pull)).ThenReturn(errors.New("cleanup err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: cleanup err")
}

func TestPost_GitlabMergeRequestClosedErrCleaningPull(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a closed gitlab merge request and an error occurs calling CleanUpPull we return a 500")
	e, _, gl, _, p, _, c, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	var event gitlab.MergeEvent
	event.ObjectAttributes.Action = "close"
	When(gl.ParseAndValidate(req, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGitlabMergeRequestEvent(event)).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(c.CleanUpPull(Any[logging.SimpleLogging](), repo, pullRequest)).ThenReturn(errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: err")
}

func TestPost_GithubClosedPullRequestSuccess(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a pull request and everything works we return a 200")
	e, v, _, _, p, _, c, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGithubPullEvent(Any[logging.SimpleLogging](), Any[*github.PullRequestEvent]())).ThenReturn(pull, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(c.CleanUpPull(Any[logging.SimpleLogging](), repo, pull)).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Pull request cleaned successfully")
}

func TestPost_GitlabMergeRequestSuccess(t *testing.T) {
	t.Skip("relies too much on mocks, should use real event parser")
	t.Log("when the event is a gitlab merge request and the cleanup works we return a 200")
	e, _, gl, _, p, _, _, _, _ := setup(t)
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
			pullCleaner := emocks.NewMockPullCleaner()
			allowlist, err := events.NewRepoAllowlistChecker("*")
			Ok(t, err)
			logger := logging.NewNoopLogger(t)
			scope, _, _ := metrics.NewLoggingScope(logger, "null")
			ec := &events_controllers.VCSEventsController{
				PullCleaner: pullCleaner,
				Parser: &events.EventParser{
					BitbucketUser:      "bb-user",
					BitbucketToken:     "bb-token",
					BitbucketServerURL: "https://bbserver.com",
				},
				RepoAllowlistChecker: allowlist,
				SupportedVCSHosts:    []models.VCSHostType{models.BitbucketServer},
				VCSClient:            nil,
				Logger:               logger,
				Scope:                scope,
			}

			// Build HTTP request.
			requestBytes, err := os.ReadFile(filepath.Join("testdata", "bb-server-pull-deleted-event.json"))
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
			ResponseContains(t, w, 200, "Pull request cleaned successfully")

			expRepo := models.Repo{
				FullName:          "project/repository",
				Owner:             "project",
				Name:              "repository",
				CloneURL:          "https://bb-user:bb-token@bbserver.com/scm/proj/repository.git",
				SanitizedCloneURL: "https://bb-user:<redacted>@bbserver.com/scm/proj/repository.git",
				VCSHost: models.VCSHost{
					Hostname: "bbserver.com",
					Type:     models.BitbucketServer,
				},
			}
			pullCleaner.VerifyWasCalledOnce().CleanUpPull(
				logger,
				expRepo, models.PullRequest{
					Num:        10,
					HeadCommit: "2d9fb6b9a46eafb1dcef7b008d1a429d45ca742c",
					URL:        "https://bbserver.com/projects/PROJ/repos/repository/pull-requests/10",
					HeadBranch: "decline-me",
					BaseBranch: "main",
					Author:     "admin",
					State:      models.OpenPullState,
					BaseRepo:   expRepo,
				})
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
			"github opened",
			models.Github,
			"opened",
		},
		{
			"gitlab opened",
			models.Gitlab,
			"open",
		},
		{
			"github synchronized",
			models.Github,
			"synchronize",
		},
		{
			"gitlab update",
			models.Gitlab,
			"update",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			e, v, gl, _, p, cr, _, _, _ := setup(t)
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
			case models.Github:
				req.Header.Set(githubHeader, "pull_request")
				event := fmt.Sprintf(`{"action": "%s"}`, c.Action)
				When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
				repo = models.Repo{}
				pullRequest = models.PullRequest{State: models.ClosedPullState}
				When(p.ParseGithubPullEvent(Any[logging.SimpleLogging](), Any[*github.PullRequestEvent]())).ThenReturn(pullRequest, models.OpenedPullEvent, repo, repo, models.User{}, nil)
			}

			w := httptest.NewRecorder()
			e.Post(w, req)
			ResponseContains(t, w, http.StatusOK, "Processing...")
			cr.VerifyWasCalledOnce().RunAutoplanCommand(models.Repo{}, models.Repo{}, models.PullRequest{State: models.ClosedPullState}, models.User{})
		})
	}
}

func setup(t *testing.T) (events_controllers.VCSEventsController, *mocks.MockGithubRequestValidator, *mocks.MockGitlabRequestParserValidator, *mocks.MockAzureDevopsRequestValidator, *emocks.MockEventParsing, *emocks.MockCommandRunner, *emocks.MockPullCleaner, *vcsmocks.MockClient, *emocks.MockCommentParsing) {
	RegisterMockTestingT(t)
	v := mocks.NewMockGithubRequestValidator()
	gl := mocks.NewMockGitlabRequestParserValidator()
	ado := mocks.NewMockAzureDevopsRequestValidator()
	p := emocks.NewMockEventParsing()
	cp := emocks.NewMockCommentParsing()
	cr := emocks.NewMockCommandRunner()
	c := emocks.NewMockPullCleaner()
	vcsmock := vcsmocks.NewMockClient()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	Ok(t, err)
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := events_controllers.VCSEventsController{
		ExecutableName:                  "atlantis",
		EmojiReaction:                   "eyes",
		TestingMode:                     true,
		Logger:                          logger,
		Scope:                           scope,
		ApplyDisabled:                   false,
		AzureDevopsWebhookBasicUser:     user,
		AzureDevopsWebhookBasicPassword: secret,
		AzureDevopsRequestValidator:     ado,
		GithubRequestValidator:          v,
		Parser:                          p,
		CommentParser:                   cp,
		CommandRunner:                   cr,
		PullCleaner:                     c,
		GithubWebhookSecret:             secret,
		SupportedVCSHosts:               []models.VCSHostType{models.Github, models.Gitlab, models.AzureDevops, models.Gitea},
		GiteaWebhookSecret:              secret,
		GitlabWebhookSecret:             secret,
		GitlabRequestParserValidator:    gl,
		RepoAllowlistChecker:            repoAllowlistChecker,
		VCSClient:                       vcsmock,
	}
	return e, v, gl, ado, p, cr, c, vcsmock, cp
}
