package event

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// Comment is our internal representation of a vcs based comment event.
type Comment struct {
	Pull      models.PullRequest
	BaseRepo  models.Repo
	HeadRepo  models.Repo
	User      models.User
	PullNum   int
	Comment   string
	VCSHost   models.VCSHostType
	Timestamp time.Time
}

// PullRequestEvent is our internal representation of a vcs based pr event
type PullRequest struct {
	Pull      models.PullRequest
	User      models.User
	EventType models.PullRequestEventType
	Timestamp time.Time
}
