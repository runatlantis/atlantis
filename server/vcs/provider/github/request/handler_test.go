package request

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/runatlantis/atlantis/server/controllers/events/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	httputils "github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/vcs/types/event"
	"github.com/stretchr/testify/assert"
)

// mocks/test implementations
type assertingPRHandler struct {
	expectedInput   event.PullRequest
	expectedRequest *httputils.CloneableRequest
	t               *testing.T
}

func (h assertingPRHandler) Handle(ctx context.Context, request *httputils.CloneableRequest, input event.PullRequest) error {
	assert.Equal(h.t, h.expectedRequest, request)
	assert.Equal(h.t, h.expectedInput, input)

	return nil
}

type assertingCommentEventHandler struct {
	expectedInput   event.Comment
	expectedRequest *httputils.CloneableRequest
	t               *testing.T
}

func (h assertingCommentEventHandler) Handle(ctx context.Context, request *httputils.CloneableRequest, input event.Comment) error {
	assert.Equal(h.t, h.expectedRequest, request)
	assert.Equal(h.t, h.expectedInput, input)
	return nil
}

type testValidator struct {
	returnedPayload []byte
	returnedError   error
}

func (d testValidator) Validate(r *httputils.CloneableRequest, secret []byte) ([]byte, error) {
	return d.returnedPayload, d.returnedError
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

type testParser struct {
	returnedEvent             interface{}
	returnedParseWebhookError error
}

func (e *testParser) Parse(r *httputils.CloneableRequest, payload []byte) (interface{}, error) {
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
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewCloneableRequest(rawRequest)
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
	scope, _, err := metrics.NewLoggingScope(logging.NewNoopLogger(t), "atlantis")
	assert.NoError(t, err)

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewCloneableRequest(rawRequest)
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
