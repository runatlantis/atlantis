package request

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/controllers/events/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	buffered "github.com/runatlantis/atlantis/server/http"
	httputils "github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/stretchr/testify/assert"
)

type assertingPushHandler struct {
	t             *testing.T
	expectedInput event.Push
}

func (h assertingPushHandler) Handle(ctx context.Context, input event.Push) error {
	assert.Equal(h.t, h.expectedInput, input)

	return nil
}

type assertingCheckRunHandler struct {
	t             *testing.T
	expectedInput event.CheckRun
}

func (h assertingCheckRunHandler) Handle(ctx context.Context, input event.CheckRun) error {
	assert.Equal(h.t, h.expectedInput, input)

	return nil
}

// mocks/test implementations
type assertingPRHandler struct {
	expectedInput   event.PullRequest
	expectedRequest *httputils.BufferedRequest
	t               *testing.T
}

func (h assertingPRHandler) Handle(ctx context.Context, request *httputils.BufferedRequest, input event.PullRequest) error {
	assert.Equal(h.t, h.expectedRequest, request)
	assert.Equal(h.t, h.expectedInput, input)

	return nil
}

type assertingCommentEventHandler struct {
	expectedInput   event.Comment
	expectedRequest *httputils.BufferedRequest
	t               *testing.T
}

func (h assertingCommentEventHandler) Handle(ctx context.Context, request *httputils.BufferedRequest, input event.Comment) error {
	assert.Equal(h.t, h.expectedRequest, request)
	assert.Equal(h.t, h.expectedInput, input)
	return nil
}

type testValidator struct {
	returnedPayload []byte
	returnedError   error
}

func (d testValidator) Validate(r *httputils.BufferedRequest, secret []byte) ([]byte, error) {
	return d.returnedPayload, d.returnedError
}

type testPushEventConverter struct {
	returnedPushEvent           event.Push
	returnedParsePushEventError error
}

func (c testPushEventConverter) Convert(_ *github.PushEvent) (event.Push, error) {
	return c.returnedPushEvent, c.returnedParsePushEventError
}

type testPullRequestReviewEventConverter struct {
	returnedPullRequestReviewEvent event.PullRequestReview
	error                          error
}

func (c testPullRequestReviewEventConverter) Convert(_ *github.PullRequestReviewEvent) (event.PullRequestReview, error) {
	return c.returnedPullRequestReviewEvent, c.error
}

type testCheckRunEventConverter struct {
	returnedCheckRunEvent           event.CheckRun
	returnedParseCheckRunEventError error
}

func (c testCheckRunEventConverter) Convert(_ *github.CheckRunEvent) (event.CheckRun, error) {
	return c.returnedCheckRunEvent, c.returnedParseCheckRunEventError
}

type testPullEventConverter struct {
	returnedPullEvent           event.PullRequest
	returnedParsePullEventError error
}

func (c testPullEventConverter) Convert(_ *github.PullRequestEvent) (event.PullRequest, error) {
	return c.returnedPullEvent, c.returnedParsePullEventError
}

type testCommentEventConverter struct {
	returnedCommentEvent                event.Comment
	returnedParseIssueCommentEventError error
}

func (c testCommentEventConverter) Convert(_ *github.IssueCommentEvent) (event.Comment, error) {
	return c.returnedCommentEvent, c.returnedParseIssueCommentEventError
}

type assertingPullRequestReviewHandler struct {
	t             *testing.T
	expectedInput event.PullRequestReview
}

func (h assertingPullRequestReviewHandler) Handle(_ context.Context, input event.PullRequestReview, _ *buffered.BufferedRequest) error {
	assert.Equal(h.t, h.expectedInput, input)
	return nil
}

type testParser struct {
	returnedEvent             interface{}
	returnedParseWebhookError error
}

func (e *testParser) Parse(r *httputils.BufferedRequest, payload []byte) (interface{}, error) {
	return e.returnedEvent, e.returnedParseWebhookError
}

func TestHandle_CommentEvent(t *testing.T) {
	repo := models.Repo{FullName: "lyft/nish"}
	user := models.User{Username: "nish"}
	payload := []byte("payload")
	commentBody := "###comment body###"
	secret := "secret"
	eventTimestamp := time.Now()

	log := logging.NewNoopCtxLogger(t)
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopCtxLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	internalEvent := event.Comment{
		BaseRepo:  repo,
		User:      user,
		PullNum:   1,
		Comment:   commentBody,
		VCSHost:   models.Github,
		Timestamp: eventTimestamp,
	}

	event := &github.IssueCommentEvent{
		Action: github.String("created"),
		Comment: &github.IssueComment{
			Body:      &commentBody,
			CreatedAt: &eventTimestamp,
		},
	}

	t.Run("success", func(t *testing.T) {
		subject := Handler{
			logger:        log,
			scope:         scope,
			webhookSecret: []byte(secret),
			commentHandler: assertingCommentEventHandler{
				expectedInput:   internalEvent,
				expectedRequest: request,
				t:               t,
			},
			validator: testValidator{
				returnedPayload: payload,
			},
			commentEventConverter: testCommentEventConverter{
				returnedCommentEvent: internalEvent,
			},
			parser: &testParser{
				returnedEvent: event,
			},
			Matcher: Matcher{},
		}

		err = subject.Handle(request)
		assert.NoError(t, err)
	})

	cases := []struct {
		returnedParseWebhookError           error
		returnedValidatorError              error
		returnedParseIssueCommentEventError error
		expectedErr                         error
		description                         string
	}{
		{
			description:               "webhook parsing error",
			returnedParseWebhookError: assert.AnError,
			expectedErr:               &errors.WebhookParsingError{},
		},
		{
			description:            "validator error",
			returnedValidatorError: assert.AnError,
			expectedErr:            &errors.RequestValidationError{},
		},
		{
			description:                         "event parsing error",
			returnedParseIssueCommentEventError: assert.AnError,
			expectedErr:                         &errors.EventParsingError{},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := Handler{
				logger:        log,
				scope:         scope,
				webhookSecret: []byte(secret),
				commentHandler: assertingCommentEventHandler{
					expectedInput:   internalEvent,
					expectedRequest: request,
					t:               t,
				},
				commentEventConverter: testCommentEventConverter{
					returnedCommentEvent:                internalEvent,
					returnedParseIssueCommentEventError: c.returnedParseIssueCommentEventError,
				},
				validator: testValidator{
					returnedPayload: payload,
					returnedError:   c.returnedValidatorError,
				},
				parser: &testParser{
					returnedEvent:             event,
					returnedParseWebhookError: c.returnedParseWebhookError,
				},
				Matcher: Matcher{},
			}

			err = subject.Handle(request)
			assert.IsType(t, c.expectedErr, err)
		})

	}
}

func TestHandle_PREvent(t *testing.T) {
	pull := models.PullRequest{}
	pullEventType := models.OpenedPullEvent
	user := models.User{Username: "nish"}
	payload := []byte("payload")
	secret := "secret"
	eventTimestamp := time.Now()

	log := logging.NewNoopCtxLogger(t)
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopCtxLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	internalEvent := event.PullRequest{
		Pull:      pull,
		User:      user,
		EventType: pullEventType,
		Timestamp: eventTimestamp,
	}

	event := &github.PullRequestEvent{
		PullRequest: &github.PullRequest{
			UpdatedAt: &eventTimestamp,
		},
		Action: github.String("open"),
	}

	t.Run("success", func(t *testing.T) {
		subject := Handler{
			logger:        log,
			scope:         scope,
			webhookSecret: []byte(secret),
			prHandler: assertingPRHandler{
				expectedInput:   internalEvent,
				expectedRequest: request,
				t:               t,
			},
			validator: testValidator{
				returnedPayload: payload,
			},
			pullEventConverter: testPullEventConverter{
				returnedPullEvent: internalEvent,
			},
			parser: &testParser{
				returnedEvent: event,
			},
			Matcher: Matcher{},
		}

		err = subject.Handle(request)
		assert.NoError(t, err)
	})

	cases := []struct {
		returnedParseWebhookError   error
		returnedValidatorError      error
		returnedParsePullEventError error
		expectedErr                 error
		description                 string
	}{
		{
			description:               "webhook parsing error",
			returnedParseWebhookError: assert.AnError,
			expectedErr:               &errors.WebhookParsingError{},
		},
		{
			description:            "validator error",
			returnedValidatorError: assert.AnError,
			expectedErr:            &errors.RequestValidationError{},
		},
		{
			description:                 "event parsing error",
			returnedParsePullEventError: assert.AnError,
			expectedErr:                 &errors.EventParsingError{},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := Handler{
				logger:        log,
				scope:         scope,
				webhookSecret: []byte(secret),
				prHandler: assertingPRHandler{
					expectedInput:   internalEvent,
					expectedRequest: request,
					t:               t,
				},
				validator: testValidator{
					returnedPayload: payload,
					returnedError:   c.returnedValidatorError,
				},
				pullEventConverter: testPullEventConverter{
					returnedPullEvent:           internalEvent,
					returnedParsePullEventError: c.returnedParsePullEventError,
				},
				parser: &testParser{
					returnedEvent:             event,
					returnedParseWebhookError: c.returnedParseWebhookError,
				},
				Matcher: Matcher{},
			}

			err = subject.Handle(request)
			assert.IsType(t, c.expectedErr, err)
		})

	}

}

func TestHandlePushEvent(t *testing.T) {
	payload := []byte("payload")
	secret := "secret"

	log := logging.NewNoopCtxLogger(t)
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopCtxLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	internalEvent := event.Push{
		Sha: "1234",
	}

	event := &github.PushEvent{
		Repo: &github.PushEventRepository{
			Name: github.String("repo"),
		},
	}

	t.Run("success", func(t *testing.T) {
		subject := Handler{
			logger:        log,
			scope:         scope,
			webhookSecret: []byte(secret),
			pushHandler: assertingPushHandler{
				expectedInput: internalEvent,
				t:             t,
			},
			validator: testValidator{
				returnedPayload: payload,
			},
			pushEventConverter: testPushEventConverter{
				returnedPushEvent: internalEvent,
			},
			parser: &testParser{
				returnedEvent: event,
			},
			Matcher: Matcher{},
		}

		err = subject.Handle(request)
		assert.NoError(t, err)
	})

	cases := []struct {
		returnedParseWebhookError   error
		returnedValidatorError      error
		returnedParsePushEventError error
		expectedErr                 error
		description                 string
	}{
		{
			description:               "webhook parsing error",
			returnedParseWebhookError: assert.AnError,
			expectedErr:               &errors.WebhookParsingError{},
		},
		{
			description:            "validator error",
			returnedValidatorError: assert.AnError,
			expectedErr:            &errors.RequestValidationError{},
		},
		{
			description:                 "event parsing error",
			returnedParsePushEventError: assert.AnError,
			expectedErr:                 &errors.EventParsingError{},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := Handler{
				logger:        log,
				scope:         scope,
				webhookSecret: []byte(secret),
				pushHandler: assertingPushHandler{
					expectedInput: internalEvent,
					t:             t,
				},
				validator: testValidator{
					returnedPayload: payload,
					returnedError:   c.returnedValidatorError,
				},
				pushEventConverter: testPushEventConverter{
					returnedPushEvent:           internalEvent,
					returnedParsePushEventError: c.returnedParsePushEventError,
				},
				parser: &testParser{
					returnedEvent:             event,
					returnedParseWebhookError: c.returnedParseWebhookError,
				},
				Matcher: Matcher{},
			}

			err = subject.Handle(request)
			assert.IsType(t, c.expectedErr, err)
		})

	}
}

func TestHandleCheckRunEvent(t *testing.T) {
	payload := []byte("payload")
	secret := "secret"

	log := logging.NewNoopCtxLogger(t)
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopCtxLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	internalEvent := event.CheckRun{
		Name: "some name",
	}

	event := &github.CheckRunEvent{
		Repo: &github.Repository{
			Name: github.String("repo"),
		},
	}

	t.Run("success", func(t *testing.T) {
		subject := Handler{
			logger:        log,
			scope:         scope,
			webhookSecret: []byte(secret),
			checkRunHandler: assertingCheckRunHandler{
				expectedInput: internalEvent,
				t:             t,
			},
			validator: testValidator{
				returnedPayload: payload,
			},
			checkRunEventConverter: testCheckRunEventConverter{
				returnedCheckRunEvent: internalEvent,
			},
			parser: &testParser{
				returnedEvent: event,
			},
			Matcher: Matcher{},
		}

		err = subject.Handle(request)
		assert.NoError(t, err)
	})

	cases := []struct {
		returnedParseWebhookError       error
		returnedValidatorError          error
		returnedParseCheckRunEventError error
		expectedErr                     error
		description                     string
	}{
		{
			description:               "webhook parsing error",
			returnedParseWebhookError: assert.AnError,
			expectedErr:               &errors.WebhookParsingError{},
		},
		{
			description:            "validator error",
			returnedValidatorError: assert.AnError,
			expectedErr:            &errors.RequestValidationError{},
		},
		{
			description:                     "event parsing error",
			returnedParseCheckRunEventError: assert.AnError,
			expectedErr:                     &errors.EventParsingError{},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := Handler{
				logger:        log,
				scope:         scope,
				webhookSecret: []byte(secret),
				checkRunHandler: assertingCheckRunHandler{
					expectedInput: internalEvent,
					t:             t,
				},
				validator: testValidator{
					returnedPayload: payload,
					returnedError:   c.returnedValidatorError,
				},
				checkRunEventConverter: testCheckRunEventConverter{
					returnedCheckRunEvent:           internalEvent,
					returnedParseCheckRunEventError: c.returnedParseCheckRunEventError,
				},
				parser: &testParser{
					returnedEvent:             event,
					returnedParseWebhookError: c.returnedParseWebhookError,
				},
				Matcher: Matcher{},
			}

			err = subject.Handle(request)
			assert.IsType(t, c.expectedErr, err)
		})

	}
}

func TestHandlePullRequestReviewEvent(t *testing.T) {
	payload := []byte("payload")
	secret := "secret"

	log := logging.NewNoopCtxLogger(t)
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopCtxLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	internalEvent := event.PullRequestReview{
		InstallationToken: 123,
		Repo:              models.Repo{},
		User:              models.User{Username: "username"},
		State:             event.Approved,
		Ref:               "abcd",
		Pull:              models.PullRequest{Num: 1},
	}

	event := &github.PullRequestReviewEvent{
		Installation: &github.Installation{
			ID: github.Int64(123),
		},
		Repo: &github.Repository{
			Name: github.String("repo"),
		},
		Sender: &github.User{
			Login: github.String("username"),
		},
		PullRequest: &github.PullRequest{
			Number: github.Int(1),
		},
		Action: github.String(event.Approved),
		Review: &github.PullRequestReview{
			CommitID: github.String("abcd"),
		},
	}

	t.Run("success", func(t *testing.T) {
		subject := Handler{
			logger:        log,
			scope:         scope,
			webhookSecret: []byte(secret),
			pullRequestReviewHandler: assertingPullRequestReviewHandler{
				expectedInput: internalEvent,
				t:             t,
			},
			validator: testValidator{
				returnedPayload: payload,
			},
			pullRequestReviewEventConverter: testPullRequestReviewEventConverter{
				returnedPullRequestReviewEvent: internalEvent,
			},
			parser: &testParser{
				returnedEvent: event,
			},
			Matcher: Matcher{},
		}

		err = subject.Handle(request)
		assert.NoError(t, err)
	})

	cases := []struct {
		returnedParseWebhookError  error
		returnedValidatorError     error
		returnedParsePRREventError error
		expectedErr                error
		description                string
	}{
		{
			description:               "webhook parsing error",
			returnedParseWebhookError: assert.AnError,
			expectedErr:               &errors.WebhookParsingError{},
		},
		{
			description:            "validator error",
			returnedValidatorError: assert.AnError,
			expectedErr:            &errors.RequestValidationError{},
		},
		{
			description:                "event parsing error",
			returnedParsePRREventError: assert.AnError,
			expectedErr:                &errors.EventParsingError{},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := Handler{
				logger:        log,
				scope:         scope,
				webhookSecret: []byte(secret),
				pullRequestReviewHandler: assertingPullRequestReviewHandler{
					expectedInput: internalEvent,
					t:             t,
				},
				validator: testValidator{
					returnedPayload: payload,
					returnedError:   c.returnedValidatorError,
				},
				pullRequestReviewEventConverter: testPullRequestReviewEventConverter{
					returnedPullRequestReviewEvent: internalEvent,
					error:                          c.returnedParsePRREventError,
				},
				parser: &testParser{
					returnedEvent:             event,
					returnedParseWebhookError: c.returnedParseWebhookError,
				},
				Matcher: Matcher{},
			}

			err = subject.Handle(request)
			assert.IsType(t, c.expectedErr, err)
		})

	}
}
