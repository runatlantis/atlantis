package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Vote identifiers
const (
	VoteApproved                = 10
	VoteApprovedWithSuggestions = 5
	VoteNone                    = 0
	VoteWaitingForAuthor        = -5
	VoteRejected                = -10
)

// CommentType enum declaration
type CommentType int

// CommentType enum declaration
const (
	// The comment type is not known.
	CommentTypeUnknown CommentType = iota
	// This is a regular user comment.
	CommentTypeText
	// The comment comes as a result of a code change.
	CommentTypeCodeChange
	// The comment represents a system message.
	CommentTypeSystem
)

func (d CommentType) String() string {
	return [...]string{"unknown", "text", "codechange", "system"}[d]
}

// CommentThreadStatus enum declaration
type CommentThreadStatus int

// CommentThreadStatus enum declaration
const (
	StatusUnknown CommentThreadStatus = iota
	StatusActive
	Fixed
	WontFix
	Closed
	ByDesign
	Pending
)

func (d CommentThreadStatus) String() string {
	return [...]string{"unknown", "active", "fixed", "wontfix", "closed", "byDesign", "pending"}[d]
}

// PullRequestAsyncStatus The current status of a pull request merge.
type PullRequestAsyncStatus int

// PullRequestAsyncStatus enum values
const (
	MergeConflicts PullRequestAsyncStatus = iota
	MergeFailure
	MergeNotSet
	MergeQueued
	MergeRejectedByPolicy
	MergeSucceeded
)

func (d PullRequestAsyncStatus) String() string {
	return [...]string{"conflicts", "failure", "notSet", "queued", "rejectedByPolicy", "succeeded"}[d]
}

// PullRequestMergeFailureType The specific type of merge request failure
type PullRequestMergeFailureType int

// PullRequestMergeFailureType enum values
const (
	NoFailure PullRequestMergeFailureType = iota
	UnknownFailure
	CaseSensitive
	ObjectTooLarge
)

func (d PullRequestMergeFailureType) String() string {
	return [...]string{"none", "unknown", "caseSensitive", "objectTooLarge"}[d]
}

// PullRequestStatus The current status of a pull request merge.
type PullRequestStatus int

// PullRequestStatus enum values
const (
	PullAbandoned PullRequestStatus = iota
	PullActive
	PullIncludeAll
	PullCompleted
	PullNotSet
)

func (d PullRequestStatus) String() string {
	return [...]string{"abandoned", "active", "all", "completed", "notSet"}[d]
}

// PullRequestsService handles communication with the pull requests methods on the API
// utilising https://docs.microsoft.com/en-us/rest/api/vsts/git/pull%20requests
type PullRequestsService struct {
	client *Client
}

// PullRequestsCommitsResponse describes a pull requests commits response
type PullRequestsCommitsResponse struct {
	Count         int             `json:"count"`
	GitCommitRefs []*GitCommitRef `json:"value"`
}

// PullRequestGetOptions describes what the request to the API should look like
type PullRequestGetOptions struct {
	IncludeCommits      bool   `url:"includeCommits,omitempty"`
	IncludeWorkItemRefs bool   `url:"includeWorkItemRefs,omitempty"`
	Project             string `url:"project,omitempty"`
	Organization        string `url:"organization,omitempty"`
	RepositoryID        string `url:"repositoryId,omitempty"`
	// maxCommentLength Not used.
	MaxCommentLength int `url:"maxCommentLength,omitempty"`
	PullRequestID    int `url:"pullRequestId,omitempty"`
	// $skip Not used.
	Skip int `url:"$skip,omitempty"`
	// $top Not used.
	Top int `url:"$top,omitempty"`
}

// PullRequestsListResponse describes a pull requests list response
type PullRequestsListResponse struct {
	Count           int               `json:"count"`
	GitPullRequests []*GitPullRequest `json:"value"`
}

// PullRequestsListCommitsResponse describes a pull requests list commits response
type PullRequestsListCommitsResponse struct {
	Count         int             `json:"count"`
	GitCommitRefs []*GitCommitRef `json:"value"`
}

// PullRequestListOptions describes what the request to the API should look like
type PullRequestListOptions struct {
	CreatorID          string `url:"searchCriteria.creatorId,omitempty"`
	IncludeLinks       string `url:"searchCriteria.includeLinks,omitempty"`
	Project            string `url:"project,omitempty"`
	RepositoryID       string `url:"searchCriteria.repositoryId,omitempty"`
	ReviewerID         string `url:"searchCriteria.reviewerId,omitempty"`
	Skip               string `url:"$skip,omitempty"`
	SourceRefName      string `url:"searchCriteria.sourceRefName,omitempty"`
	SourceRepositoryID string `url:"searchCriteria.sourceRepositoryId,omitempty"`
	Status             string `url:"searchCriteria.status,omitempty"`
	TargetRefName      string `url:"searchCriteria.targetRefName,omitempty"`
	Top                string `url:"$top,omitempty"`
}

// List returns list of pull requests in the specified Team Project with optional
// filters
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/get%20pull%20requests%20by%20project
func (s *PullRequestsService) List(ctx context.Context, owner, project string, opts *PullRequestListOptions) ([]*GitPullRequest, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/pullrequests?api-version=5.1-preview.1",
		owner,
		project,
	)
	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}

	r := new(PullRequestsListResponse)
	resp, err := s.client.Execute(ctx, req, r)
	if err != nil {
		return nil, nil, err
	}

	return r.GitPullRequests, resp, err
}

// Get returns a single pull request
// utilising https://docs.microsoft.com/en-us/rest/api/vsts/git/pull%20requests/get%20pull%20requests%20by%20project
func (s *PullRequestsService) Get(ctx context.Context, owner, project string, pullNum int, opts *PullRequestListOptions) (*GitPullRequest, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/pullrequests/%d?api-version=5.1-preview.1",
		owner,
		project,
		pullNum,
	)
	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := new(GitPullRequest)
	resp, err := s.client.Execute(ctx, req, r)

	return r, resp, err
}

// GetWithRepo returns a single pull request with additional information
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/get%20pull%20request?view=azure-devops-rest-5.1
func (s *PullRequestsService) GetWithRepo(ctx context.Context, owner, project, repo string, pullNum int, opts *PullRequestGetOptions) (*GitPullRequest, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%d?api-version=5.1-preview.1",
		owner,
		project,
		repo,
		pullNum,
	)

	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := new(GitPullRequest)
	resp, err := s.client.Execute(ctx, req, r)

	return r, resp, err
}

// Merge Completes a pull request
// pull may be nil
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/update?view=azure-devops-rest-5.1
func (s *PullRequestsService) Merge(ctx context.Context, owner, project string, repoName string, pullNum int, pull *GitPullRequest, completionOpts GitPullRequestCompletionOptions, id IdentityRef) (*GitPullRequest, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%d?api-version=5.1-preview.1",
		owner,
		project,
		repoName,
		pullNum,
	)

	/* If pull not nil, prepare for merge
	// otherwise create an empty GitPullRequest{}
	if pull != nil {
	}
	*/

	// Construct request body from supplied parameters
	body := &GitPullRequest{}
	body.AutoCompleteSetBy = &id
	body.CompletionOptions = &completionOpts

	// Now we're ready to make our API call to merge the pull request.
	request, err := s.client.NewRequest("PATCH", URL, body)
	if err != nil {
		return nil, nil, err
	}
	r := new(GitPullRequest)
	resp, err := s.client.Execute(ctx, request, r)

	return r, resp, err
}

// Create Creates a pull request
// Required fields in the GitPullRequest{} are:
// * Title
// * Description
// * SourceRefName
// * TargetRefName
//
// SourceRefName can be either the full ref name "refs/heads/branchname" or
// just "branchname".  The latter will be converted before submission.
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/create?view=azure-devops-rest-5.1
func (s *PullRequestsService) Create(ctx context.Context, owner, project string, repoName string, pull *GitPullRequest) (*GitPullRequest, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests?api-version=5.1-preview.1",
		owner,
		project,
		repoName,
	)

	if pull.GetTitle() == "" || pull.GetDescription() == "" ||
		pull.GetSourceRefName() == "" || pull.GetTargetRefName() == "" {
		return nil, nil, errors.New("PullRequests.Create: Missing required field")
	}

	formatRef(pull.SourceRefName)
	formatRef(pull.TargetRefName)

	// Now we're ready to make our API call to create the pull request.
	request, err := s.client.NewRequest("POST", URL, pull)
	if err != nil {
		return nil, nil, err
	}
	r := new(GitPullRequest)
	resp, err := s.client.Execute(ctx, request, r)

	return r, resp, err
}

// Comment Represents a comment which is one of potentially many in a comment thread.
type Comment struct {
	Links                  *map[string]Link `json:"_links,omitempty"`
	Author                 *IdentityRef     `json:"author,omitempty"`
	CommentType            *string          `json:"commentType,omitempty"`
	Content                *string          `json:"content,omitempty"`
	ID                     *int             `json:"id,omitempty"`
	IsDeleted              *bool            `json:"isDeleted,omitempty"`
	LastContentUpdatedDate *time.Time       `json:"lastContentUpdatedDate,omitempty"`
	LastUpdatedDate        *time.Time       `json:"lastUpdatedDate,omitempty"`
	ParentCommentID        *int             `json:"parentCommentId,omitempty"`
	PublishedDate          *time.Time       `json:"publishedDate,omitempty"`
	UsersLiked             []*IdentityRef   `json:"usersLiked,omitempty"`
}

// CommentPosition describes a comment position
type CommentPosition struct {
	Line   *int `json:"line,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

// GitPullRequestCommentThread Represents a comment thread of a pull request.
// A thread contains meta data about the file it was left on along with one or
// more comments (an initial comment and the subsequent replies).
type GitPullRequestCommentThread struct {
	Links                    *map[string]Link                    `json:"_links,omitempty"`
	Comments                 []*Comment                          `json:"comments,omitempty"`
	ID                       *int                                `json:"id,omitempty"`
	Identities               []*IdentityRef                      `json:"identities,omitempty"`
	IsDeleted                *bool                               `json:"isDeleted,omitempty"`
	LastUpdatedDate          *time.Time                          `json:"lastUpdatedDate,omitempty"`
	Properties               []*int                              `json:"properties,omitempty"`
	PublishedDate            *time.Time                          `json:"publishedDate,omitempty"`
	Status                   *string                             `json:"status,omitempty"`
	PullRequestThreadContext *GitPullRequestCommentThreadContext `json:"pullRequestThreadContext,omitempty"`
}

// GitPullRequestCommentThreadContext Comment thread context contains details about what
// diffs were being viewed at the time of thread creation and whether or not the thread
// has been tracked from that original diff.
type GitPullRequestCommentThreadContext struct {
	FilePath       *string          `json:"filePath,omitempty"`
	LeftFileEnd    *CommentPosition `json:"leftFileEnd,omitempty"`
	LeftFileStart  *CommentPosition `json:"leftFileStart,omitempty"`
	RightFileEnd   *CommentPosition `json:"rightFileEnd,omitempty"`
	RightFileStart *CommentPosition `json:"rightFileStart,omitempty"`
}

// GitPullRequestWithComment contains a reference to an existing pull request and a
// comment.
type GitPullRequestWithComment struct {
	Comment     *Comment        `json:"comment,omitempty"`
	PullRequest *GitPullRequest `json:"pullRequest,omitempty"`
}

// ListCommits lists the commits in a pull request.
// Azure Devops API docs: https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20request%20commits/get%20pull%20request%20commits
//
func (s *PullRequestsService) ListCommits(ctx context.Context, owner, project, repo string, pullNum int) ([]*GitCommitRef, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%d/commits?api-version=5.1-preview.1",
		owner,
		project,
		repo,
		pullNum,
	)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}

	r := new(PullRequestsListCommitsResponse)
	resp, err := s.client.Execute(ctx, req, r)
	if err != nil {
		return nil, nil, err
	}

	return r.GitCommitRefs, resp, err
}

// CreateComment adds a comment to a pull request thread.
// Azure Devops API docs: https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20request%20thread%20comments/create
//
func (s *PullRequestsService) CreateComment(ctx context.Context, owner, project, repo string, pullNum int, threadId int, comment *Comment) (*Comment, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%d/threads/%d/comments?api-version=5.1-preview.1",
		owner,
		project,
		repo,
		pullNum,
		threadId,
	)

	if comment.GetContent() == "" {
		return nil, nil, errors.New("PullRequests.CreateComment: Nil pointer or empty string in comment.Content field ")
	}

	if comment.GetCommentType() == "" {
		comment.CommentType = String("text")
	}

	req, err := s.client.NewRequest("POST", URL, comment)
	if err != nil {
		return nil, nil, err
	}

	r := new(Comment)
	resp, err := s.client.Execute(ctx, req, r)
	if err != nil {
		return nil, nil, err
	}

	return r, resp, err
}

// CreateComments adds one or more comments to a new or existing thread
// and may include additional context
// Azure Devops API docs: https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20request%20threads/create
//
func (s *PullRequestsService) CreateComments(ctx context.Context, owner, project, repo string, pullNum int, body *GitPullRequestCommentThread) (*GitPullRequestCommentThread, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%d/threads?api-version=5.1-preview.1",
		owner,
		project,
		repo,
		pullNum,
	)

	if len(body.Comments) == 0 {
		return nil, nil, errors.New("PullRequests.CreateComments: Must supply at least one comment in Comments field")
	}

	for idx, comment := range body.Comments {
		if comment.GetContent() == "" {
			return nil, nil, errors.New("PullRequests.CreateComment: Nil pointer or empty string in comment.Content field ")
		}

		if comment.GetCommentType() == "" {
			body.Comments[idx].CommentType = String("text")
		}
	}

	req, err := s.client.NewRequest("POST", URL, body)
	if err != nil {
		return nil, nil, err
	}

	r := new(GitPullRequestCommentThread)
	resp, err := s.client.Execute(ctx, req, r)
	if err != nil {
		return nil, nil, err
	}

	return r, resp, err
}
