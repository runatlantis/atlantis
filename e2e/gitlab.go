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

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabClient struct {
	client    *gitlab.Client
	username  string
	ownerName string
	repoName  string
	token     string
	projectId int
	// A mapping from branch names to MR IDs
	branchToMR map[string]int
}

func NewGitlabClient() *GitlabClient {

	gitlabUsername := os.Getenv("ATLANTIS_GITLAB_USER")
	if gitlabUsername == "" {
		log.Fatalf("ATLANTIS_GITLAB_USER cannot be empty")
	}
	gitlabToken := os.Getenv("ATLANTIS_GITLAB_TOKEN")
	if gitlabToken == "" {
		log.Fatalf("ATLANTIS_GITLAB_TOKEN cannot be empty")
	}
	ownerName := os.Getenv("GITLAB_REPO_OWNER_NAME")
	if ownerName == "" {
		ownerName = "run-atlantis"
	}
	repoName := os.Getenv("GITLAB_REPO_NAME")
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
		client:     gitlabClient,
		username:   gitlabUsername,
		ownerName:  ownerName,
		repoName:   repoName,
		token:      gitlabToken,
		projectId:  project.ID,
		branchToMR: make(map[string]int),
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

func (g GitlabClient) CreatePullRequest(ctx context.Context, title, branchName string) (string, int, error) {

	mr, _, err := g.client.MergeRequests.CreateMergeRequest(g.projectId, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(title),
		SourceBranch: gitlab.Ptr(branchName),
		TargetBranch: gitlab.Ptr("main"),
	})
	if err != nil {
		return "", 0, fmt.Errorf("error while creating new pull request: %v", err)
	}
	g.branchToMR[branchName] = mr.IID
	return mr.WebURL, mr.IID, nil

}

func (g GitlabClient) GetAtlantisStatus(ctx context.Context, branchName string) (string, error) {

	pipelineInfos, _, err := g.client.MergeRequests.ListMergeRequestPipelines(g.projectId, g.branchToMR[branchName])
	if err != nil {
		return "", err
	}
	// Possible todo: determine which status in the pipeline we care about?
	if len(pipelineInfos) != 1 {
		return "", fmt.Errorf("unexpected pipelines: %d", len(pipelineInfos))
	}
	pipelineInfo := pipelineInfos[0]
	pipeline, _, err := g.client.Pipelines.GetPipeline(g.projectId, pipelineInfo.ID)
	if err != nil {
		return "", err
	}

	return pipeline.Status, nil
}

func (g GitlabClient) ClosePullRequest(ctx context.Context, pullRequestNumber int) error {
	// clean up
	_, _, err := g.client.MergeRequests.UpdateMergeRequest(g.projectId, pullRequestNumber, &gitlab.UpdateMergeRequestOptions{
		StateEvent: gitlab.Ptr("close"),
	})
	if err != nil {
		return fmt.Errorf("error while closing new pull request: %v", err)
	}
	return nil

}
func (g GitlabClient) DeleteBranch(ctx context.Context, branchName string) error {
	_, err := g.client.Branches.DeleteBranch(g.projectId, branchName)

	if err != nil {
		return fmt.Errorf("error while deleting branch %s: %v", branchName, err)
	}
	return nil

}

func (g GitlabClient) IsAtlantisInProgress(state string) bool {
	// From https://docs.gitlab.com/ee/api/pipelines.html
	// created, waiting_for_resource, preparing, pending, running, success, failed, canceled, skipped, manual, scheduled
	for _, s := range []string{"success", "failed", "canceled", "skipped"} {
		if state == s {
			return false
		}
	}
	return true
}

func (g GitlabClient) DidAtlantisSucceed(state string) bool {
	return state == "success"
}
