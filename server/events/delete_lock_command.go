// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
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

	removeErr := l.WorkingDir.DeletePlan(logger, lock.Pull.BaseRepo, lock.Pull, lock.Workspace, lock.Project.Path, lock.Project.ProjectName)
	if removeErr != nil {
		logger.Warn("Failed to delete plan: %s", removeErr)
		return nil, removeErr
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

	return lock, nil
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
