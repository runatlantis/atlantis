package vcs

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// GithubClient is used to perform GitHub actions.
type GithubClient struct {
	client *github.Client
	ctx    context.Context
}

// NewGithubClient returns a valid GitHub client.
func NewGithubClient(hostname string, user string, pass string) (*GithubClient, error) {
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(user),
		Password: strings.TrimSpace(pass),
	}
	client := github.NewClient(tp.Client())
	// If we're using github.com then we don't need to do any additional configuration
	// for the client. It we're using Github Enterprise, then we need to manually
	// set the base url for the API.
	if hostname != "github.com" {
		baseURL := fmt.Sprintf("https://%s/api/v3/", hostname)
		base, err := url.Parse(baseURL)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid github hostname trying to parse %s", baseURL)
		}
		client.BaseURL = base
	}

	return &GithubClient{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// GetModifiedFiles returns the names of files that were modified in the pull request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt.
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

// CreateComment creates a comment on the pull request.
func (g *GithubClient) CreateComment(repo models.Repo, pull models.PullRequest, comment string) error {
	_, _, err := g.client.Issues.CreateComment(g.ctx, repo.Owner, repo.Name, pull.Num, &github.IssueComment{Body: &comment})
	return err
}

// PullIsApproved returns true if the pull request was approved.
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

// GetPullRequest returns the pull request.
func (g *GithubClient) GetPullRequest(repo models.Repo, num int) (*github.PullRequest, error) {
	pull, _, err := g.client.PullRequests.Get(g.ctx, repo.Owner, repo.Name, num)
	return pull, err
}

// UpdateStatus updates the status badge on the pull request.
// See https://github.com/blog/1227-commit-status-api.
func (g *GithubClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state CommitStatus, description string) error {
	const statusContext = "Atlantis"
	ghState := "error"
	switch state {
	case Pending:
		ghState = "pending"
	case Success:
		ghState = "success"
	case Failed:
		ghState = "failure"
	}
	status := &github.RepoStatus{
		State:       github.String(ghState),
		Description: github.String(description),
		Context:     github.String(statusContext)}
	_, _, err := g.client.Repositories.CreateStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, status)
	return err
}
