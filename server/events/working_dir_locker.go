// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"strings"
	"sync"

	"github.com/runatlantis/atlantis/server/events/command"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_working_dir_locker.go WorkingDirLocker

// WorkingDirLocker is used to prevent multiple commands from executing
// at the same time for a single repo, pull, and workspace. We need to prevent
// this from happening because a specific repo/pull/workspace has a single workspace
// on disk and we haven't written Atlantis (yet) to handle concurrent execution
// within this workspace.
type WorkingDirLocker interface {
	// TryLockPull tries to acquire a pull-level lock for a command. It is used to
	// coordinate commands that can make pull-wide plan state inconsistent.
	TryLockPull(repoFullName string, pullNum int, cmdName command.Name) (func(), error)
	// TryLock tries to acquire a lock for this repo, pull, workspace, and path.
	// It returns a function that should be used to unlock the workspace and
	// an error if the workspace is already locked. The error is expected to
	// be printed to the pull request.
	TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name) (func(), error)
	// HasCommandLock reports whether this pull request has an active lock for cmdName.
	HasCommandLock(repoFullName string, pullNum int, cmdName command.Name) bool
	// UnlockByPull unlocks all workspaces for a specific pull request
	UnlockByPull(repoFullName string, pullNum int)
}

// DefaultWorkingDirLocker implements WorkingDirLocker.
type DefaultWorkingDirLocker struct {
	// mutex prevents against multiple threads calling functions on this struct
	// concurrently. It's only used for entry/exit to each function.
	mutex sync.Mutex
	// locks is a map of workspaces showing the name of the command locking it
	locks map[string]command.Name
}

// NewDefaultWorkingDirLocker is a constructor.
func NewDefaultWorkingDirLocker() *DefaultWorkingDirLocker {
	return &DefaultWorkingDirLocker{locks: make(map[string]command.Name)}
}

func (d *DefaultWorkingDirLocker) TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	workspaceKey := d.workspaceKey(repoFullName, pullNum, workspace, path, projectName)
	if currentLock, exists := d.locks[workspaceKey]; exists {
		return func() {}, fmt.Errorf("cannot run %q: the %s workspace at path %s is currently locked for this pull request by %q.\n"+
			"Wait until the previous command is complete and try again", cmdName, workspace, path, currentLock)
	}
	d.locks[workspaceKey] = cmdName
	return func() {
		d.unlock(repoFullName, pullNum, workspace, path, projectName)
	}, nil
}

func (d *DefaultWorkingDirLocker) TryLockPull(repoFullName string, pullNum int, cmdName command.Name) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	pullKey := d.pullKey(repoFullName, pullNum)
	if currentLock, exists := d.locks[pullKey]; exists {
		return func() {}, fmt.Errorf("cannot run %q: pull request %d is currently locked by %q.\n"+
			"Wait until the previous command is complete and try again", cmdName, pullNum, currentLock)
	}
	if currentLock, exists := d.findCommandLock(repoFullName, pullNum, command.Plan); exists {
		return func() {}, fmt.Errorf("cannot run %q: pull request %d is currently locked by %q.\n"+
			"Wait until the previous command is complete and try again", cmdName, pullNum, currentLock)
	}
	d.locks[pullKey] = cmdName
	return func() {
		d.unlockPull(repoFullName, pullNum)
	}, nil
}

func (d *DefaultWorkingDirLocker) HasCommandLock(repoFullName string, pullNum int, cmdName command.Name) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, exists := d.findCommandLock(repoFullName, pullNum, cmdName)
	return exists
}

func (d *DefaultWorkingDirLocker) findCommandLock(repoFullName string, pullNum int, cmdName command.Name) (command.Name, bool) {
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	for key, currentLock := range d.locks {
		if strings.HasPrefix(key, prefix) && currentLock == cmdName {
			return currentLock, true
		}
	}
	return cmdName, false
}

// UnlockByPull unlocks all workspaces for a specific pull request
func (d *DefaultWorkingDirLocker) UnlockByPull(repoFullName string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Find and remove all locks for this pull request
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	for key := range d.locks {
		if strings.HasPrefix(key, prefix) {
			delete(d.locks, key)
		}
	}
}

// Unlock unlocks the workspace for this pull.
func (d *DefaultWorkingDirLocker) unlock(repoFullName string, pullNum int, workspace string, path string, projectName string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	workspaceKey := d.workspaceKey(repoFullName, pullNum, workspace, path, projectName)
	delete(d.locks, workspaceKey)
}

func (d *DefaultWorkingDirLocker) unlockPull(repoFullName string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.locks, d.pullKey(repoFullName, pullNum))
}

func (d *DefaultWorkingDirLocker) workspaceKey(repo string, pull int, workspace string, path string, projectName string) string {
	return strings.TrimRight(fmt.Sprintf("%s%s/%s/%s", d.pullLockPrefix(repo, pull), workspace, path, projectName), "/")
}

func (d *DefaultWorkingDirLocker) pullKey(repo string, pull int) string {
	return d.pullLockPrefix(repo, pull) + ".pull"
}

func (d *DefaultWorkingDirLocker) pullLockPrefix(repo string, pull int) string {
	return fmt.Sprintf("%d:%s/%d/", len(repo), repo, pull)
}
