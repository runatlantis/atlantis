// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/jpillora/backoff"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// maxCommentLength is the maximum number of chars allowed by Gitlab in a
// single comment, reduced by 100 to allow comments to be hidden with a summary header
// and footer.
const maxCommentLength = 1000000 - 100

type Client struct {
	Client *gitlab.Client
	// Version is set to the server version.
	Version *version.Version
	// All GitLab groups configured in allowlists and policies
	ConfiguredGroups []string
	// PollingInterval is the time between successive polls, where applicable.
	PollingInterval time.Duration
	// PollingInterval is the total duration for which to poll, where applicable.
	PollingTimeout time.Duration
	// StatusRetryEnabled enables enhanced retry logic for pipeline status updates.
	StatusRetryEnabled bool
}

// legacyMergeRequest captures fields that GitLab still returns in the API but
// gitlab.com/gitlab-org/api/client-go no longer exposes on MergeRequest.
type legacyMergeRequest struct {
	gitlab.MergeRequest
	MergeStatus          string `json:"merge_status"`
	ApprovalsBeforeMerge int    `json:"approvals_before_merge"`
}

func (m *legacyMergeRequest) UnmarshalJSON(data []byte) error {
	if err := m.MergeRequest.UnmarshalJSON(data); err != nil {
		return err
	}
	var legacyFields struct {
		MergeStatus          string `json:"merge_status"`
		ApprovalsBeforeMerge int    `json:"approvals_before_merge"`
	}
	if err := json.Unmarshal(data, &legacyFields); err != nil {
		return err
	}
	m.MergeStatus = legacyFields.MergeStatus
	m.ApprovalsBeforeMerge = legacyFields.ApprovalsBeforeMerge
	return nil
}

// commonMarkSupported is a version constraint that is true when this version of
// GitLab supports CommonMark, a markdown specification.
// See https://about.gitlab.com/2018/07/22/gitlab-11-1-released/
var commonMarkSupported = version.MustConstraints(version.NewConstraint(">=11.1"))

// gitlabClientUnderTest is true if we're running under go test.
var gitlabClientUnderTest = false

// NewClient returns a valid GitLab client.
func New(hostname string, token string, configuredGroups []string, logger logging.SimpleLogging) (*Client, error) {
	logger.Debug("Creating new GitLab client for %s", hostname)
	client := &Client{
		ConfiguredGroups: configuredGroups,
		PollingInterval:  time.Second,
		PollingTimeout:   time.Second * 30,
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
			return nil, fmt.Errorf("parsing URL %q: %w", absoluteURL, err)
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
func (g *Client) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	logger.Debug("Getting modified files for GitLab merge request %d", pull.Num)
	const maxPerPage = 100
	var files []string
	nextPage := 1
	apiURL := fmt.Sprintf("/projects/%s/merge_requests/%d", repo.FullName, pull.Num)
	pollingStart := time.Now()
	for {
		mr, resp, err := g.Client.MergeRequests.GetMergeRequest(repo.FullName, pull.Num, nil)
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
			return nil, fmt.Errorf("giving up polling %q after %s", apiURL, g.PollingTimeout.String())
		}
		time.Sleep(g.PollingInterval)
	}
	for {
		opts := &gitlab.ListMergeRequestDiffsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    nextPage,
				PerPage: maxPerPage,
			},
		}
		diffs, resp, err := g.Client.MergeRequests.ListMergeRequestDiffs(repo.FullName, pull.Num, opts)
		if resp != nil {
			logger.Debug("GET /projects/%s/merge_requests/%d/diffs returned: %d", repo.FullName, pull.Num, resp.StatusCode)
		}
		if err != nil {
			return nil, err
		}

		for _, f := range diffs {
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
func (g *Client) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	logger.Debug("Creating comment on GitLab merge request %d", pullNum)
	comments := common.SplitComment(logger, comment, maxCommentLength, 0, command)
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
func (g *Client) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	logger.Debug("Adding reaction '%s' to comment %d on GitLab merge request %d", reaction, commentID, pullNum)
	_, resp, err := g.Client.AwardEmoji.CreateMergeRequestAwardEmojiOnNote(repo.FullName, pullNum, int(commentID), &gitlab.CreateAwardEmojiOptions{Name: reaction})
	if resp != nil {
		logger.Debug("POST /projects/%s/merge_requests/%d/notes/%d/award_emoji returned: %d", repo.FullName, pullNum, commentID, resp.StatusCode)
	}
	return err
}

func (g *Client) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
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
			return fmt.Errorf("listing comments: %w", err)
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}

	currentUser, _, err := g.Client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("error getting currentuser: %w", err)
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
			return fmt.Errorf("updating comment %d: %w", comment.ID, err)
		}
	}

	return nil
}

// PullIsApproved returns true if the merge request was approved.
func (g *Client) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
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
func (g *Client) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, _ []string) (models.MergeableStatus, error) {
	logger.Debug("Checking if GitLab merge request %d is mergeable", pull.Num)
	apiURL := fmt.Sprintf("projects/%s/merge_requests/%d", gitlab.PathEscape(repo.FullName), pull.Num)
	req, err := g.Client.NewRequest("GET", apiURL, nil, nil)
	if err != nil {
		return models.MergeableStatus{}, err
	}
	lmr := &legacyMergeRequest{}
	resp, err := g.Client.Do(req, lmr)
	if resp != nil {
		logger.Debug("GET /projects/%s/merge_requests/%d returned: %d", repo.FullName, pull.Num, resp.StatusCode)
	}
	if err != nil {
		return models.MergeableStatus{}, err
	}
	mr := &lmr.MergeRequest

	// Prevent nil pointer error when mr.HeadPipeline is empty
	// See: https://github.com/runatlantis/atlantis/issues/1852
	commit := pull.HeadCommit
	if mr.HeadPipeline != nil {
		commit = mr.HeadPipeline.SHA
	}

	// Get project configuration
	project, resp, err := g.Client.Projects.GetProject(mr.ProjectID, nil)
	if resp != nil {
		logger.Debug("GET /projects/%d returned: %d", mr.ProjectID, resp.StatusCode)
	}
	if err != nil {
		return models.MergeableStatus{}, err
	}

	// Get Commit Statuses — paginate so the head-ref filter below cannot be
	// defeated by a noisy MR pushing the MR-owned status past the first page.
	const maxPerPage = 100
	var statuses []*gitlab.CommitStatus
	nextPage := 1
	for {
		statusOpts := &gitlab.GetCommitStatusesOptions{
			ListOptions: gitlab.ListOptions{Page: nextPage, PerPage: maxPerPage},
		}
		page, statusResp, err := g.Client.Commits.GetCommitStatuses(mr.ProjectID, commit, statusOpts)
		if statusResp != nil {
			logger.Debug("GET /projects/%d/commits/%s/statuses page %d returned: %d", mr.ProjectID, commit, nextPage, statusResp.StatusCode)
		}
		if err != nil {
			return models.MergeableStatus{}, err
		}
		statuses = append(statuses, page...)
		if statusResp.NextPage == 0 {
			break
		}
		nextPage = statusResp.NextPage
	}

	// GET /repository/commits/:sha/statuses returns every status posted against
	// the SHA across all refs. When the same SHA appears in multiple MRs
	// (force-pushed shared source branch, etc.) a status from another MR
	// leaks here and can self-block the current MR.
	//
	// Filter strategy:
	//   1. Statuses whose Ref equals refs/merge-requests/<iid>/head or
	//      refs/merge-requests/<iid>/merge are unambiguously owned by this MR.
	//   2. If any such MR-ref status exists for this SHA, restrict the
	//      evaluation to MR-ref + refless statuses + branch-only source_branch
	//      statuses. If the same status name appears on a current MR ref, the
	//      source_branch copy is dropped because the same branch may back several
	//      MRs and a stale status could leak.
	//   3. Otherwise fall back to source_branch + refless statuses, preserving
	//      behaviour for external CIs that post against the branch ref.
	// Empty Ref is always treated as MR-owned (backward-compat for callers
	// that post refless statuses).
	expectedHeadRef := fmt.Sprintf("refs/merge-requests/%d/head", mr.IID)
	expectedMergeRef := fmt.Sprintf("refs/merge-requests/%d/merge", mr.IID)
	isCurrentMRRef := func(ref string) bool {
		return ref == expectedHeadRef || ref == expectedMergeRef
	}
	hasMRRefStatus := false
	currentMRStatusNames := make(map[string]struct{})
	for _, status := range statuses {
		if isCurrentMRRef(status.Ref) {
			hasMRRefStatus = true
			currentMRStatusNames[status.Name] = struct{}{}
		}
	}
	isBranchOnlySourceBranchStatus := func(status *gitlab.CommitStatus) bool {
		if status.Ref != mr.SourceBranch {
			return false
		}
		_, hasCurrentMRStatusWithSameName := currentMRStatusNames[status.Name]
		return !hasCurrentMRStatusWithSameName
	}
	// Collect every blocking status (rather than returning on the first) so that
	// a per-project command requirement check can tell whether the only blockers
	// are plan statuses belonging to other projects in the same merge request.
	// See DefaultCommandRequirementHandler.
	blockingStatusState := make(map[string]string)
	var blockingStatuses []string
	for _, status := range statuses {
		if status.Ref != "" {
			if hasMRRefStatus {
				if !isCurrentMRRef(status.Ref) && !isBranchOnlySourceBranchStatus(status) {
					continue
				}
			} else if status.Ref != mr.SourceBranch {
				continue
			}
		}
		// Ignore Atlantis-owned commit statuses that can self-block apply.
		// Keep plan statuses in this check: a later failed or running specific
		// plan can leave an older .tfplan on disk, so it must still block apply.
		if isSkippableAtlantisCommitStatus(status.Name, vcsstatusname) {
			continue
		}
		if !status.AllowFailure && project.OnlyAllowMergeIfPipelineSucceeds && status.Status != "success" {
			if _, seen := blockingStatusState[status.Name]; !seen {
				blockingStatusState[status.Name] = status.Status
				blockingStatuses = append(blockingStatuses, status.Name)
			}
		}
	}
	supportsDetailedMergeStatus, err := g.SupportsDetailedMergeStatus(logger)
	if err != nil {
		return models.MergeableStatus{}, err
	}

	if supportsDetailedMergeStatus {
		logger.Debug("Detailed merge status: '%s'", mr.DetailedMergeStatus)
	} else {
		logger.Debug("Merge status: '%s'", lmr.MergeStatus)
	}

	res := isMergeable(mr, project, supportsDetailedMergeStatus, lmr.MergeStatus, lmr.ApprovalsBeforeMerge)
	if !res.IsMergeable {
		logger.Debug("Merge request is not mergeable")
		return res, nil
	}
	if len(blockingStatuses) > 0 {
		// Sort so the reported Reason and the BlockingStatuses slice are
		// deterministic, independent of the order GitLab returns statuses in.
		sort.Strings(blockingStatuses)
		res = models.MergeableStatus{
			IsMergeable:      false,
			Reason:           fmt.Sprintf("Pipeline %s has status %s", blockingStatuses[0], blockingStatusState[blockingStatuses[0]]),
			BlockingStatuses: blockingStatuses,
		}
	}
	if res.IsMergeable {
		logger.Debug("Merge request is mergeable")
	} else {
		logger.Debug("Merge request is not mergeable")
	}
	return res, nil
}

func isSkippableAtlantisCommitStatus(statusName string, vcsStatusName string) bool {
	prefix := vcsStatusName + "/"
	if !strings.HasPrefix(statusName, prefix) {
		return false
	}

	statusContext := strings.TrimPrefix(statusName, prefix)
	commandName, _, _ := strings.Cut(statusContext, ": ")
	switch commandName {
	case "apply", "policy_check", "pre_workflow_hook", "post_workflow_hook":
		return true
	default:
		return false
	}
}

// gitlabIsMergeable a pure function that encapsulates the tricky logic behind determining whether a gitlab MR is mergeable
// It doesn't make any external calls and cannot error, so is much easier to test
func isMergeable(mr *gitlab.MergeRequest, project *gitlab.Project, supportsDetailedMergeStatus bool, legacyMergeStatus string, legacyApprovalsBeforeMerge int) models.MergeableStatus {
	isPipelineSkipped := false
	if mr.HeadPipeline != nil {
		isPipelineSkipped = mr.HeadPipeline.Status == "skipped"
	}

	if legacyApprovalsBeforeMerge > 0 {
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      fmt.Sprintf("Still require %d approvals", legacyApprovalsBeforeMerge),
		}
	}
	if !mr.BlockingDiscussionsResolved {
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      "Blocking discussions unresolved",
		}
	}
	if mr.Draft || mr.WorkInProgress { //nolint:staticcheck // WorkInProgress is retained for older GitLab JSON responses.
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      "Work in progress",
		}
	}
	if isPipelineSkipped && !project.AllowMergeOnSkippedPipeline {
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      "Pipeline was skipped",
		}
	}

	if supportsDetailedMergeStatus {
		if mr.DetailedMergeStatus == "mergeable" ||
			mr.DetailedMergeStatus == "ci_still_running" ||
			mr.DetailedMergeStatus == "ci_must_pass" {
			return models.MergeableStatus{
				IsMergeable: true,
			}
		}
		return models.MergeableStatus{
			IsMergeable: false,
			Reason:      fmt.Sprintf("Merge status is %s", mr.DetailedMergeStatus),
		}
	}

	if legacyMergeStatus == "can_be_merged" {
		return models.MergeableStatus{
			IsMergeable: true,
		}
	}
	return models.MergeableStatus{
		IsMergeable: false,
		Reason:      fmt.Sprintf("Merge status is %s", legacyMergeStatus),
	}
}

func (g *Client) SupportsDetailedMergeStatus(logger logging.SimpleLogging) (bool, error) {
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
func (g *Client) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
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

	pipelineMaxAttempts := 2
	pipelineRetryer := &backoff.Backoff{
		Min: 2 * time.Second,
		Max: 2 * time.Second,
	}

	if g.StatusRetryEnabled {
		pipelineMaxAttempts = 5
		pipelineRetryer = &backoff.Backoff{
			Min:    2 * time.Second,
			Max:    5 * time.Second,
			Jitter: true,
		}
	}

	var commit *gitlab.Commit
	var resp *gitlab.Response
	var err error

	// Try a couple of times to get the pipeline ID for the commit
	for {
		attempt := int(pipelineRetryer.Attempt()) + 1
		commit, resp, err = g.Client.Commits.GetCommit(repo.FullName, pull.HeadCommit, nil)
		if resp != nil {
			logger.Debug("GET /projects/%s/repository/commits/%s: %d", pull.BaseRepo.ID(), pull.HeadCommit, resp.StatusCode)
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
		if attempt == pipelineMaxAttempts {
			// If we've exhausted all retries, set the Ref to the branch name
			logger.Info("No pipeline found for commit %s, setting Ref to %s", pull.HeadCommit, pull.HeadBranch)
			setCommitStatusOptions.Ref = gitlab.Ptr(pull.HeadBranch)
			break
		}
		sleep := pipelineRetryer.Duration()
		logger.Info("No pipeline found for commit %s, retrying in %s", pull.HeadCommit, sleep)
		time.Sleep(sleep)
	}

	var (
		maxAttempts = 10
		retryer     = &backoff.Backoff{
			Jitter: true,
			Max:    g.PollingInterval,
		}
	)

	for {
		attempt := int(retryer.Attempt()) + 1
		logger := logger.With(
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"repo", repo.FullName,
			"commit", commit.ShortID,
			"state", state.String(),
		)

		_, resp, err := g.Client.Commits.SetCommitStatus(repo.FullName, pull.HeadCommit, setCommitStatusOptions)
		if err == nil {
			if retryer.Attempt() > 0 {
				logger.Info("GitLab returned HTTP [200 OK] after updating commit status")
			}

			return nil
		}

		// If the error indicates the status is already 'running', we can treat it as a success.
		// This can happen with parallel jobs. See https://github.com/runatlantis/atlantis/issues/2685.
		if gitlabState == gitlab.Running && strings.Contains(err.Error(), "Cannot transition status via :run from :running") {
			logger.Info("Commit status is already 'running'; ignoring redundant update.")
			return nil
		}

		if attempt == maxAttempts {
			return fmt.Errorf("failed to update commit status for '%s' @ '%s' to '%s' after %d attempts: %w", repo.FullName, pull.HeadCommit, src, attempt, err)
		}

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
				logger.Warn("GitLab returned HTTP [409 Conflict] when updating commit status")
			}
		}

		sleep := retryer.Duration()

		logger.With("retry_in", sleep).Warn("GitLab errored when updating commit status: %s", err)
		time.Sleep(sleep)
	}
}

func (g *Client) GetMergeRequest(logger logging.SimpleLogging, repoFullName string, pullNum int) (*gitlab.MergeRequest, error) {
	logger.Debug("Getting GitLab merge request %d", pullNum)
	mr, resp, err := g.Client.MergeRequests.GetMergeRequest(repoFullName, pullNum, nil)
	if resp != nil {
		logger.Debug("GET /projects/%s/merge_requests/%d returned: %d", repoFullName, pullNum, resp.StatusCode)
	}
	return mr, err
}

func (g *Client) WaitForSuccessPipeline(logger logging.SimpleLogging, ctx context.Context, pull models.PullRequest) {
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
func (g *Client) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	logger.Debug("Merging GitLab merge request %d", pull.Num)
	commitMsg := common.AutomergeCommitMsg(pull.Num)

	mr, err := g.GetMergeRequest(logger, pull.BaseRepo.FullName, pull.Num)
	if err != nil {
		return fmt.Errorf("unable to merge merge request, it was not possible to retrieve the merge request: %w", err)
	}
	project, resp, err := g.Client.Projects.GetProject(mr.ProjectID, nil)
	if resp != nil {
		logger.Debug("GET /projects/%d returned: %d", mr.ProjectID, resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("unable to merge merge request, it was not possible to check the project requirements: %w", err)
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
	if err != nil {
		return fmt.Errorf("unable to merge merge request, it may not be in a mergeable state: %w", err)
	}
	return nil
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *Client) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("!%d", pull.Num), nil
}

// DiscardReviews discards all reviews on a pull request
// This is only available with a bot token and otherwise will return 401 unauthorized
// https://docs.gitlab.com/api/merge_request_approvals/#reset-approvals-of-a-merge-request
func (g *Client) DiscardReviews(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	logger.Debug("Reset approvals for merge request %d", pull.Num)
	resp, err := g.Client.MergeRequestApprovals.ResetApprovalsOfMergeRequest(repo.FullName, pull.Num)
	if resp != nil {
		logger.Debug("PUT /projects/%s/merge_requests/%d/reset_approvals returned: %d", repo.FullName, pull.Num, resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("unable to reset approvals: %w", err)
	}

	return nil
}

// GetVersion returns the version of the Gitlab server this client is using.
func (g *Client) GetVersion(logger logging.SimpleLogging) (*version.Version, error) {
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
		return nil, fmt.Errorf("parsing response to /version: %q: %w", versionResp.Version, err)
	}
	return parsedVersion, nil
}

// SupportsCommonMark returns true if the version of Gitlab this client is
// using supports the CommonMark markdown format.
func (g *Client) SupportsCommonMark() bool {
	// This function is called even if we didn't construct a gitlab client
	// so we need to handle that case.
	if g == nil {
		return false
	}

	return commonMarkSupported.Check(g.Version)
}

// GetTeamNamesForUser returns the names of the GitLab groups that the user belongs to.
// The user membership is checked in each group from configuredTeams, groups
// that the Atlantis user doesn't have access to are silently ignored.
func (g *Client) GetTeamNamesForUser(logger logging.SimpleLogging, _ models.Repo, user models.User) ([]string, error) {
	logger.Debug("Getting GitLab group names for user '%s'", user)
	var teamNames []string

	users, resp, err := g.Client.Users.ListUsers(&gitlab.ListUsersOptions{Username: &user.Username})
	if resp.StatusCode == http.StatusNotFound {
		return teamNames, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GET /users returned: %d: %w", resp.StatusCode, err)
	} else if len(users) == 0 {
		return nil, errors.New("GET /users returned no user")
	} else if len(users) > 1 {
		// Theoretically impossible, just being extra safe
		return nil, errors.New("GET /users returned more than 1 user")
	}
	userID := users[0].ID
	for _, groupName := range g.ConfiguredGroups {
		membership, resp, err := g.Client.GroupMembers.GetGroupMember(groupName, userID)
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("GET /groups/%s/members/%d returned: %d: %w", groupName, userID, resp.StatusCode, err)
		}
		if resp.StatusCode == http.StatusOK && membership.State == "active" {
			teamNames = append(teamNames, groupName)
		}
	}
	return teamNames, nil
}

// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
// The first return value indicates whether the repo contains a file or not
// if BaseRepo had a file, its content will placed on the second return value
func (g *Client) GetFileContent(logger logging.SimpleLogging, repo models.Repo, branch string, fileName string) (bool, []byte, error) {
	logger.Debug("Getting GitLab file content for file '%s'", fileName)
	opt := gitlab.GetRawFileOptions{Ref: gitlab.Ptr(branch)}

	bytes, resp, err := g.Client.RepositoryFiles.GetRawFile(repo.FullName, fileName, &opt)
	if resp != nil {
		logger.Debug("GET /projects/%s/repository/files/%s/raw returned: %d", repo.FullName, fileName, resp.StatusCode)
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, []byte{}, nil
	}

	if err != nil {
		return true, []byte{}, err
	}

	return true, bytes, nil
}

func (g *Client) SupportsSingleFileDownload(_ models.Repo) bool {
	return true
}

func (g *Client) GetCloneURL(logger logging.SimpleLogging, _ models.VCSHostType, repo string) (string, error) {
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

func (g *Client) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
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

func (g *Client) GetChildTeams(_ logging.SimpleLogging, _ models.Repo, _ string) ([]string, error) {
	return nil, nil
}
