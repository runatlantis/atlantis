package github

import (
	"context"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

type ExternalRepo interface {
	GetFullName() string
	GetCloneURL() string
	GetDefaultBranch() string
}

// without this we have an import cycle
type repoConverter interface {
	Convert(r ExternalRepo) (models.Repo, error)
}

type RepoRetriever struct {
	ClientCreator githubapp.ClientCreator
	RepoConverter repoConverter
}

func (r *RepoRetriever) Get(ctx context.Context, installationToken int64, owner, repo string) (models.Repo, error) {
	installationClient, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return models.Repo{}, errors.Wrap(err, "creating installation client")
	}

	repository, _, err := installationClient.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return models.Repo{}, errors.Wrapf(err, "getting repository")
	}

	result, err := r.RepoConverter.Convert(repository)
	if err != nil {
		return models.Repo{}, errors.Wrapf(err, "converting repository")
	}

	return result, nil
}
