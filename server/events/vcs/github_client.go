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

	"github.com/google/go-github/v51/github"
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

// GithubClient is used to perform GitHub actions.
type GithubClient struct {
	user     string
	client   *github.Client
	v4Client *githubv4.Client
	ctx      context.Context
	logger   logging.SimpleLogging
	config   GithubConfig
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
func NewGithubClient(hostname string, credentials GithubCredentials, config GithubConfig, logger logging.SimpleLogging) (*GithubClient, error) {
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

	// Use the client from shurcooL's githubv4 library for queries.
	v4Client := githubv4.NewEnterpriseClient(graphqlURL, transport)

	user, err := credentials.GetUser()
	logger.Debug("GH User: %s", user)

	if err != nil {
		return nil, errors.Wrap(err, "getting user")
	}
	return &GithubClient{
		user:     user,
		client:   client,
		v4Client: v4Client,
		ctx:      context.Background(),
		logger:   logger,
		config:   config,
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
		input := githubv4.MinimizeCommentInput{
			Classifier: githubv4.ReportedContentClassifiersOutdated,
			SubjectID:  comment.GetNodeID(),
		}
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

// isRequiredCheck is a helper function to determine if a check is required or not
func isRequiredCheck(check string, required []string) bool {
	//in go1.18 can prob replace this with slices.Contains
	for _, r := range required {
		if r == check {
			return true
		}
	}

	return false
}

// GetCombinedStatusMinusApply checks Statuses for PR, excluding atlantis apply. Returns true if all other statuses are not in failure.
func (g *GithubClient) GetCombinedStatusMinusApply(repo models.Repo, pull *github.PullRequest, vcstatusname string) (bool, error) {
	//check combined status api
	status, _, err := g.client.Repositories.GetCombinedStatus(g.ctx, repo.Owner, repo.Name, *pull.Head.Ref, nil)
	if err != nil {
		return false, errors.Wrap(err, "getting combined status")
	}

	//iterate over statuses - return false if we find one that isnt "apply" and doesnt have state = "success"
	for _, r := range status.Statuses {
		if strings.HasPrefix(*r.Context, fmt.Sprintf("%s/%s", vcstatusname, command.Apply.String())) {
			continue
		}
		if *r.State != "success" {
			return false, nil
		}
	}

	//get required status checks
	required, _, err := g.client.Repositories.GetBranchProtection(context.Background(), repo.Owner, repo.Name, *pull.Base.Ref)
	if err != nil {
		return false, errors.Wrap(err, "getting required status checks")
	}

	//check check suite/check run api
	checksuites, _, err := g.client.Checks.ListCheckSuitesForRef(context.Background(), repo.Owner, repo.Name, *pull.Head.Ref, nil)
	if err != nil {
		return false, errors.Wrap(err, "getting check suites for ref")
	}

	//iterate over check completed check suites - return false if we find one that doesnt have conclusion = "success"
	for _, c := range checksuites.CheckSuites {
		if *c.Status == "completed" {
			//iterate over the runs inside the suite
			suite, _, err := g.client.Checks.ListCheckRunsCheckSuite(context.Background(), repo.Owner, repo.Name, *c.ID, nil)
			if err != nil {
				return false, errors.Wrap(err, "getting check runs for check suite")
			}

			for _, r := range suite.CheckRuns {
				//check to see if the check is required
				if isRequiredCheck(*r.Name, required.RequiredStatusChecks.Contexts) {
					if *c.Conclusion == "success" {
						continue
					} else {
						return false, nil
					}
				} else {
					//ignore checks that arent required
					continue
				}

			}
		}
	}

	return true, nil
}

// GetPullReviewDecision gets the pull review decision, which takes into account CODEOWNERS
func (g *GithubClient) GetPullReviewDecision(repo models.Repo, pull models.PullRequest) (approvalStatus bool, err error) {
	var query struct {
		Repository struct {
			PullRequest struct {
				ReviewDecision string
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":  githubv4.String(repo.Owner),
		"name":   githubv4.String(repo.Name),
		"number": githubv4.Int(pull.Num),
	}

	err = g.v4Client.Query(g.ctx, &query, variables)
	if err != nil {
		return approvalStatus, errors.Wrap(err, "getting reviewDecision")
	}

	if query.Repository.PullRequest.ReviewDecision == "APPROVED" || len(query.Repository.PullRequest.ReviewDecision) == 0 {
		return true, nil
	}

	return false, nil
}

// PullIsMergeable returns true if the pull request is mergeable.
func (g *GithubClient) PullIsMergeable(repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error) {
	githubPR, err := g.GetPullRequest(repo, pull.Num)
	if err != nil {
		return false, errors.Wrap(err, "getting pull request")
	}

	state := githubPR.GetMergeableState()
	// We map our mergeable check to when the GitHub merge button is clickable.
	// This corresponds to the following states:
	// clean: No conflicts, all requirements satisfied.
	//        Merging is allowed (green box).
	// unstable: Failing/pending commit status that is not part of the required
	//           status checks. Merging is allowed (yellow box).
	// has_hooks: GitHub Enterprise only, if a repo has custom pre-receive
	//            hooks. Merging is allowed (green box).
	// See: https://github.com/octokit/octokit.net/issues/1763
	if state != "clean" && state != "unstable" && state != "has_hooks" {
		//mergeable bypass apply code hidden by feature flag
		if g.config.AllowMergeableBypassApply {
			g.logger.Debug("AllowMergeableBypassApply feature flag is enabled - attempting to bypass apply from mergeable requirements")
			if state == "blocked" {
				//check status excluding atlantis apply
				status, err := g.GetCombinedStatusMinusApply(repo, githubPR, vcsstatusname)
				if err != nil {
					return false, errors.Wrap(err, "getting pull request status")
				}

				//check to see if pr is approved using reviewDecision
				approved, err := g.GetPullReviewDecision(repo, pull)
				if err != nil {
					return false, errors.Wrap(err, "getting pull request reviewDecision")
				}

				//if all other status checks EXCEPT atlantis/apply are successful, and the PR is approved based on reviewDecision, let it proceed
				if status && approved {
					return true, nil
				}
			}
		}

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

// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
// The first return value indicates whether the repo contains a file or not
// if BaseRepo had a file, its content will placed on the second return value
func (g *GithubClient) GetFileContent(pull models.PullRequest, fileName string) (bool, []byte, error) {
	opt := github.RepositoryContentGetOptions{Ref: pull.HeadBranch}
	fileContent, _, resp, err := g.client.Repositories.GetContents(g.ctx, pull.BaseRepo.Owner, pull.BaseRepo.Name, fileName, &opt)

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

func (g *GithubClient) GetCloneURL(VCSHostType models.VCSHostType, repo string) (string, error) {
	parts := strings.Split(repo, "/")
	repository, _, err := g.client.Repositories.Get(g.ctx, parts[0], parts[1])
	if err != nil {
		return "", err
	}
	return repository.GetCloneURL(), nil
}
