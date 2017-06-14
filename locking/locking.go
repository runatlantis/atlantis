package locking

import (
	"time"
	"fmt"
)

type Run struct {
	RepoFullName string
	Path       string
	Env        string
	PullNum     int
	User       string
	Timestamp  time.Time
}

// StateKey returns the unique key to identify the set of infrastructure being modified by this run.
// Returns `{fullRepoName}/{tfProjectPath}/{environment}`.
// Used in locking to determine what part of the infrastructure is locked.
func (r Run) StateKey() string {
	return fmt.Sprintf("%s/%s/%s", r.RepoFullName, r.Path, r.Env)
}

type TryLockResponse struct {
	LockAcquired bool
	LockingRun   Run // what is currently holding the lock
	LockID       string
}

type Backend interface {
	TryLock(run Run) (TryLockResponse, error)
	Unlock(lockID string) error
	ListLocks() (map[string]Run, error)
	FindLocksForPull(repoFullName string, pullNum int) ([]string, error)
}
