// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"strings"
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
	unlockFn, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)

	// Now another lock for the same repo, workspace, projectName and pull should fail
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Apply, events.WorkingDirLockMetadata{})
	ErrEquals(t, "cannot run \"apply\": the default workspace at path . is currently locked for this pull request by \"plan\".\n"+
		"Wait until the previous command is complete and try again", err)

	// Unlock should work.
	unlockFn()
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestTryLockIncludesCommitMetadata(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	const commitURL = "https://github.com/owner/repo/commit/0123456789abcdef0123456789abcdef01234567"

	tests := []struct {
		name     string
		metadata events.WorkingDirLockMetadata
		want     string
	}{
		{name: "sha", metadata: events.WorkingDirLockMetadata{HeadCommit: sha}, want: "by \"plan\" for commit 0123456."},
		{name: "url remains structured", metadata: events.WorkingDirLockMetadata{HeadCommit: sha, CommitURL: commitURL}, want: "by \"plan\" for commit 0123456."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker := events.NewDefaultWorkingDirLocker()
			_, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Plan, tt.metadata)
			Ok(t, err)
			_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Apply, events.WorkingDirLockMetadata{})
			ErrContains(t, tt.want, err)
		})
	}
}

func TestTryLockPullIncludesCommitMetadata(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	const commitURL = "https://github.com/owner/repo/commit/0123456789abcdef0123456789abcdef01234567"
	locker := events.NewDefaultWorkingDirLocker()
	_, err := locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: sha, CommitURL: commitURL})
	Ok(t, err)
	_, err = locker.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "by \"plan\" for commit 0123456.", err)
	Assert(t, !strings.Contains(err.Error(), commitURL), "expected commit URL to remain out of plain error: %s", err)
}

func TestTryLockSameCommand(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	// The first lock should succeed.
	unlockFn, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Import, events.WorkingDirLockMetadata{})
	Ok(t, err)

	// Now another lock for the same repo, workspace, projectName and pull should fail
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Import, events.WorkingDirLockMetadata{})
	ErrEquals(t, "cannot run \"import\": the default workspace at path . is currently locked for this pull request by \"import\".\n"+
		"Wait until the previous command is complete and try again", err)

	// Unlock should work.
	unlockFn()
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestTryLockDifferentWorkspaces(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for the same repo and pull but different workspace should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Assert(t, err != nil, "exp err")
	_, err = locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd, events.WorkingDirLockMetadata{})
	Assert(t, err != nil, "exp err")
}

func TestTryLockDifferentRepo(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different repo but the same workspace and pull should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	newRepo := "owner/newrepo"
	_, err = locker.TryLock(newRepo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(newRepo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentPulls(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different pull but the same repo, workspace, projectName should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	newPull := 2
	_, err = locker.TryLock(repo, newPull, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, newPull, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentPaths(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different path but the same repo, pull, projectName and workspace should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	newPath := "new-path"
	_, err = locker.TryLock(repo, 1, workspace, newPath, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, 1, workspace, newPath, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentProjectNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("a lock for a different projectName but the same repo, pull, path and workspace should succeed")
	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	newProjectName := "new-project"
	_, err = locker.TryLock(repo, 1, workspace, path, newProjectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, 1, workspace, path, newProjectName, cmd, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
}

func TestUnlock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	t.Log("unlocking should work")
	unlockFn, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	unlockFn()
	_, err = locker.TryLock(repo, 1, workspace, "", projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestUnlockDifferentWorkspaces(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different workspaces")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err1)
	unlockFn2, err2 := locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, "new-workspace", path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestUnlockDifferentRepos(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different repos")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err1)
	newRepo := "owner/newrepo"
	unlockFn2, err2 := locker.TryLock(newRepo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLock(newRepo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestUnlockDifferentPulls(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different pulls")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err1)
	newPull := 2
	unlockFn2, err2 := locker.TryLock(repo, newPull, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLock(repo, newPull, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestUnlockDifferentProjectNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	t.Log("unlocking should work for different projects")
	unlockFn1, err1 := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err1)
	newProjectName := "new-project"
	unlockFn2, err2 := locker.TryLock(repo, 1, workspace, path, newProjectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, 1, workspace, path, projectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, workspace, path, newProjectName, cmd, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestWorkingDirLocker_HasCommandLockDoesNotCrossMatchSlashNestedRepoNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull("group/repo/1", 2, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlock()

	Assert(t, locker.HasCommandLock("group/repo/1", 2, command.Plan), "expected exact nested repo lock")
	Assert(t, !locker.HasCommandLock("group/repo", 1, command.Plan), "did not expect parent repo/pull prefix to match nested repo lock")
}

func TestWorkingDirLocker_TryLockPullDoesNotCrossMatchSlashNestedRepoNames(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull("group/repo/1", 2, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlock()

	unlockParent, err := locker.TryLockPull("group/repo", 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlockParent()
}

func TestWorkingDirLocker_PullLockExactRepoAndPullIsolation(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull("group/repo/12", 3, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlock()

	Assert(t, locker.HasCommandLock("group/repo/12", 3, command.Plan), "expected exact repo/pull lock")
	Assert(t, !locker.HasCommandLock("group/repo", 12, command.Plan), "did not expect pull prefix to match nested repo path")
	unlockOther, err := locker.TryLockPull("group/repo", 12, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlockOther()
}

func TestWorkingDirLocker_ProjectLocksStillUseExactWorkspaceDirProjectIdentity(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLock("group/repo/1", 2, "default", "dir", "proj", command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlock()

	Assert(t, locker.HasCommandLock("group/repo/1", 2, command.Plan), "expected exact project lock to count as command lock")
	Assert(t, !locker.HasCommandLock("group/repo", 1, command.Plan), "did not expect nested repo project lock to match parent repo/pull")
	_, err = locker.TryLock("group/repo/1", 2, "default", "dir", "proj", command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock("group/repo/1", 2, "default", "dir", "other", command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestTryLockPullSharedAllowsOtherSharedHolders(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlockFirst, err := locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	unlockSecond, err := locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)

	unlockFirst()
	_, err = locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked by \"apply\"", err)

	unlockSecond()
	_, err = locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestTryLockPullSharedBlocksExclusivePullLock(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{HeadCommit: sha})
	Ok(t, err)
	defer unlock()

	_, err = locker.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrEquals(t, "cannot run \"apply\": pull request 1 is currently locked by \"apply\" for commit 0123456.\n"+
		"Wait until the previous command is complete and try again", err)
	_, err = locker.TryLockPull(repo, 2, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestTryLockPullBlocksSharedPullLock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)

	_, err = locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "pull request 1 is currently locked by \"plan\"", err)

	unlock()
	_, err = locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestTryLockPullSharedBlocksWhenPlanInFlight(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlock()

	_, err = locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "pull request 1 is currently locked by \"plan\"", err)
}

func TestTryLockProjectApply(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlock, err := locker.TryLockProjectApply(repo, 1, "daily-1", path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)

	_, err = locker.TryLockProjectApply(repo, 1, "daily-1", path, projectName, events.WorkingDirLockMetadata{})
	ErrEquals(t, "cannot run \"apply\": the daily-1 workspace at path . is currently locked for this pull request by \"apply\".\n"+
		"Wait until the previous command is complete and try again", err)
	_, err = locker.TryLockProjectApply(repo, 1, "daily-2", path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)

	unlock()
	_, err = locker.TryLockProjectApply(repo, 1, "daily-1", path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

// The project runner takes its own working dir lock while a project apply
// claim is held, so the two must not conflict.
func TestTryLockProjectApplyDoesNotBlockWorkingDirLock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	unlockClaim, err := locker.TryLockProjectApply(repo, 1, workspace, path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlockClaim()

	unlock, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	defer unlock()
	Assert(t, !locker.HasCommandLock(repo, 1, command.Plan), "did not expect an apply claim to look like a plan")
}

func TestUnlockByPullReleasesSharedPullLocksAndApplyClaims(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	_, err := locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLockProjectApply(repo, 1, workspace, path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)

	locker.UnlockByPull(repo, 1)

	_, err = locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLockProjectApply(repo, 1, workspace, path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

// atlantis cancel releases every lock on the pull request while running
// commands continue. Their later releases must not drop newer holders' locks.
func TestStaleUnlockAfterUnlockByPullKeepsNewerHolders(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	staleUnlockShared, err := locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	staleUnlockClaim, err := locker.TryLockProjectApply(repo, 1, workspace, path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)
	locker.UnlockByPull(repo, 1)

	_, err = locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLockProjectApply(repo, 1, workspace, path, projectName, events.WorkingDirLockMetadata{})
	Ok(t, err)
	staleUnlockShared()
	staleUnlockClaim()

	_, err = locker.TryLockProjectApply(repo, 1, workspace, path, projectName, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked by \"apply\"", err)
}

func TestStaleUnlockAfterUnlockByPullKeepsNewerExclusiveAndWorkspaceLocks(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()

	staleUnlockPull, err := locker.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	staleUnlockWorkspace, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
	locker.UnlockByPull(repo, 1)

	_, err = locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Plan, events.WorkingDirLockMetadata{})
	Ok(t, err)
	staleUnlockPull()
	staleUnlockWorkspace()

	_, err = locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked by \"plan\"", err)
	_, err = locker.TryLock(repo, 1, workspace, path, projectName, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "currently locked for this pull request by \"plan\"", err)
}

// Workspace names such as ".shared-pull" and ".apply-claim" are legal, so
// workspace locks must not be able to reach the keys of pull-level locks or
// apply claims.
func TestWorkspaceLocksDoNotCollideWithReservedLockKeys(t *testing.T) {
	metadata := events.WorkingDirLockMetadata{}
	locker := events.NewDefaultWorkingDirLocker()

	// The first shared lock on a new locker has owner 1, so a key built from
	// the workspace ".shared-pull" and path "1" would spell it.
	unlockShared, err := locker.TryLockPullShared(repo, 1, command.Apply, metadata)
	Ok(t, err)
	unlockWorkspace, err := locker.TryLock(repo, 1, ".shared-pull", "1", "", command.Apply, metadata)
	Ok(t, err)
	unlockWorkspace()
	_, err = locker.TryLockPull(repo, 1, command.Plan, metadata)
	ErrContains(t, "currently locked by \"apply\"", err)
	unlockShared()

	_, err = locker.TryLock(repo, 1, ".apply-claim", workspace+"/"+path, projectName, command.Apply, metadata)
	Ok(t, err)
	_, err = locker.TryLockProjectApply(repo, 1, workspace, path, projectName, metadata)
	Ok(t, err)
}
