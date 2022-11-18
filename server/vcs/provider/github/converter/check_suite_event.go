package converter

import (
	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

type CheckSuiteEvent struct {
	RepoConverter RepoConverter
}

func (s CheckSuiteEvent) Convert(e *github.CheckSuiteEvent) (event.CheckSuite, error) {
	action := event.WrappedCheckRunAction(e.GetAction())
	installationToken := githubapp.GetInstallationIDFromEvent(e)
	repo, err := s.RepoConverter.Convert(e.GetRepo())
	if err != nil {
		return event.CheckSuite{}, errors.Wrap(err, "converting repo")
	}
	user := models.User{
		Username: e.GetSender().GetLogin(),
	}
	checkSuite := e.GetCheckSuite()
	return event.CheckSuite{
		Action:            action,
		HeadSha:           checkSuite.GetHeadSHA(),
		Repo:              repo,
		Sender:            user,
		InstallationToken: installationToken,
		Branch:            checkSuite.GetHeadBranch(),
	}, nil

}
