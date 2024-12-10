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

package events

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
	gitlab "github.com/xanzy/go-gitlab"
)

const githubHeader = "X-Github-Event"
const gitlabHeader = "X-Gitlab-Event"
const azuredevopsHeader = "Request-Id"

const giteaHeader = "X-Gitea-Event"
const giteaEventTypeHeader = "X-Gitea-Event-Type"
const giteaSignatureHeader = "X-Gitea-Signature"
const giteaRequestIDHeader = "X-Gitea-Delivery"

// bitbucketEventTypeHeader is the same in both cloud and server.
const bitbucketEventTypeHeader = "X-Event-Key"
const bitbucketCloudRequestIDHeader = "X-Request-UUID"
const bitbucketServerRequestIDHeader = "X-Request-ID"
const bitbucketServerSignatureHeader = "X-Hub-Signature"

// The URL used for Azure DevOps test webhooks
const azuredevopsTestURL = "https://fabrikam.visualstudio.com/DefaultCollection/_apis/git/repositories/4bc14d40-c903-45e2-872e-0462c7748079"

// VCSEventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub.
type VCSEventsController struct {
	CommandRunner  events.CommandRunner
	PullCleaner    events.PullCleaner
	Logger         logging.SimpleLogging
	Scope          tally.Scope
	Parser         events.EventParsing
	CommentParser  events.CommentParsing
	ApplyDisabled  bool
	EmojiReaction  string
	ExecutableName string
	// GithubWebhookSecret is the secret added to this webhook via the GitHub
	// UI that identifies this call as coming from GitHub. If empty, no
	// request validation is done.
	GithubWebhookSecret          []byte
	GithubRequestValidator       GithubRequestValidator
	GitlabRequestParserValidator GitlabRequestParserValidator
	// GitlabWebhookSecret is the secret added to this webhook via the GitLab
	// UI that identifies this call as coming from GitLab. If empty, no
	// request validation is done.
	GitlabWebhookSecret  []byte
	RepoAllowlistChecker *events.RepoAllowlistChecker
	// SilenceAllowlistErrors controls whether we write an error comment on
	// pull requests from non-allowlisted repos.
	SilenceAllowlistErrors bool
	// SupportedVCSHosts is which VCS hosts Atlantis was configured upon
	// startup to support.
	SupportedVCSHosts []models.VCSHostType
	VCSClient         vcs.Client
	TestingMode       bool
	// BitbucketWebhookSecret is the secret added to this webhook via the Bitbucket
	// UI that identifies this call as coming from Bitbucket. If empty, no
	// request validation is done.
	BitbucketWebhookSecret []byte
	// AzureDevopsWebhookUser is the Basic authentication username added to this
	// webhook via the Azure DevOps UI that identifies this call as coming from your
	// Azure DevOps Team Project. If empty, no request validation is done.
	// For more information, see https://docs.microsoft.com/en-us/azure/devops/service-hooks/services/webhooks?view=azure-devops
	AzureDevopsWebhookBasicUser []byte
	// AzureDevopsWebhookPassword is the Basic authentication password added to this
	// webhook via the Azure DevOps UI that identifies this call as coming from your
	// Azure DevOps Team Project. If empty, no request validation is done.
	AzureDevopsWebhookBasicPassword []byte
	AzureDevopsRequestValidator     AzureDevopsRequestValidator
	GiteaWebhookSecret              []byte
}

// Post handles POST webhook requests.
func (e *VCSEventsController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(giteaHeader) != "" {
		if !e.supportsHost(models.Gitea) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support Gitea")
			return
		}
		e.Logger.Debug("handling Gitea post")
		e.handleGiteaPost(w, r)
		return
	} else if r.Header.Get(githubHeader) != "" {
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
		// Bitbucket Cloud and Server use the same event type header but they
		// use different request ID headers.
		if r.Header.Get(bitbucketCloudRequestIDHeader) != "" {
			if !e.supportsHost(models.BitbucketCloud) {
				e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support Bitbucket Cloud")
				return
			}
			e.Logger.Debug("handling Bitbucket Cloud post")
			e.handleBitbucketCloudPost(w, r)
			return
		} else if r.Header.Get(bitbucketServerRequestIDHeader) != "" {
			if !e.supportsHost(models.BitbucketServer) {
				e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support Bitbucket Server")
				return
			}
			e.Logger.Debug("handling Bitbucket Server post")
			e.handleBitbucketServerPost(w, r)
			return
		}
	} else if r.Header.Get(azuredevopsHeader) != "" {
		if !e.supportsHost(models.AzureDevops) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support AzureDevops")
			return
		}
		e.Logger.Debug("handling AzureDevops post")
		e.handleAzureDevopsPost(w, r)
		return
	}
	e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request")
}

type HTTPError struct {
	err        error
	code       int
	isSilenced bool
}

type HTTPResponse struct {
	body string
	err  HTTPError
}

func (e *VCSEventsController) handleGithubPost(w http.ResponseWriter, r *http.Request) {
	// Validate the request against the optional webhook secret.
	payload, err := e.GithubRequestValidator.Validate(r, e.GithubWebhookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, "%s", err.Error())
		return
	}

	githubReqID := "X-Github-Delivery=" + html.EscapeString(r.Header.Get("X-Github-Delivery"))
	logger := e.Logger.With("gh-request-id", githubReqID)
	scope := e.Scope.SubScope("github_event")

	logger.Debug("request valid")

	event, _ := github.ParseWebHook(github.WebHookType(r), payload)

	var resp HTTPResponse

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		resp = e.HandleGithubCommentEvent(event, githubReqID, logger)
		scope = scope.SubScope(fmt.Sprintf("comment_%s", *event.Action))
		scope = vcs.SetGitScopeTags(scope, event.GetRepo().GetFullName(), event.GetIssue().GetNumber())
	case *github.PullRequestEvent:
		resp = e.HandleGithubPullRequestEvent(logger, event, githubReqID)
		scope = scope.SubScope(fmt.Sprintf("pr_%s", *event.Action))
		scope = vcs.SetGitScopeTags(scope, event.GetRepo().GetFullName(), event.GetNumber())
	default:
		resp = HTTPResponse{
			body: fmt.Sprintf("Ignoring unsupported event %s", githubReqID),
		}
	}

	if resp.err.code != 0 {
		if !resp.err.isSilenced {
			logger.Err("error handling gh post code: %d err: %s", resp.err.code, resp.err.err.Error())
		}
		scope.Counter(fmt.Sprintf("error_%d", resp.err.code)).Inc(1)
		w.WriteHeader(resp.err.code)
		fmt.Fprintln(w, resp.err.err.Error())
		return
	}

	scope.Counter(fmt.Sprintf("success_%d", http.StatusOK)).Inc(1)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, resp.body)
}

func (e *VCSEventsController) handleBitbucketCloudPost(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get(bitbucketEventTypeHeader)
	reqID := r.Header.Get(bitbucketCloudRequestIDHeader)
	defer r.Body.Close() // nolint: errcheck
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Unable to read body: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	switch eventType {
	case bitbucketcloud.PullCreatedHeader, bitbucketcloud.PullUpdatedHeader, bitbucketcloud.PullFulfilledHeader, bitbucketcloud.PullRejectedHeader:
		e.Logger.Debug("handling as pull request state changed event")
		e.handleBitbucketCloudPullRequestEvent(e.Logger, w, eventType, body, reqID)
		return
	case bitbucketcloud.PullCommentCreatedHeader:
		e.Logger.Debug("handling as comment created event")
		e.HandleBitbucketCloudCommentEvent(w, body, reqID)
		return
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event type %s %s=%s", eventType, bitbucketCloudRequestIDHeader, reqID)
	}
}

func (e *VCSEventsController) handleBitbucketServerPost(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get(bitbucketEventTypeHeader)
	reqID := r.Header.Get(bitbucketServerRequestIDHeader)
	sig := r.Header.Get(bitbucketServerSignatureHeader)
	defer r.Body.Close() // nolint: errcheck
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Unable to read body: %s %s=%s", err, bitbucketServerRequestIDHeader, reqID)
		return
	}
	if eventType == bitbucketserver.DiagnosticsPingHeader {
		// Specially handle the diagnostics:ping event because Bitbucket Server
		// doesn't send the signature with this event for some reason.
		e.respond(w, logging.Info, http.StatusOK, "Successfully received %s event %s=%s", eventType, bitbucketServerRequestIDHeader, reqID)
		return
	}
	if len(e.BitbucketWebhookSecret) > 0 {
		if err := bitbucketserver.ValidateSignature(body, sig, e.BitbucketWebhookSecret); err != nil {
			e.respond(w, logging.Warn, http.StatusBadRequest, "%s", errors.Wrap(err, "request did not pass validation").Error())
			return
		}
	}
	switch eventType {
	case bitbucketserver.PullCreatedHeader, bitbucketserver.PullFromRefUpdatedHeader, bitbucketserver.PullMergedHeader, bitbucketserver.PullDeclinedHeader, bitbucketserver.PullDeletedHeader:
		e.Logger.Debug("handling as pull request state changed event")
		e.handleBitbucketServerPullRequestEvent(e.Logger, w, eventType, body, reqID)
		return
	case bitbucketserver.PullCommentCreatedHeader:
		e.Logger.Debug("handling as comment created event")
		e.HandleBitbucketServerCommentEvent(w, body, reqID)
		return
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event type %s %s=%s", eventType, bitbucketServerRequestIDHeader, reqID)
	}
}

func (e *VCSEventsController) handleAzureDevopsPost(w http.ResponseWriter, r *http.Request) {
	// Validate the request against the optional basic auth username and password.
	payload, err := e.AzureDevopsRequestValidator.Validate(r, e.AzureDevopsWebhookBasicUser, e.AzureDevopsWebhookBasicPassword)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusUnauthorized, "%s", err.Error())
		return
	}
	e.Logger.Debug("request valid")

	azuredevopsReqID := "Request-Id=" + r.Header.Get("Request-Id")
	event, err := azuredevops.ParseWebHook(payload)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Failed parsing webhook: %v %s", err, azuredevopsReqID)
		return
	}
	switch event.PayloadType {
	case azuredevops.PullRequestCommentedEvent:
		e.Logger.Debug("handling as pull request commented event")
		e.HandleAzureDevopsPullRequestCommentedEvent(w, event, azuredevopsReqID)
	case azuredevops.PullRequestEvent:
		e.Logger.Debug("handling as pull request event")
		e.HandleAzureDevopsPullRequestEvent(w, event, azuredevopsReqID)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event: %v %s", event.PayloadType, azuredevopsReqID)
	}
}

func (e *VCSEventsController) handleGiteaPost(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get(giteaSignatureHeader)
	eventType := r.Header.Get(giteaEventTypeHeader)
	reqID := r.Header.Get(giteaRequestIDHeader)

	defer r.Body.Close() // Ensure the request body is closed

	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Unable to read body: %s %s=%s", err, "X-Gitea-Delivery", reqID)
		return
	}

	if len(e.GiteaWebhookSecret) > 0 {
		if err := gitea.ValidateSignature(body, signature, e.GiteaWebhookSecret); err != nil {
			e.respond(w, logging.Warn, http.StatusBadRequest, "%s", errors.Wrap(err, "request did not pass validation").Error())
			return
		}
	}

	logger := e.Logger.With("gitea-request-id", reqID)

	// Log the event type for debugging purposes
	logger.Debug("Received Gitea event %s with ID %s", eventType, reqID)

	// Depending on the event type, handle the event appropriately
	switch eventType {
	case "pull_request_comment":
		e.HandleGiteaPullRequestCommentEvent(w, body, reqID)
	case "pull_request":
		logger.Debug("Handling as pull_request")
		e.handleGiteaPullRequestEvent(logger, w, body, reqID)
	// Add other case handlers as necessary
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported Gitea event type: %s %s=%s", eventType, "X-Gitea-Delivery", reqID)
	}
}

func (e *VCSEventsController) handleGiteaPullRequestEvent(logger logging.SimpleLogging, w http.ResponseWriter, body []byte, reqID string) {
	logger.Debug("Entering handleGiteaPullRequestEvent")
	// Attempt to unmarshal the incoming body into the Gitea PullRequest struct
	var payload gitea.GiteaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		e.Logger.Err("Failed to unmarshal Gitea webhook payload: %v", err)
		e.respond(w, logging.Error, http.StatusBadRequest, "Failed to parse request body: %s %s=%s", err, giteaRequestIDHeader, reqID)
		return
	}

	logger.Debug("Successfully unmarshaled Gitea event")

	// Use the parser function to convert into Atlantis models
	pull, pullEventType, baseRepo, headRepo, user, err := e.Parser.ParseGiteaPullRequestEvent(payload.PullRequest)
	if err != nil {
		e.Logger.Err("Failed to parse Gitea pull request event: %v", err)
		e.respond(w, logging.Error, http.StatusInternalServerError, "Failed to process event")
		return
	}

	logger.Debug("Parsed Gitea event into Atlantis models successfully")

	// Annotate logger with repo and pull/merge request number.
	logger = logger.With(
		"repo", baseRepo.FullName,
		"pull", strconv.Itoa(pull.Num),
	)
	logger.Info("Handling Gitea Pull Request '%s' event", pullEventType.String())
	response := e.handlePullRequestEvent(logger, baseRepo, headRepo, pull, user, pullEventType)

	e.respond(w, logging.Debug, http.StatusOK, "%s", response.body)
}

// HandleGiteaPullRequestCommentEvent handles comment events from Gitea where Atlantis commands can come from.
func (e *VCSEventsController) HandleGiteaPullRequestCommentEvent(w http.ResponseWriter, body []byte, reqID string) {
	var event gitea.GiteaIssueCommentPayload
	if err := json.Unmarshal(body, &event); err != nil {
		e.Logger.Err("Failed to unmarshal Gitea comment payload: %v", err)
		e.respond(w, logging.Error, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	e.Logger.Debug("Successfully unmarshaled Gitea comment event")

	baseRepo, user, pullNum, _ := e.Parser.ParseGiteaIssueCommentEvent(event)
	// Since we're lacking headRepo and maybePull details, we'll pass nil
	// This follows the same approach as the GitHub client for handling comment events without full PR details
	response := e.handleCommentEvent(e.Logger, baseRepo, nil, nil, user, pullNum, event.Comment.Body, event.Comment.ID, models.Gitea)

	e.respond(w, logging.Debug, http.StatusOK, "%s", response.body)
}

// HandleGithubCommentEvent handles comment events from GitHub where Atlantis
// commands can come from. It's exported to make testing easier.
func (e *VCSEventsController) HandleGithubCommentEvent(event *github.IssueCommentEvent, githubReqID string, logger logging.SimpleLogging) HTTPResponse {
	if event.GetAction() != "created" {
		return HTTPResponse{
			body: fmt.Sprintf("Ignoring comment event since action was not created %s", githubReqID),
		}
	}

	baseRepo, user, pullNum, err := e.Parser.ParseGithubIssueCommentEvent(logger, event)

	wrapped := errors.Wrapf(err, "Failed parsing event: %s", githubReqID)
	if err != nil {
		return HTTPResponse{
			body: wrapped.Error(),
			err: HTTPError{
				code:       http.StatusBadRequest,
				err:        wrapped,
				isSilenced: false,
			},
		}
	}

	comment := event.GetComment()

	// We pass in nil for maybeHeadRepo because the head repo data isn't
	// available in the GithubIssueComment event.
	return e.handleCommentEvent(logger, baseRepo, nil, nil, user, pullNum, comment.GetBody(), comment.GetID(), models.Github)
}

// HandleBitbucketCloudCommentEvent handles comment events from Bitbucket.
func (e *VCSEventsController) HandleBitbucketCloudCommentEvent(w http.ResponseWriter, body []byte, reqID string) {
	pull, baseRepo, headRepo, user, comment, err := e.Parser.ParseBitbucketCloudPullCommentEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	resp := e.handleCommentEvent(e.Logger, baseRepo, &headRepo, &pull, user, pull.Num, comment, -1, models.BitbucketCloud)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

// HandleBitbucketServerCommentEvent handles comment events from Bitbucket.
func (e *VCSEventsController) HandleBitbucketServerCommentEvent(w http.ResponseWriter, body []byte, reqID string) {
	pull, baseRepo, headRepo, user, comment, err := e.Parser.ParseBitbucketServerPullCommentEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	resp := e.handleCommentEvent(e.Logger, baseRepo, &headRepo, &pull, user, pull.Num, comment, -1, models.BitbucketCloud)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

func (e *VCSEventsController) handleBitbucketCloudPullRequestEvent(logger logging.SimpleLogging, w http.ResponseWriter, eventType string, body []byte, reqID string) {
	pull, baseRepo, headRepo, user, err := e.Parser.ParseBitbucketCloudPullEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	e.Logger.Debug("SHA is %q", pull.HeadCommit)
	pullEventType := e.Parser.GetBitbucketCloudPullEventType(eventType, pull.HeadCommit, pull.URL)

	// Annotate logger with repo and pull/merge request number.
	logger = logger.With(
		"repo", baseRepo.FullName,
		"pull", strconv.Itoa(pull.Num),
	)

	logger.Info("Handling Bitbucket Cloud Pull Request '%s' event", pullEventType.String())
	resp := e.handlePullRequestEvent(e.Logger, baseRepo, headRepo, pull, user, pullEventType)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

func (e *VCSEventsController) handleBitbucketServerPullRequestEvent(logger logging.SimpleLogging, w http.ResponseWriter, eventType string, body []byte, reqID string) {
	pull, baseRepo, headRepo, user, err := e.Parser.ParseBitbucketServerPullEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketServerRequestIDHeader, reqID)
		return
	}
	pullEventType := e.Parser.GetBitbucketServerPullEventType(eventType)

	// Annotate logger with repo and pull/merge request number.
	logger = logger.With(
		"repo", baseRepo.FullName,
		"pull", strconv.Itoa(pull.Num),
	)

	logger.Info("Handling Bitbucket Server Pull Request '%s' event", pullEventType.String())
	resp := e.handlePullRequestEvent(e.Logger, baseRepo, headRepo, pull, user, pullEventType)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

// HandleGithubPullRequestEvent will delete any locks associated with the pull
// request if the event is a pull request closed event. It's exported to make
// testing easier.
func (e *VCSEventsController) HandleGithubPullRequestEvent(logger logging.SimpleLogging, pullEvent *github.PullRequestEvent, githubReqID string) HTTPResponse {
	pull, pullEventType, baseRepo, headRepo, user, err := e.Parser.ParseGithubPullEvent(logger, pullEvent)
	if err != nil {
		wrapped := errors.Wrapf(err, "Error parsing pull data: %s %s", err, githubReqID)
		return HTTPResponse{
			body: wrapped.Error(),
			err: HTTPError{
				code:       http.StatusBadRequest,
				err:        wrapped,
				isSilenced: false,
			},
		}
	}

	// Annotate logger with repo and pull/merge request number.
	logger = logger.With(
		"repo", baseRepo.FullName,
		"pull", strconv.Itoa(pull.Num),
	)

	logger.Info("Handling GitHub Pull Request '%s' event", pullEventType.String())
	return e.handlePullRequestEvent(logger, baseRepo, headRepo, pull, user, pullEventType)
}

func (e *VCSEventsController) handlePullRequestEvent(logger logging.SimpleLogging, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, eventType models.PullRequestEventType) HTTPResponse {
	if !e.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		// If the repo isn't allowlisted and we receive an opened pull request
		// event we comment back on the pull request that the repo isn't
		// allowlisted. This is because the user might be expecting Atlantis to
		// autoplan. For other events, we just ignore them.
		if eventType == models.OpenedPullEvent {
			e.commentNotAllowlisted(baseRepo, pull.Num)
		}

		err := errors.Errorf("Pull request event from non-allowlisted repo '%s/%s'", baseRepo.VCSHost.Hostname, baseRepo.FullName)

		return HTTPResponse{
			body: err.Error(),
			err: HTTPError{
				code:       http.StatusForbidden,
				err:        err,
				isSilenced: e.SilenceAllowlistErrors,
			},
		}
	}

	switch eventType {
	case models.OpenedPullEvent, models.UpdatedPullEvent:
		// If the pull request was opened or updated, we will try to autoplan.

		// Respond with success and then actually execute the command asynchronously.
		// We use a goroutine so that this function returns and the connection is
		// closed.
		if !e.TestingMode {
			go e.CommandRunner.RunAutoplanCommand(baseRepo, headRepo, pull, user)
		} else {
			// When testing we want to wait for everything to complete.
			e.CommandRunner.RunAutoplanCommand(baseRepo, headRepo, pull, user)
		}
		return HTTPResponse{
			body: "Processing...",
		}
	case models.ClosedPullEvent:
		// If the pull request was closed, we delete locks.
		logger.Info("Pull request closed, cleaning up...")
		if err := e.PullCleaner.CleanUpPull(logger, baseRepo, pull); err != nil {
			return HTTPResponse{
				body: err.Error(),
				err: HTTPError{
					code:       http.StatusForbidden,
					err:        err,
					isSilenced: false,
				},
			}
		}
		logger.Info("Locks and workspace successfully deleted")
		return HTTPResponse{
			body: "Pull request cleaned successfully",
		}
	case models.OtherPullEvent:
		// Else we ignore the event.
		return HTTPResponse{
			body: "Ignoring non-actionable pull request event",
		}
	}
	return HTTPResponse{}
}

func (e *VCSEventsController) handleGitlabPost(w http.ResponseWriter, r *http.Request) {
	event, err := e.GitlabRequestParserValidator.ParseAndValidate(r, e.GitlabWebhookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, "%s", err.Error())
		return
	}
	e.Logger.Debug("request valid")

	switch event := event.(type) {
	case gitlab.MergeCommentEvent:
		e.Logger.Debug("handling as comment event")
		e.HandleGitlabCommentEvent(w, event)
	case gitlab.MergeEvent:
		e.HandleGitlabMergeRequestEvent(e.Logger, w, event)
	case gitlab.CommitCommentEvent:
		e.Logger.Debug("comments on commits are not supported, only comments on merge requests")
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment on commit event")
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event")
	}

}

// HandleGitlabCommentEvent handles comment events from GitLab where Atlantis
// commands can come from. It's exported to make testing easier.
func (e *VCSEventsController) HandleGitlabCommentEvent(w http.ResponseWriter, event gitlab.MergeCommentEvent) {
	// todo: can gitlab return the pull request here too?
	baseRepo, headRepo, commentID, user, err := e.Parser.ParseGitlabMergeRequestCommentEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}
	resp := e.handleCommentEvent(e.Logger, baseRepo, &headRepo, nil, user, event.MergeRequest.IID, event.ObjectAttributes.Note, int64(commentID), models.Gitlab)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

func (e *VCSEventsController) handleCommentEvent(logger logging.SimpleLogging, baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, comment string, commentID int64, vcsHost models.VCSHostType) HTTPResponse {
	logger = logger.WithHistory(
		"repo", baseRepo.FullName,
		"pull", pullNum,
	)

	parseResult := e.CommentParser.Parse(comment, vcsHost)
	if parseResult.Ignore {
		truncated := comment
		truncateLen := 40
		if len(truncated) > truncateLen {
			truncated = comment[:truncateLen] + "..."
		}
		logger.Debug("Ignoring non-command comment: '%s'", truncated)
		return HTTPResponse{
			body: fmt.Sprintf("Ignoring non-command comment: %q", truncated),
		}
	}
	if parseResult.Command != nil {
		logger.Info("Handling '%s' comment", parseResult.Command.Name)
	}

	// At this point we know it's a command we're not supposed to ignore, so now
	// we check if this repo is allowed to run commands in the first place.
	if !e.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		e.commentNotAllowlisted(baseRepo, pullNum)

		err := errors.New("Repo not allowlisted")

		return HTTPResponse{
			body: err.Error(),
			err: HTTPError{
				err:        err,
				code:       http.StatusForbidden,
				isSilenced: e.SilenceAllowlistErrors,
			},
		}
	}

	// It's a comment we're going to react to so add a reaction.
	if e.EmojiReaction != "" {
		err := e.VCSClient.ReactToComment(logger, baseRepo, pullNum, commentID, e.EmojiReaction)
		if err != nil {
			logger.Warn("Failed to react to comment: %s", err)
		}
	}

	// If the command isn't valid or doesn't require processing, ex.
	// "atlantis help" then we just comment back immediately.
	// We do this here rather than earlier because we need access to the pull
	// variable to comment back on the pull request.
	if parseResult.CommentResponse != "" {
		if err := e.VCSClient.CreateComment(logger, baseRepo, pullNum, parseResult.CommentResponse, ""); err != nil {
			logger.Err("Unable to comment on pull request: %s", err)
		}
		return HTTPResponse{
			body: "Commenting back on pull request",
		}
	}
	if parseResult.Command.RepoRelDir != "" {
		logger.Info("Running comment command '%v' on dir '%v' for user '%v'.",
			parseResult.Command.Name, parseResult.Command.RepoRelDir, user.Username)
	} else {
		logger.Info("Running comment command '%v' for user '%v'.", parseResult.Command.Name, user.Username)
	}
	if !e.TestingMode {
		// Respond with success and then actually execute the command asynchronously.
		// We use a goroutine so that this function returns and the connection is
		// closed.
		go e.CommandRunner.RunCommentCommand(baseRepo, maybeHeadRepo, maybePull, user, pullNum, parseResult.Command)
	} else {
		// When testing we want to wait for everything to complete.
		e.CommandRunner.RunCommentCommand(baseRepo, maybeHeadRepo, maybePull, user, pullNum, parseResult.Command)
	}

	return HTTPResponse{
		body: "Processing...",
	}
}

// HandleGitlabMergeRequestEvent will delete any locks associated with the pull
// request if the event is a merge request closed event. It's exported to make
// testing easier.
func (e *VCSEventsController) HandleGitlabMergeRequestEvent(logger logging.SimpleLogging, w http.ResponseWriter, event gitlab.MergeEvent) {
	pull, pullEventType, baseRepo, headRepo, user, err := e.Parser.ParseGitlabMergeRequestEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}

	// Annotate logger with repo and pull/merge request number.
	logger = logger.With(
		"repo", baseRepo.FullName,
		"pull", strconv.Itoa(pull.Num),
	)
	logger.Info("Processing Gitlab merge request '%s' event", pullEventType.String())
	resp := e.handlePullRequestEvent(logger, baseRepo, headRepo, pull, user, pullEventType)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

// HandleAzureDevopsPullRequestCommentedEvent handles comment events from Azure DevOps where Atlantis
// commands can come from. It's exported to make testing easier.
// Sometimes we may want data from the parent azuredevops.Event struct, so we handle type checking here.
// Requires Resource Version 2.0 of the Pull Request Commented On webhook payload.
func (e *VCSEventsController) HandleAzureDevopsPullRequestCommentedEvent(w http.ResponseWriter, event *azuredevops.Event, azuredevopsReqID string) {
	resource, ok := event.Resource.(*azuredevops.GitPullRequestWithComment)
	if !ok || event.PayloadType != azuredevops.PullRequestCommentedEvent {
		e.respond(w, logging.Error, http.StatusBadRequest, "Event.Resource is nil or received bad event type %v; %s", event.Resource, azuredevopsReqID)
		return
	}

	if resource.Comment == nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since no comment is linked to payload; %s", azuredevopsReqID)
		return
	}

	if resource.Comment.GetIsDeleted() {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since it is linked to deleting a pull request comment; %s", azuredevopsReqID)
		return
	}

	strippedComment := bluemonday.StrictPolicy().SanitizeBytes([]byte(*resource.Comment.Content))

	if resource.PullRequest == nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since no pull request is linked to payload; %s", azuredevopsReqID)
		return
	}

	if isAzureDevOpsTestRepoURL(resource.PullRequest.GetRepository()) {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring Azure DevOps Test Event with Repo URL: %v %s", resource.PullRequest.Repository.URL, azuredevopsReqID)
		return
	}

	createdBy := resource.PullRequest.GetCreatedBy()
	user := models.User{Username: createdBy.GetUniqueName()}
	baseRepo, err := e.Parser.ParseAzureDevopsRepo(resource.PullRequest.GetRepository())
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull request repository field: %s; %s", err, azuredevopsReqID)
		return
	}
	resp := e.handleCommentEvent(e.Logger, baseRepo, nil, nil, user, resource.PullRequest.GetPullRequestID(), string(strippedComment), -1, models.AzureDevops)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

// HandleAzureDevopsPullRequestEvent will delete any locks associated with the pull
// request if the event is a pull request closed event. It's exported to make
// testing easier.
func (e *VCSEventsController) HandleAzureDevopsPullRequestEvent(w http.ResponseWriter, event *azuredevops.Event, azuredevopsReqID string) {
	prText := event.Message.GetText()
	ignoreEvents := []string{
		"changed the reviewer list",
		"approved pull request",
		"has approved and left suggestions",
		"is waiting for the author",
		"rejected pull request",
		"voted on pull request",
	}
	for _, s := range ignoreEvents {
		if strings.Contains(prText, s) {
			msg := fmt.Sprintf("pull request updated event is not a supported type [%s]", s)
			e.respond(w, logging.Debug, http.StatusOK, "%s: %s", msg, azuredevopsReqID)
			return
		}
	}

	resource, ok := event.Resource.(*azuredevops.GitPullRequest)
	if !ok || event.PayloadType != azuredevops.PullRequestEvent {
		e.respond(w, logging.Error, http.StatusBadRequest, "Event.Resource is nil or received bad event type %v; %s", event.Resource, azuredevopsReqID)
		return
	}
	if isAzureDevOpsTestRepoURL(resource.GetRepository()) {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring Azure DevOps Test Event with Repo URL: %v %s", resource.Repository.URL, azuredevopsReqID)
		return
	}

	pull, pullEventType, baseRepo, headRepo, user, err := e.Parser.ParseAzureDevopsPullEvent(*event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s", err, azuredevopsReqID)
		return
	}
	e.Logger.Info("identified event as type %q", pullEventType.String())
	resp := e.handlePullRequestEvent(e.Logger, baseRepo, headRepo, pull, user, pullEventType)

	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug
	code := http.StatusOK
	msg := resp.body
	if resp.err.code != 0 {
		lvl = logging.Error
		code = resp.err.code
		msg = resp.err.err.Error()
	}
	e.respond(w, lvl, code, "%s", msg)
}

// supportsHost returns true if h is in e.SupportedVCSHosts and false otherwise.
func (e *VCSEventsController) supportsHost(h models.VCSHostType) bool {
	for _, supported := range e.SupportedVCSHosts {
		if h == supported {
			return true
		}
	}
	return false
}

func (e *VCSEventsController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	e.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}

// commentNotAllowlisted comments on the pull request that the repo is not
// allowlisted unless allowlist error comments are disabled.
func (e *VCSEventsController) commentNotAllowlisted(baseRepo models.Repo, pullNum int) {
	if e.SilenceAllowlistErrors {
		return
	}

	errMsg := "```\nError: This repo is not allowlisted for Atlantis.\n```"
	if err := e.VCSClient.CreateComment(e.Logger, baseRepo, pullNum, errMsg, ""); err != nil {
		e.Logger.Err("unable to comment on pull request: %s", err)
	}
}

func isAzureDevOpsTestRepoURL(repository *azuredevops.GitRepository) bool {
	if repository == nil {
		return false
	}
	return repository.GetURL() == azuredevopsTestURL
}
