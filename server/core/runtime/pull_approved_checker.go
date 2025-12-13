// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate --package mocks -o mocks/mock_pull_approved_checker.go PullApprovedChecker

type PullApprovedChecker interface {
	PullIsApproved(logger logging.SimpleLogging, baseRepo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error)
}
