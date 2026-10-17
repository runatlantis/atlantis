// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	atlantisredis "github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
)

func TestLeasedWorkingDirLockerCoordinatesProcesses(t *testing.T) {
	s := miniredis.RunT(t)
	clientA := redislib.NewClient(&redislib.Options{Addr: s.Addr()})
	clientB := redislib.NewClient(&redislib.Options{Addr: s.Addr()})
	t.Cleanup(func() {
		_ = clientA.Close()
		_ = clientB.Close()
	})
	storeA, err := atlantisredis.NewWithClient(clientA, "", "")
	if err != nil {
		t.Fatal(err)
	}
	storeB, err := atlantisredis.NewWithClient(clientB, "", "")
	if err != nil {
		t.Fatal(err)
	}

	lockerA := events.NewLeasedWorkingDirLocker(events.NewDefaultWorkingDirLocker(), storeA, logging.NewNoopLogger(t), time.Minute)
	lockerB := events.NewLeasedWorkingDirLocker(events.NewDefaultWorkingDirLocker(), storeB, logging.NewNoopLogger(t), time.Minute)
	metadata := events.WorkingDirLockMetadata{
		HeadCommit: "0123456789abcdef0123456789abcdef01234567",
		CommitURL:  "https://github.com/owner/repo/commit/0123456789abcdef0123456789abcdef01234567",
	}

	unlockA, err := lockerA.TryLockPull("owner/repo", 12, command.Plan, metadata)
	if err != nil {
		t.Fatal(err)
	}

	_, err = lockerB.TryLockPull("owner/repo", 12, command.Apply, events.WorkingDirLockMetadata{})
	if err == nil {
		t.Fatal("expected the second process to be rejected")
	}
	if !strings.Contains(err.Error(), `currently locked by "plan" for commit 0123456`) {
		t.Fatalf("unexpected contention error: %s", err)
	}

	unlockA()

	unlockB, err := lockerB.TryLockPull("owner/repo", 12, command.Apply, events.WorkingDirLockMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	unlockB()
}

func TestLeasedWorkingDirLockerKeepsProjectLocksLocal(t *testing.T) {
	s := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store, err := atlantisredis.NewWithClient(client, "", "")
	if err != nil {
		t.Fatal(err)
	}
	local := events.NewDefaultWorkingDirLocker()
	locker := events.NewLeasedWorkingDirLocker(local, store, logging.NewNoopLogger(t), time.Minute)

	unlock, err := locker.TryLock("owner/repo", 12, "default", "project-a", "a", command.Plan, events.WorkingDirLockMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	if !locker.HasCommandLock("owner/repo", 12, command.Plan) {
		t.Fatal("expected project-level lock behavior to remain unchanged")
	}
}
