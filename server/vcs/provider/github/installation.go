package github

import (
	"context"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

type InstallationRetriever struct {
	ClientCreator githubapp.ClientCreator
}

type Installation struct {
	Token int64
}

func (r *InstallationRetriever) FindOrganizationInstallation(ctx context.Context, org string) (Installation, error) {
	appClient, err := r.ClientCreator.NewAppClient()
	if err != nil {
		return Installation{}, errors.Wrap(err, "creating app client")
	}

	installation, _, err := appClient.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		return Installation{}, errors.Wrapf(err, "finding organization installation")
	}

	return Installation{
		Token: installation.GetID(),
	}, nil
}
