package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate --package mocks -o mocks/mock_pull_approved_checker.go PullApprovedChecker

type PullApprovedChecker interface {
	PullIsApproved(baseRepo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error)
}
