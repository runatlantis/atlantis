// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package coordination

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// TryLockResponse results from an attempted lock.
type TryLockResponse struct {
	// LockAcquired is true if the lock was acquired from this call.
	LockAcquired bool
	// CurrLock is what project is currently holding the lock.
	CurrLock models.ProjectLock
	// LockKey is an identified by which to lookup and delete this lock.
	LockKey string
}

// LockingClient is used to perform locking actions.
type LockingClient struct {
	store Store
}

//go:generate go tool mockgen -package mocks -destination mocks/mock_locker.go . Locker

type Locker interface {
	TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error)
	Unlock(key string) (*models.ProjectLock, error)
	UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error)
	List() (map[string]models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
	GetLock(key string) (*models.ProjectLock, error)
}

// NewLockingClient returns a new locking client.
func NewLockingClient(store Store) *LockingClient {
	return &LockingClient{
		store: store,
	}
}

// TryLock attempts to acquire a lock to a project and workspace.
func (c *LockingClient) TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error) {
	lock := models.ProjectLock{
		Workspace: workspace,
		Time:      time.Now().Local(),
		Project:   p,
		User:      user,
		Pull:      pull,
	}
	lockAcquired, currLock, err := c.store.TryLock(lock)
	if err != nil {
		return TryLockResponse{}, err
	}
	return TryLockResponse{lockAcquired, currLock, c.key(p, workspace)}, nil
}

// Unlock attempts to unlock a project and workspace. If successful,
// a pointer to the now deleted lock will be returned. Else, that
// pointer will be nil. An error will only be returned if there was
// an error deleting the lock (i.e. not if there was no lock).
func (c *LockingClient) Unlock(key string) (*models.ProjectLock, error) {
	project, workspace, err := models.ParseLockKey(key)
	if err != nil {
		return nil, err
	}
	return c.store.Unlock(project, workspace)
}

// UnlockIfOwnedByPull unlocks project and workspace only when it is still owned by pullNum.
func (c *LockingClient) UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error) {
	return c.store.UnlockIfOwnedByPull(project, workspace, pullNum)
}

// List returns a map of all locks with their lock key as the map key.
// The lock key can be used in GetLock() and Unlock().
func (c *LockingClient) List() (map[string]models.ProjectLock, error) {
	m := make(map[string]models.ProjectLock)
	locks, err := c.store.List()
	if err != nil {
		return m, err
	}
	for _, lock := range locks {
		m[c.key(lock.Project, lock.Workspace)] = lock
	}
	return m, nil
}

// UnlockByPull deletes all locks associated with that pull request.
func (c *LockingClient) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return c.store.UnlockByPull(repoFullName, pullNum)
}

// GetLock attempts to get the lock stored at key. If successful,
// a pointer to the lock will be returned. Else, the pointer will be nil.
// An error will only be returned if there was an error getting the lock
// (i.e. not if there was no lock).
func (c *LockingClient) GetLock(key string) (*models.ProjectLock, error) {
	project, workspace, err := models.ParseLockKey(key)
	if err != nil {
		return nil, err
	}

	projectLock, err := c.store.GetLock(project, workspace)
	if err != nil {
		return nil, err
	}

	return projectLock, nil
}

func (c *LockingClient) key(p models.Project, workspace string) string {
	return models.GenerateLockKey(p, workspace)
}
