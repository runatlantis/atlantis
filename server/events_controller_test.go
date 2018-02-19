package server_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lkysow/go-gitlab"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	emocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/mocks"
)

const githubHeader = "X-Github-Event"
const gitlabHeader = "X-Gitlab-Event"

var secret = []byte("secret")
var eventsReq *http.Request

func TestPost_NotGithubOrGitlab(t *testing.T) {
	t.Log("when the request is not for gitlab or github a 400 is returned")
	e, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Ignoring request")
}

func TestPost_UnsupportedVCSGithub(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	eventsReq.Header.Set(githubHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitHub")
}

func TestPost_UnsupportedVCSGitlab(t *testing.T) {
	t.Log("when the request is for an unsupported vcs a 400 is returned")
	e, _, _, _, _, _ := setup(t)
	e.SupportedVCSHosts = nil
	eventsReq.Header.Set(gitlabHeader, "value")
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
}

func TestPost_InvalidGithubSecret(t *testing.T) {
	t.Log("when the github payload can't be validated a 400 is returned")
	e, v, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	eventsReq.Header.Set(githubHeader, "value")
	When(v.Validate(eventsReq, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_InvalidGitlabSecret(t *testing.T) {
	t.Log("when the gitlab payload can't be validated a 400 is returned")
	e, _, gl, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	eventsReq.Header.Set(gitlabHeader, "value")
	When(gl.Validate(eventsReq, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_UnsupportedGithubEvent(t *testing.T) {
	t.Log("when the event type is an unsupported github event we ignore it")
	e, v, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	eventsReq.Header.Set(githubHeader, "value")
	When(v.Validate(eventsReq, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_UnsupportedGitlabEvent(t *testing.T) {
	t.Log("when the event type is an unsupported gitlab event we ignore it")
	e, _, gl, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	eventsReq.Header.Set(gitlabHeader, "value")
	When(gl.Validate(eventsReq, secret)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_GithubCommentNotCreated(t *testing.T) {
	t.Log("when the event is a github comment but it's not a created event we ignore it")
	e, v, _, _, _, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "issue_comment")
	// comment action is deleted, not created
	event := `{"action": "deleted"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring comment event since action was not created")
}

func TestPost_GithubInvalidComment(t *testing.T) {
	t.Log("when the event is a github comment without all expected data we return a 400")
	e, v, _, p, _, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(models.Repo{}, models.User{}, 1, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Failed parsing event")
}

func TestPost_GitlabCommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a gitlab comment with an invalid command we ignore it")
	e, _, gl, p, _, _ := setup(t)
	eventsReq.Header.Set(gitlabHeader, "value")
	When(gl.Validate(eventsReq, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	When(p.DetermineCommand("", vcs.Gitlab)).ThenReturn(nil, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring: err")
}

func TestPost_GithubCommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a github comment with an invalid command we ignore it")
	e, v, _, p, _, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(models.Repo{}, models.User{}, 1, nil)
	When(p.DetermineCommand("", vcs.Github)).ThenReturn(nil, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring: err")
}

func TestPost_GitlabCommentSuccess(t *testing.T) {
	t.Log("when the event is a gitlab comment with a valid command we call the command handler")
	e, _, gl, _, cr, _ := setup(t)
	eventsReq.Header.Set(gitlabHeader, "value")
	When(gl.Validate(eventsReq, secret)).ThenReturn(gitlab.MergeCommentEvent{}, nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Processing...")

	// wait for 200ms so goroutine is called
	time.Sleep(200 * time.Millisecond)
	cr.VerifyWasCalledOnce().ExecuteCommand(models.Repo{}, models.Repo{}, models.User{}, 0, nil, vcs.Gitlab)
}

func TestPost_GithubCommentSuccess(t *testing.T) {
	t.Log("when the event is a github comment with a valid command we call the command handler")
	e, v, _, p, cr, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	cmd := events.Command{}
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(baseRepo, user, 1, nil)
	When(p.DetermineCommand("", vcs.Github)).ThenReturn(&cmd, nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Processing...")

	// wait for 200ms so goroutine is called
	time.Sleep(200 * time.Millisecond)
	cr.VerifyWasCalledOnce().ExecuteCommand(baseRepo, baseRepo, user, 1, &cmd, vcs.Github)
}

func TestPost_GithubPullRequestNotClosed(t *testing.T) {
	t.Log("when the event is a github pull reuqest but it's not a closed event we ignore it")
	e, v, _, _, _, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "pull_request")
	event := `{"action": "opened"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring pull request event since action was not closed")
}

func TestPost_GitlabMergeRequestNotClosed(t *testing.T) {
	t.Log("when the event is a gitlab merge request but it's not a closed event we ignore it")
	e, _, gl, p, _, _ := setup(t)
	eventsReq.Header.Set(gitlabHeader, "value")
	event := gitlab.MergeEvent{}
	When(gl.Validate(eventsReq, secret)).ThenReturn(event, nil)
	When(p.ParseGitlabMergeEvent(event)).ThenReturn(models.PullRequest{State: models.Open}, models.Repo{})
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring opened merge request event")
}

func TestPost_GithubPullRequestInvalid(t *testing.T) {
	t.Log("when the event is a github pull request with invalid data we return a 400")
	e, v, _, p, _, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(models.PullRequest{}, models.Repo{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Error parsing pull data: err")
}

func TestPost_GithubPullRequestInvalidRepo(t *testing.T) {
	t.Log("when the event is a github pull request with invalid repo data we return a 400")
	e, v, _, p, _, _ := setup(t)
	eventsReq.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(models.PullRequest{}, models.Repo{}, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(models.Repo{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Error parsing repo data: err")
}

func TestPost_GithubPullRequestErrCleaningPull(t *testing.T) {
	t.Log("when the event is a pull request and we have an error calling CleanUpPull we return a 503")
	RegisterMockTestingT(t)
	e, v, _, p, _, c := setup(t)
	eventsReq.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{}
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(pull, repo, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(repo, nil)
	When(c.CleanUpPull(repo, pull, vcs.Github)).ThenReturn(errors.New("cleanup err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: cleanup err")
}

func TestPost_GitlabMergeRequestErrCleaningPull(t *testing.T) {
	t.Log("when the event is a gitlab merge request and an error occurs calling CleanUpPull we return a 503")
	e, _, gl, p, _, c := setup(t)
	eventsReq.Header.Set(gitlabHeader, "value")
	event := gitlab.MergeEvent{}
	When(gl.Validate(eventsReq, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.Closed}
	When(p.ParseGitlabMergeEvent(event)).ThenReturn(pullRequest, repo)
	When(c.CleanUpPull(repo, pullRequest, vcs.Gitlab)).ThenReturn(errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: err")
}

func TestPost_GithubPullRequestSuccess(t *testing.T) {
	t.Log("when the event is a pull request and everything works we return a 200")
	e, v, _, p, _, c := setup(t)
	eventsReq.Header.Set(githubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{}
	When(p.ParseGithubPull(matchers.AnyPtrToGithubPullRequest())).ThenReturn(pull, repo, nil)
	When(p.ParseGithubRepo(matchers.AnyPtrToGithubRepository())).ThenReturn(repo, nil)
	When(c.CleanUpPull(repo, pull, vcs.Github)).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Pull request cleaned successfully")
}

func TestPost_GitlabMergeRequestSuccess(t *testing.T) {
	t.Log("when the event is a gitlab merge request and the cleanup works we return a 200")
	e, _, gl, p, _, _ := setup(t)
	eventsReq.Header.Set(gitlabHeader, "value")
	event := gitlab.MergeEvent{}
	When(gl.Validate(eventsReq, secret)).ThenReturn(event, nil)
	repo := models.Repo{}
	pullRequest := models.PullRequest{State: models.Closed}
	When(p.ParseGitlabMergeEvent(event)).ThenReturn(pullRequest, repo)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Merge request cleaned successfully")
}

func setup(t *testing.T) (server.EventsController, *mocks.MockGithubRequestValidator, *mocks.MockGitlabRequestParser, *emocks.MockEventParsing, *emocks.MockCommandRunner, *emocks.MockPullCleaner) {
	RegisterMockTestingT(t)
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	v := mocks.NewMockGithubRequestValidator()
	gl := mocks.NewMockGitlabRequestParser()
	p := emocks.NewMockEventParsing()
	cr := emocks.NewMockCommandRunner()
	c := emocks.NewMockPullCleaner()
	e := server.EventsController{
		Logger:                 logging.NewNoopLogger(),
		GithubRequestValidator: v,
		Parser:                 p,
		CommandRunner:          cr,
		PullCleaner:            c,
		GithubWebHookSecret:    secret,
		SupportedVCSHosts:      []vcs.Host{vcs.Github, vcs.Gitlab},
		GitlabWebHookSecret:    secret,
		GitlabRequestParser:    gl,
	}
	return e, v, gl, p, cr, c
}
