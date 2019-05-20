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
	"strconv"
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/vcs/common"

	"github.com/runatlantis/atlantis/server/events/models"

	"gopkg.in/russross/blackfriday.v2"
)

type AzureDevopsClient struct {
	Client *azuredevops.Client
	// Version is set to the server version.
	Version *version.Version
	ctx     context.Context
}

// azuredevopsClientUnderTest is true if we're running under go test.
var azuredevopsClientUnderTest = false

// NewAzureDevopsClient returns a valid Azure Devops client.
func NewAzureDevopsClient(org string, username string, project string, token string) (*AzureDevopsClient, error) {
	tp := azuredevops.BasicAuthTransport{
		Username: "",
		Password: strings.TrimSpace(token),
	}
	httpClient := tp.Client()
	httpClient.Timeout = time.Second * 10
	var adClient, err = azuredevops.NewClient(httpClient)
	if err != nil {
		return nil, errors.Wrapf(err, "azuredevops.NewClient() %p", adClient)
	}

	client := &AzureDevopsClient{
		Client: adClient,
		ctx:    context.Background(),
	}

	return client, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt.
func (g *AzureDevopsClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string
	commitIDResponse := new(azuredevops.GitPullRequest)

	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	commitIDResponse, _, _ = g.Client.PullRequests.GetWithRepo(g.ctx, repo.Owner, repo.Project, repo.Name, pull.Num, &opts)

	commitID := commitIDResponse.GetLastMergeSourceCommit().GetCommitID()

	r := new(azuredevops.GitCommitChanges)
	r, _, _ = g.Client.Git.GetChanges(g.ctx, repo.Owner, repo.Project, repo.Name, commitID)

	for _, change := range r.Changes {
		item := change.GetItem()
		files = append(files, item.GetPath())

		// If the file was renamed, we'll want to run plan in the directory
		// it was moved from as well.
		changeType := azuredevops.Rename.String()
		if change.ChangeType == &changeType {
			files = append(files, change.GetSourceServerItem())
		}
	}

	return files, nil
}

// CreateComment creates a comment on a work item linked to pullNum.
// Comments made on pull requests cannot trigger webhooks as of April, 2019.
// Since it is possible to link multiple work items to a single pull request, we:
// 1. Get the pull request
// 2. Get all linked workItems
// a. If only one, comment on it
// b. If more than one, loop through each item and
// i. Get comments for that work item
// ii. Comment on the first workItem with a valid atlantis command in a comment
// iii. Return, ignoring any other items
// iv. If no work items contain comments with a valid atlantis command, return err
// c. If anything else, return warning error
// If comment length is greater than the max comment length we split into
// multiple comments.
// Azure Devops doesn't support markdown in Work Item comments, but it will
// convert text to HTML.  We use the blackfriday library to convert Atlantis
// comment markdown before submission.
func (g *AzureDevopsClient) CreateComment(repo models.Repo, pullNum int, comment string) error {
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"

	// maxCommentLength is the maximum number of chars allowed in a single comment
	// This length was copied from the Github client - haven't found documentation
	// or tested limit in Azure Devops.
	const maxCommentLength = 65536

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart)
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	pull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, repo.Owner, repo.Project, repo.Name, pullNum, &opts)
	if err != nil {
		return err
	}
	if len(pull.WorkItemRefs) == 1 {
		workItemID, err := strconv.Atoi(*pull.WorkItemRefs[0].ID)
		if err != nil {
			return err
		}
		for _, c := range comments {
			input := blackfriday.Run([]byte(c))
			s := string(input)
			workItemComment := azuredevops.WorkItemComment{
				Text: &s,
			}
			_, _, err := g.Client.WorkItems.CreateComment(g.ctx, repo.Owner, repo.Project, workItemID, &workItemComment)
			if err != nil {
				return err
			}
		}
		return nil
	} else if len(pull.WorkItemRefs) > 1 {
		// Until we figure out a nice way to leverage CommentParser.Parse()
		// here, we're going to log an error and continue if the user has a
		// pull request linked to more than one work item. An example
		// implementation is commented below.
		return errors.New("pull request linked to more than one work item - ignoring")
		/* Example handling of PR linked to multiple work items
		for _, ref := range pull.WorkItemRefs {
			workItemID, err := strconv.Atoi(*ref.ID)
			if err != nil {
				return err
			}
			opts := azuredevops.WorkItemCommentListOptions{}
			r, _, err := g.Client.WorkItems.ListComments(g.ctx, repo.Owner, repo.Project, workItemID, &opts)
			if err != nil {
				return err
			}
			for _, comment := range r.Comments {
				fmt.Println(comment)
				parseResult := commentParser.Parse(comment, vcsHost)
				if !parseResult.Ignore {
					// This is likely the correct Work Item
					input := blackfriday.Run([]byte(comment.GetText()))
					s := string(input)
					workItemComment := azuredevops.WorkItemComment{
						Text: &s,
					}
					_, _, err = g.Client.WorkItems.CreateComment(g.ctx, repo.Owner, repo.Project, workItemID, &workItemComment)
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
		*/
	}
	return err
}

// PullIsApproved returns true if the merge request was approved.
func (g *AzureDevopsClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	// previews List() args were g.ctx, repo.Owner, repo.Name, pull.Num, nil
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, repo.Owner, repo.Project, repo.Name, pull.Num, &opts)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}
	reviews := adPull.Reviewers
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

	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, repo.Owner, repo.Project, repo.Name, pull.Num, &opts)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}
	if *adPull.MergeStatus != azuredevops.MergeConflicts.String() &&
		*adPull.MergeStatus != azuredevops.MergeRejectedByPolicy.String() {
		return true, nil
	}
	return false, nil
}

// GetPullRequest returns the pull request.
func (g *AzureDevopsClient) GetPullRequest(repo models.Repo, num int) (*azuredevops.GitPullRequest, error) {
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	pull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, repo.Owner, repo.Project, repo.Name, num, &opts)
	return pull, err
}

// UpdateStatus updates the build status of a commit.
func (g *AzureDevopsClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	adState := azuredevops.GitError.String()
	switch state {
	case models.PendingCommitStatus:
		adState = azuredevops.GitPending.String()
	case models.SuccessCommitStatus:
		adState = azuredevops.GitSucceeded.String()
	case models.FailedCommitStatus:
		adState = azuredevops.GitFailed.String()
	}

	genreStr := "Atlantis Bot"
	status := azuredevops.GitStatus{
		State:       &adState,
		Description: &description,
		Context: &azuredevops.GitStatusContext{
			Name:  &src,
			Genre: &genreStr,
		},
	}
	if url != "" {
		status.TargetURL = &url
	}
	_, _, err := g.Client.Git.CreateStatus(g.ctx, repo.Owner, repo.Project, repo.Name, pull.HeadCommit, status)
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
	// id *IdentityRef, commitMsg string, opts *PullRequestGetOptions

	descriptor := "Atlantis Terraform Pull Request Automation"
	i := "atlantis"
	imageURL := "https://github.com/runatlantis/atlantis/raw/master/runatlantis.io/.vuepress/public/hero.png"
	id := azuredevops.IdentityRef{
		Descriptor: &descriptor,
		ID:         &i,
		ImageURL:   &imageURL,
	}
	// Set default pull request completion options
	mcm := azuredevops.NoFastForward.String()
	twi := new(bool)
	*twi = true
	completionOpts := azuredevops.GitPullRequestCompletionOptions{
		BypassPolicy:            new(bool),
		BypassReason:            azuredevops.String(""),
		DeleteSourceBranch:      new(bool),
		MergeCommitMessage:      azuredevops.String(common.AutomergeCommitMsg),
		MergeStrategy:           &mcm,
		SquashMerge:             new(bool),
		TransitionWorkItems:     twi,
		TriggeredByAutoComplete: new(bool),
	}

	/*
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
	*/
	// Construct request body from supplied parameters
	mergePull := new(azuredevops.GitPullRequest)
	mergePull.AutoCompleteSetBy = &id
	mergePull.CompletionOptions = &completionOpts

	mergeResult, _, err := g.Client.PullRequests.Merge(
		g.ctx,
		pull.BaseRepo.Owner,
		pull.BaseRepo.Project,
		pull.BaseRepo.Name,
		pull.Num,
		mergePull,
		completionOpts,
		id,
	)
	if err != nil {
		return errors.Wrap(err, "merging pull request")
	}
	if *mergeResult.MergeStatus != azuredevops.MergeSucceeded.String() {
		return fmt.Errorf("could not merge pull request: %s", mergeResult.GetMergeFailureMessage())
	}
	return nil
}
