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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"

	"github.com/runatlantis/atlantis/server/events/models"
)

type AzureDevopsClient struct {
	Client *azuredevops.Client
	// Version is set to the server version.
	Version *version.Version
}

// azuredevopsClientUnderTest is true if we're running under go test.
var azuredevopsClientUnderTest = false

// NewAzureDevopsClient returns a valid Azure Devops client.
func NewAzureDevopsClient(hostname string, account string, project string, token string, logger *logging.SimpleLogger) (*AzureDevopsClient, error) {
	var httpClient = http.Client{
		Timeout: time.Second * 10,
	}
	var adClient, err = azuredevops.NewClient(account, project, token, &httpClient)
	if err != nil {
		return nil, errors.Wrapf(err, "azuredevops.NewClient() %p", adClient)
	}
	client := &AzureDevopsClient{
		Client: adClient,
	}

	// If not using dev.azure.com we need to set the URL to the API.
	if hostname != "dev.azure.com" {
		// We assume the url will be over HTTPS if the user doesn't specify a scheme.
		absoluteURL := hostname
		if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
			absoluteURL = "https://" + absoluteURL
		}

		url, err := url.Parse(absoluteURL)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing URL %q", absoluteURL)
		}

		// Warn if this hostname isn't resolvable.
		ips, err := net.LookupIP(url.Hostname())
		if err != nil {
			logger.Warn("unable to resolve %q: %s", url.Hostname(), err)
		} else if len(ips) == 0 {
			logger.Warn("found no IPs while resolving %q", url.Hostname())
		}

		// Now we're ready to construct the client.
		absoluteURL = strings.TrimSuffix(absoluteURL, "/")
	}

	return client, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt.
func (g *AzureDevopsClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string

	// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/get%20pull%20requests?view=azure-devops-rest-5.1
	apiURL := fmt.Sprintf("_apis/git/pullrequests/%d?api-version=5.1-preview.1", pull.Num)
	req, err := g.Client.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	var commitIdResponse azuredevops.GitPullRequest
	_, err = g.Client.Execute(req, &commitIdResponse)
	commitId := commitIdResponse.GetLastMergeSourceCommit().GetCommitID()

	// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/commits/get%20changes?view=azure-devops-rest-5.1
	apiURL = fmt.Sprintf("_apis/git/repositories/%s/commits/%s/changes?api-version=5.1-preview.1", url.QueryEscape(repo.Name), commitId)
	req, err = g.Client.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	var response azuredevops.GitCommitChanges
	_, err = g.Client.Execute(req, &response)

	for _, change := range response.GetChanges() {
		item := change.GetItem()
		files = append(files, item.GetPath())

		// If the file was renamed, we'll want to run plan in the directory
		// it was moved from as well.
		if change.ChangeType == azuredevops.Rename {
			files = append(files, change.GetSourceServerItem())
		}
	}

	return files, nil
}

// CreateComment creates a comment on the merge request.
// If comment length is greater than the max comment length we split into
// multiple comments.
func (g *AzureDevopsClient) CreateComment(repo models.Repo, pullNum int, comment string) error {
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"

	maxCommentLength := 72
	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart)
	adComments := make([]azuredevops.Comment, len(comments))
	for i, c := range comments {
		entry := azuredevops.Comment{
			CommentType: azuredevops.CommentTypeText,
			Content:     &c,
		}
		adComments[i] = entry
	}
	// Need to create a thread first and save the thread number from response
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20request%20threads/create#comment_on_the_pull_request
	apiURL := fmt.Sprintf("_apis/git/repositories/%s/pullRequests/%d/threads?api-version=5.1-preview.1", url.QueryEscape(repo.Name), pullNum)
	req, err := g.Client.NewRequest("POST", apiURL, nil)
	if err != nil {
		return err
	}
	var threadResp azuredevops.GitPullRequestCommentThread
	_, err = g.Client.Execute(req, &threadResp)

	threadID := threadResp.GetID()
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20request%20thread%20comments/create
	apiURL = fmt.Sprintf("_apis/git/repositories/%s/pullRequests/%d/threads/%d/comments?api-version=5.1-preview.1", url.QueryEscape(repo.Name), pullNum, threadID)
	threadResp.Comments = &adComments
	req, err = g.Client.NewRequest("POST", apiURL, threadResp)
	if err != nil {
		return err
	}
	var response azuredevops.GitPullRequestCommentThread
	_, err = g.Client.Execute(req, &response)

	return nil
}

// PullIsApproved returns true if the merge request was approved.
func (g *AzureDevopsClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	// previews List() args were g.ctx, repo.Owner, repo.Name, pull.Num, nil
	opts := &azuredevops.PullRequestListOptions{}
	adPull, _, err := g.Client.PullRequests.Get(pull.Num, opts)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}
	reviews := adPull.GetReviewers()
	if err != nil {
		return false, errors.Wrap(err, "getting list of pull request reviewers")
	}
	for _, review := range reviews {
		if *review.Vote == azuredevops.VoteApproved {
			return true, nil
		}
	}
	// what happens if no reviewers?  must have reviewers? handle both states
	//return false, nil
	return true, nil
}

// PullIsMergeable returns true if the merge request can be merged.
func (g *AzureDevopsClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {

	opts := &azuredevops.PullRequestListOptions{}
	adPull, _, err := g.Client.PullRequests.Get(pull.Num, opts)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}
	if adPull.MergeStatus != azuredevops.MergeConflicts && adPull.MergeStatus != azuredevops.MergeRejectedByPolicy {
		return false, nil
	}
	return true, nil
}

// UpdateStatus updates the build status of a commit.
func (g *AzureDevopsClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	adState := azuredevops.GitError
	switch state {
	case models.PendingCommitStatus:
		adState = azuredevops.GitPending
	case models.SuccessCommitStatus:
		adState = azuredevops.GitSucceeded
	case models.FailedCommitStatus:
		adState = azuredevops.GitFailed
	}

	genreStr := "Atlantis Bot"
	status := &azuredevops.GitStatus{
		State:       adState,
		Description: &description,
		Context: &azuredevops.GitStatusContext{
			Name:  &src,
			Genre: &genreStr,
		},
		TargetURL: &url,
	}
	_, _, err := g.Client.Git.CreateStatus(repo.Name, pull.HeadCommit, *status)
	return err
}

// MergePull merges the merge request.
func (g *AzureDevopsClient) MergePull(pull models.PullRequest) error {
	// Users can set their repo to disallow certain types of merging.
	// We detect which types aren't allowed and use the type that is.

	// Azure Devops supports squash merge and noFastForward
	// in branch policies
	// Ignoring these for now
	// https://docs.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops

	/*
		const (
			defaultMergeMethod = "merge"
		)
		method := defaultMergeMethod
	*/
	// Now we're ready to make our API call to merge the pull request.
	/*
		options := &azuredevops.GitPullRequestCompletionOptions{
			Status: "completed",
	*/
	// id *IdentityRef, commitMsg string, opts *PullRequestListOptions

	descriptor := "Atlantis Terraform Pull Request Automation"
	i := "atlantis"
	imageUrl := "https://github.com/runatlantis/atlantis/raw/master/runatlantis.io/.vuepress/public/hero.png"
	id := &azuredevops.IdentityRef{
		Descriptor: &descriptor,
		ID:         &i,
		ImageURL:   &imageUrl,
	}

	commitMsg := ""
	mergeResult, _, err := g.Client.PullRequests.Merge(
		pull.BaseRepo.Name,
		pull.Num,
		id,
		commitMsg,
	)
	if err != nil {
		return errors.Wrap(err, "merging pull request")
	}
	if mergeResult.MergeStatus != azuredevops.MergeSucceeded {
		return fmt.Errorf("could not merge pull request: %s", mergeResult.GetMergeFailureMessage())
	}
	return nil
}
