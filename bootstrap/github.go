package bootstrap

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

var githubHostname = "http://api.github.com"
var githubUsername string
var githubPassword string

// Client used for github interactions
type Client struct {
	client *github.Client
	ctx    context.Context
}

// CreateFork forks a github repo into user that is authenticated
func (g *Client) CreateFork(owner string, repoName string) error {
	// forks usually take up to 5 minutes to complete according to github
	_, _, err := g.client.Repositories.CreateFork(g.ctx, owner, repoName, nil)
	// github client returns an error even though the fork was successful
	// in order to figure out the exact error we will need to do the string evaluation below
	if err != nil && !strings.Contains(err.Error(), "job scheduled on GitHub side; try again later") {
		return err
	}
	return nil
}

// CheckForkSuccess waits for github fork to complete
func (g *Client) CheckForkSuccess(ownerName string, forkRepoName string) bool {
	for i := 0; i < 5; i++ {
		if err := g.CreateFork(ownerName, forkRepoName); err == nil {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

// CreateWebhook creates a github webhook
func (g *Client) CreateWebhook(ownerName string, repoName string, hookURL string) error {
	// create atlantis hook
	atlantisHook := &github.Hook{
		Name:   github.String("web"),
		Events: []string{"issue_comment", "pull_request", "pull_request_review", "push"},
		Config: map[string]interface{}{
			"url":          hookURL,
			"content_type": "json",
		},
		Active: github.Bool(true),
	}
	_, _, err := g.client.Repositories.CreateHook(g.ctx, ownerName, repoName, atlantisHook)
	if err != nil {
		return err
	}

	return nil
}

// CreatePullRequest creates a github pull request with custom title and description
func (g *Client) CreatePullRequest(ownerName string, repoName string, head string, base string) (string, error) {
	newPullRequest := &github.NewPullRequest{
		Title: github.String("Welcome to Atlantis!"),
		Head:  github.String(head),
		Body:  github.String(pullRequestBody),
		Base:  github.String(base),
	}

	pull, _, err := g.client.PullRequests.Create(g.ctx, ownerName, repoName, newPullRequest)
	if err != nil {
		return "", err
	}

	return pull.GetHTMLURL(), nil
}
