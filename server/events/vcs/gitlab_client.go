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

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs/common"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"

	"github.com/runatlantis/atlantis/server/events/models"
	gitlab "github.com/xanzy/go-gitlab"
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
	// logger
	logger logging.SimpleLogging
}

// commonMarkSupported is a version constraint that is true when this version of
// GitLab supports CommonMark, a markdown specification.
// See https://about.gitlab.com/2018/07/22/gitlab-11-1-released/
var commonMarkSupported = MustConstraint(">=11.1")

// gitlabClientUnderTest is true if we're running under go test.
var gitlabClientUnderTest = false

// NewGitlabClient returns a valid GitLab client.
func NewGitlabClient(hostname string, token string, logger logging.SimpleLogging) (*GitlabClient, error) {
	client := &GitlabClient{
		PollingInterval: time.Second,
		PollingTimeout:  time.Second * 30,
		logger:          logger,
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
		client.Version, err = client.GetVersion()
		if err != nil {
			return nil, err
		}
		logger.Info("determined GitLab is running version %s", client.Version.String())
	}

	return client, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *GitlabClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
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
func (g *GitlabClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	sepStart := "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"
	comments := common.SplitComment(comment, gitlabMaxCommentLength, sepEnd, sepStart)
	for _, c := range comments {
		if _, _, err := g.Client.Notes.CreateMergeRequestNote(repo.FullName, pullNum, &gitlab.CreateMergeRequestNoteOptions{Body: gitlab.String(c)}); err != nil {
			return err
		}
	}
	return nil
}

// ReactToComment adds a reaction to a comment.
func (g *GitlabClient) ReactToComment(repo models.Repo, pullNum int, commentID int64, reaction string) error {
	_, _, err := g.Client.AwardEmoji.CreateMergeRequestAwardEmojiOnNote(repo.FullName, pullNum, int(commentID), &gitlab.CreateAwardEmojiOptions{Name: reaction})
	return err
}

func (g *GitlabClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	var allComments []*gitlab.Note

	nextPage := 0
	for {
		g.logger.Debug("/projects/%v/merge_requests/%d/notes", repo.FullName, pullNum)
		comments, resp, err := g.Client.Notes.ListMergeRequestNotes(repo.FullName, pullNum,
			&gitlab.ListMergeRequestNotesOptions{
				Sort:        gitlab.String("asc"),
				OrderBy:     gitlab.String("created_at"),
				ListOptions: gitlab.ListOptions{Page: nextPage},
			})
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

		g.logger.Debug("Updating merge request note: Repo: '%s', MR: '%d', comment ID: '%d'", repo.FullName, pullNum, comment.ID)
		supersededComment := summaryHeader + lineFeed + comment.Body + lineFeed + summaryFooter + lineFeed

		if _, _, err := g.Client.Notes.UpdateMergeRequestNote(repo.FullName, pullNum, comment.ID,
			&gitlab.UpdateMergeRequestNoteOptions{Body: &supersededComment}); err != nil {
			return errors.Wrapf(err, "updating comment %d", comment.ID)
		}
	}

	return nil
}

// PullIsApproved returns true if the merge request was approved.
func (g *GitlabClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
	approvals, _, err := g.Client.MergeRequests.GetMergeRequestApprovals(repo.FullName, pull.Num)
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
func (g *GitlabClient) PullIsMergeable(repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	mr, _, err := g.Client.MergeRequests.GetMergeRequest(repo.FullName, pull.Num, nil)
	if err != nil {
		return false, err
	}

	// Prevent nil pointer error when mr.HeadPipeline is empty
	// See: https://github.com/runatlantis/atlantis/issues/1852
	commit := pull.HeadCommit
	isPipelineSkipped := true
	if mr.HeadPipeline != nil {
		commit = mr.HeadPipeline.SHA
		isPipelineSkipped = mr.HeadPipeline.Status == "skipped"
	}

	// Get project configuration
	project, _, err := g.Client.Projects.GetProject(mr.ProjectID, nil)
	if err != nil {
		return false, err
	}

	// Get Commit Statuses
	statuses, _, err := g.Client.Commits.GetCommitStatuses(mr.ProjectID, commit, nil)
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

	ok, err := g.SupportsDetailedMergeStatus()
	if err != nil {
		return false, err
	}

	if ((ok && (mr.DetailedMergeStatus == "mergeable" || mr.DetailedMergeStatus == "ci_still_running")) ||
		(!ok && mr.MergeStatus == "can_be_merged")) &&
		mr.ApprovalsBeforeMerge <= 0 &&
		mr.BlockingDiscussionsResolved &&
		!mr.WorkInProgress &&
		(allowSkippedPipeline || !isPipelineSkipped) {
		return true, nil
	}
	return false, nil
}

func (g *GitlabClient) SupportsDetailedMergeStatus() (bool, error) {
	v, err := g.GetVersion()
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
func (g *GitlabClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	gitlabState := gitlab.Pending
	switch state {
	case models.PendingCommitStatus:
		gitlabState = gitlab.Running
	case models.FailedCommitStatus:
		gitlabState = gitlab.Failed
	case models.SuccessCommitStatus:
		gitlabState = gitlab.Success
	}

	mr, err := g.GetMergeRequest(pull.BaseRepo.FullName, pull.Num)
	if err != nil {
		return err
	}
	// refTarget is set to current branch if no pipeline is assigned to the commit,
	// otherwise it is set to the pipeline created by the merge_request_event rule
	refTarget := pull.HeadBranch
	if mr.Pipeline != nil {
		switch mr.Pipeline.Source {
		case "merge_request_event":
			refTarget = fmt.Sprintf("refs/merge-requests/%d/head", pull.Num)
		}
	}
	_, _, err = g.Client.Commits.SetCommitStatus(repo.FullName, pull.HeadCommit, &gitlab.SetCommitStatusOptions{
		State:       gitlabState,
		Context:     gitlab.String(src),
		Description: gitlab.String(description),
		TargetURL:   &url,
		Ref:         gitlab.String(refTarget),
	})
	return err
}

func (g *GitlabClient) GetMergeRequest(repoFullName string, pullNum int) (*gitlab.MergeRequest, error) {
	mr, _, err := g.Client.MergeRequests.GetMergeRequest(repoFullName, pullNum, nil)
	return mr, err
}

func (g *GitlabClient) WaitForSuccessPipeline(ctx context.Context, pull models.PullRequest) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for wait := true; wait; {
		select {
		case <-ctx.Done():
			// validation check time out
			cancel()
			return //ctx.Err()

		default:
			mr, _ := g.GetMergeRequest(pull.BaseRepo.FullName, pull.Num)
			// check if pipeline has a success state to merge
			if mr.HeadPipeline.Status == "success" {
				return
			}
			time.Sleep(time.Second)
		}
	}
}

// MergePull merges the merge request.
func (g *GitlabClient) MergePull(pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	commitMsg := common.AutomergeCommitMsg(pull.Num)

	mr, err := g.GetMergeRequest(pull.BaseRepo.FullName, pull.Num)
	if err != nil {
		return errors.Wrap(
			err, "unable to merge merge request, it was not possible to retrieve the merge request")
	}
	project, _, err := g.Client.Projects.GetProject(mr.ProjectID, nil)
	if err != nil {
		return errors.Wrap(
			err, "unable to merge merge request, it was not possible to check the project requirements")
	}

	if project != nil && project.OnlyAllowMergeIfPipelineSucceeds {
		g.WaitForSuccessPipeline(context.Background(), pull)
	}

	_, _, err = g.Client.MergeRequests.AcceptMergeRequest(
		pull.BaseRepo.FullName,
		pull.Num,
		&gitlab.AcceptMergeRequestOptions{
			MergeCommitMessage:       &commitMsg,
			ShouldRemoveSourceBranch: &pullOptions.DeleteSourceBranchOnMerge,
		})
	return errors.Wrap(err, "unable to merge merge request, it may not be in a mergeable state")
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *GitlabClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("!%d", pull.Num), nil
}

func (g *GitlabClient) DiscardReviews(repo models.Repo, pull models.PullRequest) error {
	// TODO implement
	return nil
}

// GetVersion returns the version of the Gitlab server this client is using.
func (g *GitlabClient) GetVersion() (*version.Version, error) {
	versionResp, _, err := g.Client.Version.GetVersion()
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
func (g *GitlabClient) GetTeamNamesForUser(repo models.Repo, user models.User) ([]string, error) {
	return nil, nil
}

// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
// The first return value indicates whether the repo contains a file or not
// if BaseRepo had a file, its content will placed on the second return value
func (g *GitlabClient) GetFileContent(pull models.PullRequest, fileName string) (bool, []byte, error) {
	opt := gitlab.GetRawFileOptions{Ref: gitlab.String(pull.HeadBranch)}

	bytes, resp, err := g.Client.RepositoryFiles.GetRawFile(pull.BaseRepo.FullName, fileName, &opt)
	if resp.StatusCode == http.StatusNotFound {
		return false, []byte{}, nil
	}

	if err != nil {
		return true, []byte{}, err
	}

	return true, bytes, nil
}

func (g *GitlabClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return true
}

func (g *GitlabClient) GetCloneURL(VCSHostType models.VCSHostType, repo string) (string, error) {
	project, _, err := g.Client.Projects.GetProject(repo, nil)
	if err != nil {
		return "", err
	}
	return project.HTTPURLToRepo, nil
}
