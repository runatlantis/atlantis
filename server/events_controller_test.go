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
	"net/http/httptest"
	"errors"
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
}

func TestPost_CommentNotCreated(t *testing.T) {
	t.Log("when the event is a comment but it's not a created event we ignore it")
	RegisterMockTestingT(t)
}

func TestPost_CommentInvalidComment(t *testing.T) {
	t.Log("when the event is a comment without all expected data we return a 400")
	RegisterMockTestingT(t)
}

func TestPost_CommentInvalidCommand(t *testing.T) {
	t.Log("when the event is a comment with an invalid command we ignore it")
	RegisterMockTestingT(t)
}

func TestPost_CommentSuccess(t *testing.T) {
	t.Log("when the event is comment with a valid command we call the command handler")
	RegisterMockTestingT(t)
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
