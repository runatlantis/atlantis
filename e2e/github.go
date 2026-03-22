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
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v83/github"
)

type GithubClient struct {
	client    *github.Client
	username  string
	ownerName string
	repoName  string
	token     string
	// transport is set when using GitHub App auth; nil for PAT auth.
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

	// Try GitHub App auth first
	if appIDStr := os.Getenv("ATLANTIS_GH_APP_ID"); appIDStr != "" {
		return newGithubAppClient(appIDStr, ownerName, repoName)
	}

	// Fall back to PAT auth
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

	// Create an app-level transport to look up the installation for this repo
	appTransport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(appKey))
	if err != nil {
		log.Fatalf("creating GitHub App transport: %v", err)
	}

	appClient := github.NewClient(&http.Client{Transport: appTransport})
	ctx := context.Background()

	installation, _, err := appClient.Apps.FindRepositoryInstallation(ctx, ownerName, repoName)
	if err != nil {
		log.Fatalf("getting GitHub App installation for %s/%s: %v", ownerName, repoName, err)
	}

	installationID := installation.GetID()
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, []byte(appKey))
	if err != nil {
		log.Fatalf("creating GitHub App installation transport: %v", err)
	}

	ghClient := github.NewClient(&http.Client{Transport: itr})

	// Derive bot username from app slug
	username := ""
	if slug := os.Getenv("ATLANTIS_GH_APP_SLUG"); slug != "" {
		username = fmt.Sprintf("%s[bot]", slug)
	}

	log.Printf("using GitHub App auth (app ID: %d, installation ID: %d, user: %s)", appID, installationID, username)

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
	var repoURL string
	if g.transport != nil {
		// GitHub App auth: get a fresh installation token for git clone
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
	// create atlantis hook
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
