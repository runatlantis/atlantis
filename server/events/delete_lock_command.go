package events

import (
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate --package mocks -o mocks/mock_delete_lock_command.go DeleteLockCommand

// DeleteLockCommand is the first step after a command request has been parsed.
type DeleteLockCommand interface {
	DeleteLock(id string) (*models.ProjectLock, *models.ProjectLock, error)
	DeleteLocksByPull(repoFullName string, pullNum int) (int, *models.DequeueStatus, error)
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
func (l *DefaultDeleteLockCommand) DeleteLock(id string) (*models.ProjectLock, *models.ProjectLock, error) {
	// TODO(Ghais) extend the tests
	// TODO(Ghais) #9 When the lock(s) has(ve) been explicitly removed, also dequeue the next PR(s)
	lock, dequeuedLock, err := l.Locker.Unlock(id, true)
	if err != nil {
		return nil, nil, err
	}
	if lock == nil {
		return nil, nil, nil
	}

	// The locks controller currently has no implementation of Atlantis project names, so this is hardcoded to an empty string.
	projectName := ""

	removeErr := l.WorkingDir.DeletePlan(lock.Pull.BaseRepo, lock.Pull, lock.Workspace, lock.Project.Path, projectName)
	if removeErr != nil {
		l.Logger.Warn("Failed to delete plan: %s", removeErr)
		return nil, nil, removeErr
	}

	return lock, dequeuedLock, nil
}

// DeleteLocksByPull handles deleting all locks for the pull request
func (l *DefaultDeleteLockCommand) DeleteLocksByPull(repoFullName string, pullNum int) (int, *models.DequeueStatus, error) {
	locks, dequeueStatus, err := l.Locker.UnlockByPull(repoFullName, pullNum, true)
	numLocks := len(locks)
	if err != nil {
		return numLocks, dequeueStatus, err
	}
	if numLocks == 0 {
		l.Logger.Debug("No locks found for pull")
		return numLocks, dequeueStatus, nil
	}

	for i := 0; i < numLocks; i++ {
		lock := locks[i]
		l.deleteWorkingDir(lock)
	}

	return numLocks, dequeueStatus, nil
}

func (l *DefaultDeleteLockCommand) deleteWorkingDir(lock models.ProjectLock) {
	// NOTE: Because BaseRepo was added to the PullRequest model later, previous
	// installations of Atlantis will have locks in their DB that do not have
	// this field on PullRequest. We skip deleting the working dir in this case.
	if lock.Pull.BaseRepo == (models.Repo{}) {
		l.Logger.Debug("Not deleting the working dir.")
		return
	}
	unlock, err := l.WorkingDirLocker.TryLock(lock.Pull.BaseRepo.FullName, lock.Pull.Num, lock.Workspace, lock.Project.Path)
	if err != nil {
		l.Logger.Err("unable to obtain working dir lock when trying to delete old plans: %s", err)
	} else {
		defer unlock()
		// nolint: vetshadow
		if err := l.WorkingDir.DeleteForWorkspace(lock.Pull.BaseRepo, lock.Pull, lock.Workspace); err != nil {
			l.Logger.Err("unable to delete workspace: %s", err)
		}
	}
	if err := l.Backend.UpdateProjectStatus(lock.Pull, lock.Workspace, lock.Project.Path, models.DiscardedPlanStatus); err != nil {
		l.Logger.Err("unable to delete project status: %s", err)
	}
}
