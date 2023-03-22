package gateway

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	lyft_gateway "github.com/runatlantis/atlantis/server/lyft/gateway"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api"
	apiMiddleware "github.com/runatlantis/atlantis/server/neptune/gateway/api/middleware"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/request"
	commonMiddleware "github.com/runatlantis/atlantis/server/neptune/gateway/middleware"
)

func newRouter(
	logger logging.Logger,
	eventsController *lyft_gateway.VCSEventsController,
	statusController *controllers.StatusController,
	deployController *api.Controller[request.Deploy],
	globalCfg valid.GlobalCfg,
) *mux.Router {
	recovery := &commonMiddleware.Recovery{
		Logger: logger,
	}
	logging := &commonMiddleware.Logger{
		Logger: logger,
	}
	requestID := &commonMiddleware.RequestID{}

	router := mux.NewRouter()
	router.Use(requestID.Middleware, logging.Middleware, recovery.Middleware)
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/status", statusController.Get).Methods("GET")
	router.HandleFunc("/events", eventsController.Post).Methods("POST")
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)

	apiSubrouter := router.PathPrefix("/api/admin").Subrouter()
	auth := &apiMiddleware.AdminAuth{
		Admin: globalCfg.Admin,
	}

	apiSubrouter.Use(auth.Middleware)
	apiSubrouter.HandleFunc("/deploy", deployController.Handle).Methods("POST")

	return router
}
