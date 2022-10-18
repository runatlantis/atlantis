package workflows

import (
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	internal "github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/github"
	"github.com/uber-go/tally/v4"
)

type GithubActivities struct {
	*activities.Github
}

func NewGithubActivities(appConfig githubapp.Config, scope tally.Scope, dataDir string) (*GithubActivities, error) {
	clientCreator, err := githubapp.NewDefaultCachingClientCreator(
		appConfig,
		githubapp.WithClientMiddleware(
			github.ClientMetrics(scope.SubScope("app")),
		))

	if err != nil {
		return nil, errors.Wrap(err, "initializing client creator")
	}

	client := &internal.Client{
		ClientCreator: clientCreator,
	}

	githubActivities, err := activities.NewGithub(client, dataDir, activities.HashiGetter)
	if err != nil {
		return nil, errors.Wrap(err, "initializing github activities")
	}

	return &GithubActivities{
		Github: githubActivities,
	}, nil
}
