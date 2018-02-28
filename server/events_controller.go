package server

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

const githubHeader = "X-Github-Event"
const gitlabHeader = "X-Gitlab-Event"

// EventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub. It's split out from Server to make testing easier.
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
	VCSClient         vcs.ClientProxy
}

// Post handles POST webhook requests.
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

// HandleGithubCommentEvent handles comment events from GitHub where Atlantis
// commands can come from. It's exported to make testing easier.
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

	// We pass in an empty models.Repo for headRepo because we need to do additional
	// calls to get that information but we need this code path to be generic.
	// Later on in CommandHandler we detect that this is a GitHub event and
	// make the necessary calls to get the headRepo.
	e.handleCommentEvent(w, baseRepo, models.Repo{}, user, pullNum, event.Comment.GetBody(), vcs.Github)
}

// HandleGithubPullRequestEvent will delete any locks associated with the pull
// request if the event is a pull request closed event. It's exported to make
// testing easier.
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

// HandleGitlabCommentEvent handles comment events from GitLab where Atlantis
// commands can come from. It's exported to make testing easier.
func (e *EventsController) HandleGitlabCommentEvent(w http.ResponseWriter, event gitlab.MergeCommentEvent) {
	baseRepo, headRepo, user := e.Parser.ParseGitlabMergeCommentEvent(event)
	e.handleCommentEvent(w, baseRepo, headRepo, user, event.MergeRequest.IID, event.ObjectAttributes.Note, vcs.Gitlab)
}

func (e *EventsController) handleCommentEvent(w http.ResponseWriter, baseRepo models.Repo, headRepo models.Repo, user models.User, pullNum int, comment string, vcsHost vcs.Host) {
	parseResult := e.Parser.DetermineCommand(comment, vcsHost)
	if parseResult.Ignore {
		truncated := comment
		if len(truncated) > 40 {
			truncated = comment[:40] + "..."
		}
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring non-command comment: %q", truncated)
		return
	}

	// If the command isn't valid or doesn't require processing, ex.
	// "atlantis help" then we just comment back immediately.
	// We do this here rather than earlier because we need access to the pull
	// variable to comment back on the pull request.
	if parseResult.CommentResponse != "" {
		if err := e.VCSClient.CreateComment(baseRepo, pullNum, parseResult.CommentResponse, vcsHost); err != nil {
			e.Logger.Err("unable to comment on pull request: %s", err)
		}
		e.respond(w, logging.Info, http.StatusOK, "Commenting back on pull request")
		return
	}

	// Respond with success and then actually execute the command asynchronously.
	// We use a goroutine so that this function returns and the connection is
	// closed.
	fmt.Fprintln(w, "Processing...")
	go e.CommandRunner.ExecuteCommand(baseRepo, headRepo, user, pullNum, parseResult.Command, vcsHost)
}

// HandleGitlabMergeRequestEvent will delete any locks associated with the pull
// request if the event is a merge request closed event. It's exported to make
// testing easier.
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

// supportsHost returns true if h is in e.SupportedVCSHosts and false otherwise.
func (e *EventsController) supportsHost(h vcs.Host) bool {
	for _, supported := range e.SupportedVCSHosts {
		if h == supported {
			return true
		}
	}
	return false
}

func (e *EventsController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	e.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
