package server_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server"
	"github.com/hootsuite/atlantis/server/events"
	emocks "github.com/hootsuite/atlantis/server/events/mocks"
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/logging"
	"github.com/hootsuite/atlantis/server/mocks"
	. "github.com/hootsuite/atlantis/testing"
	. "github.com/petergtz/pegomock"
)

const secret = "secret"

var eventsReq *http.Request

func TestPost_InvalidSecret(t *testing.T) {
	t.Log("when the payload can't be validated a 400 is returned")
	e, v, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn(nil, errors.New("err"))
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_UnsupportedEvent(t *testing.T) {
	t.Log("when the event type is unsupported we ignore it")
	e, v, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	When(v.Validate(eventsReq, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_CommentNotCreated(t *testing.T) {
	t.Log("when the event is a comment but it's not a created event we ignore it")
	e, v, _, _, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "issue_comment")
	// comment action is deleted, not created
	event := `{"action": "deleted"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring comment event since action was not created")
}

func TestPost_CommentInvalidComment(t *testing.T) {
	t.Log("when the event is a comment without all expected data we return a 400")
	e, v, p, _, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	When(p.ExtractCommentData(AnyComment())).ThenReturn(models.Repo{}, models.User{}, models.PullRequest{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Failed parsing event")
}

func TestPost_CommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a comment with an invalid command we ignore it")
	e, v, p, _, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	When(p.ExtractCommentData(AnyComment())).ThenReturn(models.Repo{}, models.User{}, models.PullRequest{}, nil)
	When(p.DetermineCommand(AnyComment())).ThenReturn(nil, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring: err")
}

func TestPost_CommentSuccess(t *testing.T) {
	t.Log("when the event is comment with a valid command we call the command handler")
	e, v, p, cr, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	pull := models.PullRequest{}
	cmd := events.Command{}
	When(p.ExtractCommentData(AnyComment())).ThenReturn(baseRepo, user, pull, nil)
	When(p.DetermineCommand(AnyComment())).ThenReturn(&cmd, nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Processing...")

	// wait for 200ms so goroutine is called
	time.Sleep(200 * time.Millisecond)
	ctx := cr.VerifyWasCalledOnce().ExecuteCommand(AnyCommandContext()).GetCapturedArguments()
	Equals(t, baseRepo, ctx.BaseRepo)
	Equals(t, user, ctx.User)
	Equals(t, pull, ctx.Pull)
	Equals(t, cmd, *ctx.Command)
}

func TestPost_PullRequestNotClosed(t *testing.T) {
	t.Log("when the event is pull reuqest but it's not a closed event we ignore it")
	e, v, _, _, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "pull_request")
	event := `{"action": "opened"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Ignoring pull request event since action was not closed")
}

func TestPost_PullRequestInvalid(t *testing.T) {
	t.Log("when the event is pull request with invalid data we return a 400")
	e, v, p, _, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	When(p.ExtractPullData(AnyPull())).ThenReturn(models.PullRequest{}, models.Repo{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Error parsing pull data: err")
}

func TestPost_PullRequestInvalidRepo(t *testing.T) {
	t.Log("when the event is pull reuqest with invalid repo data we return a 400")
	e, v, p, _, _ := setup(t)
	eventsReq.Header.Set("X-Github-Event", "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	When(p.ExtractPullData(AnyPull())).ThenReturn(models.PullRequest{}, models.Repo{}, nil)
	When(p.ExtractRepoData(AnyRepo())).ThenReturn(models.Repo{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusBadRequest, "Error parsing repo data: err")
}

func TestPost_PullRequestErrCleaningPull(t *testing.T) {
	t.Log("when the event is a pull request and we have an error calling CleanUpPull we return a 503")
	RegisterMockTestingT(t)
	e, v, p, _, c := setup(t)
	eventsReq.Header.Set("X-Github-Event", "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{}
	When(p.ExtractPullData(AnyPull())).ThenReturn(pull, repo, nil)
	When(p.ExtractRepoData(AnyRepo())).ThenReturn(repo, nil)
	When(c.CleanUpPull(repo, pull)).ThenReturn(errors.New("cleanup err"))
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusInternalServerError, "Error cleaning pull request: cleanup err")
}

func TestPost_PullRequestSuccess(t *testing.T) {
	t.Log("when the event is a pull request and everything works we return a 200")
	e, v, p, _, c := setup(t)
	eventsReq.Header.Set("X-Github-Event", "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(eventsReq, []byte(secret))).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{}
	When(p.ExtractPullData(AnyPull())).ThenReturn(pull, repo, nil)
	When(p.ExtractRepoData(AnyRepo())).ThenReturn(repo, nil)
	When(c.CleanUpPull(repo, pull)).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, eventsReq)
	responseContains(t, w, http.StatusOK, "Pull request cleaned successfully")
}

func setup(t *testing.T) (server.EventsController, *mocks.MockGHRequestValidator, *emocks.MockEventParsing, *emocks.MockCommandRunner, *emocks.MockPullCleaner) {
	RegisterMockTestingT(t)
	eventsReq, _ = http.NewRequest("GET", "", bytes.NewBuffer(nil))
	v := mocks.NewMockGHRequestValidator()
	p := emocks.NewMockEventParsing()
	cr := emocks.NewMockCommandRunner()
	c := emocks.NewMockPullCleaner()
	e := server.EventsController{
		Logger:              logging.NewNoopLogger(),
		Validator:           v,
		Parser:              p,
		CommandRunner:       cr,
		PullCleaner:         c,
		GithubWebHookSecret: []byte(secret),
	}
	return e, v, p, cr, c
}

func AnyComment() *github.IssueCommentEvent {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&github.IssueCommentEvent{})))
	return &github.IssueCommentEvent{}
}

func AnyPull() *github.PullRequest {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&github.PullRequest{})))
	return &github.PullRequest{}
}

func AnyRepo() *github.Repository {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&github.Repository{})))
	return &github.Repository{}
}

func AnyCommandContext() *events.CommandContext {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&events.CommandContext{})))
	return &events.CommandContext{}
}
