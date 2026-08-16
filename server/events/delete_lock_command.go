// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
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

	return lock, nil
}

// DeleteLocksByPull handles deleting all locks for the pull request
func (l *DefaultDeleteLockCommand) DeleteLocksByPull(logger logging.SimpleLogging, repoFullName string, pullNum int) (int, error) {
	locks, err := l.Locker.UnlockByPull(repoFullName, pullNum)
	numLocks := len(locks)
	if err != nil {
		return numLocks, err
	}

	if numLocks == 0 {
		logger.Debug("No locks found for repo '%v', pull request: %v", repoFullName, pullNum)
	}

	return numLocks, nil
}
