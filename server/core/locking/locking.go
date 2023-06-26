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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_backend.go Backend

// Backend is an implementation of the locking API we require.
type Backend interface {
	TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
	Unlock(project models.Project, workspace string) (*models.ProjectLock, error)
	TryLockFilePath(pullKey string, workspaceKey string) (bool, error)
	UnlockFilePath(workspaceKey string) (bool, error)
	TryLockPullFilePath(pullKey string) (bool, error)
	UnlockLockPullFilePath(pullKey string) (bool, error)
	List() ([]models.ProjectLock, error)
	GetLock(project models.Project, workspace string) (*models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
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
	// LockKey is an identified by which to lookup and delete this lock.
	LockKey string
}

// Client is used to perform locking actions.
type Client struct {
	backend Backend
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_locker.go Locker

type Locker interface {
	TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error)
	Unlock(key string) (*models.ProjectLock, error)
	List() (map[string]models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
	GetLock(key string) (*models.ProjectLock, error)
	TryLockFilePath(pullKey string, workspaceKey string) (bool, error)
	UnlockFilePath(workspaceKey string) (bool, error)
	TryLockPullFilePath(pullKey string) (bool, error)
	UnlockLockPullFilePath(pullKey string) (bool, error)
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
	lockAcquired, currLock, err := c.backend.TryLock(lock)
	if err != nil {
		return TryLockResponse{}, err
	}
	return TryLockResponse{lockAcquired, currLock, c.key(p, workspace)}, nil
}

// Unlock attempts to unlock a project and workspace. If successful,
// a pointer to the now deleted lock will be returned. Else, that
// pointer will be nil. An error will only be returned if there was
// an error deleting the lock (i.e. not if there was no lock).
func (c *Client) Unlock(key string) (*models.ProjectLock, error) {
	project, workspace, err := c.lockKeyToProjectWorkspace(key)
	if err != nil {
		return nil, err
	}
	return c.backend.Unlock(project, workspace)
}

// TryLockFilePath attempts to acquire a lock to a file path
func (c *Client) TryLockFilePath(pullKey string, workspaceKey string) (bool, error) {
	isLockAcquired, err := c.backend.TryLockFilePath(pullKey, workspaceKey)

	if err != nil {
		return false, err
	}
	return isLockAcquired, nil

}

// UnlockFilePath removes a lock to a file path
func (c *Client) UnlockFilePath(workspaceKey string) (bool, error) {
	return c.backend.UnlockFilePath(workspaceKey)
}

func (c *Client) TryLockPullFilePath(pullKey string) (bool, error) {
	isLockAcquired, err := c.backend.TryLockPullFilePath(pullKey)

	if err != nil {
		return false, err
	}
	return isLockAcquired, nil
}

func (c *Client) UnlockLockPullFilePath(pullKey string) (bool, error) {
	return c.backend.UnlockLockPullFilePath(pullKey)
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

// UnlockByPull deletes all locks associated with that pull request.
func (c *Client) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return c.backend.UnlockByPull(repoFullName, pullNum)
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
	return TryLockResponse{true, models.ProjectLock{}, c.key(p, workspace)}, nil
}

// Unlock attempts to unlock a project and workspace. If successful,
// a pointer to the now deleted lock will be returned. Else, that
// pointer will be nil. An error will only be returned if there was
// an error deleting the lock (i.e. not if there was no lock).
func (c *NoOpLocker) Unlock(key string) (*models.ProjectLock, error) {
	return &models.ProjectLock{}, nil
}

// List returns a map of all locks with their lock key as the map key.
// The lock key can be used in GetLock() and Unlock().
func (c *NoOpLocker) List() (map[string]models.ProjectLock, error) {
	m := make(map[string]models.ProjectLock)
	return m, nil
}

// UnlockByPull deletes all locks associated with that pull request.
func (c *NoOpLocker) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return []models.ProjectLock{}, nil
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
