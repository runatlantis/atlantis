// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v88/github"
)

type GithubClient struct {
	client    *github.Client
	username  string
	ownerName string
	repoName  string
	token     string
	transport *ghinstallation.Transport
}

func NewGithubClient() *GithubClient {
	ownerName := os.Getenv("GITHUB_REPO_OWNER_NAME")
	if ownerName == "" {
		ownerName = "runatlantis"
	}
	repoName := os.Getenv("GITHUB_REPO_NAME")
	if repoName == "" {
		repoName = "atlantis-tests"
	}

	if appIDStr := os.Getenv("ATLANTIS_GH_APP_ID"); appIDStr != "" {
		return newGithubAppClient(appIDStr, ownerName, repoName)
	}

	return newGithubPATClient(ownerName, repoName)
}

func newGithubAppClient(appIDStr, ownerName, repoName string) *GithubClient {
	appID, err := strconv.ParseInt(strings.TrimSpace(appIDStr), 10, 64)
	if err != nil {
		log.Fatalf("ATLANTIS_GH_APP_ID is not a valid integer: %v", err)
	}
	appKey := os.Getenv("ATLANTIS_GH_APP_KEY")
	if appKey == "" {
		log.Fatalf("ATLANTIS_GH_APP_KEY cannot be empty when ATLANTIS_GH_APP_ID is set")
	}

	appTransport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(appKey))
	if err != nil {
		log.Fatalf("creating GitHub App transport: %v", err)
	}

	appClient, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: appTransport}))
	if err != nil {
		log.Fatalf("creating GitHub App client: %v", err)
	}
	ctx := context.Background()

	installation, _, err := appClient.Apps.GetRepositoryInstallation(ctx, ownerName, repoName)
	if err != nil {
		log.Fatalf("getting GitHub App installation for %s/%s: %v", ownerName, repoName, err) //nolint:gosec // env-sourced org/repo names
	}

	installationID := installation.GetID()
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, []byte(appKey))
	if err != nil {
		log.Fatalf("creating GitHub App installation transport: %v", err)
	}

	ghClient, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: itr}))
	if err != nil {
		log.Fatalf("creating GitHub App installation client: %v", err)
	}

	username := ""
	if slug := os.Getenv("ATLANTIS_GH_APP_SLUG"); slug != "" {
		username = fmt.Sprintf("%s[bot]", slug)
	}

	log.Printf("using GitHub App auth (app ID: %d, installation ID: %d, user: %s)", appID, installationID, username) //nolint:gosec // diagnostic log

	return &GithubClient{
		client:    ghClient,
		username:  username,
		ownerName: ownerName,
		repoName:  repoName,
		transport: itr,
	}
}

func newGithubPATClient(ownerName, repoName string) *GithubClient {
	githubUsername := os.Getenv("ATLANTIS_GH_USER")
	if githubUsername == "" {
		log.Fatalf("ATLANTIS_GH_USER cannot be empty")
	}
	githubToken := os.Getenv("ATLANTIS_GH_TOKEN")
	if githubToken == "" {
		log.Fatalf("ATLANTIS_GH_TOKEN cannot be empty")
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(githubUsername),
		Password: strings.TrimSpace(githubToken),
	}
	ghClient, err := github.NewClient(github.WithHTTPClient(tp.Client()))
	if err != nil {
		log.Fatalf("creating GitHub PAT client: %v", err)
	}

	return &GithubClient{
		client:    ghClient,
		username:  githubUsername,
		ownerName: ownerName,
		repoName:  repoName,
		token:     githubToken,
	}
}

func (g GithubClient) Clone(cloneDir string) error {
	var repoURL string
	if g.transport != nil {
		token, err := g.transport.Token(context.Background())
		if err != nil {
			return fmt.Errorf("getting installation token for clone: %w", err)
		}
		repoURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, g.ownerName, g.repoName)
	} else {
		repoURL = fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", g.username, g.token, g.ownerName, g.repoName)
	}

	cloneCmd := exec.Command("git", "clone", repoURL, cloneDir)
	log.Printf("git cloning into %q", cloneDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %v: %s", err, string(output))
	}
	return nil
}

func (g GithubClient) CreateAtlantisWebhook(ctx context.Context, hookURL string) (int64, error) {
	contentType := "json"
	hookConfig := &github.HookConfig{
		ContentType: &contentType,
		URL:         &hookURL,
	}
	atlantisHook := &github.Hook{
		Events: []string{"issue_comment", "pull_request", "push"},
		Config: hookConfig,
		Active: github.Ptr(true),
	}

	hook, _, err := g.client.Repositories.CreateHook(ctx, g.ownerName, g.repoName, atlantisHook)
	if err != nil {
		return 0, err
	}
	log.Println(hook.GetURL())

	return hook.GetID(), nil
}

func (g GithubClient) DeleteAtlantisHook(ctx context.Context, hookID int64) error {
	_, err := g.client.Repositories.DeleteHook(ctx, g.ownerName, g.repoName, hookID)
	if err != nil {
		return err
	}
	log.Printf("deleted webhook id %d", hookID)

	return nil
}

func (g GithubClient) CreatePullRequest(ctx context.Context, title, branchName string) (string, int, error) {
	head := fmt.Sprintf("%s:%s", g.ownerName, branchName)
	body := ""
	base := "main"
	newPullRequest := &github.NewPullRequest{Title: &title, Head: &head, Body: &body, Base: &base}

	pull, _, err := g.client.PullRequests.Create(ctx, g.ownerName, g.repoName, newPullRequest)
	if err != nil {
		return "", 0, fmt.Errorf("error while creating new pull request: %v", err)
	}

	return pull.GetHTMLURL(), pull.GetNumber(), nil
}

// PostPRComment posts an issue comment on the pull request.
func (g GithubClient) PostPRComment(ctx context.Context, pullNumber int, body string) error {
	_, _, err := g.client.Issues.CreateComment(ctx, g.ownerName, g.repoName, pullNumber, &github.IssueComment{
		Body: github.Ptr(body),
	})
	if err != nil {
		return fmt.Errorf("creating PR comment: %w", err)
	}
	return nil
}

// GetAtlantisStatus polls the aggregate "atlantis/plan" commit status.
// Used only for the polling loop to detect terminal state.
func (g GithubClient) GetAtlantisStatus(ctx context.Context, branchName string) (string, error) {
	return g.GetAtlantisCommandStatus(ctx, branchName, "plan")
}

func (g GithubClient) GetAtlantisCommandStatus(ctx context.Context, branchName string, command string) (string, error) {
	status, err := g.GetCommitStatus(ctx, branchName, atlantisCommandStatusContext(command))
	if err != nil {
		return "", err
	}
	return status.State, nil
}

func (g GithubClient) GetCommitStatus(ctx context.Context, branchName, statusContext string) (CommitStatus, error) {
	combinedStatus, _, err := g.client.Repositories.GetCombinedStatus(ctx, g.ownerName, g.repoName, branchName, nil)
	if err != nil {
		return CommitStatus{}, err
	}

	var result CommitStatus
	for _, status := range combinedStatus.Statuses {
		if status.GetContext() != statusContext {
			continue
		}
		candidate := CommitStatus{
			State:     status.GetState(),
			ID:        status.GetID(),
			UpdatedAt: status.GetUpdatedAt().Time,
		}
		if result.ID == 0 || candidate.UpdatedAt.After(result.UpdatedAt) || (candidate.UpdatedAt.Equal(result.UpdatedAt) && candidate.ID > result.ID) {
			result = candidate
		}
	}

	return result, nil
}

// GetProjectStatuses returns all per-project Atlantis plan status contexts
// and their states. These have the form "atlantis/plan: <ProjectID>".
// The aggregate "atlantis/plan" status is excluded.
func (g GithubClient) GetProjectStatuses(ctx context.Context, branchName string) (map[string]string, error) {
	combinedStatus, _, err := g.client.Repositories.GetCombinedStatus(ctx, g.ownerName, g.repoName, branchName, nil)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, status := range combinedStatus.Statuses {
		statusCtx := status.GetContext()
		if strings.HasPrefix(statusCtx, projectStatusPrefix) {
			result[statusCtx] = status.GetState()
		}
	}
	return result, nil
}

// GetPRComments returns all issue comment bodies on the pull request.
func (g GithubClient) GetPRComments(ctx context.Context, pullNumber int) ([]string, error) {
	comments, _, err := g.client.Issues.ListComments(ctx, g.ownerName, g.repoName, pullNumber, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("listing PR comments: %w", err)
	}
	var bodies []string
	for _, c := range comments {
		bodies = append(bodies, c.GetBody())
	}
	return bodies, nil
}

func (g GithubClient) ClosePullRequest(ctx context.Context, pullRequestNumber int) error {
	_, _, err := g.client.PullRequests.Edit(ctx, g.ownerName, g.repoName, pullRequestNumber, &github.PullRequest{State: github.Ptr("closed")})
	if err != nil {
		return fmt.Errorf("error while closing new pull request: %v", err)
	}
	return nil
}

func (g GithubClient) DeleteBranch(ctx context.Context, branchName string) error {
	deleteBranchName := fmt.Sprintf("%s/%s", "heads", branchName)
	_, err := g.client.Git.DeleteRef(ctx, g.ownerName, g.repoName, deleteBranchName)
	if err != nil {
		return fmt.Errorf("error while deleting branch %s: %v", branchName, err)
	}
	return nil
}

func (g GithubClient) IsAtlantisInProgress(state string) bool {
	return !slices.Contains([]string{"success", "error", "failure"}, state)
}

func (g GithubClient) DidAtlantisSucceed(state string) bool {
	return state == "success"
}

func (g GithubClient) DidAtlantisFail(state string) bool {
	return state == "failure" || state == "error"
}
