package events

import (
	"fmt"
	"sync"
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_workspace_locker.go WorkspaceLocker

// WorkspaceLocker is used to prevent multiple commands from executing at the same
// time for a single repo, pull, and environment. We need to prevent this from
// happening because a specific repo/pull/env has a single workspace on disk
// and we haven't written Atlantis (yet) to handle concurrent execution within
// this workspace.
type WorkspaceLocker interface {
	// TryLock tries to acquire a lock for this repo, env and pull.
	TryLock(repoFullName string, env string, pullNum int) bool
	// Unlock deletes the lock for this repo, env and pull. If there was no
	// lock it will do nothing.
	Unlock(repoFullName, env string, pullNum int)
}

// DefaultWorkspaceLocker implements WorkspaceLocker.
type DefaultWorkspaceLocker struct {
	mutex sync.Mutex
	locks map[string]interface{}
}

// NewDefaultWorkspaceLocker is a constructor.
func NewDefaultWorkspaceLocker() *DefaultWorkspaceLocker {
	return &DefaultWorkspaceLocker{
		locks: make(map[string]interface{}),
	}
}

// TryLock returns true if a lock is acquired for this repo, pull and env and
// false otherwise.
func (d *DefaultWorkspaceLocker) TryLock(repoFullName string, env string, pullNum int) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	key := d.key(repoFullName, env, pullNum)
	if _, ok := d.locks[key]; !ok {
		d.locks[key] = true
		return true
	}
	return false
}

// Unlock unlocks the repo, pull and env.
func (d *DefaultWorkspaceLocker) Unlock(repoFullName, env string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.locks, d.key(repoFullName, env, pullNum))
}

func (d *DefaultWorkspaceLocker) key(repo string, env string, pull int) string {
	return fmt.Sprintf("%s/%s/%d", repo, env, pull)
}
