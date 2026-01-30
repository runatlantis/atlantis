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
	"slices"
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
	config   Config
}

// NewClient returns a valid Azure DevOps client.
func New(hostname string, userName string, token string, config Config) (*Client, error) {
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
		config:   config,
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
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"

	// maxCommentLength is the maximum number of chars allowed in a single comment
	// This length was copied from the Github client - haven't found documentation
	// or tested limit in Azure DevOps.
	const maxCommentLength = 150000

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart, 0, "")
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

	for _, review := range adPull.Reviewers {
		if review == nil {
			continue
		}

		if review.GetUniqueName() == adPull.GetCreatedBy().GetUniqueName() {
			continue
		}

		if review.GetVote() == azuredevops.VoteApproved || review.GetVote() == azuredevops.VoteApprovedWithSuggestions {
			return models.ApprovalStatus{
				IsApproved: true,
			}, nil
		}
	}

	return approvalStatus, nil
}

func (g *Client) DiscardReviews(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error { //nolint: revive
	// TODO implement
	return nil
}

// PullIsMergeable returns true if the merge request can be merged.
// The vcsstatusname parameter specifies the VCS status name prefix (e.g., "atlantis").
// The ignoreVCSStatusNames parameter specifies additional status names to ignore when
// evaluating mergeability.
//
// When AllowMergeableBypassApply is enabled in the client config, the apply status check
// (with genre "Atlantis Bot/{vcsstatusname}" and name "apply") will be ignored, allowing
// the PR to be considered mergeable even if the apply check is failing.
func (g *Client) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (models.MergeableStatus, error) {
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestGetOptions{IncludeWorkItemRefs: true}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)
	if err != nil {
		return models.MergeableStatus{}, fmt.Errorf("getting pull request: %w", err)
	}

	if *adPull.MergeStatus != azuredevops.MergeSucceeded.String() {
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      fmt.Sprintf("PR has merge status: %s", *adPull.MergeStatus),
		}, nil
	}

	if *adPull.IsDraft {
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      "PR is a draft",
		}, nil
	}

	if *adPull.Status != azuredevops.PullActive.String() {
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      fmt.Sprintf("PR has status: %s", *adPull.Status),
		}, nil
	}

	projectID := *adPull.Repository.Project.ID
	artifactID := g.Client.PolicyEvaluations.GetPullRequestArtifactID(projectID, pull.Num)
	policyEvaluations, _, err := g.Client.PolicyEvaluations.List(g.ctx, owner, project, artifactID, &azuredevops.PolicyEvaluationsListOptions{})
	if err != nil {
		return models.MergeableStatus{}, fmt.Errorf("getting policy evaluations: %w", err)
	}

	// Build the expected genre for apply status based on vcsstatusname.
	// Azure DevOps status context format: genre="Atlantis Bot/{vcsstatusname}", name="apply"
	applyStatusGenre := fmt.Sprintf("Atlantis Bot/%s", vcsstatusname)

	for _, policyEvaluation := range policyEvaluations {
		if !*policyEvaluation.Configuration.IsEnabled || *policyEvaluation.Configuration.IsDeleted {
			continue
		}

		// Skip non-blocking policies - they don't affect mergeability.
		if !*policyEvaluation.Configuration.IsBlocking {
			continue
		}

		// Check if this is a status policy that should be ignored.
		settings, ok := (policyEvaluation.Configuration.Settings).(map[string]any)
		if ok {
			genre, hasGenre := settings["statusGenre"].(string)
			name, hasName := settings["statusName"].(string)

			if hasGenre && hasName {
				// Check if this status should be bypassed when AllowMergeableBypassApply is enabled.
				// This allows PRs to be mergeable even when the apply status check is failing.
				if g.config.AllowMergeableBypassApply && genre == applyStatusGenre && name == "apply" {
					logger.Debug("AllowMergeableBypassApply enabled - bypassing apply status policy (genre=%s, name=%s)", genre, name)
					continue
				}

				// Check if this status is in the ignore list.
				// The genre format is "Atlantis Bot/{vcsstatusname}", so we extract the vcsstatusname
				// and check if it matches any entry in ignoreVCSStatusNames.
				if shouldIgnoreStatus(genre, ignoreVCSStatusNames) {
					logger.Debug("Ignoring status policy as it matches ignoreVCSStatusNames (genre=%s, name=%s)", genre, name)
					continue
				}
			}
		}

		// If the policy is blocking and not approved, the PR is not mergeable.
		if *policyEvaluation.Status != azuredevops.PolicyEvaluationApproved {
			return models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("blocking policy not approved: %s", getPolicyDisplayName(policyEvaluation)),
			}, nil
		}
	}

	return models.MergeableStatus{
		IsMergeable: true,
	}, nil
}

// shouldIgnoreStatus checks if the given status should be ignored based on the
// ignoreVCSStatusNames list.
//
// The genre format is "Atlantis Bot/{vcsstatusname}" where vcsstatusname is the
// service identifier (e.g., "atlantis", "status1"). This function extracts the
// vcsstatusname and checks if it's in the ignore list.
//
// This matches GitHub's behavior where --ignore-vcs-status-names=status1 would
// ignore all checks from that service (status1/plan, status1/apply, etc.).
func shouldIgnoreStatus(genre string, ignoreVCSStatusNames []string) bool {
	// Extract the vcsstatusname from the genre.
	// Genre format: "Atlantis Bot/{vcsstatusname}"
	const prefix = "Atlantis Bot/"
	if !strings.HasPrefix(genre, prefix) {
		return false
	}

	vcsStatusName := strings.TrimPrefix(genre, prefix)
	if vcsStatusName == "" {
		return false
	}

	for _, ignoreName := range ignoreVCSStatusNames {
		if ignoreName != "" && vcsStatusName == ignoreName {
			return true
		}
	}
	return false
}

// getPolicyDisplayName returns a human-readable name for a policy evaluation.
func getPolicyDisplayName(policyEvaluation *azuredevops.PolicyEvaluationRecord) string {
	if policyEvaluation.Configuration == nil {
		return "unknown policy"
	}

	settings, ok := (policyEvaluation.Configuration.Settings).(map[string]any)
	if !ok {
		return "unknown policy"
	}

	// For status policies, return the status name.
	if genre, hasGenre := settings["statusGenre"].(string); hasGenre {
		if name, hasName := settings["statusName"].(string); hasName {
			return fmt.Sprintf("%s/%s", genre, name)
		}
	}

	// For other policies, try to get the display name from the type.
	if policyEvaluation.Configuration.Type != nil && policyEvaluation.Configuration.Type.DisplayName != nil {
		return *policyEvaluation.Configuration.Type.DisplayName
	}

	return "unknown policy"
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

// GetTeamNamesForUser returns the names of the security groups (including AAD groups)
// that the user belongs to in Azure DevOps. This is used for group-based authorization checks.
//
// The function works by:
// 1. Finding the user in Azure DevOps Graph API using their principal name (email)
// 2. Getting all group memberships for that user
// 3. Returning the display names of all groups
//
// Note: This requires the configured Azure DevOps token to have permissions to read
// the Graph API (vso.graph scope). The Graph API is hosted at vssps.dev.azure.com.
func (g *Client) GetTeamNamesForUser(logger logging.SimpleLogging, repo models.Repo, user models.User) ([]string, error) {
	owner, _, _ := SplitAzureDevopsRepoFullName(repo.FullName)

	// Step 1: Find the user in the Graph API
	userDescriptor, err := g.getUserDescriptor(owner, user.Username)
	if err != nil {
		return nil, fmt.Errorf("getting user descriptor: %w", err)
	}

	if userDescriptor == "" {
		logger.Debug("User %s not found in Azure DevOps Graph", user.Username)
		return nil, nil
	}

	logger.Debug("Found user %s with descriptor %s", user.Username, userDescriptor)

	// Step 2: Get all group memberships for the user
	memberships, err := g.getUserGroupMemberships(owner, userDescriptor)
	if err != nil {
		return nil, fmt.Errorf("getting user memberships: %w", err)
	}

	// Step 3: Get group details for each membership
	var groupNames []string
	for _, membership := range memberships {
		if membership.ContainerDescriptor == nil {
			continue
		}

		groupName, err := g.getGroupDisplayName(owner, *membership.ContainerDescriptor)
		if err != nil {
			logger.Debug("Failed to get group name for descriptor %s: %v", *membership.ContainerDescriptor, err)
			continue
		}

		if groupName != "" {
			groupNames = append(groupNames, groupName)
			logger.Debug("User %s is a member of group %s", user.Username, groupName)
		}
	}

	return groupNames, nil
}

// GraphUser represents a user in the Azure DevOps Graph API.
type GraphUser struct {
	SubjectKind    *string `json:"subjectKind,omitempty"`
	Domain         *string `json:"domain,omitempty"`
	PrincipalName  *string `json:"principalName,omitempty"`
	MailAddress    *string `json:"mailAddress,omitempty"`
	Origin         *string `json:"origin,omitempty"`
	OriginID       *string `json:"originId,omitempty"`
	DisplayName    *string `json:"displayName,omitempty"`
	Links          any     `json:"_links,omitempty"`
	URL            *string `json:"url,omitempty"`
	Descriptor     *string `json:"descriptor,omitempty"`
	MetaType       *string `json:"metaType,omitempty"`
	DirectoryAlias *string `json:"directoryAlias,omitempty"`
}

// graphUsersResponse represents the response from the Graph users API.
type graphUsersResponse struct {
	Count int         `json:"count"`
	Value []GraphUser `json:"value"`
}

// GraphMembership represents a membership relationship in the Azure DevOps Graph API.
type GraphMembership struct {
	ContainerDescriptor *string `json:"containerDescriptor,omitempty"`
	MemberDescriptor    *string `json:"memberDescriptor,omitempty"`
}

// graphMembershipsResponse represents the response from the Graph memberships API.
type graphMembershipsResponse struct {
	Count int               `json:"count"`
	Value []GraphMembership `json:"value"`
}

// GraphGroup represents a group in the Azure DevOps Graph API.
type GraphGroup struct {
	SubjectKind   *string `json:"subjectKind,omitempty"`
	Description   *string `json:"description,omitempty"`
	Domain        *string `json:"domain,omitempty"`
	PrincipalName *string `json:"principalName,omitempty"`
	MailAddress   *string `json:"mailAddress,omitempty"`
	Origin        *string `json:"origin,omitempty"`
	OriginID      *string `json:"originId,omitempty"`
	DisplayName   *string `json:"displayName,omitempty"`
	Links         any     `json:"_links,omitempty"`
	URL           *string `json:"url,omitempty"`
	Descriptor    *string `json:"descriptor,omitempty"`
}

// getUserDescriptor finds a user in Azure DevOps Graph API and returns their descriptor.
// The Graph API is hosted at vssps.dev.azure.com (different from the main API).
// API: GET https://vssps.dev.azure.com/{organization}/_apis/graph/users
func (g *Client) getUserDescriptor(organization, principalName string) (string, error) {
	// The Graph API uses a different base URL: vssps.dev.azure.com
	// We need to construct the URL based on the current base URL
	graphBaseURL := g.getGraphAPIBaseURL()

	urlStr := fmt.Sprintf("%s%s/_apis/graph/users?subjectTypes=aad,msa&api-version=7.1-preview.1",
		graphBaseURL,
		url.PathEscape(organization))

	req, err := g.Client.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	var response graphUsersResponse
	_, err = g.Client.Execute(g.ctx, req, &response)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}

	// Search for the user by principal name (email) or mail address
	for _, user := range response.Value {
		if user.Descriptor == nil {
			continue
		}

		// Match by principal name, mail address, or display name
		if (user.PrincipalName != nil && strings.EqualFold(*user.PrincipalName, principalName)) ||
			(user.MailAddress != nil && strings.EqualFold(*user.MailAddress, principalName)) ||
			(user.DisplayName != nil && strings.EqualFold(*user.DisplayName, principalName)) {
			return *user.Descriptor, nil
		}
	}

	return "", nil
}

// getUserGroupMemberships returns all group memberships for a user.
// API: GET https://vssps.dev.azure.com/{organization}/_apis/graph/memberships/{subjectDescriptor}?direction=up
func (g *Client) getUserGroupMemberships(organization, userDescriptor string) ([]GraphMembership, error) {
	graphBaseURL := g.getGraphAPIBaseURL()

	urlStr := fmt.Sprintf("%s%s/_apis/graph/memberships/%s?direction=up&api-version=7.1-preview.1",
		graphBaseURL,
		url.PathEscape(organization),
		url.PathEscape(userDescriptor))

	req, err := g.Client.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var response graphMembershipsResponse
	_, err = g.Client.Execute(g.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return response.Value, nil
}

// getGroupDisplayName returns the display name of a group given its descriptor.
// API: GET https://vssps.dev.azure.com/{organization}/_apis/graph/groups/{groupDescriptor}
func (g *Client) getGroupDisplayName(organization, groupDescriptor string) (string, error) {
	graphBaseURL := g.getGraphAPIBaseURL()

	urlStr := fmt.Sprintf("%s%s/_apis/graph/groups/%s?api-version=7.1-preview.1",
		graphBaseURL,
		url.PathEscape(organization),
		url.PathEscape(groupDescriptor))

	req, err := g.Client.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	var group GraphGroup
	_, err = g.Client.Execute(g.ctx, req, &group)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}

	if group.DisplayName != nil {
		return *group.DisplayName, nil
	}

	return "", nil
}

// getGraphAPIBaseURL returns the base URL for the Azure DevOps Graph API.
// The Graph API is hosted at vssps.dev.azure.com instead of dev.azure.com.
func (g *Client) getGraphAPIBaseURL() string {
	// Default Graph API URL
	graphBaseURL := "https://vssps.dev.azure.com/"

	// If using a custom hostname (on-premises), adjust accordingly
	currentBase := g.Client.BaseURL.String()
	if strings.Contains(currentBase, "dev.azure.com") {
		graphBaseURL = "https://vssps.dev.azure.com/"
	} else if strings.Contains(currentBase, "visualstudio.com") {
		// For legacy visualstudio.com URLs
		graphBaseURL = "https://vssps.visualstudio.com/"
	}
	// For on-premises Azure DevOps Server, the Graph API might be at the same host
	// This would need to be tested/configured for specific deployments

	return graphBaseURL
}

// ValidateBypassMerge checks if a user is allowed to perform a bypass merge and
// returns an audit message to be added as a PR comment.
//
// This method implements the vcs.BypassMergeChecker interface.
//
// When AllowMergeableBypassApply is enabled and BypassMergeRequirementTeams is configured,
// only users who are members of at least one of the specified teams can perform bypass merges.
// If BypassMergeRequirementTeams is empty, any user can perform bypass merges.
//
// The audit message includes information about:
// - Who performed the bypass merge
// - Which policies were bypassed
// - When the bypass occurred
func (g *Client) ValidateBypassMerge(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, user models.User, vcsstatusname string) (allowed bool, auditMessage string, err error) {
	// If bypass is not enabled, nothing to validate
	if !g.config.AllowMergeableBypassApply {
		return true, "", nil
	}

	// Check if bypass was actually used (apply status is failing)
	bypassUsed, err := g.isApplyStatusFailing(logger, repo, pull, vcsstatusname)
	if err != nil {
		logger.Debug("Error checking apply status: %v", err)
		// If we can't check, assume bypass wasn't used
		return true, "", nil
	}

	if !bypassUsed {
		// Apply status is passing, no bypass needed
		return true, "", nil
	}

	// Bypass was used - check team permissions if configured
	if len(g.config.BypassMergeRequirementTeams) > 0 {
		// Get user's teams
		userTeams, err := g.GetTeamNamesForUser(logger, repo, user)
		if err != nil {
			return false, "", fmt.Errorf("failed to get user teams: %w", err)
		}

		// Check if user is in any of the allowed teams
		userAllowed := false
		for _, allowedTeam := range g.config.BypassMergeRequirementTeams {
			if slices.Contains(userTeams, allowedTeam) {
				userAllowed = true
				break
			}
		}

		if !userAllowed {
			logger.Info("User %s is not in any of the bypass merge teams: %v (user teams: %v)",
				user.Username, g.config.BypassMergeRequirementTeams, userTeams)
			return false, "", nil
		}

		logger.Info("User %s is authorized to perform bypass merge (member of allowed team)", user.Username)
	}

	// Generate audit message
	auditMessage = g.generateBypassAuditMessage(user, vcsstatusname)

	return true, auditMessage, nil
}

// isApplyStatusFailing checks if the apply status check is failing for the given PR.
func (g *Client) isApplyStatusFailing(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestGetOptions{IncludeWorkItemRefs: true}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)
	if err != nil {
		return false, fmt.Errorf("getting pull request: %w", err)
	}

	projectID := *adPull.Repository.Project.ID
	artifactID := g.Client.PolicyEvaluations.GetPullRequestArtifactID(projectID, pull.Num)
	policyEvaluations, _, err := g.Client.PolicyEvaluations.List(g.ctx, owner, project, artifactID, &azuredevops.PolicyEvaluationsListOptions{})
	if err != nil {
		return false, fmt.Errorf("getting policy evaluations: %w", err)
	}

	// Build the expected genre for apply status
	applyStatusGenre := fmt.Sprintf("Atlantis Bot/%s", vcsstatusname)

	for _, policyEvaluation := range policyEvaluations {
		if !*policyEvaluation.Configuration.IsEnabled || *policyEvaluation.Configuration.IsDeleted {
			continue
		}

		settings, ok := (policyEvaluation.Configuration.Settings).(map[string]any)
		if !ok {
			continue
		}

		genre, hasGenre := settings["statusGenre"].(string)
		name, hasName := settings["statusName"].(string)

		if hasGenre && hasName && genre == applyStatusGenre && name == "apply" {
			// Found the apply status - check if it's failing
			if *policyEvaluation.Status != azuredevops.PolicyEvaluationApproved {
				return true, nil
			}
		}
	}

	return false, nil
}

// generateBypassAuditMessage creates an audit message for a bypass merge.
func (g *Client) generateBypassAuditMessage(user models.User, vcsstatusname string) string {
	return fmt.Sprintf(`:warning: **Bypass Merge Audit**

This pull request was merged with the **%s/apply** status check bypassed.

| Field | Value |
|-------|-------|
| **User** | @%s |
| **Bypassed Check** | %s/apply |

:information_source: This merge was authorized because the user is a member of an allowed bypass team.`,
		vcsstatusname, user.Username, vcsstatusname)
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
