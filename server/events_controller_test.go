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
//
package server_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lkysow/go-gitlab"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	emocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

const githubHeader = "X-Github-Event"
const gitlabHeader = "X-Gitlab-Event"

var secret = []byte("secret")

func TestPost_NotGithubOrGitlab(t *testing.T) {
	t.Log("when the request is not for gitlab or github a 400 is returned")
	e, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "Ignoring request")
}

func TestPost_UnsupportedVCSGithub(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitHub")
}

func TestPost_UnsupportedVCSGitlab(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
}

func TestPost_InvalidGithubSecret(t *testing.T) {
	t.Log("when the github payload can't be validated a 400 is returned")
	e, v, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "value")
	When(v.Validate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_InvalidGitlabSecret(t *testing.T) {
	t.Log("when the gitlab payload can't be validated a 400 is returned")
	e, _, gl, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.Validate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_UnsupportedGithubEvent(t *testing.T) {
	t.Log("when the event type is an unsupported github event we ignore it")
	e, v, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "value")
	When(v.Validate(req, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_UnsupportedGitlabEvent(t *testing.T) {
	t.Log("when the event type is an unsupported gitlab event we ignore it")
	e, _, gl, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.Validate(req, secret)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_GithubCommentNotCreated(t *testing.T) {
	t.Log("when the event is a github comment but it's not a created event we ignore it")
	e, v, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	// comment action is deleted, not created
	event := `{"action": "deleted"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring comment event since action was not created")
}

func TestPost_GithubInvalidComment(t *testing.T) {
	t.Log("when the event is a github comment without all expected data we return a 400")
	e, v, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(models.Repo{}, models.User{}, 1, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "Failed parsing event")
}

func TestPost_GitlabCommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a gitlab comment with an invalid command we ignore it")
	e, _, gl, _, _, _, _, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.Validate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(cp.Parse("", models.Gitlab)).ThenReturn(events.CommentParseResult{Ignore: true})
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring non-command comment: \"\"")
}

func TestPost_GithubCommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a github comment with an invalid command we ignore it")
	e, v, _, p, _, _, _, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(models.Repo{}, models.User{}, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{Ignore: true})
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring non-command comment: \"\"")
}

func TestPost_GitlabCommentNotWhitelisted(t *testing.T) {
	t.Log("when the event is a gitlab comment from a repo that isn't whitelisted we comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClientProxy()
	e := server.EventsController{
		Logger:              logging.NewNoopLogger(),
		CommentParser:       &events.CommentParser{},
		GitlabRequestParser: &server.DefaultGitlabRequestParser{},
		Parser:              &events.EventParser{},
		SupportedVCSHosts:   []models.VCSHostType{models.Gitlab},
		RepoWhitelist:       &events.RepoWhitelist{},
		VCSClient:           vcsClient,
	}
	requestJSON, err := ioutil.ReadFile(filepath.Join("testfixtures", "gitlabMergeCommentEvent_notWhitelisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set(gitlabHeader, "Note Hook")
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusForbidden, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	exp := "Repo not whitelisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
	expRepo, _ := models.NewRepo(models.Gitlab, "gitlabhq/gitlab-test", "https://example.com/gitlabhq/gitlab-test.git", "", "")
	vcsClient.VerifyWasCalledOnce().CreateComment(expRepo, 1, "```\nError: This repo is not whitelisted for Atlantis.\n```")
}

func TestPost_GithubCommentNotWhitelisted(t *testing.T) {
	t.Log("when the event is a github comment from a repo that isn't whitelisted we comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClientProxy()
	e := server.EventsController{
		Logger:                 logging.NewNoopLogger(),
		GithubRequestValidator: &server.DefaultGithubRequestValidator{},
		CommentParser:          &events.CommentParser{},
		Parser:                 &events.EventParser{},
		SupportedVCSHosts:      []models.VCSHostType{models.Github},
		RepoWhitelist:          &events.RepoWhitelist{},
		VCSClient:              vcsClient,
	}
	requestJSON, err := ioutil.ReadFile(filepath.Join("testfixtures", "githubIssueCommentEvent_notWhitelisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "issue_comment")
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusForbidden, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	exp := "Repo not whitelisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
	expRepo, _ := models.NewRepo(models.Github, "baxterthehacker/public-repo", "https://github.com/baxterthehacker/public-repo.git", "", "")
	vcsClient.VerifyWasCalledOnce().CreateComment(expRepo, 2, "```\nError: This repo is not whitelisted for Atlantis.\n```")
}

func TestPost_GitlabCommentResponse(t *testing.T) {
	// When the event is a gitlab comment that warrants a comment response we comment back.
	e, _, gl, _, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.Validate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(cp.Parse("", models.Gitlab)).ThenReturn(events.CommentParseResult{CommentResponse: "a comment"})
	w := httptest.NewRecorder()
	e.Post(w, req)
	vcsClient.VerifyWasCalledOnce().CreateComment(models.Repo{}, 0, "a comment")
	responseContains(t, w, http.StatusOK, "Commenting back on pull request")
}

func TestPost_GithubCommentResponse(t *testing.T) {
	t.Log("when the event is a github comment that warrants a comment response we comment back")
	e, v, _, p, _, _, vcsClient, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{CommentResponse: "a comment"})
	w := httptest.NewRecorder()

	e.Post(w, req)
	vcsClient.VerifyWasCalledOnce().CreateComment(baseRepo, 1, "a comment")
	responseContains(t, w, http.StatusOK, "Commenting back on pull request")
}

func TestPost_GitlabCommentSuccess(t *testing.T) {
	t.Log("when the event is a gitlab comment with a valid command we call the command handler")
	e, _, gl, _, cr, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	When(gl.Validate(req, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Processing...")

	// wait for 200ms so goroutine is called
	time.Sleep(200 * time.Millisecond)
	cr.VerifyWasCalledOnce().ExecuteCommand(models.Repo{}, models.Repo{}, models.User{}, 0, nil)
}

func TestPost_GithubCommentSuccess(t *testing.T) {
	t.Log("when the event is a github comment with a valid command we call the command handler")
	e, v, _, p, cr, _, _, cp := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	cmd := events.Command{}
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{Command: &cmd})
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Processing...")

	// wait for 200ms so goroutine is called
	time.Sleep(200 * time.Millisecond)
	cr.VerifyWasCalledOnce().ExecuteCommand(baseRepo, baseRepo, user, 1, &cmd)
}

func TestPost_GithubPullRequestNotClosed(t *testing.T) {
	t.Log("when the event is a github pull reuqest but it's not a closed event we ignore it")
	e, v, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")
	event := `{"action": "opened"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring opened pull request event")
}

func TestPost_GitlabMergeRequestNotClosed(t *testing.T) {
	t.Log("when the event is a gitlab merge request but it's not a closed event we ignore it")
	e, _, gl, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	event := gitlab.MergeEvent{}
	When(gl.Validate(req, secret)).ThenReturn(event, nil)
	When(p.ParseGitlabMergeEvent(event)).ThenReturn(models.PullRequest{State: models.Open}, models.Repo{}, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Ignoring opened pull request event")
}

func TestPost_GithubPullRequestInvalid(t *testing.T) {
	t.Log("when the event is a github pull request with invalid data we return a 400")
	e, v, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(models.PullRequest{}, models.Repo{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "Error parsing pull data: err")
}

func TestPost_GithubPullRequestInvalidRepo(t *testing.T) {
	t.Log("when the event is a github pull request with invalid repo data we return a 400")
	e, v, _, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(models.PullRequest{}, models.Repo{}, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(models.Repo{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusBadRequest, "Error parsing repo data: err")
}

func TestPost_GithubPullRequestNotWhitelisted(t *testing.T) {
	t.Log("when the event is a github pull request to a non-whitelisted repo we return a 400")
	e, v, _, p, _, _, _, _ := setup(t)
	e.RepoWhitelist = &events.RepoWhitelist{Whitelist: "github.com/nevermatch"}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(models.PullRequest{}, models.Repo{}, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(models.Repo{}, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusForbidden, "Ignoring pull request event from non-whitelisted repo")
}

func TestPost_GithubPullRequestErrCleaningPull(t *testing.T) {
	t.Log("when the event is a pull request and we have an error calling CleanUpPull we return a 503")
	RegisterMockTestingT(t)
	e, v, _, p, _, c, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.Closed}
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(pull, repo, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(repo, nil)
	When(c.CleanUpPull(repo, pull)).ThenReturn(errors.New("cleanup err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: cleanup err")
}

func TestPost_GitlabMergeRequestErrCleaningPull(t *testing.T) {
	t.Log("when the event is a gitlab merge request and an error occurs calling CleanUpPull we return a 503")
	e, _, gl, p, _, c, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	event := gitlab.MergeEvent{}
	When(gl.Validate(req, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.Closed}
	When(p.ParseGitlabMergeEvent(event)).ThenReturn(pullRequest, repo, nil)
	When(c.CleanUpPull(repo, pullRequest)).ThenReturn(errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: err")
}

func TestPost_GithubPullRequestSuccess(t *testing.T) {
	t.Log("when the event is a pull request and everything works we return a 200")
	e, v, _, p, _, c, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.Closed}
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(pull, repo, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(repo, nil)
	When(c.CleanUpPull(repo, pull)).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Pull request cleaned successfully")
}

func TestPost_GitlabMergeRequestSuccess(t *testing.T) {
	t.Log("when the event is a gitlab merge request and the cleanup works we return a 200")
	e, _, gl, p, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gitlabHeader, "value")
	event := gitlab.MergeEvent{}
	When(gl.Validate(req, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.Closed}
	When(p.ParseGitlabMergeEvent(event)).ThenReturn(pullRequest, repo, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	responseContains(t, w, http.StatusOK, "Pull request cleaned successfully")
}

func setup(t *testing.T) (server.EventsController, *mocks.MockGithubRequestValidator, *mocks.MockGitlabRequestParser, *emocks.MockEventParsing, *emocks.MockCommandRunner, *emocks.MockPullCleaner, *vcsmocks.MockClientProxy, *emocks.MockCommentParsing) {
	RegisterMockTestingT(t)
	v := mocks.NewMockGithubRequestValidator()
	gl := mocks.NewMockGitlabRequestParser()
	p := emocks.NewMockEventParsing()
	cp := emocks.NewMockCommentParsing()
	cr := emocks.NewMockCommandRunner()
	c := emocks.NewMockPullCleaner()
	vcsmock := vcsmocks.NewMockClientProxy()
	e := server.EventsController{
		Logger:                 logging.NewNoopLogger(),
		GithubRequestValidator: v,
		Parser:                 p,
		CommentParser:          cp,
		CommandRunner:          cr,
		PullCleaner:            c,
		GithubWebHookSecret:    secret,
		SupportedVCSHosts:      []models.VCSHostType{models.Github, models.Gitlab},
		GitlabWebHookSecret:    secret,
		GitlabRequestParser:    gl,
		RepoWhitelist: &events.RepoWhitelist{
			Whitelist: "*",
		},
		VCSClient: vcsmock,
	}
	return e, v, gl, p, cr, c, vcsmock, cp
}
