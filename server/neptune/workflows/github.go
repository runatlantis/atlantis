package workflows

import (
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/uber-go/tally/v4"
)

type GithubActivities struct {
	activities.Github
}

func NewGithubActivities(appConfig githubapp.Config, scope tally.Scope, dataDir string) (*GithubActivities, error) {
	githubActivities, err := activities.NewGithub(appConfig, scope, dataDir)
	if err != nil {
		return nil, errors.Wrap(err, "initializing github activities")
	}

	return &GithubActivities{
		Github: *githubActivities,
	}, nil
}
