package github

import (
	"context"
	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"time"
)

type CommitFetcher struct {
	ClientCreator githubapp.ClientCreator
}

func (c *CommitFetcher) FetchLatestPRCommit(ctx context.Context, installationToken int64, repo models.Repo, prNum int) (*gh.Commit, error) {
	client, err := c.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return nil, errors.Wrap(err, "creating installation client")
	}
	run := func(ctx context.Context, nextPage int) ([]*gh.RepositoryCommit, *gh.Response, error) {
		listOptions := gh.ListOptions{
			PerPage: 100,
		}
		listOptions.Page = nextPage
		return client.PullRequests.ListCommits(ctx, repo.Owner, repo.Name, prNum, &listOptions)
	}
	commits, err := Iterate(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "iterating through entries")
	}
	latestCommit := &gh.Commit{}
	latestCommitTimestamp := time.Time{}
	for _, commit := range commits {
		if commit.GetCommit() == nil {
			return nil, errors.New("getting latest commit")
		}
		if commit.GetCommit().GetCommitter() == nil {
			return nil, errors.New("getting latest committer")
		}
		commitTimestamp := commit.GetCommit().GetCommitter().GetDate()
		if commitTimestamp.After(latestCommitTimestamp) {
			latestCommitTimestamp = commitTimestamp
			latestCommit = commit.GetCommit()
		}
	}
	return latestCommit, nil
}
