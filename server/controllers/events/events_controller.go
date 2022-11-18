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
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	httputils "github.com/runatlantis/atlantis/server/http"

	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	requestErrors "github.com/runatlantis/atlantis/server/controllers/events/errors"
	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/logging"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
	github_converter "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	github_request "github.com/runatlantis/atlantis/server/vcs/provider/github/request"
	"github.com/uber-go/tally/v4"
	gitlab "github.com/xanzy/go-gitlab"
)

const githubHeader = "X-Github-Event"
const gitlabHeader = "X-Gitlab-Event"
const azuredevopsHeader = "Request-Id"

// bitbucketEventTypeHeader is the same in both cloud and server.
const bitbucketEventTypeHeader = "X-Event-Key"
const bitbucketCloudRequestIDHeader = "X-Request-UUID"
const bitbucketServerRequestIDHeader = "X-Request-ID"
const bitbucketServerSignatureHeader = "X-Hub-Signature"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_azuredevops_pull_getter.go AzureDevopsPullGetter

// AzureDevopsPullGetter makes API calls to get pull requests.
type AzureDevopsPullGetter interface {
	// GetPullRequest gets the pull request with id pullNum for the repo.
	GetPullRequest(repo models.Repo, pullNum int) (*azuredevops.GitPullRequest, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_gitlab_merge_request_getter.go GitlabMergeRequestGetter

// GitlabMergeRequestGetter makes API calls to get merge requests.
type GitlabMergeRequestGetter interface {
	// GetMergeRequest gets the pull request with the id pullNum for the repo.
	GetMergeRequest(repoFullName string, pullNum int) (*gitlab.MergeRequest, error)
}

type commentEventHandler interface {
	Handle(ctx context.Context, request *httputils.BufferedRequest, event event_types.Comment) error
}

type prEventHandler interface {
	Handle(ctx context.Context, request *httputils.BufferedRequest, event event_types.PullRequest) error
}

type unsupportedPushEventHandler struct{}

func (h unsupportedPushEventHandler) Handle(ctx context.Context, event event_types.Push) error {
	return fmt.Errorf("push events are not supported in this context")
}

type unsupportedCheckRunEventHandler struct{}

func (h unsupportedCheckRunEventHandler) Handle(ctx context.Context, event event_types.CheckRun) error {
	return fmt.Errorf("check run events are not supported in this context")
}

type unsupportedCheckSuiteEventHandler struct{}

func (h unsupportedCheckSuiteEventHandler) Handle(ctx context.Context, event event_types.CheckSuite) error {
	return fmt.Errorf("check suite events are not supported in this context")
}

func NewRequestResolvers(
	providerResolverInitializer map[models.VCSHostType]func() RequestResolver,
	supportedProviders []models.VCSHostType,
) []RequestResolver {
	var resolvers []RequestResolver
	for provider, resolverInitializer := range providerResolverInitializer {
		for _, supportedProvider := range supportedProviders {

			if provider != supportedProvider {
				continue
			}

			resolvers = append(resolvers, resolverInitializer())
		}
	}

	return resolvers
}

func NewVCSEventsController(
	scope tally.Scope,
	githubWebhookSecret []byte,
	allowDraftPRs bool,
	commandRunner events.CommandRunner,
	commentParser events.CommentParsing,
	eventParser events.EventParsing,
	pullCleaner events.PullCleaner,
	repoAllowlistChecker *events.RepoAllowlistChecker,
	vcsClient vcs.Client,
	logger logging.Logger,
	applyDisabled bool,
	gitlabWebhookSecret []byte,
	supportedVCSProviders []models.VCSHostType,
	bitbucketWebhookSecret []byte,
	azureDevopsWebhookBasicUser []byte,
	azureDevopsWebhookBasicPassword []byte,
	repoConverter github_converter.RepoConverter,
	pullConverter github_converter.PullConverter,
	githubPullGetter github_converter.PullGetter,
	azureDevopsPullGetter AzureDevopsPullGetter,
	gitlabMergeRequestGetter GitlabMergeRequestGetter,
) *VCSEventsController {

	prHandler := handlers.NewPullRequestEvent(
		repoAllowlistChecker, pullCleaner, logger, commandRunner,
	)

	commentHandler := handlers.NewCommentEvent(
		commentParser,
		repoAllowlistChecker,
		vcsClient,
		commandRunner,
		logger,
	)

	// we don't support push events in the atlantis worker and these should never make it in the queue
	// in the first place, so if it happens, let's return an error and fail fast.
	pushHandler := unsupportedPushEventHandler{}

	// lazy map of resolver providers to their resolver
	// laziness ensures we only instantiate the providers we support.
	providerResolverInitializer := map[models.VCSHostType]func() RequestResolver{
		models.Github: func() RequestResolver {
			return github_request.NewHandler(
				logger,
				scope,
				githubWebhookSecret,
				commentHandler,
				prHandler,
				pushHandler,
				unsupportedCheckRunEventHandler{},
				unsupportedCheckSuiteEventHandler{},
				allowDraftPRs,
				repoConverter,
				pullConverter,
				githubPullGetter,
			)
		},
	}

	router := &RequestRouter{
		Resolvers: NewRequestResolvers(providerResolverInitializer, supportedVCSProviders),
	}

	return &VCSEventsController{
		RequestRouter:                   router,
		Logger:                          logger,
		Scope:                           scope,
		Parser:                          eventParser,
		CommentParser:                   commentParser,
		PREventHandler:                  prHandler,
		CommentEventHandler:             commentHandler,
		ApplyDisabled:                   applyDisabled,
		GitlabRequestParserValidator:    &DefaultGitlabRequestParserValidator{},
		GitlabWebhookSecret:             gitlabWebhookSecret,
		RepoAllowlistChecker:            repoAllowlistChecker,
		SupportedVCSHosts:               supportedVCSProviders,
		VCSClient:                       vcsClient,
		BitbucketWebhookSecret:          bitbucketWebhookSecret,
		AzureDevopsWebhookBasicUser:     azureDevopsWebhookBasicUser,
		AzureDevopsWebhookBasicPassword: azureDevopsWebhookBasicPassword,
		AzureDevopsRequestValidator:     &DefaultAzureDevopsRequestValidator{},
		AzureDevopsPullGetter:           azureDevopsPullGetter,
		GitlabMergeRequestGetter:        gitlabMergeRequestGetter,
	}
}

type RequestHandler interface {
	Handle(request *httputils.BufferedRequest) error
}

type RequestMatcher interface {
	Matches(request *httputils.BufferedRequest) bool
}

type RequestResolver interface {
	RequestHandler
	RequestMatcher
}

// TODO: once VCSEventsController is fully broken down this implementation can just live in there.
type RequestRouter struct {
	Resolvers []RequestResolver
}

func (p *RequestRouter) Route(w http.ResponseWriter, r *http.Request) {
	// we do this to allow for multiple reads to the request body
	request, err := httputils.NewBufferedRequest(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		return
	}

	for _, resolver := range p.Resolvers {
		if !resolver.Matches(request) {
			continue
		}

		err := resolver.Handle(request)

		if e, ok := err.(*requestErrors.RequestValidationError); ok {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, e.Error())
			return
		}

		if e, ok := err.(*requestErrors.WebhookParsingError); ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, e.Error())
			return
		}

		if e, ok := err.(*requestErrors.EventParsingError); ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, e.Error())
			return
		}

		if e, ok := err.(*requestErrors.UnsupportedEventTypeError); ok {
			// historically we've just ignored these so for now let's just do that.
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, e.Error())
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Processing...")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "no resolver configured for request")
}

// VCSEventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub.
// TODO: migrate all provider specific request handling into packaged resolver similar to github
type VCSEventsController struct {
	Logger                       logging.Logger
	Scope                        tally.Scope
	CommentParser                events.CommentParsing
	Parser                       events.EventParsing
	PREventHandler               prEventHandler
	CommentEventHandler          commentEventHandler
	RequestRouter                *RequestRouter
	ApplyDisabled                bool
	GitlabRequestParserValidator GitlabRequestParserValidator
	// GitlabWebhookSecret is the secret added to this webhook via the GitLab
	// UI that identifies this call as coming from GitLab. If empty, no
	// request validation is done.
	GitlabWebhookSecret  []byte
	RepoAllowlistChecker *events.RepoAllowlistChecker
	// SupportedVCSHosts is which VCS hosts Atlantis was configured upon
	// startup to support.
	SupportedVCSHosts []models.VCSHostType
	VCSClient         vcs.Client
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

	GitlabMergeRequestGetter GitlabMergeRequestGetter
	AzureDevopsPullGetter    AzureDevopsPullGetter
}

// Post handles POST webhook requests.
func (e *VCSEventsController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(githubHeader) != "" {
		e.RequestRouter.Route(w, r)
		return
	} else if r.Header.Get(gitlabHeader) != "" {
		if !e.supportsHost(models.Gitlab) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support GitLab")
			return
		}
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
			e.handleBitbucketCloudPost(w, r)
			return
		} else if r.Header.Get(bitbucketServerRequestIDHeader) != "" {
			if !e.supportsHost(models.BitbucketServer) {
				e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support Bitbucket Server")
				return
			}
			e.handleBitbucketServerPost(w, r)
			return
		}
	} else if r.Header.Get(azuredevopsHeader) != "" {
		if !e.supportsHost(models.AzureDevops) {
			e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request since not configured to support AzureDevops")
			return
		}
		e.handleAzureDevopsPost(w, r)
		return
	}
	e.respond(w, logging.Debug, http.StatusBadRequest, "Ignoring request")
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
		e.handleBitbucketCloudPullRequestEvent(w, eventType, body, reqID, r)
		return
	case bitbucketcloud.PullCommentCreatedHeader:
		e.HandleBitbucketCloudCommentEvent(w, body, reqID, r)
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
			e.respond(w, logging.Warn, http.StatusBadRequest, errors.Wrap(err, "request did not pass validation").Error())
			return
		}
	}
	switch eventType {
	case bitbucketserver.PullCreatedHeader, bitbucketserver.PullMergedHeader, bitbucketserver.PullDeclinedHeader, bitbucketserver.PullDeletedHeader:
		e.handleBitbucketServerPullRequestEvent(w, eventType, body, reqID, r)
		return
	case bitbucketserver.PullCommentCreatedHeader:
		e.HandleBitbucketServerCommentEvent(w, body, reqID, r)
		return
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event type %s %s=%s", eventType, bitbucketServerRequestIDHeader, reqID)
	}
}

func (e *VCSEventsController) handleAzureDevopsPost(w http.ResponseWriter, r *http.Request) {
	// Validate the request against the optional basic auth username and password.
	payload, err := e.AzureDevopsRequestValidator.Validate(r, e.AzureDevopsWebhookBasicUser, e.AzureDevopsWebhookBasicPassword)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusUnauthorized, err.Error())
		return
	}

	azuredevopsReqID := "Request-Id=" + r.Header.Get("Request-Id")
	event, err := azuredevops.ParseWebHook(payload)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Failed parsing webhook: %v %s", err, azuredevopsReqID)
		return
	}
	switch event.PayloadType {
	case azuredevops.PullRequestCommentedEvent:
		e.HandleAzureDevopsPullRequestCommentedEvent(w, event, azuredevopsReqID, r)
	case azuredevops.PullRequestEvent:
		e.HandleAzureDevopsPullRequestEvent(w, event, azuredevopsReqID, r)
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event: %v %s", event.PayloadType, azuredevopsReqID)
	}
}

// HandleBitbucketCloudCommentEvent handles comment events from Bitbucket.
func (e *VCSEventsController) HandleBitbucketCloudCommentEvent(w http.ResponseWriter, body []byte, reqID string, request *http.Request) {
	pull, baseRepo, headRepo, user, comment, err := e.Parser.ParseBitbucketCloudPullCommentEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	eventTimestamp := time.Now()
	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.CommentEventHandler.Handle(context.TODO(), cloneableRequest, event_types.Comment{
		BaseRepo:  baseRepo,
		HeadRepo:  headRepo,
		Pull:      pull,
		User:      user,
		PullNum:   pull.Num,
		Comment:   comment,
		VCSHost:   models.BitbucketCloud,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}

	e.respond(w, lvl, http.StatusOK, err.Error())
}

// HandleBitbucketServerCommentEvent handles comment events from Bitbucket.
func (e *VCSEventsController) HandleBitbucketServerCommentEvent(w http.ResponseWriter, body []byte, reqID string, request *http.Request) {
	pull, baseRepo, headRepo, user, comment, err := e.Parser.ParseBitbucketServerPullCommentEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	eventTimestamp := time.Now()
	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.CommentEventHandler.Handle(context.TODO(), cloneableRequest, event_types.Comment{
		BaseRepo:  baseRepo,
		HeadRepo:  headRepo,
		Pull:      pull,
		User:      user,
		PullNum:   pull.Num,
		Comment:   comment,
		VCSHost:   models.BitbucketServer,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}

	e.respond(w, lvl, http.StatusOK, "")
}

func (e *VCSEventsController) handleBitbucketCloudPullRequestEvent(w http.ResponseWriter, eventType string, body []byte, reqID string, request *http.Request) {
	pull, _, _, user, err := e.Parser.ParseBitbucketCloudPullEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketCloudRequestIDHeader, reqID)
		return
	}
	pullEventType := e.Parser.GetBitbucketCloudPullEventType(eventType)
	e.Logger.Info(fmt.Sprintf("identified event as type %q", pullEventType.String()))
	eventTimestamp := time.Now()
	//TODO: move this to the outer most function similar to github
	lvl := logging.Debug

	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.PREventHandler.Handle(context.TODO(), cloneableRequest, event_types.PullRequest{

		Pull:      pull,
		User:      user,
		EventType: pullEventType,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	e.respond(w, lvl, http.StatusOK, "")
}

func (e *VCSEventsController) handleBitbucketServerPullRequestEvent(w http.ResponseWriter, eventType string, body []byte, reqID string, request *http.Request) {
	pull, _, _, user, err := e.Parser.ParseBitbucketServerPullEvent(body)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s=%s", err, bitbucketServerRequestIDHeader, reqID)
		return
	}
	pullEventType := e.Parser.GetBitbucketServerPullEventType(eventType)
	e.Logger.Info(fmt.Sprintf("identified event as type %q", pullEventType.String()))
	eventTimestamp := time.Now()
	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.PREventHandler.Handle(context.TODO(), cloneableRequest, event_types.PullRequest{
		Pull:      pull,
		User:      user,
		EventType: pullEventType,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	e.respond(w, lvl, http.StatusOK, "Processing...")
}

func (e *VCSEventsController) handleGitlabPost(w http.ResponseWriter, r *http.Request) {
	event, err := e.GitlabRequestParserValidator.ParseAndValidate(r, e.GitlabWebhookSecret)
	if err != nil {
		e.respond(w, logging.Warn, http.StatusBadRequest, err.Error())
		return
	}

	switch event := event.(type) {
	case gitlab.MergeCommentEvent:
		e.HandleGitlabCommentEvent(w, event, r)
	case gitlab.MergeEvent:
		e.HandleGitlabMergeRequestEvent(w, event, r)
	case gitlab.CommitCommentEvent:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment on commit event")
	default:
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event")
	}

}

// HandleGitlabCommentEvent handles comment events from GitLab where Atlantis
// commands can come from. It's exported to make testing easier.
func (e *VCSEventsController) HandleGitlabCommentEvent(w http.ResponseWriter, event gitlab.MergeCommentEvent, request *http.Request) {
	// todo: can gitlab return the pull request here too?
	baseRepo, headRepo, user, err := e.Parser.ParseGitlabMergeRequestCommentEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}

	pull, err := e.getGitlabData(baseRepo, event.MergeRequest.ID)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error getting merge request: %s", err)
		return
	}
	eventTimestamp := time.Now()
	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.CommentEventHandler.Handle(context.TODO(), cloneableRequest, event_types.Comment{
		BaseRepo:  baseRepo,
		HeadRepo:  headRepo,
		Pull:      pull,
		User:      user,
		PullNum:   event.MergeRequest.IID,
		Comment:   event.ObjectAttributes.Note,
		VCSHost:   models.Gitlab,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}

	e.respond(w, lvl, http.StatusOK, "Processing...")
}

// HandleGitlabMergeRequestEvent will delete any locks associated with the pull
// request if the event is a merge request closed event. It's exported to make
// testing easier.
func (e *VCSEventsController) HandleGitlabMergeRequestEvent(w http.ResponseWriter, event gitlab.MergeEvent, request *http.Request) {
	pull, pullEventType, _, _, user, err := e.Parser.ParseGitlabMergeRequestEvent(event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing webhook: %s", err)
		return
	}
	e.Logger.Info(fmt.Sprintf("identified event as type %q", pullEventType.String()))
	eventTimestamp := time.Now()

	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.PREventHandler.Handle(context.TODO(), cloneableRequest, event_types.PullRequest{
		Pull:      pull,
		User:      user,
		EventType: pullEventType,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	e.respond(w, lvl, http.StatusOK, "Processing...")
}

// HandleAzureDevopsPullRequestCommentedEvent handles comment events from Azure DevOps where Atlantis
// commands can come from. It's exported to make testing easier.
// Sometimes we may want data from the parent azuredevops.Event struct, so we handle type checking here.
// Requires Resource Version 2.0 of the Pull Request Commented On webhook payload.
func (e *VCSEventsController) HandleAzureDevopsPullRequestCommentedEvent(w http.ResponseWriter, event *azuredevops.Event, azuredevopsReqID string, request *http.Request) {
	resource, ok := event.Resource.(*azuredevops.GitPullRequestWithComment)
	if !ok || event.PayloadType != azuredevops.PullRequestCommentedEvent {
		e.respond(w, logging.Error, http.StatusBadRequest, "Event.Resource is nil or received bad event type %v; %s", event.Resource, azuredevopsReqID)
		return
	}

	if resource.Comment == nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since no comment is linked to payload; %s", azuredevopsReqID)
		return
	}
	strippedComment := bluemonday.StrictPolicy().SanitizeBytes([]byte(*resource.Comment.Content))

	if resource.PullRequest == nil {
		e.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since no pull request is linked to payload; %s", azuredevopsReqID)
		return
	}

	createdBy := resource.PullRequest.GetCreatedBy()
	user := models.User{Username: createdBy.GetUniqueName()}
	baseRepo, err := e.Parser.ParseAzureDevopsRepo(resource.PullRequest.GetRepository())
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull request repository field: %s; %s", err, azuredevopsReqID)
		return
	}
	pull, headRepo, err := e.getAzureDevopsData(baseRepo, resource.PullRequest.GetPullRequestID())
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error getting pull request %s", err)
		return
	}

	eventTimestamp := time.Now()
	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.CommentEventHandler.Handle(context.TODO(), cloneableRequest, event_types.Comment{
		BaseRepo:  baseRepo,
		HeadRepo:  headRepo,
		Pull:      pull,
		User:      user,
		PullNum:   resource.PullRequest.GetPullRequestID(),
		Comment:   string(strippedComment),
		VCSHost:   models.AzureDevops,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}

	e.respond(w, lvl, http.StatusOK, "Processing...")
}

// HandleAzureDevopsPullRequestEvent will delete any locks associated with the pull
// request if the event is a pull request closed event. It's exported to make
// testing easier.
func (e *VCSEventsController) HandleAzureDevopsPullRequestEvent(w http.ResponseWriter, event *azuredevops.Event, azuredevopsReqID string, request *http.Request) {
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

	pull, pullEventType, _, _, user, err := e.Parser.ParseAzureDevopsPullEvent(*event)
	if err != nil {
		e.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s %s", err, azuredevopsReqID)
		return
	}
	e.Logger.Info(fmt.Sprintf("identified event as type %q", pullEventType.String()))
	eventTimestamp := time.Now()
	lvl := logging.Debug
	cloneableRequest, err := httputils.NewBufferedRequest(request)
	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	err = e.PREventHandler.Handle(context.TODO(), cloneableRequest, event_types.PullRequest{
		Pull:      pull,
		User:      user,
		EventType: pullEventType,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		e.respond(w, lvl, http.StatusInternalServerError, err.Error())
		return
	}
	e.respond(w, lvl, http.StatusOK, "Processing...")
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
	switch lvl {
	case logging.Error:
		e.Logger.Error(response)
	case logging.Info:
		e.Logger.Info(response)
	case logging.Warn:
		e.Logger.Warn(response)
	case logging.Debug:
		e.Logger.Debug(response)
	default:
		e.Logger.Error(response)
	}
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}

func (e *VCSEventsController) getGitlabData(baseRepo models.Repo, pullNum int) (models.PullRequest, error) {
	if e.GitlabMergeRequestGetter == nil {
		return models.PullRequest{}, errors.New("Atlantis not configured to support GitLab")
	}
	mr, err := e.GitlabMergeRequestGetter.GetMergeRequest(baseRepo.FullName, pullNum)
	if err != nil {
		return models.PullRequest{}, errors.Wrap(err, "making merge request API call to GitLab")
	}
	pull := e.Parser.ParseGitlabMergeRequest(mr, baseRepo)
	return pull, nil
}

func (e *VCSEventsController) getAzureDevopsData(baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if e.AzureDevopsPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("atlantis not configured to support Azure DevOps")
	}
	adPull, err := e.AzureDevopsPullGetter.GetPullRequest(baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to Azure DevOps")
	}
	pull, _, headRepo, err := e.Parser.ParseAzureDevopsPull(adPull)
	if err != nil {
		return pull, headRepo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, headRepo, nil
}
