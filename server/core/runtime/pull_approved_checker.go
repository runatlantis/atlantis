package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type PullApprovedChecker interface {
	PullIsApproved(logger logging.SimpleLogging, baseRepo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error)
}
