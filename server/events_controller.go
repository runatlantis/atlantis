// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
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
	CommentParser events.CommentParsing
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
	RepoWhitelist       *events.RepoWhitelist
	// SupportedVCSHosts is which VCS hosts Atlantis was configured upon
	// startup to support.
	SupportedVCSHosts []models.VCSHostType
	VCSClient         vcs.ClientProxy
	// AtlantisGithubUser is the user that atlantis is running as for Github.
	AtlantisGithubUser models.User
	// AtlantisGitlabUser is the user that atlantis is running as for Gitlab.
	AtlantisGitlabUser models.User
}

// Post handles POST webhook requests.
func (e *EventsController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(githubHeader) != "" {
		if !e.supportsHost(models.Github) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support GitHub")
			return
		}
		e.handleGithubPost(w, r)
		return
	} else if r.Header.Get(gitlabHeader) != "" {
		if !e.supportsHost(models.Gitlab) {
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
	e.handleCommentEvent(w, baseRepo, models.Repo{}, user, pullNum, event.Comment.GetBody(), models.Github)
}

// HandleGithubPullRequestEvent will delete any locks associated with the pull
// request if the event is a pull request closed event. It's exported to make
// testing easier.
func (e *EventsController) HandleGithubPullRequestEvent(w http.ResponseWriter, pullEvent *github.PullRequestEvent, githubReqID string) {
	pull, headRepo, err := e.Parser.ParseGithubPull(pullEvent.PullRequest)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s", err, githubReqID)
		return
	}
	baseRepo, err := e.Parser.ParseGithubRepo(pullEvent.Repo)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing repo data: %s %s", err, githubReqID)
		return
	}
	var eventType string
	switch pullEvent.GetAction() {
	case "opened":
		eventType = OpenPullEvent
	case "synchronize":
		eventType = UpdatedPullEvent
	case "closed":
		eventType = ClosedPullEvent
	default:
		eventType = OtherPullEvent
	}
	e.handlePullRequestEvent(w, baseRepo, headRepo, pull, e.AtlantisGithubUser, eventType)
}

const OpenPullEvent = "opened"
const UpdatedPullEvent = "updated"
const ClosedPullEvent = "closed"
const OtherPullEvent = "other"

func (e *EventsController) handlePullRequestEvent(w http.ResponseWriter, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, eventType string) {
	if !e.RepoWhitelist.IsWhitelisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		// If the repo isn't whitelisted and we receive an opened pull request
		// event we comment back on the pull request that the repo isn't
		// whitelisted. This is because the user might be expecting Atlantis to
		// autoplan. For other events, we just ignore them.
		if eventType == OpenPullEvent {
			e.commentNotWhitelisted(w, baseRepo, pull.Num)
		}
		e.respond(w, logging.Debug, http.StatusForbidden, "Ignoring pull request event from non-whitelisted repo")
		return
	}

	switch eventType {
	case OpenPullEvent, UpdatedPullEvent:
		// If the pull request was opened or updated, we will try to autoplan.

		// Respond with success and then actually execute the command asynchronously.
		// We use a goroutine so that this function returns and the connection is
		// closed.
		fmt.Fprintln(w, "Processing...")
		// We use a Command to represent autoplanning but we set dir and
		// workspace to '*' to indicate that all applicable dirs and workspaces
		// should be planned.
		autoplanCmd := events.NewCommand("*", nil, events.Plan, false, "*", true)
		go e.CommandRunner.ExecuteCommand(baseRepo, headRepo, user, pull.Num, autoplanCmd)
		return
	case ClosedPullEvent:
		// If the pull request was closed, we delete locks.
		if err := e.PullCleaner.CleanUpPull(baseRepo, pull); err != nil {
			e.respond(w, logging.Error, http.StatusInternalServerError, "Error cleaning pull request: %s", err)
			return
		}
		e.Logger.Info("deleted locks and workspace for repo %s, pull %d", baseRepo.FullName, pull.Num)
		fmt.Fprintln(w, "Pull request cleaned successfully")
		return
	case OtherPullEvent:
		// Else we ignore the event.
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring opened pull request event")
		return
	}
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
	baseRepo, headRepo, user, err := e.Parser.ParseGitlabMergeCommentEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}
	e.handleCommentEvent(w, baseRepo, headRepo, user, event.MergeRequest.IID, event.ObjectAttributes.Note, models.Gitlab)
}

func (e *EventsController) handleCommentEvent(w http.ResponseWriter, baseRepo models.Repo, headRepo models.Repo, user models.User, pullNum int, comment string, vcsHost models.VCSHostType) {
	parseResult := e.CommentParser.Parse(comment, vcsHost)
	if parseResult.Ignore {
		truncated := comment
		truncateLen := 40
		if len(truncated) > truncateLen {
			truncated = comment[:truncateLen] + "..."
		}
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring non-command comment: %q", truncated)
		return
	}

	// At this point we know it's a command we're not supposed to ignore, so now
	// we check if this repo is allowed to run commands in the first place.
	if !e.RepoWhitelist.IsWhitelisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		e.commentNotWhitelisted(w, baseRepo, pullNum)
		e.respond(w, logging.Warn, http.StatusForbidden, "Repo not whitelisted")
		return
	}

	// If the command isn't valid or doesn't require processing, ex.
	// "atlantis help" then we just comment back immediately.
	// We do this here rather than earlier because we need access to the pull
	// variable to comment back on the pull request.
	if parseResult.CommentResponse != "" {
		if err := e.VCSClient.CreateComment(baseRepo, pullNum, parseResult.CommentResponse); err != nil {
			e.Logger.Err("unable to comment on pull request: %s", err)
		}
		e.respond(w, logging.Info, http.StatusOK, "Commenting back on pull request")
		return
	}

	// Respond with success and then actually execute the command asynchronously.
	// We use a goroutine so that this function returns and the connection is
	// closed.
	fmt.Fprintln(w, "Processing...")
	go e.CommandRunner.ExecuteCommand(baseRepo, headRepo, user, pullNum, parseResult.Command)
}

// HandleGitlabMergeRequestEvent will delete any locks associated with the pull
// request if the event is a merge request closed event. It's exported to make
// testing easier.
func (e *EventsController) HandleGitlabMergeRequestEvent(w http.ResponseWriter, event gitlab.MergeEvent) {
	pull, baseRepo, headRepo, err := e.Parser.ParseGitlabMergeEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}
	var eventType string
	switch event.ObjectAttributes.Action {
	case "open":
		eventType = OpenPullEvent
	case "update":
		eventType = UpdatedPullEvent
	case "merge", "close":
		eventType = ClosedPullEvent
	default:
		eventType = OtherPullEvent
	}
	e.handlePullRequestEvent(w, baseRepo, headRepo, pull, e.AtlantisGitlabUser, eventType)
}

// supportsHost returns true if h is in e.SupportedVCSHosts and false otherwise.
func (e *EventsController) supportsHost(h models.VCSHostType) bool {
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

// commentNotWhitelisted comments on the pull request that the repo is not
// whitelisted.
func (e *EventsController) commentNotWhitelisted(w http.ResponseWriter, baseRepo models.Repo, pullNum int) {
	errMsg := "```\nError: This repo is not whitelisted for Atlantis.\n```"
	if err := e.VCSClient.CreateComment(baseRepo, pullNum, errMsg); err != nil {
		e.Logger.Err("unable to comment on pull request: %s", err)
	}
}
