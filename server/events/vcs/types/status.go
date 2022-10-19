package types

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

type UpdateStatusRequest struct {
	Repo        models.Repo
	Ref         string
	State       models.VCSStatus
	StatusName  string
	Description string
	DetailsURL  string
	Output      string
	// if not present, should be -1
	PullNum          int
	PullCreationTime time.Time
	StatusID         string

	// Fields used to support templating project level command for github checks
	CommandName string
	Project     string
	Workspace   string
	Directory   string

	// Fields used to support templating command level operations for github checks
	NumSuccess string
	NumTotal   string
}
