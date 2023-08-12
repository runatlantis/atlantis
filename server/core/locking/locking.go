// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package locking handles locking projects when they have in-progress runs.
package locking

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate --package mocks -o mocks/mock_backend.go Backend

// Backend is an implementation of the locking API we require.
type Backend interface {
	TryLock(lock models.ProjectLock) (bool, models.ProjectLock, *models.EnqueueStatus, error)
	Unlock(project models.Project, workspace string, updateQueue bool) (*models.ProjectLock, *models.ProjectLock, error)
	List() ([]models.ProjectLock, error)
	GetLock(project models.Project, workspace string) (*models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int, updateQueue bool) ([]models.ProjectLock, *models.DequeueStatus, error)

	GetQueueByLock(project models.Project, workspace string) (models.ProjectLockQueue, error)

	UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error
	GetPullStatus(pull models.PullRequest) (*models.PullStatus, error)
	DeletePullStatus(pull models.PullRequest) error
	UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error)

	LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error)
	UnlockCommand(cmdName command.Name) error
	CheckCommandLock(cmdName command.Name) (*command.Lock, error)
}

// TryLockResponse results from an attempted lock.
type TryLockResponse struct {
	// LockAcquired is true if the lock was acquired from this call.
	LockAcquired bool
	// CurrLock is what project is currently holding the lock.
	CurrLock models.ProjectLock
	// CurrLock is what project is currently holding the lock.
	EnqueueStatus *models.EnqueueStatus
	// LockKey is an identified by which to lookup and delete this lock.
	LockKey string
}

// Client is used to perform locking actions.
type Client struct {
	backend Backend
}

//go:generate pegomock generate --package mocks -o mocks/mock_locker.go Locker

type Locker interface {
	TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error)
	Unlock(key string, updateQueue bool) (*models.ProjectLock, *models.ProjectLock, error)
	List() (map[string]models.ProjectLock, error)
	GetQueueByLock(project models.Project, workspace string) (models.ProjectLockQueue, error)
	UnlockByPull(repoFullName string, pullNum int, updateQueue bool) ([]models.ProjectLock, *models.DequeueStatus, error)
	GetLock(key string) (*models.ProjectLock, error)
}

// NewClient returns a new locking client.
func NewClient(backend Backend) *Client {
	return &Client{
		backend: backend,
	}
}

// keyRegex matches and captures {repoFullName}/{path}/{workspace} where path can have multiple /'s in it.
var keyRegex = regexp.MustCompile(`^(.*?\/.*?)\/(.*)\/(.*)$`)

// TryLock attempts to acquire a lock to a project and workspace.
func (c *Client) TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error) {
	lock := models.ProjectLock{
		Workspace: workspace,
		Time:      time.Now().Local(),
		Project:   p,
		User:      user,
		Pull:      pull,
	}
	lockAcquired, currLock, enqueueStatus, err := c.backend.TryLock(lock)
	if err != nil {
		return TryLockResponse{}, err
	}
	return TryLockResponse{lockAcquired, currLock, enqueueStatus, c.key(p, workspace)}, nil
}

// Unlock attempts to unlock a project and workspace. If successful,
// a pointer to the now deleted lock will be returned, as well as a pointer
// to the next dequeued lock from the queue (if any). Else, both
// pointers will be nil. An error will only be returned if there was
// an error deleting the lock or dequeuing (i.e. not if there was no lock).
func (c *Client) Unlock(key string, updateQueue bool) (*models.ProjectLock, *models.ProjectLock, error) {
	project, workspace, err := c.lockKeyToProjectWorkspace(key)
	if err != nil {
		return nil, nil, err
	}
	return c.backend.Unlock(project, workspace, updateQueue)
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
		m[c.key(lock.Project, lock.Workspace)] = lock
	}
	return m, nil
}

func (c *Client) GetQueueByLock(project models.Project, workspace string) (models.ProjectLockQueue, error) {
	return c.backend.GetQueueByLock(project, workspace)
}

// TODO(Ghais) extend the tests
// UnlockByPull deletes all locks associated with that pull request.
func (c *Client) UnlockByPull(repoFullName string, pullNum int, updateQueue bool) ([]models.ProjectLock, *models.DequeueStatus, error) {
	return c.backend.UnlockByPull(repoFullName, pullNum, updateQueue)
}

// GetLock attempts to get the lock stored at key. If successful,
// a pointer to the lock will be returned. Else, the pointer will be nil.
// An error will only be returned if there was an error getting the lock
// (i.e. not if there was no lock).
func (c *Client) GetLock(key string) (*models.ProjectLock, error) {
	project, workspace, err := c.lockKeyToProjectWorkspace(key)
	if err != nil {
		return nil, err
	}

	projectLock, err := c.backend.GetLock(project, workspace)
	if err != nil {
		return nil, err
	}

	return projectLock, nil
}

func (c *Client) key(p models.Project, workspace string) string {
	return fmt.Sprintf("%s/%s/%s", p.RepoFullName, p.Path, workspace)
}

func (c *Client) lockKeyToProjectWorkspace(key string) (models.Project, string, error) {
	matches := keyRegex.FindStringSubmatch(key)
	if len(matches) != 4 {
		return models.Project{}, "", errors.New("invalid key format")
	}

	return models.Project{RepoFullName: matches[1], Path: matches[2]}, matches[3], nil
}

type NoOpLocker struct{}

// NewNoOpLocker returns a new lno operation lockingclient.
func NewNoOpLocker() *NoOpLocker {
	return &NoOpLocker{}
}

// TryLock attempts to acquire a lock to a project and workspace.
func (c *NoOpLocker) TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error) {
	return TryLockResponse{true, models.ProjectLock{}, nil, c.key(p, workspace)}, nil
}

// Unlock attempts to unlock a project and workspace. If successful,
// pointers to the now deleted lock and the next dequeued lock (optional) will be returned.
// Else, both pointers will be nil. An error will only be returned if there was
// an error deleting the lock (i.e. not if there was no lock).
func (c *NoOpLocker) Unlock(string, bool) (*models.ProjectLock, *models.ProjectLock, error) {
	return &models.ProjectLock{}, &models.ProjectLock{}, nil
}

// List returns a map of all locks with their lock key as the map key.
// The lock key can be used in GetLock() and Unlock().
func (c *NoOpLocker) List() (map[string]models.ProjectLock, error) {
	m := make(map[string]models.ProjectLock)
	return m, nil
}

func (c *NoOpLocker) GetQueueByLock(models.Project, string) (models.ProjectLockQueue, error) {
	m := make(models.ProjectLockQueue, 0)
	return m, nil
}

// UnlockByPull deletes all locks associated with that pull request.
func (c *NoOpLocker) UnlockByPull(_ string, _ int, _ bool) ([]models.ProjectLock, *models.DequeueStatus, error) {
	return []models.ProjectLock{}, nil, nil
}

// GetLock attempts to get the lock stored at key. If successful,
// a pointer to the lock will be returned. Else, the pointer will be nil.
// An error will only be returned if there was an error getting the lock
// (i.e. not if there was no lock).
func (c *NoOpLocker) GetLock(key string) (*models.ProjectLock, error) {
	return nil, nil
}

func (c *NoOpLocker) key(p models.Project, workspace string) string {
	return fmt.Sprintf("%s/%s/%s", p.RepoFullName, p.Path, workspace)
}
