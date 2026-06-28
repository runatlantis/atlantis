// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
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
	client     *gitlab.Client
	username   string
	ownerName  string
	repoName   string
	token      string
	projectId  int
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
		ownerName = "runatlantis"
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

// GetAtlantisStatus for GitLab uses pipeline status which aggregates all jobs.
// GitLab pipelines cannot distinguish per-project plan counts, so
// expectedCount is ignored. Count-sensitive cases must use VCSGitHub.
func (g GitlabClient) GetAtlantisStatus(ctx context.Context, branchName string, statusPrefix string, expectedCount int) (string, error) {
	pipelineInfos, _, err := g.client.MergeRequests.ListMergeRequestPipelines(g.projectId, g.branchToMR[branchName])
	if err != nil {
		return "", err
	}
	if len(pipelineInfos) == 0 {
		return "", nil
	}
	// GitLab returns pipelines newest-first (descending ID order).
	// Use the first entry to get the most recent pipeline.
	pipelineInfo := pipelineInfos[0]
	pipeline, _, err := g.client.Pipelines.GetPipeline(g.projectId, pipelineInfo.ID)
	if err != nil {
		return "", err
	}

	return pipeline.Status, nil
}

// GetPRComments returns all MR note bodies for the merge request.
// Atlantis posts plan/apply output as MR notes on GitLab.
func (g GitlabClient) GetPRComments(ctx context.Context, pullNumber int) ([]string, error) {
	notes, _, err := g.client.Notes.ListMergeRequestNotes(g.projectId, pullNumber, &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("listing MR notes: %w", err)
	}
	var bodies []string
	for _, n := range notes {
		bodies = append(bodies, n.Body)
	}
	return bodies, nil
}

func (g GitlabClient) ClosePullRequest(ctx context.Context, pullRequestNumber int) error {
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

func (g GitlabClient) DidAtlantisFail(state string) bool {
	return state == "failed" || state == "canceled"
}
