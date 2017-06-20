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

const (
	statusContext = "Atlantis"
)

type Status int

const (
	Pending Status = iota
	Success
	Failure
	Error
)

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Success:
		return "success"
	case Failure:
		return "failure"
	case Error:
		return "error"
	}
	return "error"
}

func WorstStatus(ss []Status) Status {
	if len(ss) == 0 {
		return Success
	}
	worst := Success
	for _, s := range ss {
		if s > worst {
			worst = s
		}
	}
	return worst
}

func (g *GithubClient) UpdateStatus(repo models.Repo, pull models.PullRequest, status Status, description string) {
	repoStatus := github.RepoStatus{State: github.String(status.String()), Description: github.String(description), Context: github.String(statusContext)}
	g.client.Repositories.CreateStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, &repoStatus)
}

// GetModifiedFiles returns the names of files that were modified in the pull request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt
func (g *GithubClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string
	comparison, _, err := g.client.Repositories.CompareCommits(g.ctx, repo.Owner, repo.Name, pull.BaseCommit, pull.HeadCommit)
	if err != nil {
		return files, err
	}
	for _, file := range comparison.Files {
		files = append(files, *file.Filename)
	}
	return files, nil
}

func (g *GithubClient) CreateComment(ctx *CommandContext, comment string) error {
	_, _, err := g.client.Issues.CreateComment(g.ctx, ctx.Repo.Owner, ctx.Repo.Name, ctx.Pull.Num, &github.IssueComment{Body: &comment})
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
