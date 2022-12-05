package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"net/http"
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

	var approvalReviewers []string
	nextPage := 0
	for {
		listOptions := gh.ListOptions{
			PerPage: 100,
		}
		listOptions.Page = nextPage

		reviews, resp, err := client.PullRequests.ListReviews(ctx, repo.Owner, repo.Name, prNum, &listOptions)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching pull request reviews")
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("not ok status fetching pull request reviews: %s", resp.Status)
		}
		for _, review := range reviews {
			if review.GetState() == ApprovalState && review.GetUser() != nil {
				approvalReviewers = append(approvalReviewers, review.GetUser().GetLogin())
			}
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}
	return approvalReviewers, nil
}
