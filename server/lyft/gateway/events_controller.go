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
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	converters "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/request"
	"github.com/uber-go/tally"
)

const githubHeader = "X-Github-Event"

func NewVCSEventsController(
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
	githubClient converter.PullGetter,
) *VCSEventsController {
	pullEventWorkerProxy := gateway_handlers.NewPullEventWorkerProxy(
		snsWriter, logger,
	)

	asyncAutoplannerWorkerProxy := gateway_handlers.NewAsynchronousAutoplannerWorkerProxy(
		autoplanValidator, logger, pullEventWorkerProxy,
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
				logger,
				scope,
				webhookSecret,
				commentHandler,
				prHandler,
				allowDraftPRs,
				repoConverter,
				pullConverter,
				githubClient,
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
