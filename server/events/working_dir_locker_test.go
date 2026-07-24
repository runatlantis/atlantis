// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
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

func TestWorkingDirLocker_HasCommandLockDoesNotCrossMatchSlashNestedRepoNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull("group/repo/1", 2, command.Plan)
	Ok(t, err)
	defer unlock()

	Assert(t, locker.HasCommandLock("group/repo/1", 2, command.Plan), "expected exact nested repo lock")
	Assert(t, !locker.HasCommandLock("group/repo", 1, command.Plan), "did not expect parent repo/pull prefix to match nested repo lock")
}

func TestWorkingDirLocker_TryLockPullDoesNotCrossMatchSlashNestedRepoNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull("group/repo/1", 2, command.Plan)
	Ok(t, err)
	defer unlock()

	unlockParent, err := locker.TryLockPull("group/repo", 1, command.Apply)
	Ok(t, err)
	defer unlockParent()
}

func TestWorkingDirLocker_PullLockExactRepoAndPullIsolation(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull("group/repo/12", 3, command.Plan)
	Ok(t, err)
	defer unlock()

	Assert(t, locker.HasCommandLock("group/repo/12", 3, command.Plan), "expected exact repo/pull lock")
	Assert(t, !locker.HasCommandLock("group/repo", 12, command.Plan), "did not expect pull prefix to match nested repo path")
	unlockOther, err := locker.TryLockPull("group/repo", 12, command.Apply)
	Ok(t, err)
	defer unlockOther()
}

func TestWorkingDirLocker_ProjectLocksStillUseExactWorkspaceDirProjectIdentity(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLock("group/repo/1", 2, "default", "dir", "proj", command.Plan)
	Ok(t, err)
	defer unlock()

	Assert(t, locker.HasCommandLock("group/repo/1", 2, command.Plan), "expected exact project lock to count as command lock")
	Assert(t, !locker.HasCommandLock("group/repo", 1, command.Plan), "did not expect nested repo project lock to match parent repo/pull")
	_, err = locker.TryLock("group/repo/1", 2, "default", "dir", "proj", command.Apply)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock("group/repo/1", 2, "default", "dir", "other", command.Apply)
	Ok(t, err)
}
