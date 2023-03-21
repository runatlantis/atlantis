package request

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/middleware"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/request/external"
	internal "github.com/runatlantis/atlantis/server/vcs/provider/github"
)

func NewDeployConverter(
	repoRetriever *internal.RepoRetriever,
	branchRetriever *internal.BranchRetriever,
	InstallationRetriever *internal.InstallationRetriever,
) *JSONRequestValidationProxy[external.DeployRequest, Deploy] {
	return &JSONRequestValidationProxy[external.DeployRequest, Deploy]{
		Delegate: &DeployConverter{
			InstallationRetriever: InstallationRetriever,
			BranchRetriever:       branchRetriever,
			RepoRetriever:         repoRetriever,
		},
	}
}

type repoRetriever interface {
	Get(ctx context.Context, installationToken int64, owner, repo string) (models.Repo, error)
}

type branchRetriever interface {
	GetBranch(ctx context.Context, installationToken int64, owner, repo, branch string, followRedirects bool) (internal.Branch, error)
}

type installationRetriever interface {
	FindOrganizationInstallation(ctx context.Context, org string) (internal.Installation, error)
}

// Deploy contains everything our deploy workflow
// needs to make this request happen.
type Deploy struct {
	RootNames         []string
	Repo              models.Repo
	Branch            string
	Revision          string
	InstallationToken int64
	User              models.User
}

type DeployConverter struct {
	RepoRetriever         repoRetriever
	BranchRetriever       branchRetriever
	InstallationRetriever installationRetriever
}

func (c *DeployConverter) Convert(ctx context.Context, r external.DeployRequest) (Deploy, error) {
	// this should be set in our auth middleware
	username := ctx.Value(middleware.UsernameContextKey)
	if username == nil {
		return Deploy{}, fmt.Errorf("user not provided")
	}

	// In order to authenticate as our GH App we need to get the organization's installation token.
	installation, err := c.InstallationRetriever.FindOrganizationInstallation(ctx, r.Repo.Owner)
	if err != nil {
		return Deploy{}, errors.Wrap(err, "finding installation")
	}

	repository, branch, err := c.getRepositoryAndBranch(ctx, r, installation.Token)
	if err != nil {
		return Deploy{}, err
	}

	return Deploy{
		Repo:              repository,
		RootNames:         r.Roots,
		Branch:            branch.Name,
		Revision:          branch.Revision,
		InstallationToken: installation.Token,
		User: models.User{
			Username: username.(string),
		},
	}, nil
}

func (c *DeployConverter) getRepositoryAndBranch(ctx context.Context, r external.DeployRequest, installationToken int64) (models.Repo, internal.Branch, error) {
	repo, err := c.RepoRetriever.Get(ctx, installationToken, r.Repo.Owner, r.Repo.Name)
	if err != nil {
		return repo, internal.Branch{}, errors.Wrap(err, "getting repo")
	}

	if len(repo.DefaultBranch) == 0 {
		return repo, internal.Branch{}, fmt.Errorf("default branch was nil, this is a bug on github's side")
	}

	branch, err := c.BranchRetriever.GetBranch(ctx, installationToken, r.Repo.Owner, r.Repo.Name, repo.DefaultBranch, true)
	if err != nil {
		return repo, branch, errors.Wrap(err, "getting branch")
	}

	return repo, branch, nil
}
