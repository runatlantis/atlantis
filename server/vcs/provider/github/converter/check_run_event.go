package converter

import (
	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

type CheckRunEvent struct {
	RepoConverter RepoConverter
}

func (p CheckRunEvent) Convert(e *github.CheckRunEvent) (event.CheckRun, error) {
	var action event.CheckRunAction
	switch e.GetAction() {
	case event.RequestedActionType:
		action = event.RequestedActionChecksAction{
			Identifier: e.GetRequestedAction().Identifier,
		}
	default:
		action = event.WrappedCheckRunAction(e.GetAction())
	}

	installationToken := githubapp.GetInstallationIDFromEvent(e)

	repo, err := p.RepoConverter.Convert(e.GetRepo())
	if err != nil {
		return event.CheckRun{}, errors.Wrap(err, "converting repo")
	}

	user := models.User{
		Username: e.GetSender().GetLogin(),
	}

	return event.CheckRun{
		Name:              e.GetCheckRun().GetName(),
		Action:            action,
		ExternalID:        e.CheckRun.GetExternalID(),
		Repo:              repo,
		User:              user,
		Branch:            e.GetCheckRun().GetCheckSuite().GetHeadBranch(),
		HeadSha:           e.GetCheckRun().GetHeadSHA(),
		InstallationToken: installationToken,
	}, nil

}
