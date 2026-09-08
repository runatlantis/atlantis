// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/planstore"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/utils"
)

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_delete_lock_command.go DeleteLockCommand

// DeleteLockCommand is the first step after a command request has been parsed.
type DeleteLockCommand interface {
	DeleteLock(logger logging.SimpleLogging, id string, mode command.PublicationWriteMode) (*models.ProjectLock, bool, error)
	DeleteLocksByPull(logger logging.SimpleLogging, pull models.PullRequest, mode command.PublicationWriteMode) (int, error)
}

// PlanArtifactReaper reaps an exact artifact captured before a durable transition.
type PlanArtifactReaper interface {
	ReapPlan(logging.SimpleLogging, models.PullRequest, models.ProjectStatus) error
}

// DefaultDeleteLockCommand deletes a specific lock after a request from the LocksController.
type DefaultDeleteLockCommand struct {
	Locker            locking.Locker
	WorkingDir        WorkingDir
	WorkingDirLocker  WorkingDirLocker
	Database          db.Database
	PlanStore         planstore.PlanStore
	DataDir           string
	LocalSharePlanDir string
}

// DeleteLock handles deleting the lock at id
func (l *DefaultDeleteLockCommand) DeleteLock(logger logging.SimpleLogging, id string, mode command.PublicationWriteMode) (*models.ProjectLock, bool, error) {
	lock, err := l.Locker.GetLock(id)
	if err != nil {
		return nil, false, err
	}
	if lock == nil {
		return nil, false, nil
	}
	if l.Database == nil {
		return nil, false, fmt.Errorf("reading exact durable plan status: database is unavailable")
	}
	status, err := l.Database.GetPullStatus(lock.Pull)
	if err != nil {
		return nil, false, fmt.Errorf("reading exact durable plan status: %w", err)
	}
	if status == nil {
		return nil, false, db.ErrPlanStatusNotFound
	}
	project := findProjectInPullStatus(status, lock.Workspace, lock.Project.Path, lock.Project.ProjectName)
	if project == nil {
		return nil, false, db.ErrPlanStatusNotFound
	}
	discarded, err := l.Database.DiscardPlanStatus(status.Pull, *project, mode)
	if err != nil {
		return nil, false, err
	}
	// The durable transition happens first. A missing/mismatched status must
	// never leave only a deleted lock while an applyable plan remains.
	deleted, err := l.Locker.UnlockIfOwnedByPull(lock.Project, lock.Workspace, lock.Pull.Num)
	if err != nil {
		return nil, false, err
	}
	if deleted == nil {
		return nil, false, db.ErrPlanGenerationSuperseded
	}

	if err := l.ReapPlan(logger, status.Pull, *project); err != nil {
		return nil, false, err
	}

	return lock, discarded, nil
}

// DeleteLocksByPull handles deleting all locks for the pull request
func (l *DefaultDeleteLockCommand) DeleteLocksByPull(logger logging.SimpleLogging, pull models.PullRequest, mode command.PublicationWriteMode) (int, error) {
	var captured []models.ProjectStatus
	repoFullName, pullNum := pull.BaseRepo.FullName, pull.Num
	if l.Database != nil {
		status, err := l.Database.GetPullStatus(pull)
		if err != nil {
			return 0, err
		}
		if status != nil {
			pull = status.Pull
			captured = status.Projects
			for _, project := range captured {
				if _, err := l.Database.DiscardPlanStatus(pull, project, mode); err != nil {
					return 0, err
				}
			}
		}
	}

	locks, err := l.Locker.UnlockByPull(repoFullName, pullNum)
	numLocks := len(locks)
	if err != nil {
		return numLocks, err
	}

	for i := range numLocks {
		lock := locks[i]

		err := l.WorkingDir.DeletePlan(logger, lock.Pull.BaseRepo, lock.Pull, lock.Workspace, lock.Project.Path, lock.Project.ProjectName)
		if err != nil {
			logger.Warn("Failed to delete plan: %s", err)
			return numLocks, err
		}
	}

	for _, project := range captured {
		if project.AcceptedPlanGeneration != "" && project.ManagedPlan {
			if err := l.ReapPlan(logger, pull, project); err != nil {
				return numLocks, err
			}
		}
	}

	// Always clean up the external plan store for this pull, even when no
	// locks were found locally. Locks can be cleaned via other paths (manual
	// unlock, partial failure) which would otherwise leave orphaned S3 plans.
	if l.PlanStore != nil {
		owner, repo, ok := strings.Cut(repoFullName, "/")
		if !ok {
			logger.Warn("cannot parse owner/repo from %q; skipping external plan store cleanup", repoFullName)
		} else {
			clean := l.PlanStore.DeleteForPull
			if cleaner, ok := l.PlanStore.(planstore.LegacyPlanCleaner); ok {
				clean = cleaner.DeleteLegacyForPull
			}
			if err := clean(owner, repo, pullNum); err != nil {
				logger.Warn("deleting legacy plans from external store %v", err)
			}
		}
	}

	if numLocks == 0 {
		logger.Debug("No locks found for repo '%v', pull request: %v", repoFullName, pullNum)
	}

	return numLocks, nil
}

// ReapPlan deletes only the captured accepted artifact plus disposable canonical
// caches. Callers invalidate or supersede durable status before invoking it.
func (l *DefaultDeleteLockCommand) ReapPlan(logger logging.SimpleLogging, pull models.PullRequest, project models.ProjectStatus) error {
	removeErr := l.WorkingDir.DeletePlan(logger, pull.BaseRepo, pull, project.Workspace, project.RepoRelDir, project.ProjectName)
	if removeErr != nil {
		logger.Warn("Failed to delete plan: %s", removeErr)
		return removeErr
	}

	if l.PlanStore != nil && project.AcceptedPlanGeneration != "" && project.ManagedPlan {
		ctx := command.ProjectContext{BaseRepo: pull.BaseRepo, Pull: pull, Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, ProjectName: project.ProjectName, PlanGeneration: project.PlanGeneration, AcceptedPlanGeneration: project.AcceptedPlanGeneration, ExpectedPlanHash: project.ManagedPlanHash, LocalSharePlanDir: l.LocalSharePlanDir}
		root := l.DataDir
		if l.LocalSharePlanDir != "" {
			root = l.LocalSharePlanDir
		}
		if root == "" {
			return fmt.Errorf("reaping accepted plan: plan directory is unavailable")
		}
		projectPath := filepath.Join(planstore.PullDir(l.DataDir, ctx.BaseRepo.FullName, ctx.Pull.Num), ctx.Workspace, ctx.RepoRelDir)
		planPath := runtime.GetPlanFilePath(ctx, projectPath)
		if err := utils.EnsureSubPath(root, planPath); err != nil {
			return err
		}
		if err := l.PlanStore.Remove(ctx, planPath); err != nil {
			return fmt.Errorf("reaping discarded plan: %w", err)
		}
	} else if l.PlanStore != nil {
		if err := l.PlanStore.DeletePlanForProject(
			pull.BaseRepo.Owner,
			pull.BaseRepo.Name,
			pull.Num,
			project.Workspace,
			project.RepoRelDir,
			project.ProjectName,
		); err != nil {
			logger.Warn("Failed to delete plan from external store: %s", err)
		}
	}

	return nil
}
