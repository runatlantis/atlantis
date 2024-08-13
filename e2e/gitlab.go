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

	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
	client    *gitlab.Client
	username  string
	ownerName string
	repoName  string
	token     string
	projectId int
}

func NewGitlabClient() *GitlabClient {

	gitlabUsername := os.Getenv("ATLANTISBOT_GITLAB_USERNAME")
	if gitlabUsername == "" {
		log.Fatalf("ATLANTISBOT_GITHUB_USERNAME cannot be empty")
	}
	gitlabToken := os.Getenv("ATLANTISBOT_GITLAB_TOKEN")
	if gitlabToken == "" {
		log.Fatalf("ATLANTISBOT_GITLAB_TOKEN cannot be empty")
	}
	ownerName := os.Getenv("GITHUB_REPO_OWNER_NAME")
	if ownerName == "" {
		ownerName = "run-atlantis"
	}
	repoName := os.Getenv("GITHUB_REPO_NAME")
	if repoName == "" {
		repoName = "atlantis-tests"
	}

	gitlabClient, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	project, _, err := gitlabClient.Projects.GetProject(fmt.Sprintf("%s/%s", ownerName, repoName), &gitlab.GetProjectOptions{})
	if err != nil {
		log.Fatalf("Failed to find project: %v", err)
	}

	return &GitlabClient{
		client:    gitlabClient,
		username:  gitlabUsername,
		ownerName: ownerName,
		repoName:  repoName,
		token:     gitlabToken,
		projectId: project.ID,
	}

}

func (g GitlabClient) Clone(cloneDir string) error {

	repoURL := fmt.Sprintf("https://%s:%s@gitlab.com/%s/%s.git", g.username, g.token, g.ownerName, g.repoName)
	cloneCmd := exec.Command("git", "clone", repoURL, cloneDir)
	// git clone the repo
	log.Printf("git cloning into %q", cloneDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %v: %s", err, string(output))
	}
	return nil

}

func (g GitlabClient) CreateAtlantisWebhook(ctx context.Context, hookURL string) (int64, error) {
	hook, _, err := g.client.Projects.AddProjectHook(g.projectId, &gitlab.AddProjectHookOptions{
		URL:                 &hookURL,
		IssuesEvents:        gitlab.Ptr(true),
		MergeRequestsEvents: gitlab.Ptr(true),
		PushEvents:          gitlab.Ptr(true),
	})
	if err != nil {
		return 0, err
	}
	log.Printf("created webhook for %s", hook.URL)
	return int64(hook.ID), err
}

func (g GitlabClient) DeleteAtlantisHook(ctx context.Context, hookID int64) error {
	_, err := g.client.Projects.DeleteProjectHook(g.projectId, int(hookID))
	if err != nil {
		return err
	}
	log.Printf("deleted webhook id %d", hookID)
	return nil
}

func (g GitlabClient) CreatePullRequest(ctx context.Context, title, branchName string) (string, PullRequest, error) {
	mr, _, err := g.client.MergeRequests.CreateMergeRequest(g.projectId, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(title),
		SourceBranch: gitlab.Ptr(branchName),
		TargetBranch: gitlab.Ptr("main"),
	})
	if err != nil {
		return "", PullRequest{}, fmt.Errorf("error while creating new pull request: %v", err)
	}
	return mr.WebURL, PullRequest{
		id:     mr.IID,
		branch: branchName,
	}, nil

}

func (g GitlabClient) GetAtlantisStatus(ctx context.Context, pullRequest PullRequest) (string, error) {

	fmt.Println(pullRequest)
	res, _, err := g.client.MergeRequests.ListMergeRequestPipelines(g.projectId, pullRequest.id)
	if err != nil {
		fmt.Println("I have error", err)
		return "", err
	}
	fmt.Println(res)
	return "", nil
	/*

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
	*/
}

func (g GitlabClient) ClosePullRequest(ctx context.Context, pullRequest PullRequest) error {
	// clean up
	_, _, err := g.client.MergeRequests.UpdateMergeRequest(g.projectId, pullRequest.id, &gitlab.UpdateMergeRequestOptions{
		StateEvent: gitlab.Ptr("close"),
	})
	if err != nil {
		return fmt.Errorf("error while closing new pull request: %v", err)
	}
	return nil

}
func (g GitlabClient) DeleteBranch(ctx context.Context, branchName string) error {
	panic("I'mdeleting branch")
	/*

		_, err := g.client.Git.DeleteRef(ctx, g.ownerName, g.repoName, branchName)
		if err != nil {
			return fmt.Errorf("error while deleting branch %s: %v", branchName, err)
		}
		return nil
	*/
}
