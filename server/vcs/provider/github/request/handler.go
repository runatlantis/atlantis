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

const githubHeader = "X-Github-Event"

// interfaces used in Handler

// event handler interfaces
type commentEventHandler interface {
	Handle(ctx context.Context, request *http.CloneableRequest, event event_types.Comment) error
}

type prEventHandler interface {
	Handle(ctx context.Context, request *http.CloneableRequest, event event_types.PullRequest) error
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

func (h *Matcher) Matches(request *http.CloneableRequest) bool {
	return request.GetHeader(githubHeader) != ""
}

func NewHandler(
	logger logging.SimpleLogging,
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
	logger                logging.SimpleLogging
	scope                 tally.Scope

	Matcher
}

func (h *Handler) Handle(r *http.CloneableRequest) error {
	// Validate the request against the optional webhook secret.
	payload, err := h.validator.Validate(r, h.webhookSecret)
	if err != nil {
		return &errors.RequestValidationError{Err: err}
	}

	githubReqID := r.GetHeader("X-Github-Delivery")
	logger := h.logger.With("gh-request-id", githubReqID)
	scope := h.scope.SubScope("github.event")

	event, err := h.parser.Parse(r, payload)
	if err != nil {
		return &errors.WebhookParsingError{Err: err}
	}

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		err = h.handleGithubCommentEvent(event, logger, r)
		scope = scope.SubScope(fmt.Sprintf("comment.%s", *event.Action))
	case *github.PullRequestEvent:
		err = h.handleGithubPullRequestEvent(logger, event, r)
		scope = scope.SubScope(fmt.Sprintf("pr.%s", *event.Action))
	default:
		logger.Warnf("Ignoring unsupported event: %s", githubReqID)
	}

	if err != nil {
		scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
		return err
	}

	scope.Counter(metrics.ExecutionSuccessMetric).Inc(1)
	return nil
}

func (h *Handler) handleGithubCommentEvent(event *github.IssueCommentEvent, logger logging.SimpleLogging, request *http.CloneableRequest) error {
	if event.GetAction() != "created" {
		logger.Warnf("Ignoring comment event since action was not created")
		return nil
	}

	commentEvent, err := h.commentEventConverter.Convert(event)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}

	return h.commentHandler.Handle(context.TODO(), request, commentEvent)
}

func (h *Handler) handleGithubPullRequestEvent(logger logging.SimpleLogging, event *github.PullRequestEvent, request *http.CloneableRequest) error {
	pullEvent, err := h.pullEventConverter.Convert(event)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}

	return h.prHandler.Handle(context.TODO(), request, pullEvent)
}
