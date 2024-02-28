// Copyright 2024 Martijn van der Kleijn & Florian Beisel
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

package gitea

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Emergency break for Gitea pagination (just in case)
// Set to 500 to prevent runaway situations
// Value chosen purposely high, though randomly.
const giteaPaginationEBreak = 500

type GiteaClient struct {
	giteaClient *gitea.Client
	username    string
	token       string
	pageSize    int
	ctx         context.Context
	logger      logging.SimpleLogging
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

type GiteaPullGetter interface {
	GetPullRequest(repo models.Repo, pullNum int) (*gitea.PullRequest, error)
}

// NewClient builds a client that makes API calls to Gitea. httpClient is the
// client to use to make the requests, username and password are used as basic
// auth in the requests, baseURL is the API's baseURL, ex. https://corp.com:7990.
// Don't include the API version, ex. '/1.0'.
func NewClient(baseURL string, username string, token string, pagesize int, logger logging.SimpleLogging) (*GiteaClient, error) {
	giteaClient, err := gitea.NewClient(baseURL,
		gitea.SetToken(token),
		gitea.SetUserAgent("atlantis"),
	)

	if err != nil {
		return nil, errors.Wrap(err, "creating gitea client")
	}

	return &GiteaClient{
		giteaClient: giteaClient,
		username:    username,
		token:       token,
		pageSize:    pagesize,
		ctx:         context.Background(),
		logger:      logger,
	}, nil
}

func (c *GiteaClient) GetPullRequest(repo models.Repo, pullNum int) (*gitea.PullRequest, error) {
	pr, _, err := c.giteaClient.GetPullRequest(repo.Owner, repo.Name, int64(pullNum))

	if err != nil {
		return nil, err
	}

	return pr, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request
// relative to the repo root, e.g. parent/child/file.txt.
func (c *GiteaClient) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	changedFiles := make([]string, 0)
	page := 0
	nextPage := 1
	listOptions := gitea.ListPullRequestFilesOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: c.pageSize,
		},
	}

	for page < nextPage {
		page = +1
		listOptions.ListOptions.Page = page
		files, resp, err := c.giteaClient.ListPullRequestFiles(repo.Owner, repo.Name, int64(pull.Num), listOptions)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			changedFiles = append(changedFiles, file.Filename)
		}

		nextPage = resp.NextPage

		// Emergency break after giteaPaginationEBreak pages
		if page >= giteaPaginationEBreak {
			break
		}
	}

	return changedFiles, nil
}

// CreateComment creates a comment on the merge request. As far as we're aware, Gitea has no built in max comment length right now.
func (c *GiteaClient) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	opt := gitea.CreateIssueCommentOption{
		Body: comment,
	}

	_, _, err := c.giteaClient.CreateIssueComment(repo.Owner, repo.Name, int64(pullNum), opt)

	if err != nil {
		return err
	}

	logger.Debug("Added comment to Gitea pull request %d: %s", pullNum, comment)

	return nil
}

// ReactToComment adds a reaction to a comment.
func (c *GiteaClient) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	_, _, err := c.giteaClient.PostIssueCommentReaction(repo.Owner, repo.Name, commentID, reaction)

	if err != nil {
		return err
	}

	return nil
}

// HidePrevCommandComments hides the previous command comments from the pull
// request.
func (c *GiteaClient) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	var allComments []*gitea.Comment

	nextPage := int(1)
	for {
		// Initialize ListIssueCommentOptions with the current page
		opts := gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{
				Page:     nextPage,
				PageSize: c.pageSize,
			},
		}

		comments, resp, err := c.giteaClient.ListIssueComments(repo.Owner, repo.Name, int64(pullNum), opts)
		if err != nil {
			return err
		}

		allComments = append(allComments, comments...)

		// Break the loop if there are no more pages to fetch
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}

	currentUser, _, err := c.giteaClient.GetMyUserInfo()
	if err != nil {
		return err
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

		_, _, err := c.giteaClient.EditIssueComment(repo.Owner, repo.Name, comment.ID, gitea.EditIssueCommentOption{
			Body: supersededComment,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// PullIsApproved returns ApprovalStatus with IsApproved set to true if the pull request has a review that approved the PR.
func (c *GiteaClient) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	page := 0
	nextPage := 1

	approvalStatus := models.ApprovalStatus{
		IsApproved: false,
	}

	listOptions := gitea.ListPullReviewsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: c.pageSize,
		},
	}

	for page < nextPage {
		page = +1
		listOptions.ListOptions.Page = page
		pullReviews, resp, err := c.giteaClient.ListPullReviews(repo.Owner, repo.Name, int64(pull.Num), listOptions)

		if err != nil {
			return approvalStatus, err
		}

		for _, review := range pullReviews {
			if review.State == gitea.ReviewStateApproved {
				approvalStatus.IsApproved = true
				approvalStatus.ApprovedBy = review.Reviewer.UserName
				approvalStatus.Date = review.Submitted

				return approvalStatus, nil
			}
		}

		nextPage = resp.NextPage

		// Emergency break after giteaPaginationEBreak pages
		if page >= giteaPaginationEBreak {
			break
		}
	}

	return approvalStatus, nil
}

// PullIsMergeable returns true if the pull request is mergeable
func (c *GiteaClient) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	pullRequest, _, err := c.giteaClient.GetPullRequest(repo.Owner, repo.Name, int64(pull.Num))

	if err != nil {
		return false, err
	}

	logger.Debug("Gitea pull request is mergeable: %v (%v)", pullRequest.Mergeable, pull.Num)

	return pullRequest.Mergeable, nil
}

// UpdateStatus updates the commit status to state for pull. src is the
// source of this status. This should be relatively static across runs,
// ex. atlantis/plan or atlantis/apply.
// description is a description of this particular status update and can
// change across runs.
// url is an optional link that users should click on for more information
// about this status.
func (c *GiteaClient) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	giteaState := gitea.StatusFailure

	switch state {
	case models.PendingCommitStatus:
		giteaState = gitea.StatusPending
	case models.SuccessCommitStatus:
		giteaState = gitea.StatusSuccess
	case models.FailedCommitStatus:
		giteaState = gitea.StatusFailure
	}

	newStatusOption := gitea.CreateStatusOption{
		State:       giteaState,
		TargetURL:   url,
		Description: description,
	}

	_, _, err := c.giteaClient.CreateStatus(repo.Owner, repo.Name, pull.HeadCommit, newStatusOption)

	if err != nil {
		return err
	}

	logger.Debug("Gitea status for pull request updated: %v (%v)", state, pull.Num)

	return nil
}

// DiscardReviews discards / dismisses all pull request reviews
func (c *GiteaClient) DiscardReviews(repo models.Repo, pull models.PullRequest) error {
	page := 0
	nextPage := 1

	dismissOptions := gitea.DismissPullReviewOptions{
		Message: "Dismissed by Atlantis",
	}

	listOptions := gitea.ListPullReviewsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: c.pageSize,
		},
	}

	for page < nextPage {
		page = +1
		listOptions.ListOptions.Page = page
		pullReviews, resp, err := c.giteaClient.ListPullReviews(repo.Owner, repo.Name, int64(pull.Num), listOptions)

		if err != nil {
			return err
		}

		for _, review := range pullReviews {
			_, err := c.giteaClient.DismissPullReview(repo.Owner, repo.Name, int64(pull.Num), review.ID, dismissOptions)

			if err != nil {
				return err
			}
		}

		nextPage = resp.NextPage

		// Emergency break after giteaPaginationEBreak pages
		if page >= giteaPaginationEBreak {
			break
		}
	}

	return nil
}

func (c *GiteaClient) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	mergeOptions := gitea.MergePullRequestOption{
		Style:                  gitea.MergeStyleMerge,
		Title:                  "Atlantis merge",
		Message:                "Automatic merge by Atlantis",
		DeleteBranchAfterMerge: pullOptions.DeleteSourceBranchOnMerge,
		ForceMerge:             false,
		HeadCommitId:           pull.HeadCommit,
		MergeWhenChecksSucceed: false,
	}

	succeeded, resp, err := c.giteaClient.MergePullRequest(pull.BaseRepo.Owner, pull.BaseRepo.Name, int64(pull.Num), mergeOptions)

	if err != nil {
		return err
	}

	if !succeeded {
		return fmt.Errorf("merge failed: %s", resp.Status)
	}

	return nil
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (c *GiteaClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("#%d", pull.Num), nil
}

// GetTeamNamesForUser returns the names of the teams or groups that the user belongs to (in the organization the repository belongs to).
func (c *GiteaClient) GetTeamNamesForUser(repo models.Repo, user models.User) ([]string, error) {
	// TODO: implement
	return nil, errors.New("GetTeamNamesForUser not (yet) implemented for Gitea client")
}

// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
// The first return value indicates whether the repo contains a file or not
// if BaseRepo had a file, its content will placed on the second return value
func (c *GiteaClient) GetFileContent(logger logging.SimpleLogging, pull models.PullRequest, fileName string) (bool, []byte, error) {
	content, _, err := c.giteaClient.GetContents(pull.BaseRepo.Owner, pull.BaseRepo.Name, pull.HeadCommit, fileName)

	if err != nil {
		return false, nil, err
	}

	if content.Type == "file" {
		decodedData, err := base64.StdEncoding.DecodeString(*content.Content)
		if err != nil {
			return true, []byte{}, err
		}
		return true, decodedData, nil
	}

	return false, nil, nil
}

// SupportsSingleFileDownload returns true if the VCS supports downloading a single file
func (c *GiteaClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return true
}

// GetCloneURL returns the clone URL of the repo
func (c *GiteaClient) GetCloneURL(logger logging.SimpleLogging, _ models.VCSHostType, repo string) (string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return "", errors.New("invalid repo format, expected 'owner/repo'")
	}
	repository, _, err := c.giteaClient.GetRepo(parts[0], parts[1])
	if err != nil {
		logger.Debug("GET /repos/%v/%v returned an error: %v", parts[0], parts[1], err)
		return "", err
	}
	return repository.CloneURL, nil
}

// GetPullLabels returns the labels of a pull request
func (c *GiteaClient) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	page := 0
	nextPage := 1
	results := make([]string, 0)

	opts := gitea.ListLabelsOptions{
		ListOptions: gitea.ListOptions{
			Page:     0,
			PageSize: c.pageSize,
		},
	}

	for page < nextPage {
		page = +1
		opts.ListOptions.Page = page

		labels, resp, err := c.giteaClient.GetIssueLabels(repo.Owner, repo.Name, int64(pull.Num), opts)

		if err != nil {
			return nil, err
		}

		for _, label := range labels {
			results = append(results, label.Name)
		}

		nextPage = resp.NextPage

		// Emergency break after giteaPaginationEBreak pages
		if page >= giteaPaginationEBreak {
			break
		}
	}

	return results, nil
}

func ValidateSignature(payload []byte, signature string, secretKey []byte) error {
	isValid, err := gitea.VerifyWebhookSignature(string(secretKey), signature, payload)
	if err != nil {
		return errors.New("signature verification internal error")
	}
	if !isValid {
		return errors.New("invalid signature")
	}

	return nil
}
