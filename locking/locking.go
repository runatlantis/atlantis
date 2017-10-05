// Package locking handles locking projects when they have in-progress runs.
package locking

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hootsuite/atlantis/models"
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_backend.go Backend

// Backend is an implementation of the locking API we require.
type Backend interface {
	TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
	Unlock(project models.Project, env string) (*models.ProjectLock, error)
	List() ([]models.ProjectLock, error)
	GetLock(project models.Project, env string) (*models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
}

// TryLockResponse results from an attempted lock.
type TryLockResponse struct {
	// LockAcquired is true if the lock was acquired from this call.
	LockAcquired bool
	// CurrLock is what project is currently holding the lock.
	CurrLock models.ProjectLock
	// LockKey is an identified by which to lookup and delete this lock.
	LockKey string
}

// Client is used to perform locking actions.
type Client struct {
	backend Backend
}

type Locker interface {
	TryLock(p models.Project, env string, pull models.PullRequest, user models.User) (TryLockResponse, error)
	Unlock(key string) (*models.ProjectLock, error)
	List() (map[string]models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
	GetLock(key string) (*models.ProjectLock, error)
}

// NewClient returns a new locking client.
func NewClient(backend Backend) *Client {
	return &Client{
		backend: backend,
	}
}

// keyRegex matches and captures {repoFullName}/{path}/{env} where path can have multiple /'s in it.
var keyRegex = regexp.MustCompile(`^(.*?\/.*?)\/(.*)\/(.*)$`)

// TryLock attempts to acquire a lock to a project and environment.
func (c *Client) TryLock(p models.Project, env string, pull models.PullRequest, user models.User) (TryLockResponse, error) {
	lock := models.ProjectLock{
		Env:     env,
		Time:    time.Now().Local(),
		Project: p,
		User:    user,
		Pull:    pull,
	}
	lockAcquired, currLock, err := c.backend.TryLock(lock)
	if err != nil {
		return TryLockResponse{}, err
	}
	return TryLockResponse{lockAcquired, currLock, c.key(p, env)}, nil
}

// Unlock attempts to unlock a project and environment. If successful,
// a pointer to the now deleted lock will be returned. Else, that
// pointer will be nil. An error will only be returned if there was
// an error deleting the lock (i.e. not if there was no lock).
func (c *Client) Unlock(key string) (*models.ProjectLock, error) {
	project, env, err := c.lockKeyToProjectEnv(key)
	if err != nil {
		return nil, err
	}
	return c.backend.Unlock(project, env)
}

// List returns a map of all locks with their lock key as the map key.
// The lock key can be used in GetLock() and Unlock().
func (c *Client) List() (map[string]models.ProjectLock, error) {
	m := make(map[string]models.ProjectLock)
	locks, err := c.backend.List()
	if err != nil {
		return m, err
	}
	for _, lock := range locks {
		m[c.key(lock.Project, lock.Env)] = lock
	}
	return m, nil
}

// UnlockByPull deletes all locks associated with that pull request.
func (c *Client) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return c.backend.UnlockByPull(repoFullName, pullNum)
}

// GetLock attempts to get the lock stored at key. If successful,
// a pointer to the lock will be returned. Else, the pointer will be nil.
// An error will only be returned if there was an error getting the lock
// (i.e. not if there was no lock).
func (c *Client) GetLock(key string) (*models.ProjectLock, error) {
	project, env, err := c.lockKeyToProjectEnv(key)
	if err != nil {
		return nil, err
	}

	projectLock, err := c.backend.GetLock(project, env)
	if err != nil {
		return nil, err
	}

	return projectLock, nil
}

func (c *Client) key(p models.Project, env string) string {
	return fmt.Sprintf("%s/%s/%s", p.RepoFullName, p.Path, env)
}

func (c *Client) lockKeyToProjectEnv(key string) (models.Project, string, error) {
	matches := keyRegex.FindStringSubmatch(key)
	if len(matches) != 4 {
		return models.Project{}, "", errors.New("invalid key format")
	}

	return models.Project{RepoFullName: matches[1], Path: matches[2]}, matches[3], nil
}
