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

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_atlantis_workspace_locker.go AtlantisWorkspaceLocker

// AtlantisWorkspaceLocker is used to prevent multiple commands from executing
// at the same time for a single repo, pull, and workspace. We need to prevent
// this from happening because a specific repo/pull/workspace has a single workspace
// on disk and we haven't written Atlantis (yet) to handle concurrent execution
// within this workspace.
// This locker is called AtlantisWorkspaceLocker to differentiate it from the
// Terraform concept of workspaces, not directories on disk managed by Atlantis.
type AtlantisWorkspaceLocker interface {
	// TryLock tries to acquire a lock for this repo, workspace and pull.
	TryLock(repoFullName string, workspace string, pullNum int) bool
	// TryLock2 tries to acquire a lock for this repo, workspace and pull.
	TryLock2(repoFullName string, workspace string, pullNum int) (func(), error)
	// Unlock deletes the lock for this repo, workspace and pull. If there was no
	// lock it will do nothing.
	Unlock(repoFullName, workspace string, pullNum int)
}

// DefaultAtlantisWorkspaceLocker implements AtlantisWorkspaceLocker.
type DefaultAtlantisWorkspaceLocker struct {
	mutex sync.Mutex
	locks map[string]interface{}
}

// NewDefaultAtlantisWorkspaceLocker is a constructor.
func NewDefaultAtlantisWorkspaceLocker() *DefaultAtlantisWorkspaceLocker {
	return &DefaultAtlantisWorkspaceLocker{
		locks: make(map[string]interface{}),
	}
}

func (d *DefaultAtlantisWorkspaceLocker) TryLock2(repoFullName string, workspace string, pullNum int) (func(), error) {
	if !d.TryLock(repoFullName, workspace, pullNum) {
		return func() {}, fmt.Errorf("the %s workspace is currently locked by another"+
			" command that is running for this pull request–"+
			"wait until the previous command is complete and try again", workspace)
	}
	return func() {
		d.Unlock(repoFullName, workspace, pullNum)
	}, nil
}

// TryLock returns true if a lock is acquired for this repo, pull and workspace and
// false otherwise.
func (d *DefaultAtlantisWorkspaceLocker) TryLock(repoFullName string, workspace string, pullNum int) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	key := d.key(repoFullName, workspace, pullNum)
	if _, ok := d.locks[key]; !ok {
		d.locks[key] = true
		return true
	}
	return false
}

// Unlock unlocks the repo, pull and workspace.
func (d *DefaultAtlantisWorkspaceLocker) Unlock(repoFullName, workspace string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.locks, d.key(repoFullName, workspace, pullNum))
}

func (d *DefaultAtlantisWorkspaceLocker) key(repo string, workspace string, pull int) string {
	return fmt.Sprintf("%s/%s/%d", repo, workspace, pull)
}
