package github

// Add more fields as necessary
type PullRequest struct {
	Number int
}

type PullRequestState string

const OpenPullRequest PullRequestState = "open"
