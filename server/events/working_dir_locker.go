// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"strings"
	"sync"

	"github.com/runatlantis/atlantis/server/events/command"
)

//go:generate pegomock generate --package mocks -o mocks/mock_working_dir_locker.go WorkingDirLocker

// WorkingDirLocker is used to prevent multiple commands from executing
// at the same time for a single repo, pull, and workspace. We need to prevent
// this from happening because a specific repo/pull/workspace has a single workspace
// on disk and we haven't written Atlantis (yet) to handle concurrent execution
// within this workspace.
type WorkingDirLocker interface {
	// TryLock tries to acquire a lock for this repo, pull, workspace, and path.
	// It returns a function that should be used to unlock the workspace and
	// an error if the workspace is already locked. The error is expected to
	// be printed to the pull request.
	TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name) (func(), error)
	// UnlockByPull unlocks all workspaces for a specific pull request
	UnlockByPull(repoFullName string, pullNum int)
	// IsLockedByPull returns true if any workspace is currently locked for
	// the given pull request, indicating an operation is in progress.
	IsLockedByPull(repoFullName string, pullNum int) bool
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

// IsLockedByPull returns true if any workspace is currently locked for the
// given pull request, meaning an operation is in progress.
func (d *DefaultWorkingDirLocker) IsLockedByPull(repoFullName string, pullNum int) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	prefix := fmt.Sprintf("%s/%d/", repoFullName, pullNum)
	for key := range d.locks {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// UnlockByPull unlocks all workspaces for a specific pull request
func (d *DefaultWorkingDirLocker) UnlockByPull(repoFullName string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Find and remove all locks for this pull request
	prefix := fmt.Sprintf("%s/%d/", repoFullName, pullNum)
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

func (d *DefaultWorkingDirLocker) workspaceKey(repo string, pull int, workspace string, path string, projectName string) string {
	return strings.TrimRight(fmt.Sprintf("%s/%d/%s/%s/%s", repo, pull, workspace, path, projectName), "/")
}
