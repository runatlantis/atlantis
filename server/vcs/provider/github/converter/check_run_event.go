package converter

import (
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
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

	repo, err := p.RepoConverter.Convert(e.GetRepo())
	if err != nil {
		return event.CheckRun{}, errors.Wrap(err, "converting repo")
	}

	return event.CheckRun{
		Name:       e.GetCheckRun().GetName(),
		Action:     action,
		ExternalID: e.CheckRun.GetExternalID(),
		Repo:       repo,
		User:       e.GetSender().GetLogin(),
	}, nil

}
