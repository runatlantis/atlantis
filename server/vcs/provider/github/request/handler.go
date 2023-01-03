package request

import (
	"context"
	"fmt"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/runatlantis/atlantis/server/http"

	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/controllers/events/errors"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/uber-go/tally/v4"
)

const (
	githubHeader    = "X-Github-Event"
	requestIDHeader = "X-Github-Delivery"
)

// interfaces used in Handler

// event handler interfaces
type commentEventHandler interface {
	Handle(ctx context.Context, request *http.BufferedRequest, e event.Comment) error
}

type prEventHandler interface {
	Handle(ctx context.Context, request *http.BufferedRequest, e event.PullRequest) error
}

type pushEventHandler interface {
	Handle(ctx context.Context, e event.Push) error
}

type checkRunEventHandler interface {
	Handle(ctx context.Context, e event.CheckRun) error
}

type checkSuiteEventHandler interface {
	Handle(ctx context.Context, e event.CheckSuite) error
}

type pullRequestReviewEventHandler interface {
	Handle(ctx context.Context, e event.PullRequestReview, r *http.BufferedRequest) error
}

// converter interfaces
type pullEventConverter interface {
	Convert(event *github.PullRequestEvent) (event.PullRequest, error)
}

type commentEventConverter interface {
	Convert(event *github.IssueCommentEvent) (event.Comment, error)
}

type pushEventConverter interface {
	Convert(event *github.PushEvent) (event.Push, error)
}

type checkRunEventConverter interface {
	Convert(event *github.CheckRunEvent) (event.CheckRun, error)
}

type checkSuiteEventConverter interface {
	Convert(event *github.CheckSuiteEvent) (event.CheckSuite, error)
}

type pullRequestReviewEventConverter interface {
	Convert(event *github.PullRequestReviewEvent) (event.PullRequestReview, error)
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
	pushHandler pushEventHandler,
	pullRequestReviewEventHandler pullRequestReviewEventHandler,
	checkRunHandler checkRunEventHandler,
	checkSuiteHandler checkSuiteEventHandler,
	allowDraftPRs bool,
	repoConverter converter.RepoConverter,
	pullConverter converter.PullConverter,
	pullGetter converter.PullGetter,
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
			PullConverter: converter.PullConverter{
				RepoConverter: repoConverter,
			},
			PullGetter: pullGetter,
		},
		pushEventConverter: converter.PushEvent{
			RepoConverter: repoConverter,
		},
		pullRequestReviewEventConverter: converter.PullRequestReviewEvent{
			RepoConverter: repoConverter,
		},
		checkRunEventConverter:   converter.CheckRunEvent{},
		checkSuiteEventConverter: converter.CheckSuiteEvent{RepoConverter: repoConverter},
		pushHandler:              pushHandler,
		checkRunHandler:          checkRunHandler,
		checkSuiteHandler:        checkSuiteHandler,
		pullRequestReviewHandler: pullRequestReviewEventHandler,
		webhookSecret:            webhookSecret,
		logger:                   logger,
		scope:                    scope,
	}
}

type Handler struct {
	validator                       requestValidator
	commentHandler                  commentEventHandler
	prHandler                       prEventHandler
	pushHandler                     pushEventHandler
	checkRunHandler                 checkRunEventHandler
	checkSuiteHandler               checkSuiteEventHandler
	pullRequestReviewHandler        pullRequestReviewEventHandler
	parser                          webhookParser
	pullEventConverter              pullEventConverter
	commentEventConverter           commentEventConverter
	pushEventConverter              pushEventConverter
	checkRunEventConverter          checkRunEventConverter
	checkSuiteEventConverter        checkSuiteEventConverter
	pullRequestReviewEventConverter pullRequestReviewEventConverter
	webhookSecret                   []byte
	logger                          logging.Logger
	scope                           tally.Scope

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
		contextInternal.RequestIDKey,
		r.GetHeader(requestIDHeader),
	)

	scope := h.scope.SubScope("github.event")

	event, err := h.parser.Parse(r, payload)
	if err != nil {
		return &errors.WebhookParsingError{Err: err}
	}

	// all github app events implement this interface
	installationSource, ok := event.(githubapp.InstallationSource)

	if !ok {
		return fmt.Errorf("unable to get installation id from request %s", r.GetHeader(requestIDHeader))
	}

	installationID := githubapp.GetInstallationIDFromEvent(installationSource)

	// this will be used to create the relevant installation client
	ctx = context.WithValue(ctx, contextInternal.InstallationIDKey, installationID)

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		scope = scope.SubScope(fmt.Sprintf("comment.%s", *event.Action))
		timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
		defer timer.Stop()
		err = h.handleCommentEvent(ctx, event, r)
	case *github.PullRequestEvent:
		scope = scope.SubScope(fmt.Sprintf("pr.%s", *event.Action))
		timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
		defer timer.Stop()
		err = h.handlePullRequestEvent(ctx, event, r)
	case *github.PushEvent:
		scope = scope.SubScope("push")
		timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
		defer timer.Stop()
		err = h.handlePushEvent(ctx, event)
	case *github.CheckSuiteEvent:
		scope = scope.SubScope("checksuite")
		timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
		defer timer.Stop()
		err = h.handleCheckSuiteEvent(ctx, event)
	case *github.CheckRunEvent:
		scope = scope.SubScope("checkrun")
		timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
		defer timer.Stop()
		err = h.handleCheckRunEvent(ctx, event)
	case *github.PullRequestReviewEvent:
		scope = scope.SubScope("pullreview")
		timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
		defer timer.Stop()
		err = h.handlePullRequestReviewEvent(ctx, event, r)
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

func (h *Handler) handleCommentEvent(ctx context.Context, e *github.IssueCommentEvent, request *http.BufferedRequest) error {
	if e.GetAction() != "created" {
		h.logger.WarnContext(ctx, "Ignoring comment event since action was not created")
		return nil
	}

	commentEvent, err := h.commentEventConverter.Convert(e)
	if err != nil {
		return &errors.EventParsingError{Err: err}
	}
	ctx = context.WithValue(ctx, contextInternal.RepositoryKey, commentEvent.BaseRepo.FullName)
	ctx = context.WithValue(ctx, contextInternal.PullNumKey, commentEvent.PullNum)
	ctx = context.WithValue(ctx, contextInternal.SHAKey, commentEvent.Pull.HeadCommit)

	return h.commentHandler.Handle(ctx, request, commentEvent)
}

func (h *Handler) handlePullRequestEvent(ctx context.Context, e *github.PullRequestEvent, request *http.BufferedRequest) error {
	pullEvent, err := h.pullEventConverter.Convert(e)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}
	ctx = context.WithValue(ctx, contextInternal.RepositoryKey, pullEvent.Pull.BaseRepo.FullName)
	ctx = context.WithValue(ctx, contextInternal.PullNumKey, pullEvent.Pull.Num)
	ctx = context.WithValue(ctx, contextInternal.SHAKey, pullEvent.Pull.HeadCommit)

	return h.prHandler.Handle(ctx, request, pullEvent)
}

func (h *Handler) handlePushEvent(ctx context.Context, e *github.PushEvent) error {
	pushEvent, err := h.pushEventConverter.Convert(e)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}
	ctx = context.WithValue(ctx, contextInternal.RepositoryKey, pushEvent.Repo.FullName)
	ctx = context.WithValue(ctx, contextInternal.SHAKey, pushEvent.Sha)

	return h.pushHandler.Handle(ctx, pushEvent)
}

func (h *Handler) handleCheckRunEvent(ctx context.Context, e *github.CheckRunEvent) error {
	checkRunEvent, err := h.checkRunEventConverter.Convert(e)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}
	ctx = context.WithValue(ctx, contextInternal.RepositoryKey, checkRunEvent.Repo.FullName)

	return h.checkRunHandler.Handle(ctx, checkRunEvent)
}

func (h *Handler) handleCheckSuiteEvent(ctx context.Context, e *github.CheckSuiteEvent) error {
	checkSuiteEvent, err := h.checkSuiteEventConverter.Convert(e)

	if err != nil {
		return &errors.EventParsingError{Err: err}
	}
	ctx = context.WithValue(ctx, contextInternal.RepositoryKey, checkSuiteEvent.Repo.FullName)
	ctx = context.WithValue(ctx, contextInternal.SHAKey, checkSuiteEvent.HeadSha)
	return h.checkSuiteHandler.Handle(ctx, checkSuiteEvent)
}

func (h *Handler) handlePullRequestReviewEvent(ctx context.Context, e *github.PullRequestReviewEvent, r *http.BufferedRequest) error {
	pullRequestReviewEvent, err := h.pullRequestReviewEventConverter.Convert(e)
	if err != nil {
		return &errors.EventParsingError{Err: err}
	}
	ctx = context.WithValue(ctx, contextInternal.RepositoryKey, pullRequestReviewEvent.Repo.FullName)
	ctx = context.WithValue(ctx, contextInternal.PullNumKey, pullRequestReviewEvent.PRNum)
	ctx = context.WithValue(ctx, contextInternal.SHAKey, pullRequestReviewEvent.Ref)
	return h.pullRequestReviewHandler.Handle(ctx, pullRequestReviewEvent, r)
}
