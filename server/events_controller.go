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
	"io/ioutil"
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
const bitbucketEventTypeHeader = "X-Event-Key"
const bitbucketRequestIDHeader = "X-Request-UUID"

// EventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub.
type EventsController struct {
	CommandRunner events.CommandRunner
	PullCleaner   events.PullCleaner
	Logger        *logging.SimpleLogger
	Parser        events.EventParsing
	CommentParser events.CommentParsing
	// GithubWebHookSecret is the secret added to this webhook via the GitHub
	// UI that identifies this call as coming from GitHub. If empty, no
	// request validation is done.
	GithubWebHookSecret          []byte
	GithubRequestValidator       GithubRequestValidator
	GitlabRequestParserValidator GitlabRequestParserValidator
	// GitlabWebHookSecret is the secret added to this webhook via the GitLab
	// UI that identifies this call as coming from GitLab. If empty, no
	// request validation is done.
	GitlabWebHookSecret  []byte
	RepoWhitelistChecker *events.RepoWhitelistChecker
	// SupportedVCSHosts is which VCS hosts Atlantis was configured upon
	// startup to support.
	SupportedVCSHosts []models.VCSHostType
	VCSClient         vcs.ClientProxy
	TestingMode       bool
}

// Post handles POST webhook requests.
func (e *EventsController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(githubHeader) != "" {
		if !e.supportsHost(models.Github) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support GitHub")
			return
		}
		e.Logger.Debug("handling GitHub post")
		e.handleGithubPost(w, r)
		return
	} else if r.Header.Get(gitlabHeader) != "" {
		if !e.supportsHost(models.Gitlab) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
			return
		}
		e.Logger.Debug("handling GitLab post")
		e.handleGitlabPost(w, r)
		return
	} else if r.Header.Get(bitbucketEventTypeHeader) != "" {
		if !e.supportsHost(models.Bitbucket) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support Bitbucket")
			return
		}
		e.Logger.Debug("handling Bitbucket post")
		e.handleBitbucketPost(w, r)
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
	e.Logger.Debug("request valid")

	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")
	event, _ := github.ParseWebHook(github.WebHookType(r), payload)
	switch event := event.(type) {
	case *github.IssueCommentEvent:
		e.Logger.Debug("handling as comment event")
		e.HandleGithubCommentEvent(w, event, githubReqID)
	case *github.PullRequestEvent:
		e.Logger.Debug("handling as pull request event")
		e.HandleGithubPullRequestEvent(w, event, githubReqID)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event %s", githubReqID)
	}
}

func (e *EventsController) handleBitbucketPost(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get(bitbucketEventTypeHeader)
	reqID := r.Header.Get(bitbucketRequestIDHeader)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Unable to read body: %s %s=%s", err, bitbucketRequestIDHeader, reqID)
		return
	}
	switch eventType {
	case "pullrequest:created", "pullrequest:updated", "pullrequest:fulfilled", "pullrequest:rejected":
		e.Logger.Debug("handling as pull request state changed event")
		e.HandleBitbucketPullRequestEvent(w, eventType, body, reqID)
		return
	case "pullrequest:comment_created":
		e.Logger.Debug("handling as comment created event")
		e.HandleBitbucketCommentEvent(w, body, reqID)
		return
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event type %s %s=%s", eventType, bitbucketRequestIDHeader, reqID)
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

	// We pass in nil for maybeHeadRepo because the head repo data isn't
	// available in the GithubIssueComment event.
	e.handleCommentEvent(w, baseRepo, nil, nil, user, pullNum, event.Comment.GetBody(), models.Github)
}

// HandleBitbucketCommentEvent handles comment events from Bitbucket.
func (e *EventsController) HandleBitbucketCommentEvent(w http.ResponseWriter, body []byte, reqID string) {
	pull, baseRepo, headRepo, user, comment, err := e.Parser.ParseBitbucketCloudCommentEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketRequestIDHeader, reqID)
		return
	}
	e.handleCommentEvent(w, baseRepo, &headRepo, &pull, user, pull.Num, comment, models.Bitbucket)
}

func (e *EventsController) HandleBitbucketPullRequestEvent(w http.ResponseWriter, eventType string, body []byte, reqID string) {
	pull, baseRepo, headRepo, user, err := e.Parser.ParseBitbucketCloudPullEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketRequestIDHeader, reqID)
		return
	}
	pullEventType := e.Parser.GetBitbucketEventType(eventType)
	e.Logger.Info("identified event as type %q", pullEventType.String())
	e.handlePullRequestEvent(w, baseRepo, headRepo, pull, user, pullEventType)
}

// HandleGithubPullRequestEvent will delete any locks associated with the pull
// request if the event is a pull request closed event. It's exported to make
// testing easier.
func (e *EventsController) HandleGithubPullRequestEvent(w http.ResponseWriter, pullEvent *github.PullRequestEvent, githubReqID string) {
	pull, pullEventType, baseRepo, headRepo, user, err := e.Parser.ParseGithubPullEvent(pullEvent)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s", err, githubReqID)
		return
	}
	e.Logger.Info("identified event as type %q", pullEventType.String())
	e.handlePullRequestEvent(w, baseRepo, headRepo, pull, user, pullEventType)
}

func (e *EventsController) handlePullRequestEvent(w http.ResponseWriter, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, eventType models.PullRequestEventType) {
	if !e.RepoWhitelistChecker.IsWhitelisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		// If the repo isn't whitelisted and we receive an opened pull request
		// event we comment back on the pull request that the repo isn't
		// whitelisted. This is because the user might be expecting Atlantis to
		// autoplan. For other events, we just ignore them.
		if eventType == models.OpenedPullEvent {
			e.commentNotWhitelisted(baseRepo, pull.Num)
		}
		e.respond(w, logging.Debug, http.StatusForbidden, "Ignoring pull request event from non-whitelisted repo")
		return
	}

	switch eventType {
	case models.OpenedPullEvent, models.UpdatedPullEvent:
		// If the pull request was opened or updated, we will try to autoplan.

		// Respond with success and then actually execute the command asynchronously.
		// We use a goroutine so that this function returns and the connection is
		// closed.
		fmt.Fprintln(w, "Processing...")

		e.Logger.Info("executing autoplan")
		if !e.TestingMode {
			go e.CommandRunner.RunAutoplanCommand(baseRepo, headRepo, pull, user)
		} else {
			// When testing we want to wait for everything to complete.
			e.CommandRunner.RunAutoplanCommand(baseRepo, headRepo, pull, user)
		}
		return
	case models.ClosedPullEvent:
		// If the pull request was closed, we delete locks.
		if err := e.PullCleaner.CleanUpPull(baseRepo, pull); err != nil {
			e.respond(w, logging.Error, http.StatusInternalServerError, "Error cleaning pull request: %s", err)
			return
		}
		e.Logger.Info("deleted locks and workspace for repo %s, pull %d", baseRepo.FullName, pull.Num)
		fmt.Fprintln(w, "Pull request cleaned successfully")
		return
	case models.OtherPullEvent:
		// Else we ignore the event.
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring non-actionable pull request event")
		return
	}
}

func (e *EventsController) handleGitlabPost(w http.ResponseWriter, r *http.Request) {
	event, err := e.GitlabRequestParserValidator.ParseAndValidate(r, e.GitlabWebHookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, err.Error())
		return
	}
	e.Logger.Debug("request valid")

	switch event := event.(type) {
	case gitlab.MergeCommentEvent:
		e.Logger.Debug("handling as comment event")
		e.HandleGitlabCommentEvent(w, event)
	case gitlab.MergeEvent:
		e.Logger.Debug("handling as pull request event")
		e.HandleGitlabMergeRequestEvent(w, event)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event")
	}

}

// HandleGitlabCommentEvent handles comment events from GitLab where Atlantis
// commands can come from. It's exported to make testing easier.
func (e *EventsController) HandleGitlabCommentEvent(w http.ResponseWriter, event gitlab.MergeCommentEvent) {
	// todo: can gitlab return the pull request here too?
	baseRepo, headRepo, user, err := e.Parser.ParseGitlabMergeCommentEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}
	e.handleCommentEvent(w, baseRepo, &headRepo, nil, user, event.MergeRequest.IID, event.ObjectAttributes.Note, models.Gitlab)
}

func (e *EventsController) handleCommentEvent(w http.ResponseWriter, baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, comment string, vcsHost models.VCSHostType) {
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
	e.Logger.Info("parsed comment as %s", parseResult.Command)

	// At this point we know it's a command we're not supposed to ignore, so now
	// we check if this repo is allowed to run commands in the first place.
	if !e.RepoWhitelistChecker.IsWhitelisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		e.commentNotWhitelisted(baseRepo, pullNum)
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

	e.Logger.Debug("executing command")
	fmt.Fprintln(w, "Processing...")
	if !e.TestingMode {
		// Respond with success and then actually execute the command asynchronously.
		// We use a goroutine so that this function returns and the connection is
		// closed.
		go e.CommandRunner.RunCommentCommand(baseRepo, maybeHeadRepo, maybePull, user, pullNum, parseResult.Command)
	} else {
		// When testing we want to wait for everything to complete.
		e.CommandRunner.RunCommentCommand(baseRepo, maybeHeadRepo, maybePull, user, pullNum, parseResult.Command)
	}
}

// HandleGitlabMergeRequestEvent will delete any locks associated with the pull
// request if the event is a merge request closed event. It's exported to make
// testing easier.
func (e *EventsController) HandleGitlabMergeRequestEvent(w http.ResponseWriter, event gitlab.MergeEvent) {
	pull, pullEventType, baseRepo, headRepo, user, err := e.Parser.ParseGitlabMergeEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}
	e.Logger.Info("identified event as type %q", pullEventType.String())
	e.handlePullRequestEvent(w, baseRepo, headRepo, pull, user, pullEventType)
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
func (e *EventsController) commentNotWhitelisted(baseRepo models.Repo, pullNum int) {
	errMsg := "```\nError: This repo is not whitelisted for Atlantis.\n```"
	if err := e.VCSClient.CreateComment(baseRepo, pullNum, errMsg); err != nil {
		e.Logger.Err("unable to comment on pull request: %s", err)
	}
}
