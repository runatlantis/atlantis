// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type ApplyPlanValidator interface {
	ValidateProjectPlan(ctx command.ProjectContext, absPath string) error
}

type DefaultApplyPlanValidator struct {
	PullStatusFetcher PullStatusFetcher
}

func (v *DefaultApplyPlanValidator) ValidateProjectPlan(ctx command.ProjectContext, absPath string) error {
	if v == nil || v.PullStatusFetcher == nil {
		return nil
	}

	pullStatus, err := v.PullStatusFetcher.GetPullStatus(ctx.Pull)
	if err != nil {
		return fmt.Errorf("fetching current plan status: %w", err)
	}
	if pullStatus == nil {
		return fmt.Errorf("no current plan status found; run `atlantis plan` before apply")
	}
	if !pullStatusHeadMatchesPull(ctx.Pull, pullStatus.Pull) {
		return fmt.Errorf(
			"recorded plan status is from commit %s but current head is %s; run `atlantis plan` before apply",
			shortSHA(pullStatus.Pull.HeadCommit),
			shortSHA(ctx.Pull.HeadCommit),
		)
	}

	proj := findProjectInPullStatus(pullStatus, ctx.Workspace, ctx.RepoRelDir, ctx.ProjectName)
	if proj == nil {
		return fmt.Errorf(
			"no matching plan status exists for dir %q workspace %q project %q; run `atlantis plan`",
			ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName,
		)
	}
	if !statusAllowedForDiscoveredPlan(proj.Status) {
		return fmt.Errorf(
			"plan for dir %q workspace %q project %q has status %q and cannot be applied; run `atlantis plan`",
			ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, proj.Status.String(),
		)
	}

	planPath := filepath.Join(absPath, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	if _, err := os.Stat(planPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"plan file is missing for dir %q workspace %q project %q; run `atlantis plan`",
				ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName,
			)
		}
		return fmt.Errorf("checking plan file for dir %q workspace %q project %q: %w", ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, err)
	}

	return nil
}

func pullStatusHeadMatchesPull(pull models.PullRequest, statusPull models.PullRequest) bool {
	if pull.HeadCommit == "" || statusPull.HeadCommit == "" {
		return true
	}
	return statusPull.HeadCommit == pull.HeadCommit
}
