package events

import (
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// DeleteLockCommand deletes a specific lock after a request from the LocksController.
type DeleteLockCommand struct {
	Locker           locking.Locker
	Logger           *logging.SimpleLogger
	WorkingDir       WorkingDir
	WorkingDirLocker WorkingDirLocker
	DB               *db.BoltDB
}

// DeleteLock handles deleting the lock at id
func (l *DeleteLockCommand) DeleteLock(id string) (models.DeleteLockCommandResult, *models.ProjectLock, error) {
	lock, err := l.Locker.Unlock(id)
	if err != nil {
		return models.DeleteLockFail, nil, err
	}
	if lock == nil {
		return models.DeleteLockNotFound, nil, nil
	}

	l.deleteWorkingDir(*lock)

	return models.DeleteLockSuccess, lock, nil
}

// DeleteLocksByPull handles deleting all locks for the pull request
func (l *DeleteLockCommand) DeleteLocksByPull(repoFullName string, pullNum int) (models.DeleteLockCommandResult, error) {
	locks, err := l.Locker.UnlockByPull(repoFullName, pullNum)
	if err != nil {
		return models.DeleteLockFail, err
	}
	if len(locks) == 0 {
		return models.DeleteLockNotFound, nil
	}

	for i := 0; i < len(locks); i++ {
		lock := locks[i]
		l.deleteWorkingDir(lock)
	}

	return models.DeleteLockSuccess, nil
}

func (l *DeleteLockCommand) deleteWorkingDir(lock models.ProjectLock) {
	// NOTE: Because BaseRepo was added to the PullRequest model later, previous
	// installations of Atlantis will have locks in their DB that do not have
	// this field on PullRequest. We skip deleting the working dir in this case.
	if lock.Pull.BaseRepo != (models.Repo{}) {
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
		if err := l.DB.DeleteProjectStatus(lock.Pull, lock.Workspace, lock.Project.Path); err != nil {
			l.Logger.Err("unable to delete project status: %s", err)
		}
	}
}
