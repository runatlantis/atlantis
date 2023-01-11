package converter

import (
	"fmt"
	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"time"
)

type PullRequestReviewEvent struct {
	RepoConverter RepoConverter
	PullConverter PullConverter
}

func (p PullRequestReviewEvent) Convert(e *github.PullRequestReviewEvent) (event.PullRequestReview, error) {
	installationToken := githubapp.GetInstallationIDFromEvent(e)
	if e.GetRepo() == nil {
		return event.PullRequestReview{}, fmt.Errorf("repo is nil")
	}
	repo, err := p.RepoConverter.Convert(e.GetRepo())
	if err != nil {
		return event.PullRequestReview{}, errors.Wrap(err, "converting repo")
	}

	if e.GetPullRequest() == nil {
		return event.PullRequestReview{}, fmt.Errorf("pull request is nil")
	}
	pull, err := p.PullConverter.Convert(e.GetPullRequest())
	if err != nil {
		return event.PullRequestReview{}, errors.Wrap(err, "converting pull request")
	}

	if e.GetSender() == nil {
		return event.PullRequestReview{}, errors.Wrap(err, "sender is nil")
	}
	user := models.User{
		Username: e.GetSender().GetLogin(),
	}

	if e.GetReview() == nil {
		return event.PullRequestReview{}, errors.Wrap(err, "review is nil")
	}
	state := e.GetReview().GetState()
	ref := e.GetReview().GetCommitID()
	eventTimestamp := e.GetReview().GetSubmittedAt()
	// if submitted_at is nil, API returns time.Time{}, so we use time.Now as a backup
	if e.GetReview().GetSubmittedAt().Equal(time.Time{}) {
		eventTimestamp = time.Now()
	}

	return event.PullRequestReview{
		InstallationToken: installationToken,
		Repo:              repo,
		User:              user,
		State:             state,
		Ref:               ref,
		Timestamp:         eventTimestamp,
		Pull:              pull,
	}, nil
}
