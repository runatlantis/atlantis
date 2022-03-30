package gateway

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/aws/sns"
	"github.com/uber-go/tally"
)

const (
	GithubHeader = "X-Github-Event"
)

type HttpResponse struct {
	body string
	err  HttpError
}

type HttpError struct {
	err  error
	code int
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_event_validator.go EventValidator
type EventValidator interface {
	InstrumentedIsValid(logger logging.SimpleLogging, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) bool
}

// VCSEventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub.
type VCSEventsController struct {
	Logger        logging.SimpleLogging
	Scope         tally.Scope
	Parser        events.EventParsing
	CommentParser events.CommentParsing
	// GithubWebhookSecret is the secret added to this webhook via the GitHub
	// UI that identifies this call as coming from GitHub. If empty, no
	// request validation is done.
	GithubWebhookSecret    []byte
	GithubRequestValidator events_controllers.GithubRequestValidator
	RepoAllowlistChecker   *events.RepoAllowlistChecker
	// SilenceAllowlistErrors controls whether we write an error comment on
	// pull requests from non-allowlisted repos.
	SilenceAllowlistErrors bool
	VCSClient              vcs.Client
	SNSWriter              sns.Writer
	AutoplanValidator      EventValidator
}

// Post handles POST webhook requests.
func (g *VCSEventsController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(GithubHeader) != "" {
		g.Logger.Debugf("handling GitHub post")
		g.handleGithubPost(w, r)
		return
	}
	g.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request")
}

func (g *VCSEventsController) handleGithubPost(w http.ResponseWriter, r *http.Request) {
	// Validate the request against the optional webhook secret.
	payload, err := g.GithubRequestValidator.Validate(r, g.GithubWebhookSecret)
	if err != nil {
		g.respond(w, logging.Warn, http.StatusBadRequest, err.Error())
		return
	}
	event, _ := github.ParseWebHook(github.WebHookType(r), payload)
	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")
	logger := g.Logger.With("gh-request-id", githubReqID)
	scope := g.Scope.SubScope("github.event")

	var resp HttpResponse
	switch event := event.(type) {
	case *github.IssueCommentEvent:
		resp = g.HandleGithubCommentEvent(logger, event, r, payload)
		scope = scope.SubScope(fmt.Sprintf("comment.%s", *event.Action))
	case *github.PullRequestEvent:
		resp = g.HandleGithubPullRequestEvent(logger, event, r, payload)
		scope = scope.SubScope(fmt.Sprintf("pr.%s", *event.Action))
	default:
		resp = HttpResponse{
			body: fmt.Sprintf("Ignoring unsupported event"),
		}
	}
	if resp.err.code != 0 {
		logger.Errorf("error handling gh post code: %d err: %s", resp.err.code, resp.err.err.Error())
		scope.Counter(fmt.Sprintf("error_%d", resp.err.code)).Inc(1)
		w.WriteHeader(resp.err.code)
		fmt.Fprintln(w, resp.err.err.Error())
		return
	}
	scope.Counter(fmt.Sprintf("success_%d", http.StatusOK)).Inc(1)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, resp.body)
}

// HandleGithubCommentEvent handles comment events from GitHub where Atlantis
// commands can come from. It's exported to make testing easier.
func (g *VCSEventsController) HandleGithubCommentEvent(logger logging.SimpleLogging, event *github.IssueCommentEvent, r *http.Request, payload []byte) HttpResponse {
	if event.GetAction() != "created" {
		return HttpResponse{
			body: fmt.Sprintf("Ignoring comment event since action was not created"),
		}
	}
	baseRepo, _, pullNum, err := g.Parser.ParseGithubIssueCommentEvent(event)
	if err != nil {
		wrapped := errors.Wrap(err, "Failed parsing event")
		return HttpResponse{
			body: wrapped.Error(),
			err: HttpError{
				code: http.StatusBadRequest,
				err:  wrapped,
			},
		}
	}
	// We pass in nil for maybeHeadRepo because the head repo data isn't
	// available in the GithubIssueComment event.
	return g.handleCommentEvent(logger, baseRepo, pullNum, event.Comment.GetBody(), models.Github, r, payload)
}

func (g *VCSEventsController) handleCommentEvent(logger logging.SimpleLogging, baseRepo models.Repo, pullNum int, comment string, vcsHost models.VCSHostType, r *http.Request, payload []byte) HttpResponse {
	parseResult := g.CommentParser.Parse(comment, vcsHost)
	if parseResult.Ignore {
		truncated := comment
		truncateLen := 40
		if len(truncated) > truncateLen {
			truncated = comment[:truncateLen] + "..."
		}
		return HttpResponse{
			body: fmt.Sprintf("Ignoring non-command comment: %q", truncated),
		}
	}

	// At this point we know it's a command we're not supposed to ignore, so now
	// we check if this repo is allowed to run commands in the first place.
	if !g.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		g.commentNotAllowlisted(logger, baseRepo, pullNum)

		err := errors.New("Repo not allowlisted")
		return HttpResponse{
			body: err.Error(),
			err: HttpError{
				err:  err,
				code: http.StatusForbidden,
			},
		}
	}

	// If the command isn't valid or doesn't require processing, ex.
	// "atlantis help" then we just comment back immediately.
	// We do this here rather than earlier because we need access to the pull
	// variable to comment back on the pull request.
	if parseResult.CommentResponse != "" {
		if err := g.VCSClient.CreateComment(baseRepo, pullNum, parseResult.CommentResponse, ""); err != nil {
			logger.Errorf("unable to comment on pull request: %s", err)
		}
		return HttpResponse{
			body: "Commenting back on pull request",
		}
	}
	if err := g.SendToWorker(r, payload); err != nil {
		logger.Errorf("failed to send comment request to Atlantis worker %w", err)
		return HttpResponse{
			body: err.Error(),
			err: HttpError{
				code: http.StatusBadRequest,
				err:  err,
			},
		}
	}
	return HttpResponse{
		body: "Processing...",
	}
}

func (g *VCSEventsController) HandleGithubPullRequestEvent(logger logging.SimpleLogging, pullEvent *github.PullRequestEvent, r *http.Request, payload []byte) HttpResponse {
	pull, pullEventType, baseRepo, headRepo, user, err := g.Parser.ParseGithubPullEvent(pullEvent)
	if err != nil {
		wrapped := errors.Wrapf(err, "Error parsing pull data: %s", err)
		return HttpResponse{
			body: wrapped.Error(),
			err: HttpError{
				code: http.StatusBadRequest,
				err:  wrapped,
			},
		}
	}
	return g.handlePullRequestEvent(logger, baseRepo, headRepo, pull, user, pullEventType, r, payload)
}

func (g *VCSEventsController) handlePullRequestEvent(logger logging.SimpleLogging, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, eventType models.PullRequestEventType, request *http.Request, payload []byte) HttpResponse {
	if !g.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		// If the repo isn't allowlisted and we receive an opened pull request
		// event we comment back on the pull request that the repo isn't
		// allowlisted. This is because the user might be expecting Atlantis to
		// autoplan. For other events, we just ignore them.
		if eventType == models.OpenedPullEvent {
			g.commentNotAllowlisted(logger, baseRepo, pull.Num)
		}
		err := errors.New(fmt.Sprintf("Pull request event from non-allowlisted repo \"%s/%s\"", baseRepo.VCSHost.Hostname, baseRepo.FullName))
		return HttpResponse{
			body: err.Error(),
			err: HttpError{
				code: http.StatusForbidden,
				err:  err,
			},
		}
	}
	switch eventType {
	case models.OpenedPullEvent, models.UpdatedPullEvent:
		// If the pull request was opened or updated, we perform a pseudo-autoplan to determine if tf changes exist.
		// If it exists, then we will forward request to the worker.
		go g.handleOpenPullEvent(logger, baseRepo, headRepo, pull, user, request, payload)
		return HttpResponse{
			body: "Processing...",
		}
	case models.ClosedPullEvent:
		// If the pull request was closed, we route to worker to handle deleting locks.
		if err := g.SendToWorker(request, payload); err != nil {
			return HttpResponse{
				body: err.Error(),
				err: HttpError{
					code: http.StatusBadRequest,
					err:  err,
				},
			}
		}
	case models.OtherPullEvent:
		// Else we ignore the event.
		return HttpResponse{
			body: "Ignoring non-actionable pull request event",
		}
	}
	return HttpResponse{}
}

func (g *VCSEventsController) handleOpenPullEvent(logger logging.SimpleLogging, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, request *http.Request, payload []byte) {
	if hasTerraformChanges := g.AutoplanValidator.InstrumentedIsValid(logger, baseRepo, headRepo, pull, user); hasTerraformChanges {

		if err := g.SendToWorker(request, payload); err != nil {
			logger.Errorf("failed to send autoplan request to Atlantis worker %w", err)
		}
	}
}

func (g *VCSEventsController) SendToWorker(r *http.Request, body []byte) error {
	err := r.Body.Close()
	if err != nil {
		return errors.Wrap(err, "closing request body")
	}
	bodyBuffer := bytes.NewBuffer(body)
	copiedRequest := http.Request{
		Method:           r.Method,
		URL:              r.URL,
		Header:           r.Header,
		Body:             ioutil.NopCloser(bodyBuffer),
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Host:             r.Host,
	}
	requestBuffer := bytes.NewBuffer([]byte{})
	if err := copiedRequest.Write(requestBuffer); err != nil {
		return errors.Wrap(err, "marshalling gateway request into buffer")
	}
	if err := g.SNSWriter.Write(requestBuffer.Bytes()); err != nil {
		return errors.Wrap(err, "writing gateway request to SNS topic")
	}
	return nil
}

func (g *VCSEventsController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	g.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}

// commentNotAllowlisted comments on the pull request that the repo is not
// allowlisted unless allowlist error comments are disabled.
func (g *VCSEventsController) commentNotAllowlisted(logger logging.SimpleLogging, baseRepo models.Repo, pullNum int) {
	if g.SilenceAllowlistErrors {
		return
	}
	logger.With("repository", baseRepo.FullName, "pull-num", pullNum).Errorf("This repo is not allowlisted for Atlantis")
}
