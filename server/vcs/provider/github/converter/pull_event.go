package converter

import (
	"fmt"
	"github.com/palantir/go-githubapp/githubapp"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

type PullEventConverter struct {
	PullConverter PullConverter
	AllowDraftPRs bool
}

// Converts a github pull request event to our internal representation
func (e PullEventConverter) Convert(pullEvent *github.PullRequestEvent) (event.PullRequest, error) {
	if pullEvent.PullRequest == nil {
		return event.PullRequest{}, fmt.Errorf("pull_request is null")
	}
	pull, err := e.PullConverter.Convert(pullEvent.PullRequest)
	if err != nil {
		return event.PullRequest{}, err
	}
	if pullEvent.Sender == nil {
		return event.PullRequest{}, fmt.Errorf("sender is null")
	}
	senderUsername := pullEvent.Sender.GetLogin()
	if senderUsername == "" {
		return event.PullRequest{}, fmt.Errorf("sender.login is null")
	}

	action := pullEvent.GetAction()
	// If it's a draft PR we ignore it for auto-planning if configured to do so
	// however it's still possible for users to run plan on it manually via a
	// comment so if any draft PR is closed we still need to check if we need
	// to delete its locks.
	if pullEvent.GetPullRequest().GetDraft() && pullEvent.GetAction() != "closed" && !e.AllowDraftPRs {
		action = "other"
	}

	var pullEventType models.PullRequestEventType

	switch action {
	case "opened":
		pullEventType = models.OpenedPullEvent
	case "ready_for_review":
		// when an author takes a PR out of 'draft' state a 'ready_for_review'
		// event is triggered. We want atlantis to treat this as a freshly opened PR
		pullEventType = models.OpenedPullEvent
	case "synchronize":
		pullEventType = models.UpdatedPullEvent
	case "closed":
		pullEventType = models.ClosedPullEvent
	default:
		pullEventType = models.OtherPullEvent
	}

	eventTimestamp := time.Now()

	if pullEvent.PullRequest.UpdatedAt != nil {
		eventTimestamp = *pullEvent.PullRequest.UpdatedAt
	}

	installationToken := githubapp.GetInstallationIDFromEvent(pullEvent)

	return event.PullRequest{
		Pull:              pull,
		EventType:         pullEventType,
		User:              models.User{Username: senderUsername},
		Timestamp:         eventTimestamp,
		InstallationToken: installationToken,
	}, nil
}
