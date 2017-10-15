package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server/events"
	"github.com/hootsuite/atlantis/server/logging"
)

type EventsController struct {
	commandHandler      *events.CommandHandler
	pullClosedExecutor  *events.PullClosedExecutor
	logger              *logging.SimpleLogger
	parser              events.EventParsing
	githubWebHookSecret []byte
}

func (e *EventsController) Post(w http.ResponseWriter, r *http.Request) {
	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")
	var payload []byte

	// If we need to validate the Webhook secret, we can use go-github's
	// ValidatePayload method. Otherwise we need to parse the request ourselvee.
	if len(e.githubWebHookSecret) != 0 {
		var err error
		if payload, err = github.ValidatePayload(r, e.githubWebHookSecret); err != nil {
			e.respond(w, logging.Warn, http.StatusBadRequest, "webhook request failed secret key validation")
			return
		}
	} else {
		switch ct := r.Header.Get("Content-Type"); ct {
		case "application/json":
			var err error
			if payload, err = ioutil.ReadAll(r.Body); err != nil {
				e.respond(w, logging.Warn, http.StatusBadRequest, "could not read body: %s", err)
				return
			}
		case "application/x-www-form-urlencoded":
			// GitHub stores the json payload as a form value
			payloadForm := r.FormValue("payload")
			if payloadForm == "" {
				e.respond(w, logging.Warn, http.StatusBadRequest, "webhook request did not contain expected 'payload' form value")
				return
			}
			payload = []byte(payloadForm)
		default:
			e.respond(w, logging.Warn, http.StatusBadRequest, fmt.Sprintf("webhook request has unsupported Content-Type %q", ct))
		}
	}

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

	ctx := &events.CommandContext{}
	command, err := e.parser.DetermineCommand(event)
	if err != nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring: %s %s", err, githubReqID)
		return
	}
	ctx.Command = command

	if err = e.parser.ExtractCommentData(event, ctx); err != nil {
		e.respond(w, logging.Error, http.StatusInternalServerError, "Failed parsing event: %v %s", err, githubReqID)
		return
	}
	// Respond with success and then actually execute the command asynchronously.
	// We use a goroutine so that this function returns and the connection is
	// closed.
	fmt.Fprintln(w, "Processing...")
	go e.commandHandler.ExecuteCommand(ctx)
}

// HandlePullRequestEvent will delete any locks associated with the pull request
func (e *EventsController) HandlePullRequestEvent(w http.ResponseWriter, pullEvent *github.PullRequestEvent, githubReqID string) {
	if pullEvent.GetAction() != "closed" {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring pull request event since action was not closed %s", githubReqID)
		return
	}
	pull, _, err := e.parser.ExtractPullData(pullEvent.PullRequest)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s", err)
		return
	}
	repo, err := e.parser.ExtractRepoData(pullEvent.Repo)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing repo data: %s", err)
		return
	}

	if err := e.pullClosedExecutor.CleanUpPull(repo, pull); err != nil {
		e.respond(w, logging.Error, http.StatusInternalServerError, "Error cleaning pull request: %s", err)
		return
	}
	e.logger.Info("deleted locks and workspace for repo %s, pull %d", repo.FullName, pull.Num)
	fmt.Fprint(w, "Pull request cleaned successfully")
}

func (e *EventsController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	e.logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
