// Copyright 2024 Florian Beisel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vcs

import (
	"context"
	"fmt"
	"strings"
	"time"

	// Import the Gitea Go SDK package
	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
)

// GiteaClient is used to perform Gitea actions.
type GiteaClient struct {
	client *gitea.Client
	ctx    context.Context
	logger logging.SimpleLogging
}

type GiteaPRReviewSummary struct {
	Reviews []GiteaReview
}

type GiteaReview struct {
	ID          int64
	Body        string
	Reviewer    string
	State       gitea.ReviewStateType // e.g., "APPROVED", "PENDING", "REQUEST_CHANGES"
	SubmittedAt time.Time
}

func NewClient(baseURL string, token string, logger logging.SimpleLogging) (*GiteaClient, error) {
	client, err := gitea.NewClient(baseURL, gitea.SetToken(token))
	if err != nil {
		return nil, err
	}

	return &GiteaClient{
		client: client,
		ctx:    context.Background(),
		logger: logger,
	}, nil
}

func (g *GiteaClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var modifiedFiles []string

	// Convert pull request number to int64 as required by the Gitea SDK
	prNumber := int64(pull.Num)

	// Prepare options for listing files, if needed
	opt := gitea.ListPullRequestFilesOptions{
		// Configure options as necessary
	}

	// Retrieve the list of changed files in the pull request
	files, _, err := g.client.ListPullRequestFiles(repo.Owner, repo.Name, prNumber, opt)
	if err != nil {
		return nil, err
	}

	// Iterate over the list of changed files and add their paths to the modifiedFiles slice
	for _, file := range files {
		modifiedFiles = append(modifiedFiles, file.Filename)
	}

	return modifiedFiles, nil
}

func (g *GiteaClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	var sepStart string
	sepEnd := "\n```\n</details>" + "\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."

	if command != "" {
		sepStart = fmt.Sprintf("Continued %s output from previous comment.\n<details><summary>Show Output</summary>\n\n", command) + "```diff\n"
	} else {
		sepStart = "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" + "```diff\n"
	}

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart)
	for i := range comments {
		opt := gitea.CreateIssueCommentOption{
			Body: comments[i],
		}
		_, _, err := g.client.CreateIssueComment(repo.Owner, repo.Name, int64(pullNum), opt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GiteaClient) ReactToComment(repo models.Repo, _ int, commentID int64, reaction string) error {
	// Map the GitHub reaction to Gitea's equivalent if necessary. Gitea might have different naming conventions.
	// This example directly uses the provided reaction string assuming it matches Gitea's API.
	// You may need to adjust this mapping based on Gitea's supported reactions.

	_, _, err := g.client.PostIssueCommentReaction(repo.Owner, repo.Name, commentID, reaction)
	if err != nil {
		g.logger.Debug("POST /repos/%v/%v/issues/comments/%d/reactions returned an error: %v", repo.Owner, repo.Name, commentID, err)
		return err
	}

	return nil
}

func (g *GiteaClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string, dir string) error {
	var allComments []*gitea.Comment

	nextPage := int(1)
	for {
		// Initialize ListIssueCommentOptions with the current page
		opts := gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{
				Page:     int(nextPage),
				PageSize: 100, // Set this to a reasonable number based on your needs and Gitea's limits
			},
		}

		comments, resp, err := g.client.ListIssueComments(repo.Owner, repo.Name, int64(pullNum), opts)
		if err != nil {
			return err // Handle the error appropriately
		}

		allComments = append(allComments, comments...)

		// Break the loop if there are no more pages to fetch
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}

	currentUser, _, err := g.client.GetMyUserInfo()
	if err != nil {
		return err // Again, consider wrapping this error
	}

	summaryHeader := fmt.Sprintf("<!--- +-Superseded Command-+ ---><details><summary>Superseded Atlantis %s</summary>", command)
	summaryFooter := "</details>"
	lineFeed := "\n"

	for _, comment := range allComments {
		if comment.Poster == nil || comment.Poster.UserName != currentUser.UserName {
			continue
		}

		body := strings.Split(comment.Body, "\n")
		if len(body) == 0 || (!strings.Contains(strings.ToLower(body[0]), strings.ToLower(command)) && dir != "" && !strings.Contains(strings.ToLower(body[0]), strings.ToLower(dir))) {
			continue
		}

		supersededComment := summaryHeader + lineFeed + comment.Body + lineFeed + summaryFooter + lineFeed

		_, _, err := g.client.EditIssueComment(repo.Owner, repo.Name, comment.ID, gitea.EditIssueCommentOption{
			Body: supersededComment,
		})
		if err != nil {
			return err // Handle or wrap this error as needed
		}
	}

	return nil
}

func (g *GiteaClient) getPRReviews(repo models.Repo, pull models.PullRequest) (GiteaPRReviewSummary, error) {
	var allReviews []GiteaReview

	reviews, _, err := g.client.ListPullReviews(repo.Owner, repo.Name, int64(pull.Num), gitea.ListPullReviewsOptions{})
	if err != nil {
		return GiteaPRReviewSummary{}, err
	}

	for _, review := range reviews {
		mappedReview := GiteaReview{
			ID:          review.ID,
			Body:        review.Body,
			Reviewer:    review.Reviewer.UserName,
			State:       review.State,
			SubmittedAt: review.Submitted,
		}
		allReviews = append(allReviews, mappedReview)
	}

	return GiteaPRReviewSummary{Reviews: allReviews}, nil
}

func (g *GiteaClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	// Check if the logger is nil
	if g.logger == nil {
		return models.ApprovalStatus{}, errors.New("logger is nil")
	}

	// Initial log statement to confirm method entry
	g.logger.Debug("Entering PullIsApproved for GiteaClient")

	// Ensure client is not nil
	if g.client == nil {
		g.logger.Debug("Gitea client is nil")
		return models.ApprovalStatus{}, errors.New("Gitea client is nil")
	}

	// Ensure repo and pull information is valid
	g.logger.Debug("Checking approval for", "repo", repo.FullName, "pull number", pull.Num)

	reviews, resp, err := g.client.ListPullReviews(repo.Owner, repo.Name, int64(pull.Num), gitea.ListPullReviewsOptions{})
	if err != nil {
		g.logger.Debug("Error fetching pull reviews", "error", err)
		return models.ApprovalStatus{}, err
	}

	// Log the response status code if response is not nil
	if resp != nil {
		g.logger.Debug("Received response for pull reviews", "status code", resp.StatusCode)
	} else {
		g.logger.Debug("Response for pull reviews is nil")
	}

	// Process the reviews to check for approval
	for _, review := range reviews {
		if review != nil && review.State == "APPROVED" {
			g.logger.Debug("Pull request approved by", "user", review.Reviewer.UserName)
			return models.ApprovalStatus{
				IsApproved: true,
				ApprovedBy: review.Reviewer.UserName,
				// Ensure date parsing is correctly handled
				Date: review.Submitted,
			}, nil
		}
	}

	g.logger.Debug("Pull request not approved")
	return models.ApprovalStatus{IsApproved: false}, nil
}

func (g *GiteaClient) DiscardReviews(repo models.Repo, pull models.PullRequest) error {
	reviewStatus, err := g.getPRReviews(repo, pull)
	if err != nil {
		return err
	}

	for _, review := range reviewStatus.Reviews {
		// Check if the review state is such that it can be dismissed
		// This depends on how Gitea's API defines dismissible review states
		if review.State == gitea.ReviewStateApproved { // Assuming this state is dismissible
			opt := gitea.DismissPullReviewOptions{
				Message: "Dismissed by Atlantis due to new commit.", // Customize your message
			}
			_, err := g.client.DismissPullReview(repo.Owner, repo.Name, int64(pull.Num), review.ID, opt)
			if err != nil {
				return err // Handle the error as needed
			}
		}
	}

	return nil
}

func (g *GiteaClient) GetCloneURL(_ models.VCSHostType, repo string) (string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return "", errors.New("invalid repo format, expected 'owner/repo'")
	}
	repository, _, err := g.client.GetRepo(parts[0], parts[1])
	if err != nil {
		g.logger.Debug("GET /repos/%v/%v returned an error: %v", parts[0], parts[1], err)
		return "", err
	}
	return repository.CloneURL, nil
}

func (g *GiteaClient) GetFileContent(pull models.PullRequest, fileName string) (bool, []byte, error) {
	ref := pull.HeadBranch // Use the head branch of the pull request as the ref
	fileContent, _, err := g.client.GetFile(pull.BaseRepo.Owner, pull.BaseRepo.Name, ref, fileName)

	if err != nil {
		if strings.Contains(err.Error(), "404") { // Check if the error is because the file was not found
			return false, []byte{}, nil
		}
		return true, []byte{}, err // Return true indicating an attempt to fetch the file was made but resulted in an error other than not found
	}

	// No need to decode from base64 since Gitea's GetFile method should return the raw file content
	return true, fileContent, nil
}

func (g *GiteaClient) GetPullLabels(repo models.Repo, pull models.PullRequest) ([]string, error) {
	pullDetails, _, err := g.client.GetPullRequest(repo.Owner, repo.Name, int64(pull.Num))
	if err != nil {
		g.logger.Debug("GET /repos/%v/%v/pulls/%d returned an error: %v", repo.Owner, repo.Name, pull.Num, err)
		return nil, err
	}

	var labels []string
	for _, label := range pullDetails.Labels {
		labels = append(labels, label.Name)
	}

	return labels, nil
}

func (g *GiteaClient) GetPullRequest(repo models.Repo, num int) (*gitea.PullRequest, error) {
	pull, _, err := g.client.GetPullRequest(repo.Owner, repo.Name, int64(num))
	if err != nil {

		g.logger.Debug("GET /repos/%v/%v/pulls/%d returned an error: %v", repo.Owner, repo.Name, num, "err.Error()")
		return nil, err
	}

	return pull, nil
}

// GetTeamNamesForUser returns the names of the teams or groups that the user belongs to (in the organization the repository belongs to).
func (g *GiteaClient) GetTeamNamesForUser(_ models.Repo, _ models.User) ([]string, error) {
	return nil, nil
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *GiteaClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("!%d", pull.Num), nil
}

func (g *GiteaClient) MergePull(pull models.PullRequest, _ models.PullRequestOptions) error {
	// Fetch repository information to determine the preferred merge method
	repoInfo, _, err := g.client.GetRepo(pull.BaseRepo.Owner, pull.BaseRepo.Name)
	if err != nil {
		return fmt.Errorf("fetching repo info: %w", err)
	}

	// Use the default merge style specified in the repository settings
	mergeMethod := repoInfo.DefaultMergeStyle

	opt := gitea.MergePullRequestOption{
		Style: mergeMethod,
		// Customize other fields as necessary
		Title:                  "Auto-merged by [Your Tool Name]",
		Message:                "Merging pull request.",
		DeleteBranchAfterMerge: true,  // Assuming you want to delete the branch after merge
		ForceMerge:             false, // Set to true if you want to force the merge
		MergeWhenChecksSucceed: true,  // Assuming you want to merge only if checks succeed
	}

	merged, _, err := g.client.MergePullRequest(pull.BaseRepo.Owner, pull.BaseRepo.Name, int64(pull.Num), opt)
	if err != nil {
		return fmt.Errorf("merging pull request: %w", err)
	}
	if !merged {
		return fmt.Errorf("could not merge pull request")
	}
	return nil
}

func (g *GiteaClient) PullIsMergeable(repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	// Note: vcsstatusname is not used in this GiteaClient implementation because Gitea's API
	// does not require filtering status checks by name to determine pull request mergeability.
	// The Mergeable field in Gitea's PullRequest struct directly indicates mergeability.

	giteaPR, _, err := g.client.GetPullRequest(repo.Owner, repo.Name, int64(pull.Num))
	if err != nil {
		return false, fmt.Errorf("getting pull request: %w", err)
	}

	return giteaPR.Mergeable, nil
}

func (g *GiteaClient) SupportsSingleFileDownload(_ models.Repo) bool {
	return true
}

func (g *GiteaClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	var giteaState gitea.StatusState
	switch state {
	case models.PendingCommitStatus:
		giteaState = gitea.StatusPending
	case models.SuccessCommitStatus:
		giteaState = gitea.StatusSuccess
	case models.FailedCommitStatus:
		giteaState = gitea.StatusFailure
	default:
		giteaState = gitea.StatusError
	}

	opts := gitea.CreateStatusOption{
		State:       giteaState,
		TargetURL:   url,
		Description: description,
		Context:     src,
	}

	_, _, err := g.client.CreateStatus(repo.Owner, repo.Name, pull.HeadCommit, opts)
	if err != nil {
		return fmt.Errorf("creating status for commit %s: %w", pull.HeadCommit, err)
	}

	return nil
}
