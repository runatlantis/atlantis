package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"context"
)

type GithubClient struct {
	client *github.Client
	ctx    context.Context
}

const (
	statusContext = "Atlantis"
	PendingStatus = "pending"
	SuccessStatus = "success"
	ErrorStatus = "error"
	FailureStatus = "failure"
)

func (g *GithubClient) UpdateStatus(ctx *PullRequestContext, status string, description string) {
	repoStatus := github.RepoStatus{State: github.String(status), Description: github.String(description), Context: github.String(statusContext)}
	g.client.Repositories.CreateStatus(g.ctx, ctx.owner, ctx.repoName, ctx.head, &repoStatus)
	// todo: deal with error updating status
}

func (g *GithubClient) GetModifiedFiles(ctx *PullRequestContext) ([]string, error) {
	var files = []string{}
	comparison, _, err := g.client.Repositories.CompareCommits(g.ctx, ctx.owner, ctx.repoName, ctx.base, ctx.head)
	if err != nil {
		return files, err
	}
	for _, file := range comparison.Files {
		files = append(files, *file.Filename)
	}
	return files, nil
}

func (g *GithubClient) CreateComment(ctx *PullRequestContext, comment string) error {
	_, _, err := g.client.Issues.CreateComment(g.ctx, ctx.owner, ctx.repoName, ctx.number, &github.IssueComment{Body: &comment})
	return err
}

// CommentExists searches through comments on a pull request and returns true if one matches matcher
func (g *GithubClient) CommentExists(ctx *PullRequestContext, matcher func(*github.IssueComment) bool) (bool, error) {
	opt := &github.IssueListCommentsOptions{}
	// need to loop since there may be multiple pages of comments
	for {
		comments, resp, err := g.client.Issues.ListComments(g.ctx, ctx.owner, ctx.repoName, ctx.number, opt)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve comments: %v", err)
		}
		for _, comment := range comments {
			if matcher(comment) {
				return true, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
	return false, nil
}

func (g *GithubClient) PullIsApproved(ctx *PullRequestContext) (bool, error) {
	// todo: move back to using g.client.PullRequests.ListReviews when we update our GitHub enterprise version
	// to where we don't need to include the custom accept header
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews", ctx.owner, ctx.repoName, ctx.number)
	req, err := g.client.NewRequest("GET", u, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github.black-cat-preview+json")

	var reviews []*github.PullRequestReview
	_, err = g.client.Do(g.ctx, req, &reviews)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve reviews: %v", err)
	}
	for _, review := range reviews {
		if review != nil && review.State != nil && *review.State == "APPROVED" {
			return true, nil
		}
	}
	return false, nil
}

func (g *GithubClient) GetPullRequest(owner string, repo string, number int) (*github.PullRequest, *github.Response, error) {
	return g.client.PullRequests.Get(g.ctx, owner, repo, number)
}
