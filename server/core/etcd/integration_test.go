// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Integration tests for the etcd database contract (design §"Consistency and
// concurrency"). They run against an in-process etcd started per test.

package etcd_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func newDB(t *testing.T) *etcd.EtcdDatabase {
	backend := newTestBackend(t)
	return etcd.NewDatabase(backend, "/atlantis", 5*time.Second)
}

func projectLock(host, repo, path, workspace string, pull int) models.ProjectLock {
	return models.ProjectLock{
		Project:   models.Project{RepoFullName: repo, Path: path},
		Workspace: workspace,
		Time:      time.Now(),
		User:      models.User{Username: "u"},
		Pull: models.PullRequest{
			Num:      pull,
			BaseRepo: models.Repo{FullName: repo, VCSHost: models.VCSHost{Hostname: host}},
		},
	}
}

// TestTryLock_SingleWinner proves concurrent contention yields exactly one
// winner (design §965).
func TestTryLock_SingleWinner(t *testing.T) {
	d := newDB(t)
	lock := projectLock("github.com", "o/r", ".", "default", 1)

	const n = 12
	var wg sync.WaitGroup
	wins := make([]bool, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			acquired, _, err := d.TryLock(lock)
			if err == nil {
				wins[i] = acquired
			}
		}(i)
	}
	wg.Wait()

	won := 0
	for _, w := range wins {
		if w {
			won++
		}
	}
	Equals(t, 1, won)
}

// TestTryLock_HostIsolation proves two VCS hosts with identical repo names and
// pull numbers have distinct locks (design §968).
func TestTryLock_HostIsolation(t *testing.T) {
	d := newDB(t)
	a := projectLock("github.com", "o/r", ".", "default", 1)
	b := projectLock("gitlab.com", "o/r", ".", "default", 1)

	okA, _, err := d.TryLock(a)
	Ok(t, err)
	Assert(t, okA, "first host should acquire")
	okB, _, err := d.TryLock(b)
	Ok(t, err)
	Assert(t, okB, "second host must not collide with the first")
}

// TestUnlockIfOwnedByPull_DoesNotRemoveOthers proves conditional unlock never
// removes another owner's lock (design §970).
func TestUnlockIfOwnedByPull_DoesNotRemoveOthers(t *testing.T) {
	d := newDB(t)
	lock := projectLock("github.com", "o/r", ".", "default", 5)
	ok, _, err := d.TryLock(lock)
	Ok(t, err)
	Assert(t, ok, "lock should be acquired")

	// A different pull number must not delete it.
	removed, err := d.UnlockIfOwnedByPull(lock.Project, lock.Workspace, 999)
	Ok(t, err)
	Assert(t, removed == nil, "must not unlock a lock owned by a different pull")

	got, err := d.GetLock(lock.Project, lock.Workspace)
	Ok(t, err)
	Assert(t, got != nil, "lock must still be present")

	// The true owner unlocks.
	removed, err = d.UnlockIfOwnedByPull(lock.Project, lock.Workspace, 5)
	Ok(t, err)
	Assert(t, removed != nil, "true owner should unlock")
}

// TestUnlockByPullScope_CleaningBlocksAcquire proves the cleaning lifecycle
// record blocks concurrent same-pull lock creation and that no lock is missed
// (design §972).
func TestUnlockByPullScope_CleaningBlocksAcquire(t *testing.T) {
	d := newDB(t)
	scoped := d.Scoped()
	ctx := context.Background()

	l1 := projectLock("github.com", "o/r", "a", "default", 7)
	l2 := projectLock("github.com", "o/r", "b", "default", 7)
	for _, l := range []models.ProjectLock{l1, l2} {
		ok, _, err := d.TryLock(l)
		Ok(t, err)
		Assert(t, ok, "setup lock should acquire")
	}

	pull := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 7}
	removed, err := scoped.UnlockByPullScope(ctx, pull, false)
	Ok(t, err)
	Equals(t, 2, len(removed))

	// After manual cleanup (open), acquisition is permitted again.
	ok, _, err := d.TryLock(l1)
	Ok(t, err)
	Assert(t, ok, "reacquire should succeed after cleanup returns lifecycle to open")
}

// TestGlobalLock_SingleWinner proves the global command lock yields one winner
// (design §965).
func TestGlobalLock_SingleWinner(t *testing.T) {
	d := newDB(t)
	_, err := d.LockCommand(command.Apply, time.Now())
	Ok(t, err)
	_, err = d.LockCommand(command.Apply, time.Now())
	ErrContains(t, "lock already exists", err)

	got, err := d.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, got != nil, "lock should be present")

	Ok(t, d.UnlockCommand(command.Apply))
	got, err = d.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, got == nil, "lock should be gone")
	ErrContains(t, "no lock exists", d.UnlockCommand(command.Apply))
}

// TestPullStatus_NoLostConcurrentUpdates proves parallel status updates for
// distinct projects do not lose results under the CAS loop (design §974).
func TestPullStatus_NoLostConcurrentUpdates(t *testing.T) {
	d := newDB(t)
	pull := models.PullRequest{
		Num:        3,
		HeadCommit: "abc",
		BaseRepo:   models.Repo{FullName: "o/r", VCSHost: models.VCSHost{Hostname: "github.com"}},
	}

	// Seed the pull status so merges take the in-place merge branch.
	_, err := d.UpdatePullWithResults(pull, []command.ProjectResult{{RepoRelDir: "seed", Workspace: "default"}})
	Ok(t, err)

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			res := command.ProjectResult{RepoRelDir: dir(i), Workspace: "default"}
			_, _ = d.UpdatePullWithResults(pull, []command.ProjectResult{res})
		}(i)
	}
	wg.Wait()

	got, err := d.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, got != nil, "status should exist")
	// seed + n distinct project dirs, none lost.
	Equals(t, n+1, len(got.Projects))
}

func dir(i int) string { return "proj-" + string(rune('a'+i)) }
