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
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// maxCommentLength is the maximum number of chars allowed in a single comment
// by GitHub.
const maxCommentLength = 65536

// GithubClient is used to perform GitHub actions.
type GithubClient struct {
	user           string
	client         *github.Client
	v4MutateClient *graphql.Client
	v4QueryClient  *githubv4.Client
	ctx            context.Context
	logger         logging.SimpleLogging
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
func NewGithubClient(hostname string, credentials GithubCredentials, logger logging.SimpleLogging) (*GithubClient, error) {
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
	token, err := credentials.GetToken()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get GitHub token")
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	// Use the client from shurcooL's githubv4 library for queries.
	v4QueryClient := githubv4.NewEnterpriseClient(graphqlURL, httpClient)

	user, err := credentials.GetUser()
	logger.Debug("GH User: %s", user)

	if err != nil {
		return nil, errors.Wrap(err, "getting user")
	}
	return &GithubClient{
		user:           user,
		client:         client,
		v4MutateClient: v4MutateClient,
		v4QueryClient:  v4QueryClient,
		ctx:            context.Background(),
		logger:         logger,
	}, nil
}

// GetModifiedFiles returns the names of files that were modified in the pull request
// relative to the repo root, e.g. parent/child/file.txt.
func (g *GithubClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
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

			g.logger.Debug("[attempt %d] GET /repos/%v/%v/pulls/%d/files", i+1, repo.Owner, repo.Name, pull.Num)
			pageFiles, resp, err := g.client.PullRequests.ListFiles(g.ctx, repo.Owner, repo.Name, pull.Num, &opts)
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
		g.logger.Debug("POST /repos/%v/%v/issues/%d/comments", repo.Owner, repo.Name, pullNum)
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
		g.logger.Debug("GET /repos/%v/%v/issues/%d/comments", repo.Owner, repo.Name, pullNum)
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
		g.logger.Debug("GET /repos/%v/%v/pulls/%d/reviews", repo.Owner, repo.Name, pull.Num)
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
	state := githubPR.GetMergeableState()
	status, _ := g.GetCombinedStatus(repo, githubPR.GetHead().GetSHA())

	// We map our mergeable check to when the GitHub merge button is clickable.
	// This corresponds to the following states:
	// clean: No conflicts, all requirements satisfied.
	//        Merging is allowed (green box).
	// unstable: Failing/pending commit status that is not part of the required
	//           status checks. Merging is allowed (yellow box).
	// has_hooks: GitHub Enterprise only, if a repo has custom pre-receive
	//            hooks. Merging is allowed (green box).
	// See: https://github.com/octokit/octokit.net/issues/1763
	//
	// We should not dismiss the PR if the only our "atlantis/apply" status is pending/failing
	if state == "blocked" {
		applyStatus := false
		for _, s := range status.Statuses {
			if strings.Contains(s.GetContext(), "atlantis/apply") {
				applyStatus = true
				continue
			}
			if s.GetContext() != "atlantis/apply" && s.GetState() != "success" {
				// If any other status is pending/failing mark as non-mergeable
				return false, nil
			}
		}
		return applyStatus, nil
	}

	if state != "clean" && state != "unstable" && state != "has_hooks" {
		return false, nil
	}

	return true, nil
}

// GetPullRequest returns the pull request.
func (g *GithubClient) GetPullRequest(repo models.Repo, num int) (*github.PullRequest, error) {
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

		pull, _, err = g.client.PullRequests.Get(g.ctx, repo.Owner, repo.Name, num)
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

func (g *GithubClient) GetCombinedStatus(repo models.Repo, ref string) (*github.CombinedStatus, error) {
	opts := github.ListOptions{
		PerPage: 300,
	}
	status, _, err := g.client.Repositories.GetCombinedStatus(g.ctx, repo.Owner, repo.Name, ref, &opts)
	return status, err
}

// UpdateStatus updates the status badge on the pull request.
// See https://github.com/blog/1227-commit-status-api.
func (g *GithubClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	ghState := "error"
	switch state {
	case models.PendingCommitStatus:
		ghState = "pending"
	case models.SuccessCommitStatus:
		ghState = "success"
	case models.FailedCommitStatus:
		ghState = "failure"
	}

	status := &github.RepoStatus{
		State:       github.String(ghState),
		Description: github.String(description),
		Context:     github.String(src),
		TargetURL:   &url,
	}
	_, _, err := g.client.Repositories.CreateStatus(g.ctx, repo.Owner, repo.Name, pull.HeadCommit, status)
	return err
}

// MergePull merges the pull request.
func (g *GithubClient) MergePull(pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	// Users can set their repo to disallow certain types of merging.
	// We detect which types aren't allowed and use the type that is.
	g.logger.Debug("GET /repos/%v/%v", pull.BaseRepo.Owner, pull.BaseRepo.Name)
	repo, _, err := g.client.Repositories.Get(g.ctx, pull.BaseRepo.Owner, pull.BaseRepo.Name)
	if err != nil {
		return errors.Wrap(err, "fetching repo info")
	}
	const (
		defaultMergeMethod = "merge"
		rebaseMergeMethod  = "rebase"
		squashMergeMethod  = "squash"
	)
	method := defaultMergeMethod
	if !repo.GetAllowMergeCommit() {
		if repo.GetAllowRebaseMerge() {
			method = rebaseMergeMethod
		} else if repo.GetAllowSquashMerge() {
			method = squashMergeMethod
		}
	}

	// Now we're ready to make our API call to merge the pull request.
	options := &github.PullRequestOptions{
		MergeMethod: method,
	}
	g.logger.Debug("PUT /repos/%v/%v/pulls/%d/merge", repo.Owner, repo.Name, pull.Num)
	mergeResult, _, err := g.client.PullRequests.Merge(
		g.ctx,
		pull.BaseRepo.Owner,
		pull.BaseRepo.Name,
		pull.Num,
		// NOTE: Using the empty string here causes GitHub to autogenerate
		// the commit message as it normally would.
		"",
		options)
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
		err := g.v4QueryClient.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}
		for _, edge := range q.Organization.Teams.Edges {
			teamNames = append(teamNames, edge.Node.Name)
		}
		if !q.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["teamCursor"] = githubv4.NewString(q.Organization.Teams.PageInfo.EndCursor)
	}
	return teamNames, nil
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

func (g *GithubClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return true
}
