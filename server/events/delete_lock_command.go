package events

import (
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_delete_lock_command.go DeleteLockCommand

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
	DB               *db.BoltDB
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

	l.deleteWorkingDir(*lock)
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
		l.Logger.Debug("No locks found for pull")
		return numLocks, nil
	}

	for i := 0; i < numLocks; i++ {
		lock := locks[i]
		l.deleteWorkingDir(lock)
	}

	return numLocks, nil
}

func (l *DefaultDeleteLockCommand) deleteWorkingDir(lock models.ProjectLock) {
	// NOTE: Because BaseRepo was added to the PullRequest model later, previous
	// installations of Atlantis will have locks in their DB that do not have
	// this field on PullRequest. We skip deleting the working dir in this case.
	if lock.Pull.BaseRepo == (models.Repo{}) {
		l.Logger.Debug("Not deleting the working dir.")
		return
	}
	unlock, err := l.WorkingDirLocker.TryLock(lock.Pull.BaseRepo.FullName, lock.Pull.Num, lock.Workspace)
	if err != nil {
		l.Logger.Err("unable to obtain working dir lock when trying to delete old plans: %s", err)
	} else {
		defer unlock()
		// nolint: vetshadow
		if err := l.WorkingDir.DeleteForWorkspace(lock.Pull.BaseRepo, lock.Pull, lock.Workspace); err != nil {
			l.Logger.Err("unable to delete workspace: %s", err)
		}
	}
	if err := l.DB.UpdateProjectStatus(lock.Pull, lock.Workspace, lock.Project.Path, models.DiscardedPlanStatus); err != nil {
		l.Logger.Err("unable to delete project status: %s", err)
	}
}
