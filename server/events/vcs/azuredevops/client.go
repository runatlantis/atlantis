// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/drmaxgit/go-azuredevops/azuredevops"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
)

// Client represents an Azure DevOps VCS client
type Client struct {
	Client   *azuredevops.Client
	ctx      context.Context
	UserName string
}

// NewClient returns a valid Azure DevOps client.
func New(hostname string, userName string, token string) (*Client, error) {
	tp := azuredevops.BasicAuthTransport{
		Username: "",
		Password: strings.TrimSpace(token),
	}
	httpClient := tp.Client()
	httpClient.Timeout = time.Second * 10
	var adClient, err = azuredevops.NewClient(httpClient)
	if err != nil {
		return nil, err
	}

	if hostname != "dev.azure.com" {
		baseURL := fmt.Sprintf("https://%s/", hostname)
		base, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid azure devops hostname trying to parse %s: %w", baseURL, err)
		}
		adClient.BaseURL = *base
	}

	client := &Client{
		Client:   adClient,
		UserName: userName,
		ctx:      context.Background(),
	}

	return client, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *Client) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string

	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	pullRequest, _, _ := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)

	targetRefName := strings.Replace(pullRequest.GetTargetRefName(), "refs/heads/", "", 1)
	sourceRefName := strings.Replace(pullRequest.GetSourceRefName(), "refs/heads/", "", 1)

	const pageSize = 100 // Number of files from diff call
	var skip int

	for {
		r, resp, err := g.Client.Git.GetDiffs(g.ctx, owner, project, repoName, targetRefName, sourceRefName, &azuredevops.GitDiffListOptions{
			Top:  pageSize,
			Skip: skip,
		})
		if err != nil {
			return nil, fmt.Errorf("getting pull request: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http response code %d getting diff %s to %s: %w", resp.StatusCode, sourceRefName, targetRefName, err)
		}

		for _, change := range r.Changes {
			item := change.GetItem()
			// Convert the path to a relative path from the repo's root.
			relativePath := filepath.Clean("./" + item.GetPath())
			files = append(files, relativePath)

			// If the file was renamed, we'll want to run plan in the directory
			// it was moved from as well.
			changeType := azuredevops.Rename.String()
			if change.ChangeType == &changeType {
				relativePath = filepath.Clean("./" + change.GetSourceServerItem())
				files = append(files, relativePath)
			}
		}

		if len(r.Changes) < pageSize {
			break // Break if we have reached the end
		}
		skip += pageSize // Move to next page
	}

	return files, nil
}

// CreateComment creates a comment on a pull request.
//
// If comment length is greater than the max comment length we split into
// multiple comments.
func (g *Client) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error { //nolint: revive
	// maxCommentLength is the maximum number of chars allowed in a single comment
	// This length was copied from the Github client - haven't found documentation
	// or tested limit in Azure DevOps.
	const maxCommentLength = 150000

	comments := common.SplitComment(logger, comment, maxCommentLength, 0, command)
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	for i := range comments {
		commentType := "text"
		parentCommentID := 0

		prComment := azuredevops.Comment{
			CommentType:     &commentType,
			Content:         &comments[i],
			ParentCommentID: &parentCommentID,
		}
		prComments := []*azuredevops.Comment{&prComment}
		body := azuredevops.GitPullRequestCommentThread{
			Comments: prComments,
		}
		_, _, err := g.Client.PullRequests.CreateComments(g.ctx, owner, project, repoName, pullNum, &body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Client) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error { //nolint: revive
	return nil
}

func (g *Client) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error { //nolint: revive
	return nil
}

// PullIsApproved returns true if the merge request was approved by another reviewer.
// https://docs.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops#require-a-minimum-number-of-reviewers
func (g *Client) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)
	if err != nil {
		return approvalStatus, fmt.Errorf("getting pull request: %w", err)
	}

	numApprovals := 0
	for _, review := range adPull.Reviewers {
		if review == nil {
			continue
		}

		if review.GetUniqueName() == adPull.GetCreatedBy().GetUniqueName() {
			continue
		}

        if review.GetVote() == azuredevops.VoteApproved || review.GetVote() == azuredevops.VoteApprovedWithSuggestions {
            numApprovals++
        }
    }

    if numApprovals > 0 {
        return models.ApprovalStatus{
            IsApproved: true,
            NumApprovals: numApprovals,
        }, nil
    }

	return approvalStatus, nil
}

func (g *Client) DiscardReviews(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error { //nolint: revive
	// TODO implement
	return nil
}

// PullIsMergeable returns true if the merge request can be merged.
func (g *Client) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, _ string, _ []string) (models.MergeableStatus, error) { //nolint: revive
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestGetOptions{IncludeWorkItemRefs: true}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)
	if err != nil {
		return models.MergeableStatus{}, fmt.Errorf("getting pull request: %w", err)
	}

	if *adPull.MergeStatus != azuredevops.MergeSucceeded.String() {
		return models.MergeableStatus{
			IsMergeable: false,
		}, nil
	}

	if *adPull.IsDraft {
		return models.MergeableStatus{
			IsMergeable: false,
		}, nil
	}

	if *adPull.Status != azuredevops.PullActive.String() {
		return models.MergeableStatus{
			IsMergeable: false,
		}, nil
	}

	projectID := *adPull.Repository.Project.ID
	artifactID := g.Client.PolicyEvaluations.GetPullRequestArtifactID(projectID, pull.Num)
	policyEvaluations, _, err := g.Client.PolicyEvaluations.List(g.ctx, owner, project, artifactID, &azuredevops.PolicyEvaluationsListOptions{})
	if err != nil {
		return models.MergeableStatus{}, fmt.Errorf("getting policy evaluations: %w", err)
	}

	for _, policyEvaluation := range policyEvaluations {
		if !*policyEvaluation.Configuration.IsEnabled || *policyEvaluation.Configuration.IsDeleted {
			continue
		}

		// Ignore the Atlantis status, even if its set as a blocker.
		// This status should not be considered when evaluating if the pull request can be applied.
		settings := (policyEvaluation.Configuration.Settings).(map[string]any)
		if genre, ok := settings["statusGenre"]; ok && genre == "Atlantis Bot/atlantis" {
			if name, ok := settings["statusName"]; ok && name == "apply" {
				continue
			}
		}

		if *policyEvaluation.Configuration.IsBlocking && *policyEvaluation.Status != azuredevops.PolicyEvaluationApproved {
			return models.MergeableStatus{
				IsMergeable: false,
			}, nil
		}
	}

	return models.MergeableStatus{
		IsMergeable: true,
	}, nil
}

// GetPullRequest returns the pull request.
func (g *Client) GetPullRequest(logger logging.SimpleLogging, repo models.Repo, num int) (*azuredevops.GitPullRequest, error) {
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)
	pull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, num, &opts)
	return pull, err
}

// UpdateStatus updates the build status of a commit.
func (g *Client) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	adState := azuredevops.GitError.String()
	switch state {
	case models.PendingCommitStatus:
		adState = azuredevops.GitPending.String()
	case models.SuccessCommitStatus:
		adState = azuredevops.GitSucceeded.String()
	case models.FailedCommitStatus:
		adState = azuredevops.GitFailed.String()
	}

	logger.Info("Updating Azure DevOps commit status for '%s' to '%s'", src, adState)

	status := azuredevops.GitPullRequestStatus{}
	status.Context = gitStatusContextFromSrc(src)
	status.Description = &description
	status.State = &adState
	if url != "" {
		status.TargetURL = &url
	}

	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestListOptions{}
	source, resp, err := g.Client.PullRequests.Get(g.ctx, owner, project, pull.Num, &opts)
	if err != nil {
		return fmt.Errorf("getting pull request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http response code %d getting pull request", resp.StatusCode)
	}
	if source.GetSupportsIterations() {
		opts := azuredevops.PullRequestIterationsListOptions{}
		iterations, resp, err := g.Client.PullRequests.ListIterations(g.ctx, owner, project, repoName, pull.Num, &opts)
		if err != nil {
			return fmt.Errorf("listing pull request iterations: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http response code %d listing pull request iterations", resp.StatusCode)
		}
		for _, iteration := range iterations {
			if sourceRef := iteration.GetSourceRefCommit(); sourceRef != nil {
				if *sourceRef.CommitID == pull.HeadCommit {
					status.IterationID = iteration.ID
					break
				}
			}
		}
		if iterationID := status.IterationID; iterationID != nil {
			if *iterationID < 1 {
				return errors.New("supportsIterations was true but got invalid iteration ID or no matching iteration commit SHA was found")
			}
		}
	}
	_, resp, err = g.Client.PullRequests.CreateStatus(g.ctx, owner, project, repoName, pull.Num, &status)
	if err != nil {
		return fmt.Errorf("creating pull request status: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http response code %d creating pull request status", resp.StatusCode)
	}
	return err
}

// MergePull merges the merge request using the default no fast-forward strategy
// If the user has set a branch policy that disallows no fast-forward, the merge will fail
// until we handle branch policies
// https://docs.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops
func (g *Client) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	owner, project, repoName := SplitAzureDevopsRepoFullName(pull.BaseRepo.FullName)
	descriptor := "Atlantis Terraform Pull Request Automation"

	userID, err := g.Client.UserEntitlements.GetUserID(g.ctx, g.UserName, owner)
	if err != nil {
		return fmt.Errorf("getting user id, User name: %s Organization %s : %w", g.UserName, owner, err)
	}
	if userID == nil {
		return fmt.Errorf("the user %s is not found in the organization %s", g.UserName, owner)
	}

	imageURL := "https://raw.githubusercontent.com/runatlantis/atlantis/main/runatlantis.io/public/hero.png"
	id := azuredevops.IdentityRef{
		Descriptor: &descriptor,
		ID:         userID,
		ImageURL:   &imageURL,
	}
	// Set default pull request completion options
	mcm := azuredevops.NoFastForward.String()
	twi := new(bool)
	*twi = true
	completionOpts := azuredevops.GitPullRequestCompletionOptions{
		BypassPolicy:            new(bool),
		BypassReason:            azuredevops.String(""),
		DeleteSourceBranch:      &pullOptions.DeleteSourceBranchOnMerge,
		MergeCommitMessage:      azuredevops.String(common.AutomergeCommitMsg(pull.Num)),
		MergeStrategy:           &mcm,
		SquashMerge:             new(bool),
		TransitionWorkItems:     twi,
		TriggeredByAutoComplete: new(bool),
	}

	// Construct request body from supplied parameters
	mergePull := new(azuredevops.GitPullRequest)
	mergePull.AutoCompleteSetBy = &id
	mergePull.CompletionOptions = &completionOpts

	mergeResult, _, err := g.Client.PullRequests.Merge(
		g.ctx,
		owner,
		project,
		repoName,
		pull.Num,
		mergePull,
		completionOpts,
		id,
	)
	if err != nil {
		return fmt.Errorf("merging pull request: %w", err)
	}
	if *mergeResult.MergeStatus != azuredevops.MergeSucceeded.String() {
		return fmt.Errorf("could not merge pull request: %s", mergeResult.GetMergeFailureMessage())
	}
	return nil
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *Client) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("!%d", pull.Num), nil
}

// SplitAzureDevopsRepoFullName splits a repo full name up into its owner,
// repo and project name segments. If the repoFullName is malformed, may
// return empty strings for owner, repo, or project.  Azure DevOps uses
// repoFullName format owner/project/repo.
//
// Ex. runatlantis/atlantis => (runatlantis, atlantis)
//
//	gitlab/subgroup/runatlantis/atlantis => (gitlab/subgroup/runatlantis, atlantis)
//	azuredevops/project/atlantis => (azuredevops, project, atlantis)
func SplitAzureDevopsRepoFullName(repoFullName string) (owner string, project string, repo string) {
	firstSlashIdx := strings.Index(repoFullName, "/")
	lastSlashIdx := strings.LastIndex(repoFullName, "/")
	slashCount := strings.Count(repoFullName, "/")
	if lastSlashIdx == -1 || lastSlashIdx == len(repoFullName)-1 {
		return "", "", ""
	}
	if firstSlashIdx != lastSlashIdx && slashCount == 2 {
		return repoFullName[:firstSlashIdx],
			repoFullName[firstSlashIdx+1 : lastSlashIdx],
			repoFullName[lastSlashIdx+1:]
	}
	return repoFullName[:lastSlashIdx], "", repoFullName[lastSlashIdx+1:]
}

// GetTeamNamesForUser returns the names of the teams or groups that the user belongs to (in the organization the repository belongs to).
func (g *Client) GetTeamNamesForUser(_ logging.SimpleLogging, _ models.Repo, _ models.User) ([]string, error) { //nolint: revive
	return nil, nil
}

func (g *Client) SupportsSingleFileDownload(repo models.Repo) bool { //nolint: revive
	return false
}

func (g *Client) GetFileContent(_ logging.SimpleLogging, _ models.Repo, _ string, _ string) (bool, []byte, error) { //nolint: revive
	return false, []byte{}, fmt.Errorf("not implemented")
}

// GitStatusContextFromSrc parses an Atlantis formatted src string into a context suitable
// for the status update API. In the AzureDevops branch policy UI there is a single string
// field used to drive these contexts where all text preceding the final '/' character is
// treated as the 'genre'.
func gitStatusContextFromSrc(src string) *azuredevops.GitStatusContext {
	lastSlashIdx := strings.LastIndex(src, "/")
	genre := "Atlantis Bot"
	name := src
	if lastSlashIdx != -1 {
		genre = fmt.Sprintf("%s/%s", genre, src[:lastSlashIdx])
		name = src[lastSlashIdx+1:]
	}

	return &azuredevops.GitStatusContext{
		Name:  &name,
		Genre: &genre,
	}
}

func (g *Client) GetCloneURL(_ logging.SimpleLogging, VCSHostType models.VCSHostType, repo string) (string, error) { //nolint: revive
	return "", fmt.Errorf("not yet implemented")
}

func (g *Client) GetPullLabels(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest) ([]string, error) {
	return nil, fmt.Errorf("not yet implemented")
}
