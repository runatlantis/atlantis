package github

import (
	"context"
	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

const ApprovalState = "APPROVED"

type PRReviewerFetcher struct {
	ClientCreator githubapp.ClientCreator
}

func (r *PRReviewerFetcher) ListApprovalReviewers(ctx context.Context, installationToken int64, repo models.Repo, prNum int) ([]string, error) {
	client, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return nil, errors.Wrap(err, "creating installation client")
	}

	run := func(ctx context.Context, nextPage int) ([]*gh.PullRequestReview, *gh.Response, error) {
		listOptions := gh.ListOptions{
			PerPage: 100,
		}
		listOptions.Page = nextPage
		return client.PullRequests.ListReviews(ctx, repo.Owner, repo.Name, prNum, &listOptions)
	}
	reviews, err := Iterate(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "iterating through entries")
	}
	return findLatestApprovals(reviews), nil
}

// only return an approval from a user if it is their most recent review
// this is because a user can approve a PR then request more changes later on
func findLatestApprovals(reviews []*gh.PullRequestReview) []string {
	var approvalReviewers []string
	reviewers := make(map[string]bool)

	//reviews are returned chronologically
	for i := len(reviews) - 1; i >= 0; i-- {
		review := reviews[i]
		reviewer := review.GetUser()
		if reviewer == nil {
			continue
		}
		// add reviewer if an approval + we have not already processed their most recent review
		if review.GetState() == ApprovalState && !reviewers[reviewer.GetLogin()] {
			approvalReviewers = append(approvalReviewers, reviewer.GetLogin())
		}
		reviewers[reviewer.GetLogin()] = true
	}
	return approvalReviewers
}
