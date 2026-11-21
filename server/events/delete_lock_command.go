// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/planstore"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_delete_lock_command.go DeleteLockCommand

// DeleteLockCommand is the first step after a command request has been parsed.
type DeleteLockCommand interface {
	DeleteLock(logger logging.SimpleLogging, id string) (*models.ProjectLock, error)
	DeleteLocksByPull(logger logging.SimpleLogging, repoFullName string, pullNum int) (int, error)
	DeleteLocksByProject(logger logging.SimpleLogging, repoFullName string, pullNum int, projectName string) (int, error)
}

// DefaultDeleteLockCommand deletes a specific lock after a request from the LocksController.
type DefaultDeleteLockCommand struct {
	Locker           locking.Locker
	WorkingDir       WorkingDir
	WorkingDirLocker WorkingDirLocker
	Database         db.Database
	PlanStore        planstore.PlanStore
}

// DeleteLock handles deleting the lock at id
func (l *DefaultDeleteLockCommand) DeleteLock(logger logging.SimpleLogging, id string) (*models.ProjectLock, error) {
	lock, err := l.Locker.Unlock(id)
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, nil
	}

	if err := l.deleteProjectPlan(logger, lock); err != nil {
		return nil, err
	}
	return lock, nil
}

// deleteProjectPlan is shared by UI and comment-triggered project unlocks.
func (l *DefaultDeleteLockCommand) deleteProjectPlan(logger logging.SimpleLogging, lock *models.ProjectLock) error {
	removeErr := l.WorkingDir.DeletePlan(logger, lock.Pull.BaseRepo, lock.Pull, lock.Workspace, lock.Project.Path, lock.Project.ProjectName)
	if removeErr != nil {
		logger.Warn("Failed to delete plan: %s", removeErr)
		return removeErr
	}

	if l.PlanStore != nil {
		if err := l.PlanStore.DeletePlanForProject(
			lock.Pull.BaseRepo.Owner,
			lock.Pull.BaseRepo.Name,
			lock.Pull.Num,
			lock.Workspace,
			lock.Project.Path,
			lock.Project.ProjectName,
		); err != nil {
			logger.Warn("Failed to delete plan from external store: %s", err)
		}
	}

	return nil
}

// DeleteLocksByProject discards only locks and plans belonging to the named
// project in this pull request. It uses saved locks rather than checking out
// the repo, so projects removed from the current config can still be unlocked.
func (l *DefaultDeleteLockCommand) DeleteLocksByProject(logger logging.SimpleLogging, repoFullName string, pullNum int, projectName string) (int, error) {
	if strings.TrimSpace(projectName) == "" {
		return 0, fmt.Errorf("project name must not be empty")
	}
	locks, err := l.Locker.List()
	if err != nil {
		return 0, fmt.Errorf("listing project locks: %w", err)
	}

	var numLocks int
	for _, candidate := range locks {
		if candidate.Project.RepoFullName != repoFullName || candidate.Pull.Num != pullNum || candidate.Project.ProjectName != projectName {
			continue
		}
		// Ownership may have changed since List. Never unlock another PR's lock.
		lock, err := l.Locker.UnlockIfOwnedByPull(candidate.Project, candidate.Workspace, pullNum)
		if err != nil {
			return numLocks, fmt.Errorf("unlocking project %q: %w", projectName, err)
		}
		if lock == nil {
			continue
		}
		numLocks++
		if err := l.deleteProjectPlan(logger, lock); err != nil {
			return numLocks, fmt.Errorf("discarding plan for project %q: %w", projectName, err)
		}
		if err := l.Database.UpdateProjectStatus(lock.Pull, lock.Workspace, lock.Project.Path, lock.Project.ProjectName, models.DiscardedPlanStatus); err != nil {
			return numLocks, fmt.Errorf("updating status for project %q: %w", projectName, err)
		}
	}
	if numLocks == 0 {
		return 0, fmt.Errorf("no Atlantis lock found for project %q in this PR", projectName)
	}
	return numLocks, nil
}

// DeleteLocksByPull handles deleting all locks for the pull request
func (l *DefaultDeleteLockCommand) DeleteLocksByPull(logger logging.SimpleLogging, repoFullName string, pullNum int) (int, error) {
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

	// Always clean up the external plan store for this pull, even when no
	// locks were found locally. Locks can be cleaned via other paths (manual
	// unlock, partial failure) which would otherwise leave orphaned S3 plans.
	if l.PlanStore != nil {
		owner, repo, ok := strings.Cut(repoFullName, "/")
		if !ok {
			logger.Warn("cannot parse owner/repo from %q; skipping external plan store cleanup", repoFullName)
		} else if err := l.PlanStore.DeleteForPull(owner, repo, pullNum); err != nil {
			logger.Warn("Failed to delete plans from external store: %s", err)
		}
	}

	if numLocks == 0 {
		logger.Debug("No locks found for repo '%v', pull request: %v", repoFullName, pullNum)
	}

	return numLocks, nil
}
