package events

import (
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate --package mocks -o mocks/mock_delete_lock_command.go DeleteLockCommand

// DeleteLockCommand is the first step after a command request has been parsed.
type DeleteLockCommand interface {
	DeleteLock(id string) (*models.ProjectLock, error)
	DeleteLocksByPull(repoFullName string, pullNum int) (int, error)
}

// DefaultDeleteLockCommand deletes a specific lock after a request from the LocksController.
type DefaultDeleteLockCommand struct {
	Locker           locking.Locker
	Logger           logging.SimpleLogging
	WorkingDir       WorkingDir
	WorkingDirLocker WorkingDirLocker
	Backend          locking.Backend
}

// DeleteLock handles deleting the lock at id
func (l *DefaultDeleteLockCommand) DeleteLock(id string) (*models.ProjectLock, error) {
	lock, err := l.Locker.Unlock(id)
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, nil
	}

	removeErr := l.WorkingDir.DeletePlan(lock.Pull.BaseRepo, lock.Pull, lock.Workspace, lock.Project.Path, lock.Project.ProjectName)
	if removeErr != nil {
		l.Logger.Warn("Failed to delete plan: %s", removeErr)
		return nil, removeErr
	}

	return lock, nil
}

// DeleteLocksByPull handles deleting all locks for the pull request
func (l *DefaultDeleteLockCommand) DeleteLocksByPull(repoFullName string, pullNum int) (int, error) {
	locks, err := l.Locker.UnlockByPull(repoFullName, pullNum)
	numLocks := len(locks)
	if err != nil {
		return numLocks, err
	}
	if numLocks == 0 {
		l.Logger.Debug("No locks found for repo '%v', pull request: %v", repoFullName, pullNum)
		return numLocks, nil
	}

	for i := 0; i < numLocks; i++ {
		lock := locks[i]

		err := l.WorkingDir.DeletePlan(lock.Pull.BaseRepo, lock.Pull, lock.Workspace, lock.Project.Path, lock.Project.ProjectName)
		if err != nil {
			l.Logger.Warn("Failed to delete plan: %s", err)
			return numLocks, err
		}
	}

	return numLocks, nil
}
