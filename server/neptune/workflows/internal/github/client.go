package github

import (
	"context"
	"net/url"

	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
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
