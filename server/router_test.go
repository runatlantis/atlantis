package server_test

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRouter_GenerateLockURL(t *testing.T) {
	queryParam := "queryparam"
	routeName := "routename"
	atlantisURL := "https://example.com"

	underlyingRouter := mux.NewRouter()
	underlyingRouter.HandleFunc("/lock", func(_ http.ResponseWriter, _ *http.Request) {}).Methods("GET").Queries(queryParam, "{queryparam}").Name(routeName)

	router := &server.Router{
		AtlantisURL:               atlantisURL,
		LockViewRouteIDQueryParam: queryParam,
		LockViewRouteName:         routeName,
		Underlying:                underlyingRouter,
	}
	Equals(t, "https://example.com/lock?queryparam=myid", router.GenerateLockURL("myid"))
}
