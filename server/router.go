package server

import (
	"fmt"
	"net/url"

	"github.com/gorilla/mux"
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
	// LockViewRouteIDQueryParam is the query parameter needed to construct the
	// lock view: underlying.Get(LockViewRouteName).URL(LockViewRouteIDQueryParam, "my id").
	LockViewRouteIDQueryParam string
	// AtlantisURL is the fully qualified URL (scheme included) that Atlantis is
	// being served at, ex: https://example.com.
	AtlantisURL url.URL
}

// GenerateLockURL returns a fully qualified URL to view the lock at lockID.
func (r *Router) GenerateLockURL(lockID string) string {
	path, _ := r.Underlying.Get(r.LockViewRouteName).URL(r.LockViewRouteIDQueryParam, url.QueryEscape(lockID))
	return fmt.Sprintf("%s%s", r.AtlantisURL.String(), path)
}
