package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

type PullApprovedChecker interface {
	PullIsApproved(baseRepo models.Repo, pull models.PullRequest) (bool, error)
}
