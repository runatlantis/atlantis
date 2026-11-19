// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"strings"
	"testing"
	"time"

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

const (
	oldSHA   = "1111111111111111111111111111111111111111"
	newSHA   = "2222222222222222222222222222222222222222"
	newerSHA = "3333333333333333333333333333333333333333"
)

type lockPullResult struct {
	unlock func()
	err    error
}

// lockPullAfterStalePlanAsync starts LockPullAfterStalePlan in a goroutine
// and returns once the call is either waiting or finished.
func lockPullAfterStalePlanAsync(locker *events.DefaultWorkingDirLocker, headCommit string) (<-chan lockPullResult, <-chan events.WorkingDirLockMetadata) {
	result := make(chan lockPullResult, 1)
	waiting := make(chan events.WorkingDirLockMetadata, 1)
	go func() {
		unlock, err := locker.LockPullAfterStalePlan(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: headCommit}, func(holder events.WorkingDirLockMetadata) {
			waiting <- holder
		})
		result <- lockPullResult{unlock: unlock.Unlock, err: err}
	}()
	return result, waiting
}

func awaitWaiting(t *testing.T, waiting <-chan events.WorkingDirLockMetadata, result <-chan lockPullResult) events.WorkingDirLockMetadata {
	t.Helper()
	select {
	case holder := <-waiting:
		return holder
	case r := <-result:
		t.Fatalf("expected LockPullAfterStalePlan to wait, it returned err=%v", r.err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for LockPullAfterStalePlan to start waiting")
	}
	return events.WorkingDirLockMetadata{}
}

func awaitResult(t *testing.T, result <-chan lockPullResult) lockPullResult {
	t.Helper()
	select {
	case r := <-result:
		return r
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for LockPullAfterStalePlan to return")
	}
	return lockPullResult{}
}

func TestLockPullAfterStalePlan_LocksImmediatelyWhenFree(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	unlock, err := locker.LockPullAfterStalePlan(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: newSHA}, func(events.WorkingDirLockMetadata) {
		t.Fatal("did not expect to wait")
	})
	Ok(t, err)
	Assert(t, locker.HasCommandLock(repo, 1, command.Plan), "expected plan lock")
	unlock.Unlock()
	Assert(t, !locker.HasCommandLock(repo, 1, command.Plan), "expected plan lock to be released")
}

func TestLockPullAfterStalePlan_WaitsForPlanOfOlderCommit(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	unlockOld, err := locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: oldSHA})
	Ok(t, err)

	result, waiting := lockPullAfterStalePlanAsync(locker, newSHA)
	holder := awaitWaiting(t, waiting, result)
	Equals(t, oldSHA, holder.HeadCommit)

	unlockOld()
	r := awaitResult(t, result)
	Ok(t, r.err)

	// The new commit now owns the pull lock.
	_, err = locker.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "for commit 2222222", err)
	r.unlock()
	_, err = locker.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	Ok(t, err)
}

func TestLockPullAfterStalePlan_WaitsForProjectPlanOfOlderCommit(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	unlockOld, err := locker.TryLock(repo, 1, workspace, path, projectName, command.Plan, events.WorkingDirLockMetadata{HeadCommit: oldSHA})
	Ok(t, err)

	result, waiting := lockPullAfterStalePlanAsync(locker, newSHA)
	awaitWaiting(t, waiting, result)

	unlockOld()
	r := awaitResult(t, result)
	Ok(t, r.err)
	r.unlock()
}

func TestLockPullAfterStalePlan_DoesNotWaitForOtherConflicts(t *testing.T) {
	cases := []struct {
		name   string
		holder command.Name
		sha    string
	}{
		{name: "plan for the same commit", holder: command.Plan, sha: newSHA},
		{name: "plan without commit metadata", holder: command.Plan, sha: ""},
		{name: "apply for an older commit", holder: command.Apply, sha: oldSHA},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			locker := events.NewDefaultWorkingDirLocker()
			unlock, err := locker.TryLockPull(repo, 1, c.holder, events.WorkingDirLockMetadata{HeadCommit: c.sha})
			Ok(t, err)
			defer unlock()

			_, err = locker.LockPullAfterStalePlan(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: newSHA}, func(events.WorkingDirLockMetadata) {
				t.Fatal("did not expect to wait")
			})
			ErrContains(t, "is currently locked by", err)
		})
	}
}

func TestLockPullAfterStalePlan_NewerCommitSupersedesWaiter(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	unlockOld, err := locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: oldSHA})
	Ok(t, err)

	first, firstWaiting := lockPullAfterStalePlanAsync(locker, newSHA)
	awaitWaiting(t, firstWaiting, first)
	second, secondWaiting := lockPullAfterStalePlanAsync(locker, newerSHA)
	awaitWaiting(t, secondWaiting, second)

	r := awaitResult(t, first)
	Assert(t, errors.Is(r.err, events.ErrPullLockSuperseded), "expected superseded, got %v", r.err)
	Assert(t, locker.HasCommandLock(repo, 1, command.Plan), "old plan should still hold the lock")

	unlockOld()
	r = awaitResult(t, second)
	Ok(t, r.err)
	_, err = locker.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{})
	ErrContains(t, "for commit 3333333", err)
	r.unlock()
}

func TestLockPullAfterStalePlan_UnlockByPullCancelsWaiter(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	_, err := locker.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: oldSHA})
	Ok(t, err)

	result, waiting := lockPullAfterStalePlanAsync(locker, newSHA)
	awaitWaiting(t, waiting, result)

	locker.UnlockByPull(repo, 1)
	r := awaitResult(t, result)
	Assert(t, errors.Is(r.err, events.ErrPullLockCancelled), "expected cancelled, got %v", r.err)
	Assert(t, !locker.HasCommandLock(repo, 1, command.Plan), "expected no plan lock after cancellation")
}

func TestTryLockPullShared_TargetedAppliesShareThePullLock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	metadata := events.WorkingDirLockMetadata{HeadCommit: newSHA}

	unlockFirst, err := locker.TryLockPullShared(repo, 1, command.Apply, metadata)
	Ok(t, err)
	unlockSecond, err := locker.TryLockPullShared(repo, 1, command.Apply, metadata)
	Ok(t, err)

	// Plans and full applies still have to wait for every targeted apply.
	_, err = locker.TryLockPull(repo, 1, command.Plan, metadata)
	ErrContains(t, `currently locked by "apply"`, err)
	_, err = locker.TryLockPull(repo, 1, command.Apply, metadata)
	ErrContains(t, `currently locked by "apply"`, err)

	unlockFirst()
	unlockFirst() // releasing twice must not drop the other holder
	_, err = locker.TryLockPull(repo, 1, command.Plan, metadata)
	ErrContains(t, `currently locked by "apply"`, err)

	unlockSecond()
	unlockPlan, err := locker.TryLockPull(repo, 1, command.Plan, metadata)
	Ok(t, err)
	unlockPlan()
}

func TestTryLockPullShared_ConflictsWithOtherLocks(t *testing.T) {
	cases := []struct {
		name string
		hold func(*events.DefaultWorkingDirLocker) (func(), error)
	}{
		{
			name: "full apply",
			hold: func(l *events.DefaultWorkingDirLocker) (func(), error) {
				return l.TryLockPull(repo, 1, command.Apply, events.WorkingDirLockMetadata{HeadCommit: newSHA})
			},
		},
		{
			name: "plan",
			hold: func(l *events.DefaultWorkingDirLocker) (func(), error) {
				return l.TryLockPull(repo, 1, command.Plan, events.WorkingDirLockMetadata{HeadCommit: newSHA})
			},
		},
		{
			name: "project plan",
			hold: func(l *events.DefaultWorkingDirLocker) (func(), error) {
				return l.TryLock(repo, 1, workspace, path, projectName, command.Plan, events.WorkingDirLockMetadata{HeadCommit: newSHA})
			},
		},
		{
			name: "targeted apply for another commit",
			hold: func(l *events.DefaultWorkingDirLocker) (func(), error) {
				return l.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{HeadCommit: oldSHA})
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			locker := events.NewDefaultWorkingDirLocker()
			unlock, err := c.hold(locker)
			Ok(t, err)
			defer unlock()

			_, err = locker.TryLockPullShared(repo, 1, command.Apply, events.WorkingDirLockMetadata{HeadCommit: newSHA})
			ErrContains(t, "is currently locked by", err)
		})
	}
}

func TestTryLockPullShared_StaleUnlockDoesNotReleaseNewLock(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	metadata := events.WorkingDirLockMetadata{HeadCommit: newSHA}
	staleUnlock, err := locker.TryLockPullShared(repo, 1, command.Apply, metadata)
	Ok(t, err)

	locker.UnlockByPull(repo, 1)
	unlock, err := locker.TryLockPullShared(repo, 1, command.Apply, metadata)
	Ok(t, err)
	defer unlock()

	staleUnlock()
	_, err = locker.TryLockPull(repo, 1, command.Plan, metadata)
	ErrContains(t, `currently locked by "apply"`, err)
}
