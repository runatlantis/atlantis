package server_test

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRouter_GenerateLockURL(t *testing.T) {
	cases := []struct {
		AtlantisURL string
		ExpURL      string
	}{
		{
			"http://localhost:4141",
			"http://localhost:4141/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
		},
		{
			"https://localhost:4141",
			"https://localhost:4141/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
		},
		{
			"https://localhost:4141/",
			"https://localhost:4141/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
		},
		{
			"https://example.com/basepath",
			"https://example.com/basepath/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
		},
		{
			"https://example.com/basepath/",
			"https://example.com/basepath/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
		},
		{
			"https://example.com/path/1/",
			"https://example.com/path/1/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
		},
	}

	queryParam := "id"
	routeName := "routename"
	underlyingRouter := mux.NewRouter()
	underlyingRouter.HandleFunc("/lock", func(_ http.ResponseWriter, _ *http.Request) {}).Methods("GET").Queries(queryParam, "{id}").Name(routeName)

	for _, c := range cases {
		t.Run(c.AtlantisURL, func(t *testing.T) {
			atlantisURL, err := server.ParseAtlantisURL(c.AtlantisURL)
			Ok(t, err)

			router := &server.Router{
				AtlantisURL:               atlantisURL,
				LockViewRouteIDQueryParam: queryParam,
				LockViewRouteName:         routeName,
				Underlying:                underlyingRouter,
			}
			Equals(t, c.ExpURL, router.GenerateLockURL("lkysow/atlantis-example/./default"))
		})
	}
}
