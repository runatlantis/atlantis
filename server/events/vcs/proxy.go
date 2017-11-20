package vcs

import (
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/pkg/errors"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_proxy.go ClientProxy

// ClientProxy proxies calls to the correct VCS client depending on which
// VCS host is required.
type ClientProxy interface {
	GetModifiedFiles(repo models.Repo, pull models.PullRequest, host Host) ([]string, error)
	CreateComment(repo models.Repo, pull models.PullRequest, comment string, host Host) error
	PullIsApproved(repo models.Repo, pull models.PullRequest, host Host) (bool, error)
	UpdateStatus(repo models.Repo, pull models.PullRequest, state CommitStatus, description string, host Host) error
}

// DefaultClientProxy proxies calls to the correct VCS client depending on which
// VCS host is required.
type DefaultClientProxy struct {
	GithubClient Client
	GitlabClient Client
}

func NewDefaultClientProxy(githubClient Client, gitlabClient Client) *DefaultClientProxy {
	if githubClient == nil {
		githubClient = &NotConfiguredVCSClient{}
	}
	if gitlabClient == nil {
		gitlabClient = &NotConfiguredVCSClient{}
	}
	return &DefaultClientProxy{
		GitlabClient: gitlabClient,
		GithubClient: githubClient,
	}
}

var invalidVCSErr = errors.New("Invalid VCS Host. This is a bug!")

func (d *DefaultClientProxy) GetModifiedFiles(repo models.Repo, pull models.PullRequest, host Host) ([]string, error) {
	switch host {
	case Github:
		return d.GithubClient.GetModifiedFiles(repo, pull)
	case Gitlab:
		return d.GitlabClient.GetModifiedFiles(repo, pull)
	}
	return nil, invalidVCSErr
}

func (d *DefaultClientProxy) CreateComment(repo models.Repo, pull models.PullRequest, comment string, host Host) error {
	switch host {
	case Github:
		return d.GithubClient.CreateComment(repo, pull, comment)
	case Gitlab:
		return d.GitlabClient.CreateComment(repo, pull, comment)
	}
	return invalidVCSErr
}

func (d *DefaultClientProxy) PullIsApproved(repo models.Repo, pull models.PullRequest, host Host) (bool, error) {
	switch host {
	case Github:
		return d.GithubClient.PullIsApproved(repo, pull)
	case Gitlab:
		return d.GitlabClient.PullIsApproved(repo, pull)
	}
	return false, invalidVCSErr
}

func (d *DefaultClientProxy) UpdateStatus(repo models.Repo, pull models.PullRequest, state CommitStatus, description string, host Host) error {
	switch host {
	case Github:
		return d.GithubClient.UpdateStatus(repo, pull, state, description)
	case Gitlab:
		return d.GitlabClient.UpdateStatus(repo, pull, state, description)
	}
	return invalidVCSErr
}
