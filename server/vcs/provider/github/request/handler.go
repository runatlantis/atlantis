package request

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/http"

	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	event_types "github.com/runatlantis/atlantis/server/vcs/types/event"

	"github.com/google/go-github/v31/github"
	"github.com/runatlantis/atlantis/server/controllers/events/errors"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/uber-go/tally"
)

const (
	githubHeader    = "X-Github-Event"
	requestIDHeader = "X-Github-Delivery"
)

// interfaces used in Handler

// event handler interfaces
type commentEventHandler interface {
	Handle(ctx context.Context, request *http.BufferedRequest, event event_types.Comment) error
}

type prEventHandler interface {
	Handle(ctx context.Context, request *http.BufferedRequest, event event_types.PullRequest) error
}

// converter interfaces
type pullEventConverter interface {
	Convert(event *github.PullRequestEvent) (event_types.PullRequest, error)
}

type commentEventConverter interface {
	Convert(event *github.IssueCommentEvent) (event_types.Comment, error)
}

// Matcher matches a provided request against some condition
type Matcher struct{}

func (h *Matcher) Matches(request *http.BufferedRequest) bool {
	return request.GetHeader(githubHeader) != ""
}

func NewHandler(
	logger logging.Logger,
	scope tally.Scope,
	webhookSecret []byte,
	commentHandler *handlers.CommentEvent,
	prHandler *handlers.PullRequestEvent,
	allowDraftPRs bool,
	repoConverter converter.RepoConverter,
	pullConverter converter.PullConverter,
) *Handler {
	return &Handler{
		Matcher:        Matcher{},
		validator:      validator{},
		commentHandler: commentHandler,
		prHandler:      prHandler,
		parser:         &parser{},
		pullEventConverter: converter.PullEventConverter{
			PullConverter: pullConverter,
			AllowDraftPRs: allowDraftPRs,
		},
		commentEventConverter: converter.CommentEventConverter{
			RepoConverter: repoConverter,
		},
		webhookSecret: webhookSecret,
		logger:        logger,
		scope:         scope,
	}
}

type Handler struct {
	validator             requestValidator
	commentHandler        commentEventHandler
	prHandler             prEventHandler
	parser                webhookParser
	pullEventConverter    pullEventConverter
	commentEventConverter commentEventConverter
	webhookSecret         []byte
	logger                logging.Logger
	scope                 tally.Scope

	Matcher
}

func (h *Handler) Handle(r *http.BufferedRequest) error {
	// Validate the request against the optional webhook secret.
	payload, err := h.validator.Validate(r, h.webhookSecret)
	if err != nil {
		return &errors.RequestValidationError{Err: err}
	}

	ctx := context.WithValue(
		r.GetRequest().Context(),
		logging.RequestIDKey,
		r.GetHeader(requestIDHeader),
	)

	scope := h.scope.SubScope("github.event")

	event, err := h.parser.Parse(r, payload)
	if err != nil {
		return &errors.WebhookParsingError{Err: err}
	}

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		err = h.handleGithubCommentEvent(ctx, event, r)
		scope = scope.SubScope(fmt.Sprintf("comment.%s", *event.Action))
	case *github.PullRequestEvent:
		err = h.handleGithubPullRequestEvent(ctx, event, r)
		scope = scope.SubScope(fmt.Sprintf("pr.%s", *event.Action))
	default:
		h.logger.WarnContext(ctx, "Ignoring unsupported event")
	}

	if err != nil {
		scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
		return err
	}

	scope.Counter(metrics.ExecutionSuccessMetric).Inc(1)
	return nil
}

func (h *Handler) handleGithubCommentEvent(ctx context.Context, event *github.IssueCommentEvent, request *http.BufferedRequest) error {
	if event.GetAction() != "created" {
		h.logger.WarnContext(ctx, "Ignoring comment event since action was not created")
		return nil
	}

	commentEvent, err := h.commentEventConverter.Convert(event)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}

	return h.commentHandler.Handle(ctx, request, commentEvent)
}

func (h *Handler) handleGithubPullRequestEvent(ctx context.Context, event *github.PullRequestEvent, request *http.BufferedRequest) error {
	pullEvent, err := h.pullEventConverter.Convert(event)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}

	return h.prHandler.Handle(ctx, request, pullEvent)
}
