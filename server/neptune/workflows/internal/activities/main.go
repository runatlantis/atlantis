package activities

import (
	"net/url"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/github"
	repo "github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github/link"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
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
}

func NewDeploy(config githubapp.Config, scope tally.Scope) (*Deploy, error) {
	return &Deploy{
		dbActivities: &dbActivities{},
	}, nil
}

type Terraform struct {
	*terraformActivities
	*executeCommandActivities
	*workerInfoActivity
	*notifyActivities
	*cleanupActivities
}

func NewTerraform(serverURL *url.URL) *Terraform {
	return &Terraform{
		executeCommandActivities: &executeCommandActivities{},
		workerInfoActivity: &workerInfoActivity{
			ServerURL: serverURL,
		},
	}
}

type Github struct {
	*githubActivities
}

type LinkBuilder interface {
	BuildDownloadLinkFromArchive(archiveURL *url.URL, root root.Root, repo repo.Repo, revision string) string
}

func NewGithub(config githubapp.Config, scope tally.Scope, dataDir string) (*Github, error) {
	clientCreator, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientMiddleware(
			github.ClientMetrics(scope.SubScope("app")),
		))

	if err != nil {
		return nil, errors.Wrap(err, "initializing client creator")
	}
	return &Github{
		githubActivities: &githubActivities{
			ClientCreator: clientCreator,
			DataDir:       dataDir,
			LinkBuilder:   link.Builder{},
		},
	}, nil
}
