package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

type GiteaClient struct {
	giteaClient *gitea.Client
	username    string
	token       string
}

// NewClient builds a client that makes API calls to Gitea. httpClient is the
// client to use to make the requests, username and password are used as basic
// auth in the requests, baseURL is the API's baseURL, ex. https://corp.com:7990.
// Don't include the API version, ex. '/1.0'.
func NewClient(baseURL string, username string, token string) (*GiteaClient, error) {
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
	}, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request
// relative to the repo root, e.g. parent/child/file.txt.
func (c *GiteaClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	changedFiles := make([]string, 0)
	page := 0
	nextPage := 1
	listOptions := gitea.ListPullRequestFilesOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
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

		// Emergency break after 500 pages
		if page >= 500 {
			break
		}
	}

	return changedFiles, nil
}

// CreateComment creates a comment on the merge request. It will write multiple
// comments if a single comment is too long.
func (c *GiteaClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	opt := gitea.CreateIssueCommentOption{
		Body: comment,
	}

	_, _, err := c.giteaClient.CreateIssueComment(repo.Owner, repo.Name, int64(pullNum), opt)

	if err != nil {
		return err
	}

	return nil
}

// ReactToComment adds a reaction to a comment.
func (c *GiteaClient) ReactToComment(repo models.Repo, pullNum int, commentID int64, reaction string) error {
	_, _, err := c.giteaClient.PostIssueCommentReaction(repo.Owner, repo.Name, int64(commentID), reaction)

	if err != nil {
		return err
	}

	return nil
}

// HidePrevCommandComments hides the previous command comments from the pull
// request.
func (c *GiteaClient) HidePrevCommandComments(_ models.Repo, _ int, _ string, _ string) error {
	return nil
}

// PullIsApproved returns ApprovalStatus with IsApproved set to true if the pull request has a review that approved the PR.
func (c *GiteaClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	page := 0
	nextPage := 1

	approvalStatus := models.ApprovalStatus{
		IsApproved: false,
	}

	listOptions := gitea.ListPullReviewsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
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

		// Emergency break after 500 pages
		if page >= 500 {
			break
		}
	}

	return approvalStatus, nil
}

// PullIsMergeable returns true if the pull request is mergeable
func (c *GiteaClient) PullIsMergeable(repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	pullRequest, _, err := c.giteaClient.GetPullRequest(repo.Owner, repo.Name, int64(pull.Num))

	if err != nil {
		return false, err
	}

	return pullRequest.Mergeable, nil
}

// UpdateStatus updates the commit status to state for pull. src is the
// source of this status. This should be relatively static across runs,
// ex. atlantis/plan or atlantis/apply.
// description is a description of this particular status update and can
// change across runs.
// url is an optional link that users should click on for more information
// about this status.
func (c *GiteaClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
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
			PageSize: 100,
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
			_, err := c.giteaClient.DismissPullReview(repo.Owner, repo.Name, int64(pull.Num), int64(review.ID), dismissOptions)

			if err != nil {
				return err
			}
		}

		nextPage = resp.NextPage

		// Emergency break after 500 pages
		if page >= 500 {
			break
		}
	}

	return nil
}

func (c *GiteaClient) MergePull(pull models.PullRequest, pullOptions models.PullRequestOptions) error {
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
func (c *GiteaClient) GetFileContent(pull models.PullRequest, fileName string) (bool, []byte, error) {
	content, _, err := c.giteaClient.GetContents(pull.BaseRepo.Owner, pull.BaseRepo.Name, pull.HeadCommit, fileName)

	if err != nil {
		return false, nil, err
	}

	if content.Type == "file" {
		return true, []byte(*content.Content), nil
	}

	return false, nil, nil
}

// SupportsSingleFileDownload returns true if the VCS supports downloading a single file
func (c *GiteaClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return true
}

// GetCloneURL returns the clone URL of the repo
func (c *GiteaClient) GetCloneURL(VCSHostType models.VCSHostType, repo string) (string, error) {
	return "", errors.New("GetCloneURL not (yet) implemented for Gitea client")
}

// GetPullLabels returns the labels of a pull request
func (c *GiteaClient) GetPullLabels(repo models.Repo, pull models.PullRequest) ([]string, error) {
	page := 0
	nextPage := 1
	results := make([]string, 0)

	opts := gitea.ListLabelsOptions{
		ListOptions: gitea.ListOptions{
			Page:     0,
			PageSize: 100,
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

		// Emergency break after 500 pages
		if page >= 500 {
			break
		}
	}

	return results, nil
}
