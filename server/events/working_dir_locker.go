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
//
package events

import (
	"fmt"
	"sync"
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_working_dir_locker.go WorkingDirLocker

// WorkingDirLocker is used to prevent multiple commands from executing
// at the same time for a single repo, pull, and workspace. We need to prevent
// this from happening because a specific repo/pull/workspace has a single workspace
// on disk and we haven't written Atlantis (yet) to handle concurrent execution
// within this workspace.
type WorkingDirLocker interface {
	// TryLock tries to acquire a lock for this repo, workspace and pull.
	// It returns a function that should be used to unlock the workspace and
	// an error if the workspace is already locked. The error is expected to
	// be printed to the pull request.
	TryLock(repoFullName string, workspace string, pullNum int) (func(), error)
	TryLockPull(repoFullName string, pullNum int) (func(), error)
	// Unlock deletes the lock for this repo, workspace and pull. If there was no
	// lock it will do nothing.
	Unlock(repoFullName, workspace string, pullNum int)
}

// DefaultWorkingDirLocker implements WorkingDirLocker.
type DefaultWorkingDirLocker struct {
	mutex sync.Mutex
	locks map[string]interface{}
}

// NewDefaultWorkingDirLocker is a constructor.
func NewDefaultWorkingDirLocker() *DefaultWorkingDirLocker {
	return &DefaultWorkingDirLocker{
		locks: make(map[string]interface{}),
	}
}

func (d *DefaultWorkingDirLocker) TryLockPull(repoFullName string, pullNum int) (func(), error) {
	// todo: implement
	return func() {}, nil
}

func (d *DefaultWorkingDirLocker) TryLock(repoFullName string, workspace string, pullNum int) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	key := d.key(repoFullName, workspace, pullNum)
	_, exists := d.locks[key]
	if exists {
		return func() {}, fmt.Errorf("the %s workspace is currently locked by another"+
			" command that is running for this pull request–"+
			"wait until the previous command is complete and try again", workspace)
	}
	d.locks[key] = true
	return func() {
		d.Unlock(repoFullName, workspace, pullNum)
	}, nil
}

// Unlock unlocks the repo, pull and workspace.
func (d *DefaultWorkingDirLocker) Unlock(repoFullName, workspace string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.locks, d.key(repoFullName, workspace, pullNum))
}

func (d *DefaultWorkingDirLocker) key(repo string, workspace string, pull int) string {
	return fmt.Sprintf("%s/%s/%d", repo, workspace, pull)
}
