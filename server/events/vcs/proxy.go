// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package vcs

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_proxy.go ClientProxy

// ClientProxy proxies calls to the correct VCS client depending on which
// VCS host is required.
type ClientProxy interface {
	GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error)
	CreateComment(repo models.Repo, pullNum int, comment string) error
	PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error)
	PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error)
	UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, description string) error
}

// DefaultClientProxy proxies calls to the correct VCS client depending on which
// VCS host is required.
type DefaultClientProxy struct {
	// clients maps from the vcs host type to the client that implements the
	// api for that host type, ex. github -> github client.
	clients map[models.VCSHostType]Client
}

func NewDefaultClientProxy(githubClient Client, gitlabClient Client, bitbucketCloudClient Client, bitbucketServerClient Client) *DefaultClientProxy {
	if githubClient == nil {
		githubClient = &NotConfiguredVCSClient{}
	}
	if gitlabClient == nil {
		gitlabClient = &NotConfiguredVCSClient{}
	}
	if bitbucketCloudClient == nil {
		bitbucketCloudClient = &NotConfiguredVCSClient{}
	}
	if bitbucketServerClient == nil {
		bitbucketServerClient = &NotConfiguredVCSClient{}
	}
	return &DefaultClientProxy{
		clients: map[models.VCSHostType]Client{
			models.Github:          githubClient,
			models.Gitlab:          gitlabClient,
			models.BitbucketCloud:  bitbucketCloudClient,
			models.BitbucketServer: bitbucketServerClient,
		},
	}
}

func (d *DefaultClientProxy) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	return d.clients[repo.VCSHost.Type].GetModifiedFiles(repo, pull)
}

func (d *DefaultClientProxy) CreateComment(repo models.Repo, pullNum int, comment string) error {
	return d.clients[repo.VCSHost.Type].CreateComment(repo, pullNum, comment)
}

func (d *DefaultClientProxy) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	return d.clients[repo.VCSHost.Type].PullIsApproved(repo, pull)
}

func (d *DefaultClientProxy) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	return d.clients[repo.VCSHost.Type].PullIsMergeable(repo, pull)
}

func (d *DefaultClientProxy) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, description string) error {
	return d.clients[repo.VCSHost.Type].UpdateStatus(repo, pull, state, description)
}
