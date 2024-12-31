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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/xanzy/go-gitlab"
)

// gitlabMaxCommentLength is the maximum number of chars allowed by Gitlab in a
// single comment, reduced by 100 to allow comments to be hidden with a summary header
// and footer.
const gitlabMaxCommentLength = 1000000 - 100

type GitlabClient struct {
	Client *gitlab.Client
	// Version is set to the server version.
	Version *version.Version
	// PollingInterval is the time between successive polls, where applicable.
	PollingInterval time.Duration
	// PollingInterval is the total duration for which to poll, where applicable.
	PollingTimeout time.Duration
}

// commonMarkSupported is a version constraint that is true when this version of
// GitLab supports CommonMark, a markdown specification.
// See https://about.gitlab.com/2018/07/22/gitlab-11-1-released/
var commonMarkSupported = MustConstraint(">=11.1")

// gitlabClientUnderTest is true if we're running under go test.
var gitlabClientUnderTest = false

// NewGitlabClient returns a valid GitLab client.
func NewGitlabClient(hostname string, token string, logger logging.SimpleLogging) (*GitlabClient, error) {
	logger.Debug("Creating new GitLab client for %s", hostname)
	client := &GitlabClient{
		PollingInterval: time.Second,
		PollingTimeout:  time.Second * 30,
	}

	// Create the client differently depending on the base URL.
	if hostname == "gitlab.com" {
		glClient, err := gitlab.NewClient(token)
		if err != nil {
			return nil, err
		}
		client.Client = glClient
	} else {
		// We assume the url will be over HTTPS if the user doesn't specify a scheme.
		absoluteURL := hostname
		if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
			absoluteURL = "https://" + absoluteURL
		}

		url, err := url.Parse(absoluteURL)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing URL %q", absoluteURL)
		}

		// Warn if this hostname isn't resolvable. The GitLab client
		// doesn't give good error messages in this case.
		ips, err := net.LookupIP(url.Hostname())
		if err != nil {
			logger.Warn("unable to resolve %q: %s", url.Hostname(), err)
		} else if len(ips) == 0 {
			logger.Warn("found no IPs while resolving %q", url.Hostname())
		}

		// Now we're ready to construct the client.
		absoluteURL = strings.TrimSuffix(absoluteURL, "/")
		apiURL := fmt.Sprintf("%s/api/v4/", absoluteURL)
		glClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(apiURL))
		if err != nil {
			return nil, err
		}
		client.Client = glClient
	}

	// Determine which version of GitLab is running.
	if !gitlabClientUnderTest {
		var err error
		client.Version, err = client.GetVersion(logger)
		if err != nil {
			return nil, err
		}
		logger.Info("GitLab host '%s' is running version %s", client.Client.BaseURL().Host, client.Version.String())
	}

	return client, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *GitlabClient) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	logger.Debug("Getting modified files for GitLab merge request %d", pull.Num)
	const maxPerPage = 100
	var files []string
	nextPage := 1
	// Constructing the api url by hand so we can do pagination.
	apiURL := fmt.Sprintf("projects/%s/merge_requests/%d/changes", url.QueryEscape(repo.FullName), pull.Num)
	for {
		opts := gitlab.ListOptions{
			Page:    nextPage,
			PerPage: maxPerPage,
		}
		req, err := g.Client.NewRequest("GET", apiURL, opts, nil)
		if err != nil {
			return nil, err
		}
		resp := new(gitlab.Response)
		mr := new(gitlab.MergeRequest)
		pollingStart := time.Now()
		for {
			resp, err = g.Client.Do(req, mr)
			if resp != nil {
				logger.Debug("GET %s returned: %d", apiURL, resp.StatusCode)
			}
			if err != nil {
				return nil, err
			}
			if mr.ChangesCount != "" {
				break
			}
			if time.Since(pollingStart) > g.PollingTimeout {
				return nil, errors.Errorf("giving up polling %q after %s", apiURL, g.PollingTimeout.String())
			}
			time.Sleep(g.PollingInterval)
		}

		for _, f := range mr.Changes {
			files = append(files, f.NewPath)

			// If the file was renamed, we'll want to run plan in the directory
			// it was moved from as well.
			if f.RenamedFile {
				files = append(files, f.OldPath)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}

	return files, nil
}

// CreateComment creates a comment on the merge request.
func (g *GitlabClient) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, _ string) error {
	logger.Debug("Creating comment on GitLab merge request %d", pullNum)
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"
	comments := common.SplitComment(comment, gitlabMaxCommentLength, sepEnd, sepStart, 0, "")
	for _, c := range comments {
		_, resp, err := g.Client.Notes.CreateMergeRequestNote(repo.FullName, pullNum, &gitlab.CreateMergeRequestNoteOptions{Body: gitlab.Ptr(c)})
		if resp != nil {
			logger.Debug("POST /projects/%s/merge_requests/%d/notes returned: %d", repo.FullName, pullNum, resp.StatusCode)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ReactToComment adds a reaction to a comment.
func (g *GitlabClient) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	logger.Debug("Adding reaction '%s' to comment %d on GitLab merge request %d", reaction, commentID, pullNum)
	_, resp, err := g.Client.AwardEmoji.CreateMergeRequestAwardEmojiOnNote(repo.FullName, pullNum, int(commentID), &gitlab.CreateAwardEmojiOptions{Name: reaction})
	if resp != nil {
		logger.Debug("POST /projects/%s/merge_requests/%d/notes/%d/award_emoji returned: %d", repo.FullName, pullNum, commentID, resp.StatusCode)
	}
	return err
}

func (g *GitlabClient) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	logger.Debug("Hiding previous command comments on GitLab merge request %d", pullNum)
	var allComments []*gitlab.Note

	nextPage := 0
	for {
		logger.Debug("/projects/%v/merge_requests/%d/notes", repo.FullName, pullNum)
		comments, resp, err := g.Client.Notes.ListMergeRequestNotes(repo.FullName, pullNum,
			&gitlab.ListMergeRequestNotesOptions{
				Sort:        gitlab.Ptr("asc"),
				OrderBy:     gitlab.Ptr("created_at"),
				ListOptions: gitlab.ListOptions{Page: nextPage},
			})
		if resp != nil {
			logger.Debug("GET /projects/%s/merge_requests/%d/notes returned: %d", repo.FullName, pullNum, resp.StatusCode)
		}
		if err != nil {
			return errors.Wrap(err, "listing comments")
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}

	currentUser, _, err := g.Client.Users.CurrentUser()
	if err != nil {
		return errors.Wrap(err, "error getting currentuser")
	}

	summaryHeader := fmt.Sprintf("<!--- +-Superseded Command-+ ---><details><summary>Superseded Atlantis %s</summary>", command)
	summaryFooter := "</details>"
	lineFeed := "\n"

	for _, comment := range allComments {
		// Only process non-system comments authored by the Atlantis user
		if comment.System || (comment.Author.Username != "" && !strings.EqualFold(comment.Author.Username, currentUser.Username)) {
			continue
		}

		body := strings.Split(comment.Body, "\n")
		if len(body) == 0 {
			continue
		}
		firstLine := strings.ToLower(body[0])
		// Skip processing comments that don't contain the command or contain the summary header in the first line
		if !strings.Contains(firstLine, strings.ToLower(command)) || firstLine == strings.ToLower(summaryHeader) {
			continue
		}

		// If dir was specified, skip processing comments that don't contain the dir in the first line
		if dir != "" && !strings.Contains(firstLine, strings.ToLower(dir)) {
			continue
		}

		logger.Debug("Updating merge request note: Repo: '%s', MR: '%d', comment ID: '%d'", repo.FullName, pullNum, comment.ID)
		supersededComment := summaryHeader + lineFeed + comment.Body + lineFeed + summaryFooter + lineFeed

		_, resp, err := g.Client.Notes.UpdateMergeRequestNote(repo.FullName, pullNum, comment.ID, &gitlab.UpdateMergeRequestNoteOptions{Body: &supersededComment})
		if resp != nil {
			logger.Debug("PUT /projects/%s/merge_requests/%d/notes/%d returned: %d", repo.FullName, pullNum, comment.ID, resp.StatusCode)
		}
		if err != nil {
			return errors.Wrapf(err, "updating comment %d", comment.ID)
		}
	}

	return nil
}

// PullIsApproved returns true if the merge request was approved.
func (g *GitlabClient) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
	logger.Debug("Checking if GitLab merge request %d is approved", pull.Num)
	approvals, resp, err := g.Client.MergeRequests.GetMergeRequestApprovals(repo.FullName, pull.Num)
	if resp != nil {
		logger.Debug("GET /projects/%s/merge_requests/%d/approvals returned: %d", repo.FullName, pull.Num, resp.StatusCode)
	}
	if err != nil {
		return approvalStatus, err
	}
	if approvals.ApprovalsLeft > 0 {
		return approvalStatus, nil
	}
	return models.ApprovalStatus{
		IsApproved: true,
	}, nil
}

// PullIsMergeable returns true if the merge request can be merged.
// In GitLab, there isn't a single field that tells us if the pull request is
// mergeable so for now we check the merge_status and approvals_before_merge
// fields.
// In order to check if the repo required these, we'd need to make another API
// call to get the repo settings.
// It's also possible that GitLab implements their own "mergeable" field in
// their API in the future.
// See:
// - https://gitlab.com/gitlab-org/gitlab-ee/issues/3169
// - https://gitlab.com/gitlab-org/gitlab-ce/issues/42344
func (g *GitlabClient) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, _ []string) (bool, error) {
	logger.Debug("Checking if GitLab merge request %d is mergeable", pull.Num)
	mr, resp, err := g.Client.MergeRequests.GetMergeRequest(repo.FullName, pull.Num, nil)
	if resp != nil {
		logger.Debug("GET /projects/%s/merge_requests/%d returned: %d", repo.FullName, pull.Num, resp.StatusCode)
	}
	if err != nil {
		return false, err
	}

	// Prevent nil pointer error when mr.HeadPipeline is empty
	// See: https://github.com/runatlantis/atlantis/issues/1852
	commit := pull.HeadCommit
	isPipelineSkipped := false
	if mr.HeadPipeline != nil {
		commit = mr.HeadPipeline.SHA
		isPipelineSkipped = mr.HeadPipeline.Status == "skipped"
	}

	// Get project configuration
	project, resp, err := g.Client.Projects.GetProject(mr.ProjectID, nil)
	if resp != nil {
		logger.Debug("GET /projects/%d returned: %d", mr.ProjectID, resp.StatusCode)
	}
	if err != nil {
		return false, err
	}

	// Get Commit Statuses
	statuses, _, err := g.Client.Commits.GetCommitStatuses(mr.ProjectID, commit, nil)
	if resp != nil {
		logger.Debug("GET /projects/%d/commits/%s/statuses returned: %d", mr.ProjectID, commit, resp.StatusCode)
	}
	if err != nil {
		return false, err
	}

	for _, status := range statuses {
		// Ignore any commit statuses with 'atlantis/apply' as prefix
		if strings.HasPrefix(status.Name, fmt.Sprintf("%s/%s", vcsstatusname, command.Apply.String())) {
			continue
		}
		if !status.AllowFailure && project.OnlyAllowMergeIfPipelineSucceeds && status.Status != "success" {
			return false, nil
		}
	}

	allowSkippedPipeline := project.AllowMergeOnSkippedPipeline && isPipelineSkipped

	supportsDetailedMergeStatus, err := g.SupportsDetailedMergeStatus(logger)
	if err != nil {
		return false, err
	}

	if supportsDetailedMergeStatus {
		logger.Debug("Detailed merge status: '%s'", mr.DetailedMergeStatus)
	} else {
		logger.Debug("Merge status: '%s'", mr.MergeStatus) //nolint:staticcheck // Need to reference deprecated field for backwards compatibility
	}

	if ((supportsDetailedMergeStatus &&
		(mr.DetailedMergeStatus == "mergeable" ||
			mr.DetailedMergeStatus == "ci_still_running" ||
			mr.DetailedMergeStatus == "ci_must_pass" ||
			mr.DetailedMergeStatus == "need_rebase")) ||
		(!supportsDetailedMergeStatus &&
			mr.MergeStatus == "can_be_merged")) && //nolint:staticcheck // Need to reference deprecated field for backwards compatibility
		mr.ApprovalsBeforeMerge <= 0 &&
		mr.BlockingDiscussionsResolved &&
		!mr.WorkInProgress &&
		(allowSkippedPipeline || !isPipelineSkipped) {

		logger.Debug("Merge request is mergeable")
		return true, nil
	}
	logger.Debug("Merge request is not mergeable")
	return false, nil
}

func (g *GitlabClient) SupportsDetailedMergeStatus(logger logging.SimpleLogging) (bool, error) {
	logger.Debug("Checking if GitLab supports detailed merge status")
	v, err := g.GetVersion(logger)
	if err != nil {
		return false, err
	}

	cons, err := version.NewConstraint(">= 15.6")
	if err != nil {
		return false, err
	}
	return cons.Check(v), nil
}

// UpdateStatus updates the build status of a commit.
func (g *GitlabClient) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	gitlabState := gitlab.Pending
	switch state {
	case models.PendingCommitStatus:
		gitlabState = gitlab.Running
	case models.FailedCommitStatus:
		gitlabState = gitlab.Failed
	case models.SuccessCommitStatus:
		gitlabState = gitlab.Success
	}

	logger.Info("Updating GitLab commit status for '%s' to '%s'", src, gitlabState)

	setCommitStatusOptions := &gitlab.SetCommitStatusOptions{
		State:       gitlabState,
		Context:     gitlab.Ptr(src),
		Description: gitlab.Ptr(description),
		TargetURL:   &url,
	}

	retries := 1
	delay := 2 * time.Second
	var commit *gitlab.Commit
	var resp *gitlab.Response
	var err error

	// Try a couple of times to get the pipeline ID for the commit
	for i := 0; i <= retries; i++ {
		commit, resp, err = g.Client.Commits.GetCommit(repo.FullName, pull.HeadCommit, nil)
		if resp != nil {
			logger.Debug("GET /projects/%s/repository/commits/%d: %d", pull.BaseRepo.ID(), pull.HeadCommit, resp.StatusCode)
		}
		if err != nil {
			return err
		}
		if commit.LastPipeline != nil {
			logger.Info("Pipeline found for commit %s, setting pipeline ID to %d", pull.HeadCommit, commit.LastPipeline.ID)
			// Set the pipeline ID to the last pipeline that ran for the commit
			setCommitStatusOptions.PipelineID = gitlab.Ptr(commit.LastPipeline.ID)
			break
		}
		if i != retries {
			logger.Info("No pipeline found for commit %s, retrying in %s", pull.HeadCommit, delay)
			time.Sleep(delay)
		} else {
			// If we've exhausted all retries, set the Ref to the branch name
			logger.Info("No pipeline found for commit %s, setting Ref to %s", pull.HeadCommit, pull.HeadBranch)
			setCommitStatusOptions.Ref = gitlab.Ptr(pull.HeadBranch)
		}
	}

	var (
		maxAttempts = 10
		retryer     = &backoff.Backoff{
			Jitter: true,
			Max:    g.PollingInterval,
		}
	)

	for i := 0; i < maxAttempts; i++ {
		logger := logger.With(
			"attempt", i+1,
			"max_attempts", maxAttempts,
			"repo", repo.FullName,
			"commit", commit.ShortID,
			"state", state.String(),
		)

		_, resp, err = g.Client.Commits.SetCommitStatus(repo.FullName, pull.HeadCommit, setCommitStatusOptions)

		if resp != nil {
			logger.Debug("POST /projects/%s/statuses/%s returned: %d", repo.FullName, pull.HeadCommit, resp.StatusCode)

			// GitLab returns a `409 Conflict` status when the commit pipeline status is being changed/locked by another request,
			// which is likely to happen if you use [`--parallel-pool-size > 1`] and [`parallel-plan|apply`].
			//
			// The likelihood of this happening is increased when the number of parallel apply jobs is increased.
			//
			// Returning the [err] without retrying will permanently leave the GitLab commit status in a "running" state,
			// which would prevent Atlantis from merging the merge request on [apply].
			//
			// GitLab does not allow merge requests to be merged when the pipeline status is "running."

			if resp.StatusCode == http.StatusConflict {
				sleep := retryer.ForAttempt(float64(i))

				logger.With("retry_in", sleep).Warn("GitLab returned HTTP [409 Conflict] when updating commit status")
				time.Sleep(sleep)

				continue
			}
		}

		// Log we got a 200 OK response from GitLab after at least one retry to help with debugging/understanding delays/errors.
		if err == nil && i > 0 {
			logger.Info("GitLab returned HTTP [200 OK] after updating commit status")
		}

		// Return the err, which might be nil if everything worked out
		return err
	}

	// If we got here, we've exhausted all attempts to update the commit status and still failed, so return the error upstream
	return errors.Wrap(err, fmt.Sprintf("failed to update commit status for '%s' @ '%s' to '%s' after %d attempts", repo.FullName, pull.HeadCommit, src, maxAttempts))
}

func (g *GitlabClient) GetMergeRequest(logger logging.SimpleLogging, repoFullName string, pullNum int) (*gitlab.MergeRequest, error) {
	logger.Debug("Getting GitLab merge request %d", pullNum)
	mr, resp, err := g.Client.MergeRequests.GetMergeRequest(repoFullName, pullNum, nil)
	if resp != nil {
		logger.Debug("GET /projects/%s/merge_requests/%d returned: %d", repoFullName, pullNum, resp.StatusCode)
	}
	return mr, err
}

func (g *GitlabClient) WaitForSuccessPipeline(logger logging.SimpleLogging, ctx context.Context, pull models.PullRequest) {
	logger.Debug("Waiting for GitLab success pipeline for merge request %d", pull.Num)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for wait := true; wait; {
		select {
		case <-ctx.Done():
			// validation check time out
			cancel()
			return // ctx.Err()

		default:
			mr, _ := g.GetMergeRequest(logger, pull.BaseRepo.FullName, pull.Num)
			// check if pipeline has a success state to merge
			if mr.HeadPipeline.Status == "success" {
				return
			}
			time.Sleep(time.Second)
		}
	}
}

// MergePull merges the merge request.
func (g *GitlabClient) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	logger.Debug("Merging GitLab merge request %d", pull.Num)
	commitMsg := common.AutomergeCommitMsg(pull.Num)

	mr, err := g.GetMergeRequest(logger, pull.BaseRepo.FullName, pull.Num)
	if err != nil {
		return errors.Wrap(
			err, "unable to merge merge request, it was not possible to retrieve the merge request")
	}
	project, resp, err := g.Client.Projects.GetProject(mr.ProjectID, nil)
	if resp != nil {
		logger.Debug("GET /projects/%d returned: %d", mr.ProjectID, resp.StatusCode)
	}
	if err != nil {
		return errors.Wrap(
			err, "unable to merge merge request, it was not possible to check the project requirements")
	}

	if project != nil && project.OnlyAllowMergeIfPipelineSucceeds {
		g.WaitForSuccessPipeline(logger, context.Background(), pull)
	}

	_, resp, err = g.Client.MergeRequests.AcceptMergeRequest(
		pull.BaseRepo.FullName,
		pull.Num,
		&gitlab.AcceptMergeRequestOptions{
			MergeCommitMessage:       &commitMsg,
			ShouldRemoveSourceBranch: &pullOptions.DeleteSourceBranchOnMerge,
		})
	if resp != nil {
		logger.Debug("PUT /projects/%s/merge_requests/%d/merge returned: %d", pull.BaseRepo.FullName, pull.Num, resp.StatusCode)
	}
	return errors.Wrap(err, "unable to merge merge request, it may not be in a mergeable state")
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *GitlabClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("!%d", pull.Num), nil
}

func (g *GitlabClient) DiscardReviews(_ models.Repo, _ models.PullRequest) error {
	// TODO implement
	return nil
}

// GetVersion returns the version of the Gitlab server this client is using.
func (g *GitlabClient) GetVersion(logger logging.SimpleLogging) (*version.Version, error) {
	logger.Debug("Getting GitLab version")
	versionResp, resp, err := g.Client.Version.GetVersion()
	if resp != nil {
		logger.Debug("GET /version returned: %d", resp.StatusCode)
	}
	if err != nil {
		return nil, err
	}
	// We need to strip any "-ee" or similar from the resulting version because go-version
	// uses that in its constraints and it breaks the comparison we're trying
	// to do for Common Mark.
	split := strings.Split(versionResp.Version, "-")
	parsedVersion, err := version.NewVersion(split[0])
	if err != nil {
		return nil, errors.Wrapf(err, "parsing response to /version: %q", versionResp.Version)
	}
	return parsedVersion, nil
}

// SupportsCommonMark returns true if the version of Gitlab this client is
// using supports the CommonMark markdown format.
func (g *GitlabClient) SupportsCommonMark() bool {
	// This function is called even if we didn't construct a gitlab client
	// so we need to handle that case.
	if g == nil {
		return false
	}

	return commonMarkSupported.Check(g.Version)
}

// MustConstraint returns a constraint. It panics on error.
func MustConstraint(constraint string) version.Constraints {
	c, err := version.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return c
}

// GetTeamNamesForUser returns the names of the teams or groups that the user belongs to (in the organization the repository belongs to).
func (g *GitlabClient) GetTeamNamesForUser(_ models.Repo, _ models.User) ([]string, error) {
	return nil, nil
}

// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
// The first return value indicates whether the repo contains a file or not
// if BaseRepo had a file, its content will placed on the second return value
func (g *GitlabClient) GetFileContent(logger logging.SimpleLogging, pull models.PullRequest, fileName string) (bool, []byte, error) {
	logger.Debug("Getting GitLab file content for file '%s'", fileName)
	opt := gitlab.GetRawFileOptions{Ref: gitlab.Ptr(pull.HeadBranch)}

	bytes, resp, err := g.Client.RepositoryFiles.GetRawFile(pull.BaseRepo.FullName, fileName, &opt)
	if resp != nil {
		logger.Debug("GET /projects/%s/repository/files/%s/raw returned: %d", pull.BaseRepo.FullName, fileName, resp.StatusCode)
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, []byte{}, nil
	}

	if err != nil {
		return true, []byte{}, err
	}

	return true, bytes, nil
}

func (g *GitlabClient) SupportsSingleFileDownload(_ models.Repo) bool {
	return true
}

func (g *GitlabClient) GetCloneURL(logger logging.SimpleLogging, _ models.VCSHostType, repo string) (string, error) {
	logger.Debug("Getting GitLab clone URL for repo '%s'", repo)
	project, resp, err := g.Client.Projects.GetProject(repo, nil)
	if resp != nil {
		logger.Debug("GET /projects/%s returned: %d", repo, resp.StatusCode)
	}
	if err != nil {
		return "", err
	}
	return project.HTTPURLToRepo, nil
}

func (g *GitlabClient) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	logger.Debug("Getting GitLab labels for merge request %d", pull.Num)
	mr, resp, err := g.Client.MergeRequests.GetMergeRequest(repo.FullName, pull.Num, nil)
	if resp != nil {
		logger.Debug("GET /projects/%s/merge_requests/%d returned: %d", repo.FullName, pull.Num, resp.StatusCode)
	}

	if err != nil {
		return nil, err
	}

	return mr.Labels, nil
}
