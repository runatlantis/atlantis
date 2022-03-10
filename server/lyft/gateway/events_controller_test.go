package gateway_test

import (
	"bytes"
	"errors"
	. "github.com/petergtz/pegomock"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/events/mocks"
	"github.com/runatlantis/atlantis/server/events"
	emocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	sns_mocks "github.com/runatlantis/atlantis/server/lyft/aws/sns/mocks"
	sns_matchers "github.com/runatlantis/atlantis/server/lyft/aws/sns/mocks/matchers"
	"github.com/runatlantis/atlantis/server/lyft/gateway"
	ev_mocks "github.com/runatlantis/atlantis/server/lyft/gateway/mocks"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var secret = []byte("secret")

func TestPost_NotGithub(t *testing.T) {
	t.Log("when the request is not for github a 400 is returned")
	e, _, _, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Ignoring request")
}

func TestPost_InvalidGithubSecret(t *testing.T) {
	t.Log("when the github payload can't be validated a 400 is returned")
	e, v, _, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "value")
	When(v.Validate(req, secret)).ThenReturn(nil, errors.New("err"))
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "err")
}

func TestPost_UnsupportedGithubEvent(t *testing.T) {
	t.Log("when the event type is an unsupported github event we ignore it")
	e, v, _, _, _, _, _, _, _, _ := setup(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "value")
	When(v.Validate(req, nil)).ThenReturn([]byte(`{"not an event": ""}`), nil)
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring unsupported event")
}

func TestPost_GithubCommentNotCreated(t *testing.T) {
	t.Log("when the event is a github comment but it's not a created event we ignore it")
	e, v, _, _, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	// comment action is deleted, not created
	event := `{"action": "deleted"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusOK, "Ignoring comment event since action was not created")
}

func TestPost_GithubInvalidComment(t *testing.T) {
	t.Log("when the event is a github comment without all expected data we return a 400")
	e, v, _, p, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(models.Repo{}, models.User{}, 1, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Failed parsing event")
}

func TestPost_GithubCommentNotAllowlisted(t *testing.T) {
	t.Log("when the event is a github comment from a repo that isn't allowlisted we comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := gateway.VCSEventsController{
		Logger:                 logger,
		Scope:                  scope,
		GithubRequestValidator: &events_controllers.DefaultGithubRequestValidator{},
		CommentParser:          &events.CommentParser{},
		Parser:                 &events.EventParser{},
		RepoAllowlistChecker:   &events.RepoAllowlistChecker{},
		VCSClient:              vcsClient,
	}
	requestJSON, err := ioutil.ReadFile(filepath.Join("../../controllers/events/testfixtures", "githubIssueCommentEvent_notAllowlisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusForbidden, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	exp := "Repo not allowlisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
}

func TestPost_GithubCommentNotAllowlistedWithSilenceErrors(t *testing.T) {
	t.Log("when the event is a github comment from a repo that isn't allowlisted and we are silencing errors, do not comment with an error")
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := gateway.VCSEventsController{
		Logger:                 logger,
		Scope:                  scope,
		GithubRequestValidator: &events_controllers.DefaultGithubRequestValidator{},
		CommentParser:          &events.CommentParser{},
		Parser:                 &events.EventParser{},
		RepoAllowlistChecker:   &events.RepoAllowlistChecker{},
		VCSClient:              vcsClient,
		SilenceAllowlistErrors: true,
	}
	requestJSON, err := ioutil.ReadFile(filepath.Join("../../controllers/events/testfixtures", "githubIssueCommentEvent_notAllowlisted.json"))
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	w := httptest.NewRecorder()
	e.Post(w, req)

	Equals(t, http.StatusForbidden, w.Result().StatusCode)
	body, _ := ioutil.ReadAll(w.Result().Body)
	exp := "Repo not allowlisted"
	Assert(t, strings.Contains(string(body), exp), "exp %q to be contained in %q", exp, string(body))
}

func TestPost_GithubCommentResponse(t *testing.T) {
	t.Log("when the event is a github comment that warrants a comment response we comment back")
	e, v, _, p, _, _, vcsClient, cp, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{CommentResponse: "a comment"})
	w := httptest.NewRecorder()

	e.Post(w, req)
	vcsClient.VerifyWasCalledOnce().CreateComment(baseRepo, 1, "a comment", "")
	ResponseContains(t, w, http.StatusOK, "Commenting back on pull request")
}

func TestPost_GithubPullRequestInvalid(t *testing.T) {
	t.Log("when the event is a github pull request with invalid data we return a 400")
	e, v, _, p, _, _, _, _, _, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	When(p.ParseGithubPullEvent(matchers.AnyPtrToGithubPullRequestEvent())).ThenReturn(models.PullRequest{}, models.OpenedPullEvent, models.Repo{}, models.Repo{}, models.User{}, errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusBadRequest, "Error parsing pull data: err")
}

func TestPost_GithubPullRequestNotAllowlisted(t *testing.T) {
	t.Log("when the event is a github pull request to a non-allowlisted repo we return a 400")
	e, v, _, _, _, _, _, _, _, _ := setup(t)
	var err error
	e.RepoAllowlistChecker, err = events.NewRepoAllowlistChecker("github.com/nevermatch")
	Ok(t, err)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "pull_request")

	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ResponseContains(t, w, http.StatusForbidden, "Pull request event from non-allowlisted repo")
}

func setup(t *testing.T) (gateway.VCSEventsController, *mocks.MockGithubRequestValidator, *mocks.MockGitlabRequestParserValidator, *emocks.MockEventParsing, *emocks.MockCommandRunner, *emocks.MockPullCleaner, *vcsmocks.MockClient, *emocks.MockCommentParsing, *sns_mocks.MockWriter, *ev_mocks.MockEventValidator) {
	RegisterMockTestingT(t)
	v := mocks.NewMockGithubRequestValidator()
	gl := mocks.NewMockGitlabRequestParserValidator()
	p := emocks.NewMockEventParsing()
	cp := emocks.NewMockCommentParsing()
	cr := emocks.NewMockCommandRunner()
	c := emocks.NewMockPullCleaner()
	vcsmock := vcsmocks.NewMockClient()
	snsMock := sns_mocks.NewMockWriter()
	evMock := ev_mocks.NewMockEventValidator()
	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	Ok(t, err)
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "null")
	e := gateway.VCSEventsController{
		Logger:                 logger,
		Scope:                  scope,
		GithubRequestValidator: v,
		Parser:                 p,
		CommentParser:          cp,
		GithubWebhookSecret:    secret,
		RepoAllowlistChecker:   repoAllowlistChecker,
		VCSClient:              vcsmock,
		SNSWriter:              snsMock,
		AutoplanValidator:      evMock,
	}
	return e, v, gl, p, cr, c, vcsmock, cp, snsMock, evMock
}

func TestPost_GithubCommentSuccess(t *testing.T) {
	t.Log("when the event is a github comment with a valid command we send request to SNS")
	e, v, _, p, _, _, _, cp, sns, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	cmd := events.CommentCommand{}
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{Command: &cmd})
	When(sns.Write(sns_matchers.AnySliceOfByte())).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	sns.VerifyWasCalledOnce().Write(sns_matchers.AnySliceOfByte())
	ResponseContains(t, w, http.StatusOK, "Processing...")
}

func TestPost_GithubCommentFailure(t *testing.T) {
	t.Log("when the event is a github comment with a valid command we send request to SNS (failure)")
	e, v, _, p, _, _, _, cp, sns, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "issue_comment")
	event := `{"action": "created"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	baseRepo := models.Repo{}
	user := models.User{}
	cmd := events.CommentCommand{}
	When(p.ParseGithubIssueCommentEvent(matchers.AnyPtrToGithubIssueCommentEvent())).ThenReturn(baseRepo, user, 1, nil)
	When(cp.Parse("", models.Github)).ThenReturn(events.CommentParseResult{Command: &cmd})
	When(sns.Write(sns_matchers.AnySliceOfByte())).ThenReturn(errors.New("err"))
	w := httptest.NewRecorder()
	e.Post(w, req)
	sns.VerifyWasCalledOnce().Write(sns_matchers.AnySliceOfByte())
	ResponseContains(t, w, http.StatusBadRequest, "writing gateway request to SNS topic: err")
}

func TestPost_GithubIgnorePR(t *testing.T) {
	t.Log("when the event is not PR open/update/close, we ignore it")
	e, v, _, p, _, _, _, _, sns, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "pull_request")
	event := `{"action": "other"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGithubPullEvent(matchers.AnyPtrToGithubPullRequestEvent())).ThenReturn(pull, models.OtherPullEvent, repo, repo, models.User{}, nil)
	When(sns.Write(sns_matchers.AnySliceOfByte())).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	sns.VerifyWasCalled(Never()).Write(sns_matchers.AnySliceOfByte())
	ResponseContains(t, w, http.StatusOK, "Ignoring non-actionable pull request event")
}

func TestPost_GithubClosePR(t *testing.T) {
	t.Log("when the event is a PR closed, we send request to SNS")
	e, v, _, p, _, _, _, _, sns, _ := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "pull_request")
	event := `{"action": "closed"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.ClosedPullState}
	When(p.ParseGithubPullEvent(matchers.AnyPtrToGithubPullRequestEvent())).ThenReturn(pull, models.ClosedPullEvent, repo, repo, models.User{}, nil)
	When(sns.Write(sns_matchers.AnySliceOfByte())).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	sns.VerifyWasCalledOnce().Write(sns_matchers.AnySliceOfByte())
	ResponseContains(t, w, http.StatusOK, "")
}

func TestPost_GithubOpenPR_WithTerraformChanges(t *testing.T) {
	t.Log("when the event is a PR opened, we send request to SNS if there are tf changes")
	e, v, _, p, _, _, _, _, sns, ev := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "pull_request")
	event := `{"action": "open"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.OpenPullState}
	When(p.ParseGithubPullEvent(matchers.AnyPtrToGithubPullRequestEvent())).ThenReturn(pull, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(ev.InstrumentedIsValid(matchers.AnyModelsRepo(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsUser())).ThenReturn(true)
	When(sns.Write(sns_matchers.AnySliceOfByte())).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ev.VerifyWasCalledEventually(Once(), 500*time.Millisecond).InstrumentedIsValid(matchers.AnyModelsRepo(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsUser())
	sns.VerifyWasCalledEventually(Once(), 500*time.Millisecond).Write(sns_matchers.AnySliceOfByte())
	ResponseContains(t, w, http.StatusOK, "")
}

func TestPost_GithubOpenPR_NoTerraformChanges(t *testing.T) {
	t.Log("when the event is a PR opened, we don't send a request to SNS if there are no tf changes")
	e, v, _, p, _, _, _, _, sns, ev := setup(t)
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	req.Header.Set(gateway.GithubHeader, "pull_request")
	event := `{"action": "open"}`
	When(v.Validate(req, secret)).ThenReturn([]byte(event), nil)
	repo := models.Repo{}
	pull := models.PullRequest{State: models.OpenPullState}
	When(p.ParseGithubPullEvent(matchers.AnyPtrToGithubPullRequestEvent())).ThenReturn(pull, models.OpenedPullEvent, repo, repo, models.User{}, nil)
	When(ev.InstrumentedIsValid(matchers.AnyModelsRepo(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsUser())).ThenReturn(false)
	When(sns.Write(sns_matchers.AnySliceOfByte())).ThenReturn(nil)
	w := httptest.NewRecorder()
	e.Post(w, req)
	ev.VerifyWasCalledEventually(Once(), 500*time.Millisecond).InstrumentedIsValid(matchers.AnyModelsRepo(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyModelsUser())
	sns.VerifyWasCalledEventually(Never(), 500*time.Millisecond).Write(sns_matchers.AnySliceOfByte())
	ResponseContains(t, w, http.StatusOK, "")
}
