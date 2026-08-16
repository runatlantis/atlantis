// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PullStatusFreshForPull returns true when the recorded pull status
// matches the head commit and base branch of the given pull.
func PullStatusFreshForPull(pull PullRequest, statusPull PullRequest) bool {
	return PullStatusFreshnessError(pull, statusPull, "recorded plan status") == nil
}

// PullStatusFreshnessError explains why the recorded pull status is stale
// for the given pull, or returns nil when it is fresh. Missing recorded
// fields are tolerated.
func PullStatusFreshnessError(pull PullRequest, statusPull PullRequest, subject string) error {
	return pullStatusFreshnessErrorWithMissingFields(pull, statusPull, subject, false)
}

// PullStatusApplyEligibilityError is PullStatusFreshnessError with missing
// recorded fields rejected, for gating applies.
func PullStatusApplyEligibilityError(pull PullRequest, statusPull PullRequest, subject string) error {
	return pullStatusFreshnessErrorWithMissingFields(pull, statusPull, subject, true)
}

func pullStatusFreshnessErrorWithMissingFields(pull PullRequest, statusPull PullRequest, subject string, rejectMissing bool) error {
	verb := "is"
	if subject == "plans" {
		verb = "are"
	}
	if rejectMissing && pull.HeadCommit != "" && statusPull.HeadCommit == "" {
		return fmt.Errorf(
			"%s %s missing a recorded head commit while current head is %s; run `atlantis plan` before apply",
			subject,
			verb,
			ShortSHA(pull.HeadCommit),
		)
	}
	if pull.HeadCommit != "" && statusPull.HeadCommit != "" && statusPull.HeadCommit != pull.HeadCommit {
		return fmt.Errorf(
			"%s %s from commit %s but current head is %s; run `atlantis plan` before apply",
			subject,
			verb,
			ShortSHA(statusPull.HeadCommit),
			ShortSHA(pull.HeadCommit),
		)
	}
	currentBase := strings.TrimSpace(pull.BaseBranch)
	statusBase := strings.TrimSpace(statusPull.BaseBranch)
	if rejectMissing && currentBase != "" && statusBase == "" {
		return fmt.Errorf(
			"%s %s missing a recorded base branch while current base branch is %q; run `atlantis plan` before apply",
			subject,
			verb,
			currentBase,
		)
	}
	if currentBase != "" && statusBase != "" && statusBase != currentBase {
		return fmt.Errorf(
			"%s %s for base branch %q but current base branch is %q; run `atlantis plan` before apply",
			subject,
			verb,
			statusBase,
			currentBase,
		)
	}
	return nil
}

// ShortSHA abbreviates a commit SHA for messages.
func ShortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// FindProject returns the project status matching workspace, dir, and
// (when set) project name, or nil.
func (p *PullStatus) FindProject(workspace, repoRelDir, projectName string) *ProjectStatus {
	cleanDir := filepath.Clean(repoRelDir)
	for i := range p.Projects {
		proj := &p.Projects[i]
		if proj.Workspace != workspace || filepath.Clean(proj.RepoRelDir) != cleanDir {
			continue
		}
		if projectName != "" {
			if proj.ProjectName == projectName {
				return proj
			}
			continue
		}
		if proj.ProjectName == "" {
			return proj
		}
	}
	return nil
}

// StatusAllowsDiscoveredPlan returns true if a discovered .tfplan file
// with this status is valid to build an apply command for. Includes
// PlannedNoChangesPlanStatus (no-op plan can leave a .tfplan on disk) and
// ErroredPolicyCheckStatus (let apply build the command and fail through
// existing per-project apply-requirement handling).
func StatusAllowsDiscoveredPlan(status ProjectPlanStatus) bool {
	switch status {
	case PlannedPlanStatus, PassedPolicyCheckStatus, ErroredApplyStatus,
		PlannedNoChangesPlanStatus, ErroredPolicyCheckStatus:
		return true
	default:
		return false
	}
}
