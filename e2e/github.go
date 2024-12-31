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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v66/github"
)

type GithubClient struct {
	client    *github.Client
	username  string
	ownerName string
	repoName  string
	token     string
}

func NewGithubClient() *GithubClient {

	githubUsername := os.Getenv("ATLANTIS_GH_USER")
	if githubUsername == "" {
		log.Fatalf("ATLANTIS_GH_USER cannot be empty")
	}
	githubToken := os.Getenv("ATLANTIS_GH_TOKEN")
	if githubToken == "" {
		log.Fatalf("ATLANTIS_GH_TOKEN cannot be empty")
	}
	ownerName := os.Getenv("GITHUB_REPO_OWNER_NAME")
	if ownerName == "" {
		ownerName = "runatlantis"
	}
	repoName := os.Getenv("GITHUB_REPO_NAME")
	if repoName == "" {
		repoName = "atlantis-tests"
	}

	// create github client
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(githubUsername),
		Password: strings.TrimSpace(githubToken),
	}
	ghClient := github.NewClient(tp.Client())

	return &GithubClient{
		client:    ghClient,
		username:  githubUsername,
		ownerName: ownerName,
		repoName:  repoName,
		token:     githubToken,
	}

}

func (g GithubClient) Clone(cloneDir string) error {

	repoURL := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", g.username, g.token, g.ownerName, g.repoName)
	cloneCmd := exec.Command("git", "clone", repoURL, cloneDir)
	// git clone the repo
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
	// create atlantis hook
	atlantisHook := &github.Hook{
		Events: []string{"issue_comment", "pull_request", "push"},
		Config: hookConfig,
		Active: github.Bool(true),
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

	// set pull request url
	return pull.GetHTMLURL(), pull.GetNumber(), nil

}

func (g GithubClient) GetAtlantisStatus(ctx context.Context, branchName string) (string, error) {
	// check repo status
	combinedStatus, _, err := g.client.Repositories.GetCombinedStatus(ctx, g.ownerName, g.repoName, branchName, nil)
	if err != nil {
		return "", err
	}

	for _, status := range combinedStatus.Statuses {
		if status.GetContext() == "atlantis/plan" {
			return status.GetState(), nil
		}
	}

	return "", nil
}

func (g GithubClient) ClosePullRequest(ctx context.Context, pullRequestNumber int) error {
	// clean up
	_, _, err := g.client.PullRequests.Edit(ctx, g.ownerName, g.repoName, pullRequestNumber, &github.PullRequest{State: github.String("closed")})
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
	for _, s := range []string{"success", "error", "failure"} {
		if state == s {
			return false
		}
	}
	return true
}

func (g GithubClient) DidAtlantisSucceed(state string) bool {
	return state == "success"
}
