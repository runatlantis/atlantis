package event

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// Comment is our internal representation of a vcs based comment event.
type Comment struct {
	//TODO: enforce the existence of this so that we can eliminate
	// a number of other fields in this struct
	// This requires calling for the pull request info when this object is created
	// in the converter itself.
	MaybePull *models.PullRequest

	BaseRepo      models.Repo
	MaybeHeadRepo *models.Repo
	User          models.User
	PullNum       int
	Comment       string
	VCSHost       models.VCSHostType
	Timestamp     time.Time
}

// PullRequestEvent is our internal representation of a vcs based pr event
type PullRequest struct {
	Pull      models.PullRequest
	User      models.User
	EventType models.PullRequestEventType
	Timestamp time.Time
}
