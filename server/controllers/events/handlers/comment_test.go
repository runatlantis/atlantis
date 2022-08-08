package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	httputils "github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/stretchr/testify/assert"
)

type testCommentParser struct {
	returnedResult events.CommentParseResult
}

type testCommentCreator struct{}

func (c *testCommentCreator) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	return nil
}

func (p *testCommentParser) Parse(comment string, vcsHost models.VCSHostType) events.CommentParseResult {
	return p.returnedResult
}

type assertingCommentHandler struct {
	expectedEvent   event_types.Comment
	expectedRequest *httputils.BufferedRequest
	expectedCommand *command.Comment
	t               *testing.T
}

func (h *assertingCommentHandler) Handle(ctx context.Context, request *httputils.BufferedRequest, event event_types.Comment, command *command.Comment) error {
	assert.Equal(h.t, h.expectedRequest, request)
	assert.Equal(h.t, h.expectedEvent, event)
	assert.Equal(h.t, h.expectedCommand, command)

	return nil
}

func TestCommentHandler(t *testing.T) {
	command := &command.Comment{
		Name: command.Apply,
	}

	rawRequest, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "",
		bytes.NewBuffer([]byte("body")),
	)
	assert.NoError(t, err)

	request, err := httputils.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	event := event_types.Comment{}

	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		commentEventHandler := handlers.NewCommentEventWithCommandHandler(
			&testCommentParser{
				returnedResult: events.CommentParseResult{
					Command: command,
				},
			},
			repoAllowlistChecker,
			&testCommentCreator{},
			&assertingCommentHandler{expectedEvent: event, expectedRequest: request, t: t, expectedCommand: command},
			logging.NewNoopCtxLogger(t),
		)

		err = commentEventHandler.Handle(context.Background(), request, event)
		assert.NoError(t, err)
	})

	t.Run("ignore non-atlantis comment", func(t *testing.T) {
		commentEventHandler := handlers.NewCommentEventWithCommandHandler(
			&testCommentParser{
				returnedResult: events.CommentParseResult{
					Command: command,
					Ignore:  true,
				},
			},
			repoAllowlistChecker,
			&testCommentCreator{},

			// by not passing in any data here, we're basically simulating VerifyNeverCalled
			&assertingCommentHandler{},
			logging.NewNoopCtxLogger(t),
		)

		err = commentEventHandler.Handle(context.Background(), request, event)
		assert.NoError(t, err)

	})

}
