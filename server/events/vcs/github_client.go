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
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Laisky/graphql"
	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/shurcooL/githubv4"
)

// maxCommentLength is the maximum number of chars allowed in a single comment
// by GitHub.
const (
	maxCommentLength = 65536
)

// allows for custom handling of github 404s
type PullRequestNotFound struct {
	Err error
}

func (p *PullRequestNotFound) Error() string {
	return "Pull request not found: " + p.Err.Error()
}

// GithubClient is used to perform GitHub actions.
type GithubClient struct {
	user                string
	client              *github.Client
	v4MutateClient      *graphql.Client
	ctx                 context.Context
	logger              logging.Logger
	mergeabilityChecker MergeabilityChecker
}

// GithubAppTemporarySecrets holds app credentials obtained from github after creation.
type GithubAppTemporarySecrets struct {
	// ID is the app id.
	ID int64
	// Key is the app's PEM-encoded key.
	Key string
	// Name is the app name.
	Name string
	// WebhookSecret is the generated webhook secret for this app.
	WebhookSecret string
	// URL is a link to the app, like https://github.com/apps/octoapp.
	URL string
}

// NewGithubClient returns a valid GitHub client.
func NewGithubClient(hostname string, credentials GithubCredentials, logger logging.Logger, mergeabilityChecker MergeabilityChecker) (*GithubClient, error) {
	transport, err := credentials.Client()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing github authentication transport")
	}

	var graphqlURL string
	var client *github.Client
	if hostname == "github.com" {
		client = github.NewClient(transport)
		graphqlURL = "https://api.github.com/graphql"
	} else {
		apiURL := resolveGithubAPIURL(hostname)
		client, err = github.NewEnterpriseClient(apiURL.String(), apiURL.String(), transport)
		if err != nil {
			return nil, err
		}
		graphqlURL = fmt.Sprintf("https://%s/api/graphql", apiURL.Host)
	}

	// shurcooL's githubv4 library has a client ctor, but it doesn't support schema
	// previews, which need custom Accept headers (https://developer.github.com/v4/previews)
	// So for now use the graphql client, since the githubv4 library was basically
	// a simple wrapper around it. And instead of using shurcooL's graphql lib, use
	// Laisky's, since shurcooL's doesn't support custom headers.
	// Once the Minimize Comment schema is official, this can revert back to using
	// shurcooL's libraries completely.
	v4MutateClient := graphql.NewClient(
		graphqlURL,
		transport,
		graphql.WithHeader("Accept", "application/vnd.github.queen-beryl-preview+json"),
	)

	user, err := credentials.GetUser()

	if err != nil {
		return nil, errors.Wrap(err, "getting user")
	}
	return &GithubClient{
		user:                user,
		client:              client,
		v4MutateClient:      v4MutateClient,
		ctx:                 context.Background(),
		logger:              logger,
		mergeabilityChecker: mergeabilityChecker,
	}, nil
}

func (g *GithubClient) GetRateLimits() (*github.RateLimits, error) {
	rateLimits, resp, err := g.client.RateLimits(g.ctx)

	if err != nil {
		g.logger.Error("error retrieving rate limits", map[string]interface{}{"err": err})
		return nil, errors.Wrap(err, "retrieving rate limits")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error retrieving rate limits: %s", resp.Status)
	}
	return rateLimits, nil
}

// GetModifiedFiles returns the names of files that were modified in the pull request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *GithubClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string
	nextPage := 0
	for {
		opts := github.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		pageFiles, resp, err := g.client.PullRequests.ListFiles(g.ctx, repo.Owner, repo.Name, pull.Num, &opts)
		if err != nil {
			return files, err
		}
		for _, f := range pageFiles {
			files = append(files, f.GetFilename())

			// If the file was renamed, we'll want to run plan in the directory
			// it was moved from as well.
			if f.GetStatus() == "renamed" {
				files = append(files, f.GetPreviousFilename())
			}
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}
	return files, nil
}

// CreateComment creates a comment on the pull request.
// If comment length is greater than the max comment length we split into
// multiple comments.
func (g *GithubClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	var sepStart string

	sepEnd := "\n```\n</details>" +
		"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."

	if command != "" {
		sepStart = fmt.Sprintf("Continued %s output from previous comment.\n<details><summary>Show Output</summary>\n\n", command) +
			"```diff\n"
	} else {
		sepStart = "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n" +
			"```diff\n"
	}

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart)
	for i := range comments {
		_, _, err := g.client.Issues.CreateComment(g.ctx, repo.Owner, repo.Name, pullNum, &github.IssueComment{Body: &comments[i]})
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GithubClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	var allComments []*github.IssueComment
	nextPage := 0
	for {
		comments, resp, err := g.client.Issues.ListComments(g.ctx, repo.Owner, repo.Name, pullNum, &github.IssueListCommentsOptions{
			Sort:        github.String("created"),
			Direction:   github.String("asc"),
			ListOptions: github.ListOptions{Page: nextPage},
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

	for _, comment := range allComments {
		// Using a case insensitive compare here because usernames aren't case
		// sensitive and users may enter their atlantis users with different
		// cases.
		if comment.User != nil && !strings.EqualFold(comment.User.GetLogin(), g.user) {
			continue
		}
		// Crude filtering: The comment templates typically include the command name
		// somewhere in the first line. It's a bit of an assumption, but seems like
		// a reasonable one, given we've already filtered the comments by the
		// configured Atlantis user.
		body := strings.Split(comment.GetBody(), "\n")
		if len(body) == 0 {
			continue
		}
		firstLine := strings.ToLower(body[0])
		if !strings.Contains(firstLine, strings.ToLower(command)) {
			continue
		}
		var m struct {
			MinimizeComment struct {
				MinimizedComment struct {
					IsMinimized       githubv4.Boolean
					MinimizedReason   githubv4.String
					ViewerCanMinimize githubv4.Boolean
				}
			} `graphql:"minimizeComment(input:$input)"`
		}
		input := map[string]interface{}{
			"input": githubv4.MinimizeCommentInput{
				Classifier: githubv4.ReportedContentClassifiersOutdated,
				SubjectID:  comment.GetNodeID(),
			},
		}
		if err := g.v4MutateClient.Mutate(g.ctx, &m, input); err != nil {
			return errors.Wrapf(err, "minimize comment %s", comment.GetNodeID())
		}
	}

	return nil
}

// PullIsApproved returns true if the pull request was approved.
func (g *GithubClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
	nextPage := 0
	for {
		opts := github.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		pageReviews, resp, err := g.client.PullRequests.ListReviews(g.ctx, repo.Owner, repo.Name, pull.Num, &opts)
		if err != nil {
			return approvalStatus, errors.Wrap(err, "getting reviews")
		}
		for _, review := range pageReviews {
			if review != nil && review.GetState() == "APPROVED" {
				return models.ApprovalStatus{
					IsApproved: true,
					ApprovedBy: *review.User.Login,
					Date:       *review.SubmittedAt,
				}, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}
	return approvalStatus, nil
}

// PullIsMergeable returns true if the pull request is mergeable.
func (g *GithubClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	githubPR, err := g.GetPullRequest(repo, pull.Num)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}

	statuses, err := g.GetRepoStatuses(repo, pull)

	if err != nil {
		return false, errors.Wrap(err, "getting commit statuses")
	}

	checks, err := g.GetRepoChecks(repo, pull)

	if err != nil {
		return false, errors.Wrapf(err, "getting check runs")
	}

	return g.mergeabilityChecker.Check(githubPR, statuses, checks), nil
}

func (g *GithubClient) GetPullRequestFromName(repoName string, repoOwner string, num int) (*github.PullRequest, error) {
	var err error
	var pull *github.PullRequest

	// GitHub has started to return 404's here (#1019) even after they send the webhook.
	// They've got some eventual consistency issues going on so we're just going
	// to retry up to 3 times with a 1s sleep.
	numRetries := 3
	retryDelay := 1 * time.Second
	for i := 0; i < numRetries; i++ {
		pull, _, err = g.client.PullRequests.Get(g.ctx, repoOwner, repoName, num)
		if err == nil {
			return pull, nil
		}
		ghErr, ok := err.(*github.ErrorResponse)
		if !ok || ghErr.Response.StatusCode != http.StatusNotFound {
			return pull, err
		}
		time.Sleep(retryDelay)
	}

	ghErr, ok := err.(*github.ErrorResponse)
	if ok && ghErr.Response.StatusCode == http.StatusNotFound {
		return pull, &PullRequestNotFound{Err: err}
	}
	return pull, err
}

// GetPullRequest returns the pull request.
func (g *GithubClient) GetPullRequest(repo models.Repo, num int) (*github.PullRequest, error) {
	return g.GetPullRequestFromName(repo.Name, repo.Owner, num)
}

func (g *GithubClient) GetRepoChecks(repo models.Repo, pull models.PullRequest) ([]*github.CheckRun, error) {
	nextPage := 0

	var results []*github.CheckRun

	for {
		opts := &github.ListCheckRunsOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		if nextPage != 0 {
			opts.Page = nextPage
		}

		result, response, err := g.client.Checks.ListCheckRunsForRef(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, opts)

		if err != nil {
			return results, errors.Wrapf(err, "getting check runs for page %d", nextPage)
		}

		results = append(results, result.CheckRuns...)

		if response.NextPage == 0 {
			break
		}
		nextPage = response.NextPage
	}

	return results, nil
}

func (g *GithubClient) GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error) {
	// Get Combined statuses

	nextPage := 0

	var result []*github.RepoStatus

	for {
		opts := github.ListOptions{
			// explicit default
			// https://developer.github.com/v3/repos/statuses/#list-commit-statuses-for-a-reference
			PerPage: 100,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}

		combinedStatus, response, err := g.client.Repositories.GetCombinedStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, &opts)
		result = append(result, combinedStatus.Statuses...)

		if err != nil {
			return nil, err
		}
		if response.NextPage == 0 {
			break
		}
		nextPage = response.NextPage
	}

	return result, nil
}

// UpdateStatus updates the status badge on the pull request.
// See https://github.com/blog/1227-commit-status-api.
func (g *GithubClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) error {
	ghState := "error"
	switch request.State {
	case models.PendingCommitStatus:
		ghState = "pending"
	case models.SuccessCommitStatus:
		ghState = "success"
	case models.FailedCommitStatus:
		ghState = "failure"
	}

	status := &github.RepoStatus{
		State:       github.String(ghState),
		Description: github.String(request.Description),
		Context:     github.String(request.StatusName),
		TargetURL:   &request.DetailsURL,
	}
	_, _, err := g.client.Repositories.CreateStatus(ctx, request.Repo.Owner, request.Repo.Name, request.Ref, status)
	return err
}

// [WENGINES-4643] TODO: Move the checks implementation to UpdateStatus once github checks is stable
// UpdateChecksStatus updates the status check
func (g *GithubClient) UpdateChecksStatus(ctx context.Context, request types.UpdateStatusRequest) error {
	// TODO: Implement updating github checks
	// - Get all checkruns for this SHA
	// - Match the UpdateReqIdentifier with the check run. If it exists, update the checkrun. If it does not, create a new check run.

	// Checks uses Status and Conlusion. Need to map models.CommitStatus to Status and Conclusion
	// Status -> queued, in_progress, completed
	// Conclusion -> failure, neutral, cancelled, timed_out, or action_required. (Optional. Required if you provide a status of "completed".)
	return nil
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *GithubClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("#%d", pull.Num), nil
}

// ExchangeCode returns a newly created app's info
func (g *GithubClient) ExchangeCode(code string) (*GithubAppTemporarySecrets, error) {
	ctx := context.Background()
	cfg, _, err := g.client.Apps.CompleteAppManifest(ctx, code)
	data := &GithubAppTemporarySecrets{
		ID:            cfg.GetID(),
		Key:           cfg.GetPEM(),
		WebhookSecret: cfg.GetWebhookSecret(),
		Name:          cfg.GetName(),
		URL:           cfg.GetHTMLURL(),
	}

	return data, err
}

// DownloadRepoConfigFile return `atlantis.yaml` content from VCS (which support fetch a single file from repository)
// The first return value indicate that repo contain atlantis.yaml or not
// if BaseRepo had one repo config file, its content will placed on the second return value
func (g *GithubClient) DownloadRepoConfigFile(pull models.PullRequest) (bool, []byte, error) {
	opt := github.RepositoryContentGetOptions{Ref: pull.HeadBranch}
	fileContent, _, resp, err := g.client.Repositories.GetContents(g.ctx, pull.BaseRepo.Owner, pull.BaseRepo.Name, config.AtlantisYAMLFilename, &opt)

	if resp.StatusCode == http.StatusNotFound {
		return false, []byte{}, nil
	}
	if err != nil {
		return true, []byte{}, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return true, []byte{}, err
	}

	return true, decodedData, nil
}

func (g *GithubClient) GetContents(owner, repo, branch, path string) ([]byte, error) {
	opt := github.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, resp, err := g.client.Repositories.GetContents(g.ctx, owner, repo, path, &opt)
	if err != nil {
		return []byte{}, errors.Wrap(err, "fetching file contents")
	}

	if resp.StatusCode == http.StatusNotFound {
		return []byte{}, fmt.Errorf("%s not found in %s/%s", path, owner, repo)
	}

	decodedData, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "decoding file content")
	}

	return decodedData, nil
}

func (g *GithubClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return true
}
