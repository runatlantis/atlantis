package github

import (
	"context"
	"net/url"

	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	gh_helper "github.com/runatlantis/atlantis/server/vcs/provider/github"
)

type Client struct {
	ClientCreator githubapp.ClientCreator
}

type Context interface {
	GetInstallationToken() int64
	context.Context
}

type contextWithToken struct {
	InstallationToken int64
	context.Context
}

func (c *contextWithToken) GetInstallationToken() int64 {
	return c.InstallationToken
}

func ContextWithInstallationToken(ctx context.Context, installationToken int64) Context {
	return &contextWithToken{
		InstallationToken: installationToken,
		Context:           ctx,
	}
}

func (c *Client) CreateCheckRun(ctx Context, owner, repo string, opts github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	client, err := c.ClientCreator.NewInstallationClient(ctx.GetInstallationToken())

	if err != nil {
		return nil, nil, errors.Wrap(err, "creating client from installation")
	}

	return client.Checks.CreateCheckRun(ctx, owner, repo, opts)
}
func (c *Client) UpdateCheckRun(ctx Context, owner, repo string, checkRunID int64, opts github.UpdateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	client, err := c.ClientCreator.NewInstallationClient(ctx.GetInstallationToken())

	if err != nil {
		return nil, nil, errors.Wrap(err, "creating client from installation")
	}

	return client.Checks.UpdateCheckRun(ctx, owner, repo, checkRunID, opts)
}
func (c *Client) GetArchiveLink(ctx Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error) {
	client, err := c.ClientCreator.NewInstallationClient(ctx.GetInstallationToken())

	if err != nil {
		return nil, nil, errors.Wrap(err, "creating client from installation")
	}

	return client.Repositories.GetArchiveLink(ctx, owner, repo, archiveformat, opts, followRedirects)

}

func (c *Client) CompareCommits(ctx Context, owner, repo string, base, head string, opts *github.ListOptions) (*github.CommitsComparison, *github.Response, error) {
	client, err := c.ClientCreator.NewInstallationClient(ctx.GetInstallationToken())

	if err != nil {
		return nil, nil, errors.Wrap(err, "creating client from installation")
	}

	return client.Repositories.CompareCommits(ctx, owner, repo, base, head, opts)
}

func (c *Client) ListPullRequests(ctx Context, owner, repo, base, state string) ([]*github.PullRequest, error) {
	client, err := c.ClientCreator.NewInstallationClient(ctx.GetInstallationToken())
	if err != nil {
		return nil, errors.Wrap(err, "creating client from installation")
	}

	run := func(ctx context.Context, nextPage int) ([]*github.PullRequest, *github.Response, error) {
		prListOptions := github.PullRequestListOptions{
			State: state,
			Base:  base,
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		prListOptions.ListOptions.Page = nextPage
		return client.PullRequests.List(ctx, owner, repo, &prListOptions)
	}

	return gh_helper.Iterate(ctx, run)
}
