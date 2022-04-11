package gateway

import (
	"net/http"

	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	gateway_handlers "github.com/runatlantis/atlantis/server/lyft/gateway/events/handlers"
	converters "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/request"
	"github.com/uber-go/tally"
)

const githubHeader = "X-Github-Event"

func NewVCSEventsController(
	legacyLogger logging.SimpleLogging,
	scope tally.Scope,
	webhookSecret []byte,
	allowDraftPRs bool,
	autoplanValidator gateway_handlers.EventValidator,
	snsWriter gateway_handlers.Writer,
	commentParser events.CommentParsing,
	repoAllowlistChecker *events.RepoAllowlistChecker,
	vcsClient vcs.Client,
	logger logging.Logger,
	supportedVCSProviders []models.VCSHostType,
	repoConverter converters.RepoConverter,
	pullConverter converters.PullConverter,
) *VCSEventsController {
	pullEventWorkerProxy := gateway_handlers.NewPullEventWorkerProxy(
		snsWriter, logger,
	)

	asyncAutoplannerWorkerProxy := gateway_handlers.NewAsynchronousAutoplannerWorkerProxy(
		autoplanValidator, logger, legacyLogger, pullEventWorkerProxy,
	)

	prHandler := handlers.NewPullRequestEventWithEventTypeHandlers(
		repoAllowlistChecker,
		asyncAutoplannerWorkerProxy,
		asyncAutoplannerWorkerProxy,
		pullEventWorkerProxy,
	)

	commentHandler := handlers.NewCommentEventWithCommandHandler(
		commentParser,
		repoAllowlistChecker,
		vcsClient,
		gateway_handlers.NewCommentEventWorkerProxy(logger, snsWriter),
		logger,
	)

	// lazy map of resolver providers to their resolver
	// laziness ensures we only instantiate the providers we support.
	providerResolverInitializer := map[models.VCSHostType]func() events_controllers.RequestResolver{
		models.Github: func() events_controllers.RequestResolver {
			return request.NewHandler(
				legacyLogger,
				scope,
				webhookSecret,
				commentHandler,
				prHandler,
				allowDraftPRs,
				repoConverter,
				pullConverter,
			)
		},
	}

	router := &events_controllers.RequestRouter{
		Resolvers: events_controllers.NewRequestResolvers(providerResolverInitializer, supportedVCSProviders),
	}

	return &VCSEventsController{
		router: router,
	}
}

// TODO: remove this once event_controllers.VCSEventsController has the same function
// VCSEventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub.
type VCSEventsController struct {
	router *events_controllers.RequestRouter
}

// Post handles POST webhook requests.
func (g *VCSEventsController) Post(w http.ResponseWriter, r *http.Request) {
	g.router.Route(w, r)
}
