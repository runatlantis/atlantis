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
//
package vcs

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_proxy.go ClientProxy

// ClientProxy proxies calls to the correct VCS client depending on which
// VCS host is required.
type ClientProxy interface {
	GetModifiedFiles(repo models.Repo, pull models.PullRequest, host Host) ([]string, error)
	CreateComment(repo models.Repo, pullNum int, comment string, host Host) error
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

func (d *DefaultClientProxy) CreateComment(repo models.Repo, pullNum int, comment string, host Host) error {
	switch host {
	case Github:
		return d.GithubClient.CreateComment(repo, pullNum, comment)
	case Gitlab:
		return d.GitlabClient.CreateComment(repo, pullNum, comment)
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
