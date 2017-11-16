package server

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server/events"
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/events/vcs"
	"github.com/hootsuite/atlantis/server/logging"
	"github.com/lkysow/go-gitlab"
)

const githubHeader = "X-Github-Event"
const gitlabHeader = "X-Gitlab-Event"

type EventsController struct {
	CommandRunner events.CommandRunner
	PullCleaner   events.PullCleaner
	Logger        *logging.SimpleLogger
	Parser        events.EventParsing
	// GithubWebHookSecret is the secret added to this webhook via the GitHub
	// UI that identifies this call as coming from GitHub. If empty, no
	// request validation is done.
	GithubWebHookSecret    []byte
	GithubRequestValidator GithubRequestValidator
	GitlabRequestParser    GitlabRequestParser
	// GitlabWebHookSecret is the secret added to this webhook via the GitLab
	// UI that identifies this call as coming from GitLab. If empty, no
	// request validation is done.
	GitlabWebHookSecret []byte
	// SupportedVCSHosts is which VCS hosts Atlantis was configured upon
	// startup to support.
	SupportedVCSHosts []vcs.Host
}

func (e *EventsController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(githubHeader) != "" {
		if !e.supportsHost(vcs.Github) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support GitHub")
			return
		}
		e.handleGithubPost(w, r)
		return
	} else if r.Header.Get(gitlabHeader) != "" {
		if !e.supportsHost(vcs.Gitlab) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
			return
		}
		e.handleGitlabPost(w, r)
		return
	}
	e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request")
}

// supportsHost returns true if h is in e.SupportedVCSHosts and false otherwise
func (e *EventsController) supportsHost(h vcs.Host) bool {
	for _, supported := range e.SupportedVCSHosts {
		if h == supported {
			return true
		}
	}
	return false
}

func (e *EventsController) handleGitlabPost(w http.ResponseWriter, r *http.Request) {
	event, err := e.GitlabRequestParser.Validate(r, e.GitlabWebHookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, err.Error())
		return
	}
	switch event := event.(type) {
	case gitlab.MergeCommentEvent:
		e.HandleGitlabCommentEvent(w, event)
	case gitlab.MergeEvent:
		e.HandleGitlabMergeRequestEvent(w, event)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event")
	}

}

func (e *EventsController) handleGithubPost(w http.ResponseWriter, r *http.Request) {
	// Validate the request against the optional webhook secret.
	payload, err := e.GithubRequestValidator.Validate(r, e.GithubWebHookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, err.Error())
		return
	}

	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")
	event, _ := github.ParseWebHook(github.WebHookType(r), payload)
	switch event := event.(type) {
	case *github.IssueCommentEvent:
		e.HandleGithubCommentEvent(w, event, githubReqID)
	case *github.PullRequestEvent:
		e.HandleGithubPullRequestEvent(w, event, githubReqID)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event %s", githubReqID)
	}
}

func (e *EventsController) HandleGithubCommentEvent(w http.ResponseWriter, event *github.IssueCommentEvent, githubReqID string) {
	if event.GetAction() != "created" {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since action was not created %s", githubReqID)
		return
	}

	baseRepo, user, pullNum, err := e.Parser.ParseGithubIssueCommentEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Failed parsing event: %v %s", err, githubReqID)
		return
	}

	command, err := e.Parser.DetermineCommand(event.Comment.GetBody(), vcs.Github)
	if err != nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring: %s %s", err, githubReqID)
		return
	}

	// Respond with success and then actually execute the command asynchronously.
	// We use a goroutine so that this function returns and the connection is
	// closed.
	fmt.Fprintln(w, "Processing...")
	go e.CommandRunner.ExecuteCommand(baseRepo, models.Repo{}, user, pullNum, command, vcs.Github)
}

func (e *EventsController) HandleGitlabCommentEvent(w http.ResponseWriter, event gitlab.MergeCommentEvent) {
	baseRepo, headRepo, user := e.Parser.ParseGitlabMergeCommentEvent(event)
	command, err := e.Parser.DetermineCommand(event.ObjectAttributes.Note, vcs.Gitlab)
	if err != nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring: %s", err)
		return
	}

	// Respond with success and then actually execute the command asynchronously.
	// We use a goroutine so that this function returns and the connection is
	// closed.
	fmt.Fprintln(w, "Processing...")
	go e.CommandRunner.ExecuteCommand(baseRepo, headRepo, user, event.MergeRequest.IID, command, vcs.Gitlab)
}

// HandleGitlabMergeRequestEvent will delete any locks associated with the merge request
func (e *EventsController) HandleGitlabMergeRequestEvent(w http.ResponseWriter, event gitlab.MergeEvent) {
	pull, repo := e.Parser.ParseGitlabMergeEvent(event)
	if pull.State != models.Closed {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring opened merge request event")
		return
	}
	if err := e.PullCleaner.CleanUpPull(repo, pull, vcs.Gitlab); err != nil {
		e.respond(w, logging.Error, http.StatusInternalServerError, "Error cleaning pull request: %s", err)
		return
	}
	e.Logger.Info("deleted locks and workspace for repo %s, pull %d", repo.FullName, pull.Num)
	fmt.Fprintln(w, "Merge request cleaned successfully")
}

// HandleGithubPullRequestEvent will delete any locks associated with the pull request
func (e *EventsController) HandleGithubPullRequestEvent(w http.ResponseWriter, pullEvent *github.PullRequestEvent, githubReqID string) {
	if pullEvent.GetAction() != "closed" {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring pull request event since action was not closed %s", githubReqID)
		return
	}
	pull, _, err := e.Parser.ParseGithubPull(pullEvent.PullRequest)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s", err)
		return
	}
	repo, err := e.Parser.ParseGithubRepo(pullEvent.Repo)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing repo data: %s", err)
		return
	}

	if err := e.PullCleaner.CleanUpPull(repo, pull, vcs.Github); err != nil {
		e.respond(w, logging.Error, http.StatusInternalServerError, "Error cleaning pull request: %s", err)
		return
	}
	e.Logger.Info("deleted locks and workspace for repo %s, pull %d", repo.FullName, pull.Num)
	fmt.Fprintln(w, "Pull request cleaned successfully")
}

func (e *EventsController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	e.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
