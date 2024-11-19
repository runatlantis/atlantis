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
	"maps"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/shurcooL/githubv4"
)

// maxCommentLength is the maximum number of chars allowed in a single comment
// by GitHub.
const maxCommentLength = 65536

var (
	clientMutationID            = githubv4.NewString("atlantis")
	pullRequestDismissalMessage = *githubv4.NewString("Dismissing reviews because of plan changes")
)

type GithubRepoIdCacheEntry struct {
	RepoId     githubv4.Int
	LookupTime time.Time
}

type GitHubRepoIdCache struct {
	cache map[githubv4.String]GithubRepoIdCacheEntry
}

func NewGitHubRepoIdCache() GitHubRepoIdCache {
	return GitHubRepoIdCache{
		cache: make(map[githubv4.String]GithubRepoIdCacheEntry),
	}
}

func (c *GitHubRepoIdCache) Get(key githubv4.String) (githubv4.Int, bool) {
	entry, ok := c.cache[key]
	if !ok {
		return githubv4.Int(0), false
	}
	if time.Since(entry.LookupTime) > time.Hour {
		delete(c.cache, key)
		return githubv4.Int(0), false
	}
	return entry.RepoId, true
}

func (c *GitHubRepoIdCache) Set(key githubv4.String, value githubv4.Int) {
	c.cache[key] = GithubRepoIdCacheEntry{
		RepoId:     value,
		LookupTime: time.Now(),
	}
}

// GithubClient is used to perform GitHub actions.
type GithubClient struct {
	user                  string
	client                *github.Client
	v4Client              *githubv4.Client
	ctx                   context.Context
	config                GithubConfig
	maxCommentsPerCommand int
	repoIdCache           GitHubRepoIdCache
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

type GithubReview struct {
	ID          githubv4.ID
	SubmittedAt githubv4.DateTime
	Author      struct {
		Login githubv4.String
	}
}

type GithubPRReviewSummary struct {
	ReviewDecision githubv4.String
	Reviews        []GithubReview
}

// NewGithubClient returns a valid GitHub client.

func NewGithubClient(hostname string, credentials GithubCredentials, config GithubConfig, maxCommentsPerCommand int, logger logging.SimpleLogging) (*GithubClient, error) {
	logger.Debug("Creating new GitHub client for host: %s", hostname)
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
		// TODO: Deprecated: Use NewClient(httpClient).WithEnterpriseURLs(baseURL, uploadURL) instead
		client, err = github.NewEnterpriseClient(apiURL.String(), apiURL.String(), transport) //nolint:staticcheck
		if err != nil {
			return nil, err
		}
		graphqlURL = fmt.Sprintf("https://%s/api/graphql", apiURL.Host)
	}

	// Use the client from shurcooL's githubv4 library for queries.
	v4Client := githubv4.NewEnterpriseClient(graphqlURL, transport)

	user, err := credentials.GetUser()
	logger.Debug("GH User: %s", user)

	if err != nil {
		return nil, errors.Wrap(err, "getting user")
	}

	return &GithubClient{
		user:                  user,
		client:                client,
		v4Client:              v4Client,
		ctx:                   context.Background(),
		config:                config,
		maxCommentsPerCommand: maxCommentsPerCommand,
		repoIdCache:           NewGitHubRepoIdCache(),
	}, nil
}

// GetModifiedFiles returns the names of files that were modified in the pull request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *GithubClient) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	logger.Debug("Getting modified files for GitHub pull request %d", pull.Num)
	var files []string
	nextPage := 0

listloop:
	for {
		opts := github.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		// GitHub has started to return 404's sometimes. They've got some
		// eventual consistency issues going on so we're just going to attempt
		// up to 5 times for each page with exponential backoff.
		maxAttempts := 5
		attemptDelay := 0 * time.Second
		for i := 0; i < maxAttempts; i++ {
			// First don't sleep, then sleep 1, 3, 7, etc.
			time.Sleep(attemptDelay)
			attemptDelay = 2*attemptDelay + 1*time.Second

			pageFiles, resp, err := g.client.PullRequests.ListFiles(g.ctx, repo.Owner, repo.Name, pull.Num, &opts)
			if resp != nil {
				logger.Debug("[attempt %d] GET /repos/%v/%v/pulls/%d/files returned: %v", i+1, repo.Owner, repo.Name, pull.Num, resp.StatusCode)
			}
			if err != nil {
				ghErr, ok := err.(*github.ErrorResponse)
				if ok && ghErr.Response.StatusCode == 404 {
					// (hopefully) transient 404, retry after backoff
					continue
				}
				// something else, give up
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
				break listloop
			}
			nextPage = resp.NextPage
			break
		}
	}
	return files, nil
}

// CreateComment creates a comment on the pull request.
// If comment length is greater than the max comment length we split into
// multiple comments.
func (g *GithubClient) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	logger.Debug("Creating comment on GitHub pull request %d", pullNum)
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

	truncationHeader := "> [!WARNING]\n" +
		"> **Warning**: Command output is larger than the maximum number of comments per command. Output truncated.\n<details><summary>Show Output</summary>\n\n" +
		"```diff\n"

	comments := common.SplitComment(comment, maxCommentLength, sepEnd, sepStart, g.maxCommentsPerCommand, truncationHeader)
	for i := range comments {
		_, resp, err := g.client.Issues.CreateComment(g.ctx, repo.Owner, repo.Name, pullNum, &github.IssueComment{Body: &comments[i]})
		if resp != nil {
			logger.Debug("POST /repos/%v/%v/issues/%d/comments returned: %v", repo.Owner, repo.Name, pullNum, resp.StatusCode)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ReactToComment adds a reaction to a comment.
func (g *GithubClient) ReactToComment(logger logging.SimpleLogging, repo models.Repo, _ int, commentID int64, reaction string) error {
	logger.Debug("Adding reaction to GitHub pull request comment %d", commentID)
	_, resp, err := g.client.Reactions.CreateIssueCommentReaction(g.ctx, repo.Owner, repo.Name, commentID, reaction)
	if resp != nil {
		logger.Debug("POST /repos/%v/%v/issues/comments/%d/reactions returned: %v", repo.Owner, repo.Name, commentID, resp.StatusCode)
	}
	return err
}

func (g *GithubClient) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	logger.Debug("Hiding previous command comments on GitHub pull request %d", pullNum)
	var allComments []*github.IssueComment
	nextPage := 0
	for {
		comments, resp, err := g.client.Issues.ListComments(g.ctx, repo.Owner, repo.Name, pullNum, &github.IssueListCommentsOptions{
			Sort:        github.String("created"),
			Direction:   github.String("asc"),
			ListOptions: github.ListOptions{Page: nextPage},
		})
		if resp != nil {
			logger.Debug("GET /repos/%v/%v/issues/%d/comments returned: %v", repo.Owner, repo.Name, pullNum, resp.StatusCode)
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

		// If dir was specified, skip processing comments that don't contain the dir in the first line
		if dir != "" && !strings.Contains(firstLine, strings.ToLower(dir)) {
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
		input := githubv4.MinimizeCommentInput{
			Classifier: githubv4.ReportedContentClassifiersOutdated,
			SubjectID:  comment.GetNodeID(),
		}
		logger.Debug("Hiding comment %s", comment.GetNodeID())
		if err := g.v4Client.Mutate(g.ctx, &m, input, nil); err != nil {
			return errors.Wrapf(err, "minimize comment %s", comment.GetNodeID())
		}
	}

	return nil
}

// getPRReviews Retrieves PR reviews for a pull request on a specific repository.
// The reviews are being retrieved using pages with the size of 10 reviews.
func (g *GithubClient) getPRReviews(repo models.Repo, pull models.PullRequest) (GithubPRReviewSummary, error) {
	var query struct {
		Repository struct {
			PullRequest struct {
				ReviewDecision githubv4.String
				Reviews        struct {
					Nodes []GithubReview
					// contains pagination information
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage githubv4.Boolean
					}
				} `graphql:"reviews(first: $entries, after: $reviewCursor, states: $reviewState)"`
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":        githubv4.String(repo.Owner),
		"name":         githubv4.String(repo.Name),
		"number":       githubv4.Int(pull.Num),
		"entries":      githubv4.Int(10),
		"reviewState":  []githubv4.PullRequestReviewState{githubv4.PullRequestReviewStateApproved},
		"reviewCursor": (*githubv4.String)(nil), // initialize the reviewCursor with null
	}

	var allReviews []GithubReview
	for {
		err := g.v4Client.Query(g.ctx, &query, variables)
		if err != nil {
			return GithubPRReviewSummary{
				query.Repository.PullRequest.ReviewDecision,
				allReviews,
			}, errors.Wrap(err, "getting reviewDecision")
		}

		allReviews = append(allReviews, query.Repository.PullRequest.Reviews.Nodes...)
		// if we don't have a NextPage pointer, we have requested all pages
		if !query.Repository.PullRequest.Reviews.PageInfo.HasNextPage {
			break
		}
		// set the end cursor, so the next batch of reviews is going to be requested and not the same again
		variables["reviewCursor"] = githubv4.NewString(query.Repository.PullRequest.Reviews.PageInfo.EndCursor)
	}
	return GithubPRReviewSummary{
		query.Repository.PullRequest.ReviewDecision,
		allReviews,
	}, nil
}

// PullIsApproved returns true if the pull request was approved.
func (g *GithubClient) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (approvalStatus models.ApprovalStatus, err error) {
	logger.Debug("Checking if GitHub pull request %d is approved", pull.Num)
	nextPage := 0
	for {
		opts := github.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		pageReviews, resp, err := g.client.PullRequests.ListReviews(g.ctx, repo.Owner, repo.Name, pull.Num, &opts)
		if resp != nil {
			logger.Debug("GET /repos/%v/%v/pulls/%d/reviews returned: %v", repo.Owner, repo.Name, pull.Num, resp.StatusCode)
		}
		if err != nil {
			return approvalStatus, errors.Wrap(err, "getting reviews")
		}
		for _, review := range pageReviews {
			if review != nil && review.GetState() == "APPROVED" {
				return models.ApprovalStatus{
					IsApproved: true,
					ApprovedBy: *review.User.Login,
					Date:       review.SubmittedAt.Time,
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

// DiscardReviews dismisses all reviews on a pull request
func (g *GithubClient) DiscardReviews(repo models.Repo, pull models.PullRequest) error {
	reviewStatus, err := g.getPRReviews(repo, pull)
	if err != nil {
		return err
	}

	// https://docs.github.com/en/graphql/reference/input-objects#dismisspullrequestreviewinput
	var mutation struct {
		DismissPullRequestReview struct {
			PullRequestReview struct {
				ID githubv4.ID
			}
		} `graphql:"dismissPullRequestReview(input: $input)"`
	}

	// dismiss every review one by one.
	// currently there is no way to dismiss them in one mutation.
	for _, review := range reviewStatus.Reviews {
		input := githubv4.DismissPullRequestReviewInput{
			PullRequestReviewID: review.ID,
			Message:             pullRequestDismissalMessage,
			ClientMutationID:    clientMutationID,
		}
		mutationResult := &mutation
		err := g.v4Client.Mutate(g.ctx, mutationResult, input, nil)
		if err != nil {
			return errors.Wrap(err, "dismissing reviewDecision")
		}
	}
	return nil
}

type PageInfo struct {
	EndCursor   *githubv4.String
	HasNextPage githubv4.Boolean
}

type WorkflowFileReference struct {
	Path         githubv4.String
	RepositoryId githubv4.Int
	Sha          *githubv4.String
}

func (original WorkflowFileReference) Copy() WorkflowFileReference {
	copy := WorkflowFileReference{
		Path:         original.Path,
		RepositoryId: original.RepositoryId,
		Sha:          new(githubv4.String),
	}
	if original.Sha != nil {
		*copy.Sha = *original.Sha
	}
	return copy
}

type WorkflowRun struct {
	File struct {
		Path              githubv4.String
		RepositoryFileUrl githubv4.String
		RepositoryName    githubv4.String
	}
	RunNumber githubv4.Int
}

type CheckRun struct {
	Name       githubv4.String
	Conclusion githubv4.String
	// Not currently used: GitHub API classifies as required if coming from ruleset, even when the ruleset is not enforced!
	IsRequired githubv4.Boolean `graphql:"isRequired(pullRequestNumber: $number)"`
	CheckSuite struct {
		WorkflowRun *WorkflowRun
	}
}

func (original CheckRun) Copy() CheckRun {
	copy := CheckRun{
		Name:       original.Name,
		Conclusion: original.Conclusion,
		IsRequired: original.IsRequired,
		CheckSuite: original.CheckSuite,
	}
	if original.CheckSuite.WorkflowRun != nil {
		copy.CheckSuite.WorkflowRun = new(WorkflowRun)
		*copy.CheckSuite.WorkflowRun = *original.CheckSuite.WorkflowRun
	}
	return copy
}

type StatusContext struct {
	Context githubv4.String
	State   githubv4.String
	// Not currently used: GitHub API classifies as required if coming from ruleset, even when the ruleset is not enforced!
	IsRequired githubv4.Boolean `graphql:"isRequired(pullRequestNumber: $number)"`
}

func (g *GithubClient) LookupRepoId(repo githubv4.String) (githubv4.Int, error) {
	// This function may get many calls for the same repo, and repo names are not often changed
	// Utilize caching to reduce the number of API calls to GitHub
	if repoId, ok := g.repoIdCache.Get(repo); ok {
		return repoId, nil
	}

	repoSplit := strings.Split(string(repo), "/")
	if len(repoSplit) != 2 {
		return githubv4.Int(0), fmt.Errorf("invalid repository name: %s", repo)
	}

	var query struct {
		Repository struct {
			DatabaseId githubv4.Int
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(repoSplit[0]),
		"name":  githubv4.String(repoSplit[1]),
	}

	err := g.v4Client.Query(g.ctx, &query, variables)

	if err != nil {
		return githubv4.Int(0), errors.Wrap(err, "getting repository id from GraphQL")
	}

	g.repoIdCache.Set(repo, query.Repository.DatabaseId)

	return query.Repository.DatabaseId, nil
}

func (g *GithubClient) WorkflowRunMatchesWorkflowFileReference(workflowRun WorkflowRun, workflowFileReference WorkflowFileReference) (bool, error) {
	// Unfortunately, the GitHub API doesn't expose the repositoryId for the WorkflowRunFile from the statusCheckRollup.
	// Conversely, it doesn't expose the repository name for the WorkflowFileReference from the RepositoryRuleConnection.
	// Therefore, a second query is required to lookup the association between repositoryId and repositoryName.
	repoId, err := g.LookupRepoId(workflowRun.File.RepositoryName)
	if err != nil {
		return false, err
	}

	if !(repoId == workflowFileReference.RepositoryId && workflowRun.File.Path == workflowFileReference.Path) {
		return false, nil
	} else if workflowFileReference.Sha != nil {
		return strings.Contains(string(workflowRun.File.RepositoryFileUrl), string(*workflowFileReference.Sha)), nil
	} else {
		return true, nil
	}
}

func (g *GithubClient) GetPullRequestMergeabilityInfo(
	repo models.Repo,
	pull *github.PullRequest,
) (
	reviewDecision githubv4.String,
	requiredChecks []githubv4.String,
	requiredWorkflows []WorkflowFileReference,
	checkRuns []CheckRun,
	statusContexts []StatusContext,
	err error,
) {
	var query struct {
		Repository struct {
			PullRequest struct {
				ReviewDecision githubv4.String
				BaseRef        struct {
					BranchProtectionRule struct {
						RequiredStatusChecks []struct {
							Context githubv4.String
						}
					}
					Rules struct {
						PageInfo PageInfo
						Nodes    []struct {
							Type              githubv4.String
							RepositoryRuleset struct {
								Enforcement githubv4.String
							}
							Parameters struct {
								RequiredStatusChecksParameters struct {
									RequiredStatusChecks []struct {
										Context githubv4.String
									}
								} `graphql:"... on RequiredStatusChecksParameters"`
								WorkflowsParameters struct {
									Workflows []WorkflowFileReference
								} `graphql:"... on WorkflowsParameters"`
							}
						}
					} `graphql:"rules(first: 100, after: $ruleCursor)"`
				}
				Commits struct {
					Nodes []struct {
						Commit struct {
							StatusCheckRollup struct {
								Contexts struct {
									PageInfo PageInfo
									Nodes    []struct {
										Typename      githubv4.String `graphql:"__typename"`
										CheckRun      CheckRun        `graphql:"... on CheckRun"`
										StatusContext StatusContext   `graphql:"... on StatusContext"`
									}
								} `graphql:"contexts(first: 100, after: $contextCursor)"`
							}
						}
					}
				} `graphql:"commits(last: 1)"`
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":         githubv4.String(repo.Owner),
		"name":          githubv4.String(repo.Name),
		"number":        githubv4.Int(*pull.Number),
		"ruleCursor":    (*githubv4.String)(nil),
		"contextCursor": (*githubv4.String)(nil),
	}

	requiredChecksSet := make(map[githubv4.String]any)

pagination:
	for {
		err = g.v4Client.Query(g.ctx, &query, variables)

		if err != nil {
			break pagination
		}

		reviewDecision = query.Repository.PullRequest.ReviewDecision

		for _, rule := range query.Repository.PullRequest.BaseRef.BranchProtectionRule.RequiredStatusChecks {
			requiredChecksSet[rule.Context] = struct{}{}
		}

		for _, rule := range query.Repository.PullRequest.BaseRef.Rules.Nodes {
			if rule.RepositoryRuleset.Enforcement != "ACTIVE" {
				continue
			}
			switch rule.Type {
			case "REQUIRED_STATUS_CHECKS":
				for _, context := range rule.Parameters.RequiredStatusChecksParameters.RequiredStatusChecks {
					requiredChecksSet[context.Context] = struct{}{}
				}
			case "WORKFLOWS":
				for _, workflow := range rule.Parameters.WorkflowsParameters.Workflows {
					requiredWorkflows = append(requiredWorkflows, workflow.Copy())
				}
			default:
				continue
			}
		}

		if len(query.Repository.PullRequest.Commits.Nodes) == 0 {
			err = errors.New("no commits found on PR")
			break pagination
		}

		for _, context := range query.Repository.PullRequest.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes {
			switch context.Typename {
			case "CheckRun":
				checkRuns = append(checkRuns, context.CheckRun.Copy())
			case "StatusContext":
				statusContexts = append(statusContexts, context.StatusContext)
			default:
				err = fmt.Errorf("unknown type of status check, %q", context.Typename)
				break pagination
			}
		}

		if !query.Repository.PullRequest.BaseRef.Rules.PageInfo.HasNextPage &&
			!query.Repository.PullRequest.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.PageInfo.HasNextPage {
			break pagination
		}

		if query.Repository.PullRequest.BaseRef.Rules.PageInfo.EndCursor != nil {
			variables["ruleCursor"] = query.Repository.PullRequest.BaseRef.Rules.PageInfo.EndCursor
		}
		if query.Repository.PullRequest.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.PageInfo.EndCursor != nil {
			variables["contextCursor"] = query.Repository.PullRequest.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.PageInfo.EndCursor
		}
	}

	if err != nil {
		return "", nil, nil, nil, nil, errors.Wrap(err, "fetching rulesets, branch protections and status checks from GraphQL")
	}

	for context := range requiredChecksSet {
		requiredChecks = append(requiredChecks, context)
	}

	return reviewDecision, requiredChecks, requiredWorkflows, checkRuns, statusContexts, nil
}

func CheckRunPassed(checkRun CheckRun) bool {
	return checkRun.Conclusion == "SUCCESS" || checkRun.Conclusion == "SKIPPED" || checkRun.Conclusion == "NEUTRAL"
}

func StatusContextPassed(statusContext StatusContext, vcsstatusname string) bool {
	return statusContext.State == "SUCCESS"
}

func ExpectedCheckPassed(expectedContext githubv4.String, checkRuns []CheckRun, statusContexts []StatusContext, vcsstatusname string) bool {
	// If there's no WorkflowRun, we assume there's only one CheckRun with the given name.
	// In this case, we evaluate and return the status of this CheckRun.
	// If there is WorkflowRun, we assume there can be multiple checkRuns with the given name,
	// so we retrieve the latest checkRun and evaluate and return the status of the latest CheckRun.
	latestCheckRunNumber := githubv4.Int(-1)
	var latestCheckRun *CheckRun
	for _, checkRun := range checkRuns {
		if checkRun.Name != expectedContext {
			continue
		}
		if checkRun.CheckSuite.WorkflowRun == nil {
			return CheckRunPassed(checkRun)
		}
		if checkRun.CheckSuite.WorkflowRun.RunNumber > latestCheckRunNumber {
			latestCheckRunNumber = checkRun.CheckSuite.WorkflowRun.RunNumber
			latestCheckRun = &checkRun
		}
	}

	if latestCheckRun != nil {
		return CheckRunPassed(*latestCheckRun)
	}

	for _, statusContext := range statusContexts {
		if statusContext.Context == expectedContext {
			return StatusContextPassed(statusContext, vcsstatusname)
		}
	}

	return false
}

func (g *GithubClient) ExpectedWorkflowPassed(expectedWorkflow WorkflowFileReference, checkRuns []CheckRun) (bool, error) {
	// If there's no WorkflowRun, we just skip evaluation for given CheckRun.
	// If there is WorkflowRun, we assume there can be multiple checkRuns with the given name,
	// so we retrieve the latest checkRun and evaluate and return the status of the latest CheckRun.
	latestCheckRunNumber := githubv4.Int(-1)
	var latestCheckRun *CheckRun
	for _, checkRun := range checkRuns {
		if checkRun.CheckSuite.WorkflowRun == nil {
			continue
		}
		match, err := g.WorkflowRunMatchesWorkflowFileReference(*checkRun.CheckSuite.WorkflowRun, expectedWorkflow)
		if err != nil {
			return false, err
		}
		if match {
			if checkRun.CheckSuite.WorkflowRun.RunNumber > latestCheckRunNumber {
				latestCheckRunNumber = checkRun.CheckSuite.WorkflowRun.RunNumber
				latestCheckRun = &checkRun
			}
		}
	}

	if latestCheckRun != nil {
		return CheckRunPassed(*latestCheckRun), nil
	}

	return false, nil
}

// IsMergeableMinusApply checks review decision (which takes into account CODEOWNERS) and required checks for PR (excluding the atlantis apply check).
func (g *GithubClient) IsMergeableMinusApply(logger logging.SimpleLogging, repo models.Repo, pull *github.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (bool, error) {
	if pull.Number == nil {
		return false, errors.New("pull request number is nil")
	}
	reviewDecision, requiredChecks, requiredWorkflows, checkRuns, statusContexts, err := g.GetPullRequestMergeabilityInfo(repo, pull)
	if err != nil {
		return false, err
	}

	notMergeablePrefix := fmt.Sprintf("Pull Request %s/%s:%s is not mergeable", repo.Owner, repo.Name, strconv.Itoa(*pull.Number))

	// Review decision takes CODEOWNERS into account
	// Empty review decision means review is not required
	if reviewDecision != "APPROVED" && len(reviewDecision) != 0 {
		logger.Debug("%s: Review Decision: %s", notMergeablePrefix, reviewDecision)
		return false, nil
	}

	// The statusCheckRollup does not always contain all required checks
	// For example, if a check was made required after the pull request was opened, it would be missing
	// Go through all checks and workflows required by branch protection or rulesets
	// Make sure that they can all be found in the statusCheckRollup and that they all pass
	for _, requiredCheck := range requiredChecks {
		if strings.HasPrefix(string(requiredCheck), fmt.Sprintf("%s/%s", vcsstatusname, command.Apply.String())) {
			// Ignore atlantis apply check(s)
			continue
		}
		if !slices.Contains(ignoreVCSStatusNames, GetVCSStatusNameFromRequiredCheck(requiredCheck)) && !ExpectedCheckPassed(requiredCheck, checkRuns, statusContexts, vcsstatusname) {
			logger.Debug("%s: Expected Required Check: %s VCS Status Name: %s Ignore VCS Status Names: %s", notMergeablePrefix, requiredCheck, vcsstatusname, ignoreVCSStatusNames)
			return false, nil
		}
	}
	for _, requiredWorkflow := range requiredWorkflows {
		passed, err := g.ExpectedWorkflowPassed(requiredWorkflow, checkRuns)
		if err != nil {
			return false, err
		}
		if !passed {
			logger.Debug("%s: Expected Required Workflow: RepositoryId: %d Path: %s", notMergeablePrefix, requiredWorkflow.RepositoryId, requiredWorkflow.Path)
			return false, nil
		}
	}

	return true, nil
}

func GetVCSStatusNameFromRequiredCheck(requiredCheck githubv4.String) string {
	return strings.Split(string(requiredCheck), "/")[0]
}

// PullIsMergeable returns true if the pull request is mergeable.
func (g *GithubClient) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (bool, error) {
	logger.Debug("Checking if GitHub pull request %d is mergeable", pull.Num)
	githubPR, err := g.GetPullRequest(logger, repo, pull.Num)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}

	// We map our mergeable check to when the GitHub merge button is clickable.
	// This corresponds to the following states:
	// clean: No conflicts, all requirements satisfied.
	//        Merging is allowed (green box).
	// unstable: Failing/pending commit status that is not part of the required
	//           status checks. Merging is allowed (yellow box).
	// has_hooks: GitHub Enterprise only, if a repo has custom pre-receive
	//            hooks. Merging is allowed (green box).
	// See: https://github.com/octokit/octokit.net/issues/1763
	switch githubPR.GetMergeableState() {
	case "clean", "unstable", "has_hooks":
		return true, nil
	case "blocked":
		if g.config.AllowMergeableBypassApply {
			logger.Debug("AllowMergeableBypassApply feature flag is enabled - attempting to bypass apply from mergeable requirements")
			isMergeableMinusApply, err := g.IsMergeableMinusApply(logger, repo, githubPR, vcsstatusname, ignoreVCSStatusNames)
			if err != nil {
				return false, errors.Wrap(err, "getting pull request status")
			}
			return isMergeableMinusApply, nil
		}
		return false, nil
	default:
		return false, nil
	}
}

// GetPullRequest returns the pull request.
func (g *GithubClient) GetPullRequest(logger logging.SimpleLogging, repo models.Repo, num int) (*github.PullRequest, error) {
	logger.Debug("Getting GitHub pull request %d", num)
	var err error
	var pull *github.PullRequest

	// GitHub has started to return 404's here (#1019) even after they send the webhook.
	// They've got some eventual consistency issues going on so we're just going
	// to attempt up to 5 times with exponential backoff.
	maxAttempts := 5
	attemptDelay := 0 * time.Second
	for i := 0; i < maxAttempts; i++ {
		// First don't sleep, then sleep 1, 3, 7, etc.
		time.Sleep(attemptDelay)
		attemptDelay = 2*attemptDelay + 1*time.Second

		pull, resp, err := g.client.PullRequests.Get(g.ctx, repo.Owner, repo.Name, num)
		if resp != nil {
			logger.Debug("GET /repos/%v/%v/pulls/%d returned: %v", repo.Owner, repo.Name, num, resp.StatusCode)
		}
		if err == nil {
			return pull, nil
		}
		ghErr, ok := err.(*github.ErrorResponse)
		if !ok || ghErr.Response.StatusCode != 404 {
			return pull, err
		}
	}
	return pull, err
}

// UpdateStatus updates the status badge on the pull request.
// See https://github.com/blog/1227-commit-status-api.
func (g *GithubClient) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	ghState := "error"
	switch state {
	case models.PendingCommitStatus:
		ghState = "pending"
	case models.SuccessCommitStatus:
		ghState = "success"
	case models.FailedCommitStatus:
		ghState = "failure"
	}

	logger.Info("Updating GitHub Check status for '%s' to '%s'", src, ghState)

	status := &github.RepoStatus{
		State:       github.String(ghState),
		Description: github.String(description),
		Context:     github.String(src),
		TargetURL:   &url,
	}
	_, resp, err := g.client.Repositories.CreateStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, status)
	if resp != nil {
		logger.Debug("POST /repos/%v/%v/statuses/%s returned: %v", repo.Owner, repo.Name, pull.HeadCommit, resp.StatusCode)
	}
	return err
}

// MergePull merges the pull request.
func (g *GithubClient) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	logger.Debug("Merging GitHub pull request %d", pull.Num)
	// Users can set their repo to disallow certain types of merging.
	// We detect which types aren't allowed and use the type that is.
	repo, resp, err := g.client.Repositories.Get(g.ctx, pull.BaseRepo.Owner, pull.BaseRepo.Name)
	if resp != nil {
		logger.Debug("GET /repos/%v/%v returned: %v", pull.BaseRepo.Owner, pull.BaseRepo.Name, resp.StatusCode)
	}
	if err != nil {
		return errors.Wrap(err, "fetching repo info")
	}

	const (
		defaultMergeMethod = "merge"
		rebaseMergeMethod  = "rebase"
		squashMergeMethod  = "squash"
	)

	mergeMethodsAllow := map[string]func() bool{
		defaultMergeMethod: repo.GetAllowMergeCommit,
		rebaseMergeMethod:  repo.GetAllowRebaseMerge,
		squashMergeMethod:  repo.GetAllowSquashMerge,
	}

	mergeMethodsName := slices.Collect(maps.Keys(mergeMethodsAllow))
	sort.Strings(mergeMethodsName)

	var method string
	if pullOptions.MergeMethod != "" {
		method = pullOptions.MergeMethod

		isMethodAllowed, isMethodExist := mergeMethodsAllow[method]
		if !isMethodExist {
			return fmt.Errorf("Merge method '%s' is unknown. Specify one of the valid values: '%s'", method, strings.Join(mergeMethodsName, ", "))
		}

		if !isMethodAllowed() {
			return fmt.Errorf("Merge method '%s' is not allowed by the repository Pull Request settings", method)
		}
	} else {
		method = defaultMergeMethod
		if !repo.GetAllowMergeCommit() {
			if repo.GetAllowRebaseMerge() {
				method = rebaseMergeMethod
			} else if repo.GetAllowSquashMerge() {
				method = squashMergeMethod
			}
		}
	}

	// Now we're ready to make our API call to merge the pull request.
	options := &github.PullRequestOptions{
		MergeMethod: method,
	}
	logger.Debug("PUT /repos/%v/%v/pulls/%d/merge", repo.Owner, repo.Name, pull.Num)
	mergeResult, resp, err := g.client.PullRequests.Merge(
		g.ctx,
		pull.BaseRepo.Owner,
		pull.BaseRepo.Name,
		pull.Num,
		// NOTE: Using the empty string here causes GitHub to autogenerate
		// the commit message as it normally would.
		"",
		options)
	if resp != nil {
		logger.Debug("POST /repos/%v/%v/pulls/%d/merge returned: %v", repo.Owner, repo.Name, pull.Num, resp.StatusCode)
	}
	if err != nil {
		return errors.Wrap(err, "merging pull request")
	}
	if !mergeResult.GetMerged() {
		return fmt.Errorf("could not merge pull request: %s", mergeResult.GetMessage())
	}
	return nil
}

// MarkdownPullLink specifies the string used in a pull request comment to reference another pull request.
func (g *GithubClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return fmt.Sprintf("#%d", pull.Num), nil
}

// GetTeamNamesForUser returns the names of the teams or groups that the user belongs to (in the organization the repository belongs to).
// https://docs.github.com/en/graphql/reference/objects#organization
func (g *GithubClient) GetTeamNamesForUser(repo models.Repo, user models.User) ([]string, error) {
	orgName := repo.Owner
	variables := map[string]interface{}{
		"orgName":    githubv4.String(orgName),
		"userLogins": []githubv4.String{githubv4.String(user.Username)},
		"teamCursor": (*githubv4.String)(nil),
	}
	var q struct {
		Organization struct {
			Teams struct {
				Edges []struct {
					Node struct {
						Name string
						Slug string
					}
				}
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"teams(first:100, after: $teamCursor, userLogins: $userLogins)"`
		} `graphql:"organization(login: $orgName)"`
	}
	var teamNames []string
	ctx := context.Background()
	for {
		err := g.v4Client.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}
		for _, edge := range q.Organization.Teams.Edges {
			teamNames = append(teamNames, edge.Node.Name, edge.Node.Slug)
		}
		if !q.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["teamCursor"] = githubv4.NewString(q.Organization.Teams.PageInfo.EndCursor)
	}
	return teamNames, nil
}

// ExchangeCode returns a newly created app's info
func (g *GithubClient) ExchangeCode(logger logging.SimpleLogging, code string) (*GithubAppTemporarySecrets, error) {
	logger.Debug("Exchanging code for app secrets")
	ctx := context.Background()
	cfg, resp, err := g.client.Apps.CompleteAppManifest(ctx, code)
	if resp != nil {
		logger.Debug("POST /app-manifests/%s/conversions returned: %v", code, resp.StatusCode)
	}
	data := &GithubAppTemporarySecrets{
		ID:            cfg.GetID(),
		Key:           cfg.GetPEM(),
		WebhookSecret: cfg.GetWebhookSecret(),
		Name:          cfg.GetName(),
		URL:           cfg.GetHTMLURL(),
	}

	return data, err
}

// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
// The first return value indicates whether the repo contains a file or not
// if BaseRepo had a file, its content will placed on the second return value
func (g *GithubClient) GetFileContent(logger logging.SimpleLogging, pull models.PullRequest, fileName string) (bool, []byte, error) {
	logger.Debug("Getting file content for %s in GitHub pull request %d", fileName, pull.Num)
	opt := github.RepositoryContentGetOptions{Ref: pull.HeadBranch}
	fileContent, _, resp, err := g.client.Repositories.GetContents(g.ctx, pull.BaseRepo.Owner, pull.BaseRepo.Name, fileName, &opt)
	if resp != nil {
		logger.Debug("GET /repos/%v/%v/contents/%s returned: %v", pull.BaseRepo.Owner, pull.BaseRepo.Name, fileName, resp.StatusCode)
	}

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

func (g *GithubClient) SupportsSingleFileDownload(_ models.Repo) bool {
	return true
}

func (g *GithubClient) GetCloneURL(logger logging.SimpleLogging, _ models.VCSHostType, repo string) (string, error) {
	logger.Debug("Getting clone URL for %s", repo)
	parts := strings.Split(repo, "/")
	repository, resp, err := g.client.Repositories.Get(g.ctx, parts[0], parts[1])
	if resp != nil {
		logger.Debug("GET /repos/%v/%v returned: %v", parts[0], parts[1], resp.StatusCode)
	}
	if err != nil {
		return "", err
	}
	return repository.GetCloneURL(), nil
}

func (g *GithubClient) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	logger.Debug("Getting labels for GitHub pull request %d", pull.Num)
	pullDetails, resp, err := g.client.PullRequests.Get(g.ctx, repo.Owner, repo.Name, pull.Num)
	if resp != nil {
		logger.Debug("GET /repos/%v/%v/pulls/%d returned: %v", repo.Owner, repo.Name, pull.Num, resp.StatusCode)
	}
	if err != nil {
		return nil, err
	}

	var labels []string

	for _, label := range pullDetails.Labels {
		labels = append(labels, *label.Name)
	}

	return labels, nil
}
