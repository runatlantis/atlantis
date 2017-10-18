package server

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server/events"
	"github.com/hootsuite/atlantis/server/logging"
)

type EventsController struct {
	CommandRunner      events.CommandRunner
	PullClosedExecutor *events.PullClosedExecutor
	Logger             *logging.SimpleLogger
	Parser             events.EventParsing
	// GithubWebHookSecret is the secret added to this webhook via the GitHub
	// UI that identifies this call as coming from GitHub. If empty, no
	// request validation is done.
	GithubWebHookSecret []byte
	Validator           GHRequestValidator
}

func (e *EventsController) Post(w http.ResponseWriter, r *http.Request) {
	// Validate the request against the optional webhook secret.
	payload, err := e.Validator.Validate(r, e.GithubWebHookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, err.Error())
		return
	}

	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")
	event, _ := github.ParseWebHook(github.WebHookType(r), payload)
	switch event := event.(type) {
	case *github.IssueCommentEvent:
		e.HandleCommentEvent(w, event, githubReqID)
	case *github.PullRequestEvent:
		e.HandlePullRequestEvent(w, event, githubReqID)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event %s", githubReqID)
	}
}

func (e *EventsController) HandleCommentEvent(w http.ResponseWriter, event *github.IssueCommentEvent, githubReqID string) {
	if event.GetAction() != "created" {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since action was not created %s", githubReqID)
		return
	}

	baseRepo, user, pull, err := e.Parser.ExtractCommentData(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Failed parsing event: %v %s", err, githubReqID)
		return
	}
	ctx := &events.CommandContext{}
	ctx.BaseRepo = baseRepo
	ctx.User = user
	ctx.Pull = pull

	command, err := e.Parser.DetermineCommand(event)
	if err != nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring: %s %s", err, githubReqID)
		return
	}
	ctx.Command = command

	// Respond with success and then actually execute the command asynchronously.
	// We use a goroutine so that this function returns and the connection is
	// closed.
	fmt.Fprintln(w, "Processing...")
	go e.CommandRunner.ExecuteCommand(ctx)
}

// HandlePullRequestEvent will delete any locks associated with the pull request
func (e *EventsController) HandlePullRequestEvent(w http.ResponseWriter, pullEvent *github.PullRequestEvent, githubReqID string) {
	if pullEvent.GetAction() != "closed" {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring pull request event since action was not closed %s", githubReqID)
		return
	}
	pull, _, err := e.Parser.ExtractPullData(pullEvent.PullRequest)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s", err)
		return
	}
	repo, err := e.Parser.ExtractRepoData(pullEvent.Repo)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing repo data: %s", err)
		return
	}

	if err := e.PullClosedExecutor.CleanUpPull(repo, pull); err != nil {
		e.respond(w, logging.Error, http.StatusServiceUnavailable, "Error cleaning pull request: %s", err)
		return
	}
	e.Logger.Info("deleted locks and workspace for repo %s, pull %d", repo.FullName, pull.Num)
	fmt.Fprint(w, "Pull request cleaned successfully")
}

func (e *EventsController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	e.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
