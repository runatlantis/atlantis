package github

import (
	"context"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

type BranchRetriever struct {
	ClientCreator githubapp.ClientCreator
}

type Branch struct {
	Name     string
	Revision string
}

func (r *BranchRetriever) GetBranch(ctx context.Context, installationToken int64, owner, repo, branch string, followRedirects bool) (Branch, error) {
	client, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return Branch{}, errors.Wrap(err, "creating installation client")
	}

	b, _, err := client.Repositories.GetBranch(ctx, owner, repo, branch, followRedirects)
	if err != nil {
		return Branch{}, errors.Wrap(err, "getting branch")
	}

	branchName := b.GetName()
	revision := b.GetCommit().GetSHA()

	if len(branchName) == 0 {
		return Branch{}, errors.Wrap(err, "branch name returned is empty, this is bug with github")
	}

	if len(revision) == 0 {
		return Branch{}, errors.Wrap(err, "revision returned is empty, this is bug with github")
	}

	return Branch{
		Name:     branchName,
		Revision: revision,
	}, nil
}
