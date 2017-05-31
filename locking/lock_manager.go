package locking

import (
	"time"
	"fmt"
	"crypto/sha256"
)

type Run struct {
	RepoOwner  string
	RepoName   string
	Path       string
	Env        string
	PullID     int
	User       string
	Timestamp  time.Time
}

// StateKey returns the unique key to identify the set of infrastructure being modified by this run.
// Used in locking to determine what part of the infrastructure is locked.
func (r Run) StateKey() []byte {
	// we combine the repository, path into the repo where terraform is being run and the environment
	// to make up the state key and then hash it with sha256
	key := fmt.Sprintf("%s/%s/%s/%s", r.RepoOwner, r.RepoName, r.Path, r.Env)
	h := sha256.New()
	h.Write([]byte(key))
	return h.Sum(nil)
}

type TryLockResponse struct {
	LockAcquired bool
	LockingRun   Run // what is currently holding the lock
	LockID       string
}

type LockManager interface {
	TryLock(run Run) (TryLockResponse, error)
	Unlock(lockID string) error
	ListLocks() (map[string]Run, error)
}
