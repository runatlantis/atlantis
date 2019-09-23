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
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
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
	TryLock(ctx *models.CommandContext, workspace string, tryCancel bool) (func(), error)
	// TryLockPull tries to acquire a lock for all the workspaces in this repo
	// and pull.
	// It returns a function that should be used to unlock the workspace and
	// an error if the workspace is already locked. The error is expected to
	// be printed to the pull request.
	TryLockPull(ctx *models.CommandContext) (func(), error)
}

// DefaultWorkingDirLocker implements WorkingDirLocker.
type DefaultWorkingDirLocker struct {
	// mutex prevents against multiple threads calling functions on this struct
	// concurrently. It's only used for entry/exit to each function.
	mutex sync.Mutex
	// locks is a list of the keys that are locked. We then use prefix
	// matching to determine if something is locked. It's naive but that's okay
	// because there won't be many locks at one time.
	locks map[string]*models.CommandContext
}

// NewDefaultWorkingDirLocker is a constructor.
func NewDefaultWorkingDirLocker() *DefaultWorkingDirLocker {
	return &DefaultWorkingDirLocker{
		locks: map[string]*models.CommandContext{},
	}
}

func (d *DefaultWorkingDirLocker) TryLockPull(ctx *models.CommandContext) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	pullKey := d.pullKey(ctx.BaseRepo.FullName, ctx.Pull.Num)
	for l := range d.locks {
		if l == pullKey || strings.HasPrefix(l, pullKey+"/") {
			return func() {}, fmt.Errorf("the Atlantis working dir is currently locked by another" +
				" command that is running for this pull request–" +
				"wait until the previous command is complete and try again")
		}
	}
	d.locks[pullKey] = ctx
	return func() {
		d.UnlockPull(ctx.BaseRepo.FullName, ctx.Pull.Num)
	}, nil
}

func (d *DefaultWorkingDirLocker) TryLock(ctx *models.CommandContext, workspace string, tryCancel bool) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	pullKey := d.pullKey(ctx.BaseRepo.FullName, ctx.Pull.Num)
	workspaceKey := d.workspaceKey(ctx.BaseRepo.FullName, ctx.Pull.Num, workspace)

	pullLockCtx, pullLockOK := d.locks[pullKey]
	workspaceLockCtx, workspaceLockOK := d.locks[workspaceKey]

	if tryCancel {
		var triedCancel bool
		if pullLockOK && pullLockCtx != nil {
			pullLockCtx.Cancel()
			triedCancel = true
		}
		if workspaceLockOK && workspaceLockCtx != nil {
			workspaceLockCtx.Cancel()
			triedCancel = true
		}

		if triedCancel {
			time.Sleep(5 * time.Second)
			pullLockCtx, pullLockOK = d.locks[pullKey]
			workspaceLockCtx, workspaceLockOK = d.locks[workspaceKey]
		}
	}

	if pullLockOK || workspaceLockOK {
		return func() {}, fmt.Errorf("the %s workspace is currently locked by another"+
			" command that is running for this pull request–"+
			"wait until the previous command is complete and try again", workspace)
	}
	d.locks[workspaceKey] = ctx
	return func() {
		d.unlock(ctx.BaseRepo.FullName, ctx.Pull.Num, workspace)
	}, nil
}

// Unlock unlocks the workspace for this pull.
func (d *DefaultWorkingDirLocker) unlock(repoFullName string, pullNum int, workspace string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	workspaceKey := d.workspaceKey(repoFullName, pullNum, workspace)
	d.removeLock(workspaceKey)
}

// Unlock unlocks all workspaces for this pull.
func (d *DefaultWorkingDirLocker) UnlockPull(repoFullName string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	pullKey := d.pullKey(repoFullName, pullNum)
	d.removeLock(pullKey)
}

func (d *DefaultWorkingDirLocker) removeLock(key string) {
	delete(d.locks, key)
}

func (d *DefaultWorkingDirLocker) workspaceKey(repo string, pull int, workspace string) string {
	return fmt.Sprintf("%s/%s", d.pullKey(repo, pull), workspace)
}

func (d *DefaultWorkingDirLocker) pullKey(repo string, pull int) string {
	return fmt.Sprintf("%s/%d", repo, pull)
}
