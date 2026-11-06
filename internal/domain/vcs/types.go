package vcs

import "time"

// Repository represents a VCS repository
type Repository struct {
	FullName    string
	Owner       string
	Name        string
	HTMLURL     string
	CloneURL    string
	DefaultBranch string
}

// PullRequest represents a pull/merge request
type PullRequest struct {
	Number       int
	Title        string
	Author       string
	HeadSHA      string
	HeadBranch   string
	BaseBranch   string
	State        PullRequestState
	URL          string
	UpdatedAt    time.Time
	Mergeable    *bool
}

type PullRequestState string

const (
	PullRequestOpen   PullRequestState = "open"
	PullRequestClosed PullRequestState = "closed"
	PullRequestMerged PullRequestState = "merged"
)

// CommitStatus represents a commit status check
type CommitStatus struct {
	State       CommitState
	Description string
	Context     string
	TargetURL   string
}

type CommitState string

const (
	CommitPending CommitState = "pending"
	CommitSuccess CommitState = "success"
	CommitFailure CommitState = "failure"
	CommitError   CommitState = "error"
) 