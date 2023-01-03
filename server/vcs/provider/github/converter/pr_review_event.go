package converter

import (
	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

type PullRequestReviewEvent struct {
	RepoConverter RepoConverter
}

func (p PullRequestReviewEvent) Convert(e *github.PullRequestReviewEvent) (event.PullRequestReview, error) {
	installationToken := githubapp.GetInstallationIDFromEvent(e)
	repo, err := p.RepoConverter.Convert(e.GetRepo())
	if err != nil {
		return event.PullRequestReview{}, errors.Wrap(err, "converting repo")
	}
	user := models.User{
		Username: e.GetSender().GetLogin(),
	}
	prNum := e.GetPullRequest().GetNumber()
	action := e.GetAction()
	ref := e.GetReview().GetCommitID()
	return event.PullRequestReview{
		InstallationToken: installationToken,
		Repo:              repo,
		PRNum:             prNum,
		User:              user,
		Action:            action,
		Ref:               ref,
	}, nil

}
