package types

import "github.com/runatlantis/atlantis/server/events/models"

type UpdateStatusRequest struct {
	Repo models.Repo
	// if not present, should be -1
	PullNum     int
	Ref         string
	State       models.CommitStatus
	StatusName  string
	Description string
	DetailsURL  string
	Output      string
}
