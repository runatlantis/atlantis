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

package vcs

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/runatlantis/atlantis/server/events/vcs/common"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// maxCommentLength is the maximum number of chars allowed in a single comment
// by GitHub.
const maxCommentLength = 65536

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

			// If the file was renamed, we'll want to run plan in the directory
			// it was moved from as well.
			if f.GetStatus() == "renamed" {
				files = append(files, f.GetPreviousFilename())
			}
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}
	return files, nil
}

// CreateComment creates a comment on the pull request.
// If comment length is greater than the max comment length we split into
// multiple comments.
func (g *GithubClient) CreateComment(repo models.Repo, pullNum int, comment string) error {
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart)
	for _, c := range comments {
		_, _, err := g.client.Issues.CreateComment(g.ctx, repo.Owner, repo.Name, pullNum, &github.IssueComment{Body: &c})
		if err != nil {
			return err
		}
	}
	return nil
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

// PullIsMergeable returns true if the pull request is mergeable.
func (g *GithubClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	githubPR, err := g.GetPullRequest(repo, pull.Num)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}
	state := githubPR.GetMergeableState()
	// We map our mergeable check to when the GitHub merge button is clickable.
	// This corresponds to the following states:
	// clean: No conflicts, all requirements satisfied.
	//        Merging is allowed (green box).
	// unstable: Failing/pending commit status that is not part of the required
	//           status checks. Merging is allowed (yellow box).
	// has_hooks: GitHub Enterprise only, if a repo has custom pre-receive
	//            hooks. Merging is allowed (green box).
	// See: https://github.com/octokit/octokit.net/issues/1763
	if state != "clean" && state != "unstable" && state != "has_hooks" {
		return false, nil
	}
	return true, nil
}

// GetPullRequest returns the pull request.
func (g *GithubClient) GetPullRequest(repo models.Repo, num int) (*github.PullRequest, error) {
	pull, _, err := g.client.PullRequests.Get(g.ctx, repo.Owner, repo.Name, num)
	return pull, err
}

// UpdateStatus updates the status badge on the pull request.
// See https://github.com/blog/1227-commit-status-api.
func (g *GithubClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	ghState := "error"
	switch state {
	case models.PendingCommitStatus:
		ghState = "pending"
	case models.SuccessCommitStatus:
		ghState = "success"
	case models.FailedCommitStatus:
		ghState = "failure"
	}

	status := &github.RepoStatus{
		State:       github.String(ghState),
		Description: github.String(description),
		Context:     github.String(src),
		TargetURL:   &url,
	}
	_, _, err := g.client.Repositories.CreateStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, status)
	return err
}

// MergePull merges the pull request.
func (g *GithubClient) MergePull(pull models.PullRequest) error {
	// Users can set their repo to disallow certain types of merging.
	// We detect which types aren't allowed and use the type that is.
	repo, _, err := g.client.Repositories.Get(g.ctx, pull.BaseRepo.Owner, pull.BaseRepo.Name)
	if err != nil {
		return errors.Wrap(err, "fetching repo info")
	}
	const (
		defaultMergeMethod = "merge"
		rebaseMergeMethod  = "rebase"
		squashMergeMethod  = "squash"
	)
	method := defaultMergeMethod
	if !repo.GetAllowMergeCommit() {
		if repo.GetAllowRebaseMerge() {
			method = rebaseMergeMethod
		} else if repo.GetAllowSquashMerge() {
			method = squashMergeMethod
		}
	}

	// Now we're ready to make our API call to merge the pull request.
	options := &github.PullRequestOptions{
		MergeMethod: method,
	}
	mergeResult, _, err := g.client.PullRequests.Merge(
		g.ctx,
		pull.BaseRepo.Owner,
		pull.BaseRepo.Name,
		pull.Num,
		common.AutomergeCommitMsg,
		options)
	if err != nil {
		return errors.Wrap(err, "merging pull request")
	}
	if !mergeResult.GetMerged() {
		return fmt.Errorf("could not merge pull request: %s", mergeResult.GetMessage())
	}
	return nil
}
