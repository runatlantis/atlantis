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

	// TODO: look at latest reviews only if user reviewed multiple times
	var approvalReviewers []string
	for _, review := range reviews {
		if review.GetState() == ApprovalState && review.GetUser() != nil {
			approvalReviewers = append(approvalReviewers, review.GetUser().GetLogin())
		}
	}
	return approvalReviewers, nil
}
