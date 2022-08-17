package activities

import (
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/github"
	"github.com/uber-go/tally/v4"
)

// Exported Activites should be here.
// The convention should be one exported struct per workflow
// This guarantees function naming uniqueness within a given workflow
// which is a requirement at a per worker level
//
// Note: This doesn't prevent issues with naming duplication that can come up when
// registering multiple workflows to the same worker
type Deploy struct {
	*dbActivities
	*githubActivities
}

func NewDeploy(config githubapp.Config, scope tally.Scope) (*Deploy, error) {
	clientCreator, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientMiddleware(
			github.ClientMetrics(scope),
		))

	if err != nil {
		return nil, errors.Wrap(err, "initializing client creator")
	}

	return &Deploy{
		dbActivities: &dbActivities{},
		githubActivities: &githubActivities{
			ClientCreator: clientCreator,
		},
	}, nil
}
