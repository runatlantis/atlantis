package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate mockgen -destination=mocks/mock_pull_approved_checker.go -package=mocks . PullApprovedChecker

type PullApprovedChecker interface {
	PullIsApproved(logger logging.SimpleLogging, baseRepo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error)
}
