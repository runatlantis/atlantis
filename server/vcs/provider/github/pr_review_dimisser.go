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

const DismissReason = "**New plans have triggered policy failures that must be approved by policy owners.**"

type PRReviewDismisser struct {
	ClientCreator githubapp.ClientCreator
}

func (r *PRReviewDismisser) Dismiss(ctx context.Context, installationToken int64, repo models.Repo, prNum int, reviewID int64) error {
	client, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return errors.Wrap(err, "creating installation client")
	}
	dismissRequest := &gh.PullRequestReviewDismissalRequest{
		Message: gh.String(DismissReason),
	}

	_, resp, err := client.PullRequests.DismissReview(ctx, repo.Owner, repo.Name, prNum, reviewID, dismissRequest)
	if err != nil {
		return errors.Wrap(err, "error running gh api call")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not ok status running gh api call: %s", resp.Status)
	}
	return nil
}
