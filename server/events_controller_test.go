package server_test

import (
	"testing"
	. "github.com/hootsuite/atlantis/testing_util"
	. "github.com/petergtz/pegomock"
	"github.com/hootsuite/atlantis/server"
	"github.com/hootsuite/atlantis/server/logging"
	"net/http"
	"bytes"
	"github.com/hootsuite/atlantis/server/mocks"
	emocks "github.com/hootsuite/atlantis/server/events/mocks"
	"net/http/httptest"
	"errors"
	"strings"
	"io/ioutil"
	"reflect"
	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/events"
	"time"
)

func TestPost_InvalidSecret(t *testing.T) {
	t.Log("when the payload can't be validated a 400 is returned")
	RegisterMockTestingT(t)
	v := mocks.NewMockGHRequestValidator()
	secret := []byte("secret")
	e := server.EventsController{
		Logger:              logging.NewNoopLogger(),
		GithubWebHookSecret: secret,
		Validator:           v,
	}
	req, err := http.NewRequest("GET", "http://localhost/event", bytes.NewBuffer(nil))
	Ok(t, err)
	w := httptest.NewRecorder()
	When(v.Validate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestPost_UnsupportedEvent(t *testing.T) {
	t.Log("when the event type is unsupported we ignore it")
	RegisterMockTestingT(t)
	v := mocks.NewMockGHRequestValidator()
	e := server.EventsController{
		Logger:    logging.NewNoopLogger(),
		Validator: v,
	}
	req, err := http.NewRequest("GET", "http://localhost/event", bytes.NewBuffer(nil))
	Ok(t, err)
	w := httptest.NewRecorder()
	When(v.Validate(req, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)

	Equals(t, http.StatusOK, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	Assert(t, strings.Contains(string(body), "Ignoring unsupported event"), "Response body was: %s", string(body))
}

func TestPost_CommentNotCreated(t *testing.T) {
	t.Log("when the event is a comment but it's not a created event we ignore it")
	RegisterMockTestingT(t)
	v := mocks.NewMockGHRequestValidator()
	e := server.EventsController{
		Logger:    logging.NewNoopLogger(),
		Validator: v,
	}
	req, err := http.NewRequest("GET", "http://localhost/event", bytes.NewBuffer(nil))
	req.Header.Set("X-Github-Event", "issue_comment")
	Ok(t, err)

	// comment action is deleted, not created
	event := `{"action": "deleted"}`
	When(v.Validate(req, nil)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusOK, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	Assert(t, strings.Contains(string(body), "Ignoring comment event since action was not created"), "Response body was: %s", string(body))
}

func TestPost_CommentInvalidComment(t *testing.T) {
	t.Log("when the event is a comment without all expected data we return a 400")
	RegisterMockTestingT(t)
	v := mocks.NewMockGHRequestValidator()
	p := emocks.NewMockEventParsing()
	e := server.EventsController{
		Logger:    logging.NewNoopLogger(),
		Validator: v,
		Parser: p,
	}
	req, err := http.NewRequest("GET", "http://localhost/event", bytes.NewBuffer(nil))
	req.Header.Set("X-Github-Event", "issue_comment")
	Ok(t, err)

	event := `{"action": "created"}`
	When(v.Validate(req, nil)).ThenReturn([]byte(event), nil)
	When(p.ExtractCommentData(AnyComment())).ThenReturn(models.Repo{}, models.User{}, models.PullRequest{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	Assert(t, strings.Contains(string(body), "Failed parsing event"), "Response body was: %s", string(body))
}

func TestPost_CommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a comment with an invalid command we ignore it")
	RegisterMockTestingT(t)
	v := mocks.NewMockGHRequestValidator()
	p := emocks.NewMockEventParsing()
	e := server.EventsController{
		Logger:    logging.NewNoopLogger(),
		Validator: v,
		Parser: p,
	}
	req, err := http.NewRequest("GET", "http://localhost/event", bytes.NewBuffer(nil))
	req.Header.Set("X-Github-Event", "issue_comment")
	Ok(t, err)

	event := `{"action": "created"}`
	When(v.Validate(req, nil)).ThenReturn([]byte(event), nil)
	When(p.ExtractCommentData(AnyComment())).ThenReturn(models.Repo{}, models.User{}, models.PullRequest{}, nil)
	When(p.DetermineCommand(AnyComment())).ThenReturn(nil, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusOK, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	Assert(t, strings.Contains(string(body), "Ignoring: err"), "Response body was: %s", string(body))
}

func TestPost_CommentSuccess(t *testing.T) {
	t.Log("when the event is comment with a valid command we call the command handler")
	RegisterMockTestingT(t)
	v := mocks.NewMockGHRequestValidator()
	p := emocks.NewMockEventParsing()
	cr := emocks.NewMockCommandRunner()
	e := server.EventsController{
		Logger:        logging.NewNoopLogger(),
		Validator:     v,
		Parser:        p,
		CommandRunner: cr,
	}
	req, err := http.NewRequest("GET", "http://localhost/event", bytes.NewBuffer(nil))
	req.Header.Set("X-Github-Event", "issue_comment")
	Ok(t, err)

	event := `{"action": "created"}`
	When(v.Validate(req, nil)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	pull := models.PullRequest{}
	cmd := events.Command{}
	When(p.ExtractCommentData(AnyComment())).ThenReturn(baseRepo, user, pull, nil)
	When(p.DetermineCommand(AnyComment())).ThenReturn(&cmd, nil)
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusOK, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	Equals(t, "Processing...\n", string(body))

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
	RegisterMockTestingT(t)
}

func TestPost_PullRequestInvalid(t *testing.T) {
	t.Log("when the event is pull reuqest with invalid data we return a 400")
	RegisterMockTestingT(t)
}

func TestPost_PullRequestInvalidRepo(t *testing.T) {
	t.Log("when the event is pull reuqest with invalid repo data we return a 400")
	RegisterMockTestingT(t)
}

func TestPost_PullRequestErrCleaningPull(t *testing.T) {
	t.Log("when the event is a pull request and we have an error calling CleanUpPull we return a 503")
	RegisterMockTestingT(t)
}

func TestPost_PullRequestSuccess(t *testing.T) {
	t.Log("when the event is a pull request and everything works we return a 200")
	RegisterMockTestingT(t)
}

func AnyComment() *github.IssueCommentEvent {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&github.IssueCommentEvent{})))
	return &github.IssueCommentEvent{}
}

func AnyCommandContext() *events.CommandContext {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&events.CommandContext{})))
	return &events.CommandContext{}
}
