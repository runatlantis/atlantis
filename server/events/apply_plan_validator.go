// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/utils"
)

type ApplyPlanValidator interface {
	ValidateProjectPlan(ctx command.ProjectContext, absPath string) error
}

type LivePullHeadFetcher interface {
	GetLiveHeadCommit(ctx command.ProjectContext) (string, error)
}

type DefaultApplyPlanValidator struct {
	PullStatusFetcher   PullStatusFetcher
	LivePullHeadFetcher LivePullHeadFetcher
}

var errStaleCommandHead = errors.New("stale command head")

func (v *DefaultApplyPlanValidator) ValidateProjectPlan(ctx command.ProjectContext, absPath string) error {
	if v == nil || v.PullStatusFetcher == nil {
		return nil
	}
	planPath, err := safePlanFilePath(ctx, absPath)
	if err != nil {
		return err
	}

	pullStatus, err := v.pullStatusForApply(ctx)
	if err != nil {
		return fmt.Errorf("fetching current plan status: %w", err)
	}
	if pullStatus == nil {
		return rejectProjectPlan(planPath, "no current plan status found; run `atlantis plan` before apply")
	}

	liveHead, err := v.getLiveHeadCommit(ctx)
	if err != nil {
		return fmt.Errorf("fetching live pull request head: %w", err)
	}
	if liveHead != "" {
		if pullStatus.Pull.HeadCommit == "" {
			return fmt.Errorf("recorded plan status has no head commit; run `atlantis plan` before apply")
		}
		if pullStatus.Pull.HeadCommit != liveHead {
			return fmt.Errorf(
				"recorded plan status is from commit %s but live pull request head is %s; run `atlantis plan` before apply",
				shortSHA(pullStatus.Pull.HeadCommit),
				shortSHA(liveHead),
			)
		}
		if ctx.Pull.HeadCommit != "" && looksLikeCommitSHA(ctx.Pull.HeadCommit) && ctx.Pull.HeadCommit != liveHead {
			return fmt.Errorf(
				"%w: pull request head changed from %s to %s; run `atlantis plan` before apply",
				errStaleCommandHead,
				shortSHA(ctx.Pull.HeadCommit),
				shortSHA(liveHead),
			)
		}
	} else if !pullStatusHeadMatchesPull(ctx.Pull, pullStatus.Pull) {
		return rejectProjectPlan(planPath,
			"recorded plan status is from commit %s but current head is %s; run `atlantis plan` before apply",
			shortSHA(pullStatus.Pull.HeadCommit),
			shortSHA(ctx.Pull.HeadCommit),
		)
	}

	proj := findProjectInPullStatus(pullStatus, ctx.Workspace, ctx.RepoRelDir, ctx.ProjectName)
	if proj == nil {
		return rejectProjectPlan(planPath,
			"no matching plan status exists for dir %q workspace %q project %q; run `atlantis plan`",
			ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName,
		)
	}
	if !statusAllowedForApplyExecution(proj.Status) {
		if proj.Status == models.ErroredPolicyCheckStatus {
			return rejectProjectPlan(planPath,
				"policy checks have errored for dir %q workspace %q project %q and cannot be applied; run `atlantis plan`",
				ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName,
			)
		}
		return rejectProjectPlan(planPath,
			"plan for dir %q workspace %q project %q has status %q and cannot be applied; run `atlantis plan`",
			ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, proj.Status.String(),
		)
	}

	if _, err := os.Stat(planPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"plan file is missing for dir %q workspace %q project %q; run `atlantis plan`",
				ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName,
			)
		}
		return fmt.Errorf("checking plan file for dir %q workspace %q project %q: %w", ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, err)
	}

	if ctx.ExpectedPlanHash != "" {
		actualHash, err := hashFile(absPath, planPath)
		if err != nil {
			return fmt.Errorf("hashing plan file for dir %q workspace %q project %q: %w", ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, err)
		}
		if actualHash != ctx.ExpectedPlanHash {
			return fmt.Errorf(
				"plan file changed for dir %q workspace %q project %q; run `atlantis plan` before apply",
				ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName,
			)
		}
	}

	return nil
}

func (v *DefaultApplyPlanValidator) pullStatusForApply(ctx command.ProjectContext) (*models.PullStatus, error) {
	if ctx.API && ctx.PullStatus != nil {
		return ctx.PullStatus, nil
	}
	pullStatus, err := v.PullStatusFetcher.GetPullStatus(ctx.Pull)
	if err != nil {
		return nil, err
	}
	if pullStatus != nil {
		return pullStatus, nil
	}
	return nil, nil
}

func (v *DefaultApplyPlanValidator) getLiveHeadCommit(ctx command.ProjectContext) (string, error) {
	if v.LivePullHeadFetcher == nil {
		return "", nil
	}
	if ctx.API && ctx.Pull.Num <= 0 {
		return "", nil
	}
	liveHead, err := v.LivePullHeadFetcher.GetLiveHeadCommit(ctx)
	if err != nil {
		return "", err
	}
	if liveHead == "" {
		return "", fmt.Errorf("live pull request head is empty")
	}
	return liveHead, nil
}

func statusAllowedForApplyExecution(status models.ProjectPlanStatus) bool {
	switch status {
	case models.PlannedPlanStatus, models.PassedPolicyCheckStatus, models.ErroredApplyStatus,
		models.PlannedNoChangesPlanStatus:
		return true
	default:
		return false
	}
}

func planFilePath(ctx command.ProjectContext, absPath string) string {
	return filepath.Join(absPath, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
}

func safePlanFilePath(ctx command.ProjectContext, absPath string) (string, error) {
	planPath := planFilePath(ctx, absPath)
	if _, err := containedPlanRelPath(absPath, planPath); err != nil {
		return "", err
	}
	return planPath, nil
}

func pendingPlanFilePath(plan PendingPlan) (string, error) {
	absPath := filepath.Join(plan.RepoDir, plan.RepoRelDir)
	planPath := filepath.Join(absPath, runtime.GetPlanFilename(plan.Workspace, plan.ProjectName))
	if _, err := containedPlanRelPath(absPath, planPath); err != nil {
		return "", err
	}
	return planPath, nil
}

func hashFile(baseDir, path string) (string, error) {
	relPath, err := containedPlanRelPath(baseDir, path)
	if err != nil {
		return "", err
	}
	file, err := os.OpenInRoot(filepath.Clean(baseDir), relPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func containedPlanRelPath(baseDir, path string) (string, error) {
	cleanBase := filepath.Clean(baseDir)
	cleanPath := filepath.Clean(path)
	relPath, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil {
		return "", fmt.Errorf("checking plan path containment: %w", err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relPath) {
		return "", fmt.Errorf("plan path traversal detected: %w", utils.ErrPathEscapesBase)
	}
	return relPath, nil
}

func rejectProjectPlan(planPath string, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if removeErr := utils.RemoveIgnoreNonExistent(planPath); removeErr != nil {
		return fmt.Errorf("%w; deleting rejected plan file %q: %v", err, planPath, removeErr)
	}
	return err
}

func looksLikeCommitSHA(s string) bool {
	if len(s) < 7 || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func pullStatusHeadMatchesPull(pull models.PullRequest, statusPull models.PullRequest) bool {
	if pull.HeadCommit == "" || statusPull.HeadCommit == "" {
		return true
	}
	return statusPull.HeadCommit == pull.HeadCommit
}
