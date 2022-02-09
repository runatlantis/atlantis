package server_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
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

func setupJobsRouter(t *testing.T) *server.Router {
	atlantisURL, err := server.ParseAtlantisURL("http://localhost:4141")
	Ok(t, err)

	underlyingRouter := mux.NewRouter()
	underlyingRouter.HandleFunc("/jobs/{job-id}", func(_ http.ResponseWriter, _ *http.Request) {}).Methods("GET").Name("project-jobs-detail")

	return &server.Router{
		AtlantisURL:              atlantisURL,
		Underlying:               underlyingRouter,
		ProjectJobsViewRouteName: "project-jobs-detail",
	}
}

func TestGenerateProjectJobURL_ShouldGenerateURLWhenJobIDSpecified(t *testing.T) {
	router := setupJobsRouter(t)
	jobID := uuid.New().String()
	ctx := models.ProjectCommandContext{
		JobID: jobID,
	}
	expectedURL := fmt.Sprintf("http://localhost:4141/jobs/%s", jobID)
	gotURL, err := router.GenerateProjectJobURL(ctx)
	Ok(t, err)

	Equals(t, expectedURL, gotURL)
}

func TestGenerateProjectJobURL_ShouldReturnErrorWhenJobIDNotSpecified(t *testing.T) {
	router := setupJobsRouter(t)
	ctx := models.ProjectCommandContext{
		Pull: models.PullRequest{
			BaseRepo: models.Repo{
				Owner: "test-owner",
				Name:  "test-repo",
			},
			Num: 1,
		},
		RepoRelDir: "ops/terraform/",
	}
	expectedErrString := "no job id in ctx"
	gotURL, err := router.GenerateProjectJobURL(ctx)
	assert.EqualError(t, err, expectedErrString)
	Equals(t, "", gotURL)
}
