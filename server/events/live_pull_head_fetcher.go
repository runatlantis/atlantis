// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/logging"
)

type PullIdentityGetter interface {
	GetPullRequestIdentity(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.PullRequest, error)
}

type DefaultLivePullHeadFetcher struct {
	EventParser               EventParsing
	GithubPullGetter          GithubPullGetter
	GitlabMergeRequestGetter  GitlabMergeRequestGetter
	AzureDevopsPullGetter     AzureDevopsPullGetter
	GiteaPullGetter           *gitea.Client
	BitbucketCloudPullGetter  PullIdentityGetter
	BitbucketServerPullGetter PullIdentityGetter
}

func (f *DefaultLivePullHeadFetcher) GetLivePullIdentity(ctx command.ProjectContext) (models.PullRequest, error) {
	switch ctx.Pull.BaseRepo.VCSHost.Type {
	case models.Github:
		if f.GithubPullGetter == nil || f.EventParser == nil {
			return models.PullRequest{}, errors.New("atlantis is not configured to fetch live GitHub pull requests")
		}
		ghPull, err := f.GithubPullGetter.GetPullRequest(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("making pull request API call to GitHub: %w", err)
		}
		pull, _, _, err := f.EventParser.ParseGithubPull(ctx.Log, ghPull)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("extracting GitHub pull request identity: %w", err)
		}
		return pull, nil
	case models.Gitlab:
		if f.GitlabMergeRequestGetter == nil || f.EventParser == nil {
			return models.PullRequest{}, errors.New("atlantis is not configured to fetch live GitLab merge requests")
		}
		mr, err := f.GitlabMergeRequestGetter.GetMergeRequest(ctx.Log, ctx.Pull.BaseRepo.FullName, ctx.Pull.Num)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("making merge request API call to GitLab: %w", err)
		}
		pull := f.EventParser.ParseGitlabMergeRequest(mr, ctx.Pull.BaseRepo)
		return pull, nil
	case models.AzureDevops:
		if f.AzureDevopsPullGetter == nil || f.EventParser == nil {
			return models.PullRequest{}, errors.New("atlantis is not configured to fetch live Azure DevOps pull requests")
		}
		adPull, err := f.AzureDevopsPullGetter.GetPullRequest(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("making pull request API call to Azure DevOps: %w", err)
		}
		pull, _, _, err := f.EventParser.ParseAzureDevopsPull(adPull)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("extracting Azure DevOps pull request identity: %w", err)
		}
		return pull, nil
	case models.Gitea:
		if f.GiteaPullGetter == nil || f.EventParser == nil {
			return models.PullRequest{}, errors.New("atlantis is not configured to fetch live Gitea pull requests")
		}
		giteaPull, err := f.GiteaPullGetter.GetPullRequest(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("making pull request API call to Gitea: %w", err)
		}
		pull, _, _, err := f.EventParser.ParseGiteaPull(giteaPull)
		if err != nil {
			return models.PullRequest{}, fmt.Errorf("extracting Gitea pull request identity: %w", err)
		}
		return pull, nil
	case models.BitbucketCloud:
		if f.BitbucketCloudPullGetter == nil {
			return models.PullRequest{}, errors.New("atlantis is not configured to fetch live Bitbucket Cloud pull requests")
		}
		return f.BitbucketCloudPullGetter.GetPullRequestIdentity(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull)
	case models.BitbucketServer:
		if f.BitbucketServerPullGetter == nil {
			return models.PullRequest{}, errors.New("atlantis is not configured to fetch live Bitbucket Server pull requests")
		}
		return f.BitbucketServerPullGetter.GetPullRequestIdentity(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull)
	default:
		return models.PullRequest{}, fmt.Errorf("unsupported vcs host type %q", ctx.Pull.BaseRepo.VCSHost.Type.String())
	}
}
