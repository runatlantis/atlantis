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
	"github.com/runatlantis/atlantis/server/logging"
)

// ClientProxy proxies calls to the correct VCS client depending on which
// VCS host is required.
type ClientProxy struct {
	// clients maps from the vcs host type to the client that implements the
	// api for that host type, ex. github -> github client.
	clients map[models.VCSHostType]Client
}

func NewClientProxy(githubClient Client, gitlabClient Client, bitbucketCloudClient Client, bitbucketServerClient Client, azuredevopsClient Client, giteaClient Client) *ClientProxy {
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
	if azuredevopsClient == nil {
		azuredevopsClient = &NotConfiguredVCSClient{}
	}
	if giteaClient == nil {
		giteaClient = &NotConfiguredVCSClient{}
	}
	return &ClientProxy{
		clients: map[models.VCSHostType]Client{
			models.Github:          githubClient,
			models.Gitlab:          gitlabClient,
			models.BitbucketCloud:  bitbucketCloudClient,
			models.BitbucketServer: bitbucketServerClient,
			models.AzureDevops:     azuredevopsClient,
			models.Gitea:           giteaClient,
		},
	}
}

func (d *ClientProxy) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	return d.clients[repo.VCSHost.Type].GetModifiedFiles(logger, repo, pull)
}

func (d *ClientProxy) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	return d.clients[repo.VCSHost.Type].CreateComment(logger, repo, pullNum, comment, command)
}

func (d *ClientProxy) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	return d.clients[repo.VCSHost.Type].HidePrevCommandComments(logger, repo, pullNum, command, dir)
}

func (d *ClientProxy) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	return d.clients[repo.VCSHost.Type].ReactToComment(logger, repo, pullNum, commentID, reaction)
}

func (d *ClientProxy) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	return d.clients[repo.VCSHost.Type].PullIsApproved(logger, repo, pull)
}

func (d *ClientProxy) DiscardReviews(repo models.Repo, pull models.PullRequest) error {
	return d.clients[repo.VCSHost.Type].DiscardReviews(repo, pull)
}

func (d *ClientProxy) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	return d.clients[repo.VCSHost.Type].PullIsMergeable(logger, repo, pull, vcsstatusname)
}

func (d *ClientProxy) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	return d.clients[repo.VCSHost.Type].UpdateStatus(logger, repo, pull, state, src, description, url)
}

func (d *ClientProxy) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	return d.clients[pull.BaseRepo.VCSHost.Type].MergePull(logger, pull, pullOptions)
}

func (d *ClientProxy) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return d.clients[pull.BaseRepo.VCSHost.Type].MarkdownPullLink(pull)
}

func (d *ClientProxy) GetTeamNamesForUser(repo models.Repo, user models.User) ([]string, error) {
	return d.clients[repo.VCSHost.Type].GetTeamNamesForUser(repo, user)
}

func (d *ClientProxy) GetFileContent(logger logging.SimpleLogging, repo models.Repo, branch string, fileName string) (bool, []byte, error) {
	return d.clients[repo.VCSHost.Type].GetFileContent(logger, repo, branch, fileName)
}

func (d *ClientProxy) SupportsSingleFileDownload(repo models.Repo) bool {
	return d.clients[repo.VCSHost.Type].SupportsSingleFileDownload(repo)
}

func (d *ClientProxy) GetCloneURL(logger logging.SimpleLogging, VCSHostType models.VCSHostType, repo string) (string, error) {
	return d.clients[VCSHostType].GetCloneURL(logger, VCSHostType, repo)
}

func (d *ClientProxy) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	return d.clients[repo.VCSHost.Type].GetPullLabels(logger, repo, pull)
}
