package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	httputils "github.com/runatlantis/atlantis/server/http"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/stretchr/testify/assert"
)

type assertingPRHandler struct {
	expectedEvent   event_types.PullRequest
	expectedRequest *httputils.BufferedRequest
	t               *testing.T
}

func (h *assertingPRHandler) Handle(ctx context.Context, request *httputils.BufferedRequest, event event_types.PullRequest) error {
	assert.Equal(h.t, h.expectedRequest, request)
	assert.Equal(h.t, h.expectedEvent, event)

	return nil
}

func TestPREventHandler(t *testing.T) {
	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	event := event_types.PullRequest{
		Pull: models.PullRequest{Num: 1},
	}

	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	assert.NoError(t, err)

	openPREventHandler := &assertingPRHandler{expectedEvent: event, expectedRequest: request, t: t}

	t.Run("invokes open event type handler", func(t *testing.T) {
		prEventHandler := handlers.NewPullRequestEventWithEventTypeHandlers(
			repoAllowlistChecker,
			openPREventHandler,

			// keeping these with empty fields in order to simulate VerifyWasNeverCalled
			&assertingPRHandler{},
			&assertingPRHandler{},
		)
		err = prEventHandler.Handle(context.Background(), request, event)
		assert.NoError(t, err)
	})
}
