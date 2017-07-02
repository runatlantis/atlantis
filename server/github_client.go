package server

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/models"
	"github.com/pkg/errors"
)

type GithubClient struct {
	client *github.Client
	ctx    context.Context
}

// GetModifiedFiles returns the names of files that were modified in the pull request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt
func (g *GithubClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string
	nextPage := 0
	for {
		opts := github.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		pageFiles, resp, err := g.client.PullRequests.ListFiles(g.ctx, repo.Owner, repo.Name, pull.Num, &opts)
		if err != nil {
			return files, err
		}
		for _, f := range pageFiles {
			files = append(files, f.GetFilename())
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}
	return files, nil
}

func (g *GithubClient) CreateComment(repo models.Repo, pull models.PullRequest, comment string) error {
	_, _, err := g.client.Issues.CreateComment(g.ctx, repo.Owner, repo.Name, pull.Num, &github.IssueComment{Body: &comment})
	return err
}

func (g *GithubClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	reviews, _, err := g.client.PullRequests.ListReviews(g.ctx, repo.Owner, repo.Name, pull.Num, nil)
	if err != nil {
		return false, errors.Wrap(err, "getting reviews")
	}
	for _, review := range reviews {
		if review != nil && review.GetState() == "APPROVED" {
			return true, nil
		}
	}
	return false, nil
}

func (g *GithubClient) GetPullRequest(repo models.Repo, num int) (*github.PullRequest, *github.Response, error) {
	return g.client.PullRequests.Get(g.ctx, repo.Owner, repo.Name, num)
}

func (g *GithubClient) UpdateStatus(repo models.Repo, pull models.PullRequest, status *github.RepoStatus) error {
	_, _, err := g.client.Repositories.CreateStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, status)
	return err
}
