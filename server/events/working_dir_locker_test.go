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

package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/testing"
)

var repo = "repo/owner"
var workspace = "default"
var path = "."
var projectName = "testProjectName"
var cmd = command.Plan

func TestTryLock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	// The first lock should succeed.
	unlockFn, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)

	// Now another lock for the same repo, workspace, projectName and pull should fail
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Apply)
	ErrEquals(t, "cannot run \"apply\": the default workspace at path . is currently locked for this pull request by \"plan\".\n"+
		"Wait until the previous command is complete and try again", err)

	// Unlock should work.
	unlockFn()
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
}

func TestTryLockSameCommand(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	// The first lock should succeed.
	unlockFn, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Import)
	Ok(t, err)

	// Now another lock for the same repo, workspace, projectName and pull should fail
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Import)
	ErrEquals(t, "cannot run \"import\": the default workspace at path . is currently locked for this pull request by \"import\".\n"+
		"Wait until the previous command is complete and try again", err)

	// Unlock should work.
	unlockFn()
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
}

func TestTryLockDifferentWorkspaces(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for the same repo and pull but different workspace should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Assert(t, err != nil, "exp err")
	_, err = locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd)
	Assert(t, err != nil, "exp err")
}

func TestTryLockDifferentRepo(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different repo but the same workspace and pull should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	newRepo := "owner/newrepo"
	_, err = locker.TryLock(newRepo, 1, workspace, path, projectName, cmd)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(newRepo, 1, workspace, path, projectName, cmd)
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentPulls(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different pull but the same repo, workspace, projectName should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	newPull := 2
	_, err = locker.TryLock(repo, newPull, workspace, path, projectName, cmd)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, newPull, workspace, path, projectName, cmd)
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentPaths(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different path but the same repo, pull, projectName and workspace should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	newPath := "new-path"
	_, err = locker.TryLock(repo, 1, workspace, newPath, projectName, cmd)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, 1, workspace, newPath, projectName, cmd)
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentProjectNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different projectName but the same repo, pull, path and workspace should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	newProjectName := "new-project"
	_, err = locker.TryLock(repo, 1, workspace, path, newProjectName, cmd)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, 1, workspace, path, newProjectName, cmd)
	ErrContains(t, "currently locked", err)
}

func TestUnlock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("unlocking should work")
	unlockFn, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	unlockFn()
	_, err = locker.TryLock(repo, 1, workspace, "", projectName, cmd)
	Ok(t, err)
}

func TestUnlockDifferentWorkspaces(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different workspaces")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err1)
	unlockFn2, err2 := locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd)
	Ok(t, err)
}

func TestUnlockDifferentRepos(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different repos")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err1)
	newRepo := "owner/newrepo"
	unlockFn2, err2 := locker.TryLock(newRepo, 1, workspace, path, projectName, cmd)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	_, err = locker.TryLock(newRepo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
}

func TestUnlockDifferentPulls(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different pulls")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err1)
	newPull := 2
	unlockFn2, err2 := locker.TryLock(repo, newPull, workspace, path, projectName, cmd)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	_, err = locker.TryLock(repo, newPull, workspace, path, projectName, cmd)
	Ok(t, err)
}

func TestUnlockDifferentProjectNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different projects")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err1)
	newProjectName := "new-project"
	unlockFn2, err2 := locker.TryLock(repo, 1, workspace, path, newProjectName, cmd)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, workspace, path, newProjectName, cmd)
	Ok(t, err)
}

func TestIsLockedByPull_NoLocks(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("IsLockedByPull returns false when no locks are held")
	Assert(t, !locker.IsLockedByPull(repo, 1), "exp not locked")
}

func TestIsLockedByPull_LockedPull(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("IsLockedByPull returns true when a workspace is locked for the given pull")
	unlock, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	Assert(t, locker.IsLockedByPull(repo, 1), "exp locked")

	unlock()
	Assert(t, !locker.IsLockedByPull(repo, 1), "exp not locked after unlock")
}

func TestIsLockedByPull_DifferentPull(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("IsLockedByPull returns false for a different pull request")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	Assert(t, !locker.IsLockedByPull(repo, 2), "exp not locked for pull 2")
	Assert(t, locker.IsLockedByPull(repo, 1), "exp locked for pull 1")
}

func TestIsLockedByPull_DifferentRepo(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("IsLockedByPull returns false for a different repo")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	Assert(t, !locker.IsLockedByPull("other/repo", 1), "exp not locked for other repo")
	Assert(t, locker.IsLockedByPull(repo, 1), "exp locked for original repo")
}

func TestIsLockedByPull_MultipleWorkspaces(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("IsLockedByPull returns true if any workspace is locked")
	unlock1, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	unlock2, err := locker.TryLock(repo, 1, "other-workspace", path, projectName, cmd)
	Ok(t, err)

	Assert(t, locker.IsLockedByPull(repo, 1), "exp locked")

	unlock1()
	Assert(t, locker.IsLockedByPull(repo, 1), "exp still locked after first unlock")

	unlock2()
	Assert(t, !locker.IsLockedByPull(repo, 1), "exp not locked after both unlocks")
}

func TestIsLockedByPull_UnlockByPull(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("IsLockedByPull returns false after UnlockByPull clears all locks")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd)
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, "other-workspace", path, projectName, cmd)
	Ok(t, err)

	Assert(t, locker.IsLockedByPull(repo, 1), "exp locked before unlock")
	locker.UnlockByPull(repo, 1)
	Assert(t, !locker.IsLockedByPull(repo, 1), "exp not locked after UnlockByPull")
}
