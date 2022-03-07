package server

import (
	"fmt"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// Router can be used to retrieve Atlantis URLs. It acts as an intermediary
// between the underlying router and the rest of Atlantis that might need to
// construct URLs to different resources.
type Router struct {
	// Underlying is the router that the routes have been constructed on.
	Underlying *mux.Router
	// LockViewRouteName is the named route for the lock view that can be Get'd
	// from the Underlying router.
	LockViewRouteName string
	// ProjectJobsViewRouteName is the named route for the projects active jobs
	ProjectJobsViewRouteName string
	// LockViewRouteIDQueryParam is the query parameter needed to construct the
	// lock view: underlying.Get(LockViewRouteName).URL(LockViewRouteIDQueryParam, "my id").
	LockViewRouteIDQueryParam string
	// AtlantisURL is the fully qualified URL that Atlantis is
	// accessible from externally.
	AtlantisURL *url.URL
}

// GenerateLockURL returns a fully qualified URL to view the lock at lockID.
func (r *Router) GenerateLockURL(lockID string) string {
	lockURL, _ := r.Underlying.Get(r.LockViewRouteName).URL(r.LockViewRouteIDQueryParam, url.QueryEscape(lockID))
	// At this point, lockURL will just be a path because r.Underlying isn't
	// configured with host or scheme information. So to generate the fully
	// qualified LockURL we just append the router's url to our base url.
	// We're not doing anything fancy here with the actual url object because
	// golang likes to double escape the lockURL path when using url.Parse().
	return r.AtlantisURL.String() + lockURL.String()
}

func (r *Router) GenerateProjectJobURL(ctx models.ProjectCommandContext) (string, error) {
	if ctx.JobID == "" {
		return "", fmt.Errorf("no job id in ctx")
	}
	jobURL, err := r.Underlying.Get((r.ProjectJobsViewRouteName)).URL(
		"job-id", ctx.JobID,
	)
	if err != nil {
		return "", errors.Wrapf(err, "creating job url for %s", ctx.JobID)
	}

	return r.AtlantisURL.String() + jobURL.String(), nil
}
