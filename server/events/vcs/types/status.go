package types

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

type UpdateStatusRequest struct {
	Repo        models.Repo
	Ref         string
	State       models.CommitStatus
	StatusName  string
	Description string
	DetailsURL  string
	Output      string
	// if not present, should be -1
	PullNum          int
	PullCreationTime time.Time
}
