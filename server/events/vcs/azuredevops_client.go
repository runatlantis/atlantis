package vcs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
)

// AzureDevopsClient represents an Azure DevOps VCS client
type AzureDevopsClient struct {
	Client   *azuredevops.Client
	ctx      context.Context
	UserName string
}

// NewAzureDevopsClient returns a valid Azure DevOps client.
func NewAzureDevopsClient(hostname string, userName string, token string) (*AzureDevopsClient, error) {
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
			return nil, errors.Wrapf(err, "invalid azure devops hostname trying to parse %s", baseURL)
		}
		adClient.BaseURL = *base
	}

	client := &AzureDevopsClient{
		Client:   adClient,
		UserName: userName,
		ctx:      context.Background(),
	}

	return client, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *AzureDevopsClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string

	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	pullRequest, _, _ := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)

	targetRefName := strings.Replace(pullRequest.GetTargetRefName(), "refs/heads/", "", 1)
	sourceRefName := strings.Replace(pullRequest.GetSourceRefName(), "refs/heads/", "", 1)

	r, resp, err := g.Client.Git.GetDiffs(g.ctx, owner, project, repoName, targetRefName, sourceRefName)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "http response code %d getting diff %s to %s", resp.StatusCode, sourceRefName, targetRefName)
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
			// Convert the path to a relative path from the repo's root.
			relativePath = filepath.Clean("./" + change.GetSourceServerItem())
			files = append(files, relativePath)
		}
	}

	return files, nil
}

// CreateComment creates a comment on a pull request.
//
// If comment length is greater than the max comment length we split into
// multiple comments.
func (g *AzureDevopsClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"

	// maxCommentLength is the maximum number of chars allowed in a single comment
	// This length was copied from the Github client - haven't found documentation
	// or tested limit in Azure DevOps.
	const maxCommentLength = 150000

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart)
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

func (g *AzureDevopsClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	return nil
}

// PullIsApproved returns true if the merge request was approved by another reviewer.
// https://docs.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops#require-a-minimum-number-of-reviewers
func (g *AzureDevopsClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)
	if err != nil {
		return approvalStatus, errors.Wrap(err, "getting pull request")
	}

	for _, review := range adPull.Reviewers {
		if review == nil {
			continue
		}

		if review.IdentityRef.GetUniqueName() == adPull.GetCreatedBy().GetUniqueName() {
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

// PullIsMergeable returns true if the merge request can be merged.
func (g *AzureDevopsClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestGetOptions{IncludeWorkItemRefs: true}
	adPull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, pull.Num, &opts)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}

	if *adPull.MergeStatus != azuredevops.MergeSucceeded.String() {
		return false, nil
	}

	if *adPull.IsDraft {
		return false, nil
	}

	if *adPull.Status != azuredevops.PullActive.String() {
		return false, nil
	}

	projectID := *adPull.Repository.Project.ID
	artifactID := g.Client.PolicyEvaluations.GetPullRequestArtifactID(projectID, pull.Num)
	policyEvaluations, _, err := g.Client.PolicyEvaluations.List(g.ctx, owner, project, artifactID, &azuredevops.PolicyEvaluationsListOptions{})
	if err != nil {
		return false, errors.Wrap(err, "getting policy evaluations")
	}

	for _, policyEvaluation := range policyEvaluations {
		if !*policyEvaluation.Configuration.IsEnabled || *policyEvaluation.Configuration.IsDeleted {
			continue
		}

		// Ignore the Atlantis status, even if its set as a blocker.
		// This status should not be considered when evaluating if the pull request can be applied.
		settings := (policyEvaluation.Configuration.Settings).(map[string]interface{})
		if genre, ok := settings["statusGenre"]; ok && genre == "Atlantis Bot/atlantis" {
			if name, ok := settings["statusName"]; ok && name == "apply" {
				continue
			}
		}

		if *policyEvaluation.Configuration.IsBlocking && *policyEvaluation.Status != azuredevops.PolicyEvaluationApproved {
			return false, nil
		}
	}

	return true, nil
}

// GetPullRequest returns the pull request.
func (g *AzureDevopsClient) GetPullRequest(repo models.Repo, num int) (*azuredevops.GitPullRequest, error) {
	opts := azuredevops.PullRequestGetOptions{
		IncludeWorkItemRefs: true,
	}
	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)
	pull, _, err := g.Client.PullRequests.GetWithRepo(g.ctx, owner, project, repoName, num, &opts)
	return pull, err
}

// UpdateStatus updates the build status of a commit.
func (g *AzureDevopsClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	adState := azuredevops.GitError.String()
	switch request.State {
	case models.PendingVCSStatus:
		adState = azuredevops.GitPending.String()
	case models.SuccessVCSStatus:
		adState = azuredevops.GitSucceeded.String()
	case models.FailedVCSStatus:
		adState = azuredevops.GitFailed.String()
	}

	status := azuredevops.GitPullRequestStatus{}
	status.Context = GitStatusContextFromSrc(request.StatusName)
	status.Description = &request.Description
	status.State = &adState

	repo := request.Repo

	url := request.DetailsURL
	if url != "" {
		status.TargetURL = &url
	}

	owner, project, repoName := SplitAzureDevopsRepoFullName(repo.FullName)

	opts := azuredevops.PullRequestListOptions{}
	source, resp, err := g.Client.PullRequests.Get(g.ctx, owner, project, request.PullNum, &opts)
	if err != nil {
		return "", errors.Wrap(err, "getting pull request")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrapf(err, "http response code %d getting pull request", resp.StatusCode)
	}
	if source.GetSupportsIterations() {
		opts := azuredevops.PullRequestIterationsListOptions{}
		iterations, resp, err := g.Client.PullRequests.ListIterations(g.ctx, owner, project, repoName, request.PullNum, &opts)
		if err != nil {
			return "", errors.Wrap(err, "listing pull request iterations")
		}
		if resp.StatusCode != http.StatusOK {
			return "", errors.Wrapf(err, "http response code %d listing pull request iterations", resp.StatusCode)
		}
		for _, iteration := range iterations {
			if sourceRef := iteration.GetSourceRefCommit(); sourceRef != nil {
				if *sourceRef.CommitID == request.Ref {
					status.IterationID = iteration.ID
					break
				}
			}
		}
		if iterationID := status.IterationID; iterationID != nil {
			if !(*iterationID >= 1) {
				return "", errors.New("supportsIterations was true but got invalid iteration ID or no matching iteration commit SHA was found")
			}
		}
	}
	_, resp, err = g.Client.PullRequests.CreateStatus(g.ctx, owner, project, repoName, request.PullNum, &status)
	if err != nil {
		return "", errors.Wrap(err, "creating pull request status")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrapf(err, "http response code %d creating pull request status", resp.StatusCode)
	}
	return "", err
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *AzureDevopsClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
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

func (g *AzureDevopsClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return false
}

func (g *AzureDevopsClient) DownloadRepoConfigFile(pull models.PullRequest) (bool, []byte, error) {
	return false, []byte{}, fmt.Errorf("Not Implemented")
}

// GitStatusContextFromSrc parses an Atlantis formatted src string into a context suitable
// for the status update API. In the AzureDevops branch policy UI there is a single string
// field used to drive these contexts where all text preceding the final '/' character is
// treated as the 'genre'.
func GitStatusContextFromSrc(src string) *azuredevops.GitStatusContext {
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
