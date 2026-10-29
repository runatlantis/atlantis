// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
)

type renewalFailureStore struct {
	mutex       sync.Mutex
	renewCalls  int
	releaseCall chan struct{}
	renew       func(context.Context) (bool, error)
}

func (s *renewalFailureStore) TryAcquirePullLease(context.Context, string, int, string, time.Duration) (bool, string, error) {
	return true, "", nil
}

func (s *renewalFailureStore) RenewPullLease(ctx context.Context, _ string, _ int, _ string, _ time.Duration) (bool, error) {
	s.mutex.Lock()
	s.renewCalls++
	s.mutex.Unlock()
	if s.renew != nil {
		return s.renew(ctx)
	}
	return false, errors.New("redis unavailable")
}

func (s *renewalFailureStore) ReleasePullLease(context.Context, string, int, string) (bool, error) {
	close(s.releaseCall)
	return true, nil
}

func TestLeasedWorkingDirLockerFailsBeforeUnconfirmedLeaseExpires(t *testing.T) {
	store := &renewalFailureStore{releaseCall: make(chan struct{})}
	ttl := 300 * time.Millisecond
	locker := NewLeasedWorkingDirLocker(
		NewDefaultWorkingDirLocker(),
		store,
		logging.NewNoopLogger(t),
		ttl,
	)
	leaseLost := make(chan error, 1)
	locker.leaseLossHandler = func(err error) {
		leaseLost <- err
	}

	unlock, err := locker.TryLockPull("owner/repo", 1, command.Plan, WorkingDirLockMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	startedAt := time.Now()
	select {
	case err := <-leaseLost:
		if !strings.Contains(err.Error(), "before expiry") {
			t.Fatalf("expected expiry error, got %q", err)
		}
		if time.Since(startedAt) >= ttl {
			t.Fatalf("lease loss handler ran after the lease could expire")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lease loss handler")
	}

	store.mutex.Lock()
	renewCalls := store.renewCalls
	store.mutex.Unlock()
	if renewCalls < 2 {
		t.Fatalf("expected renewal retries, got %d call(s)", renewCalls)
	}
}

func TestLeasedWorkingDirLockerDoesNotFailStopAfterUnlock(t *testing.T) {
	renewStarted := make(chan struct{})
	store := &renewalFailureStore{
		releaseCall: make(chan struct{}),
		renew: func(ctx context.Context) (bool, error) {
			close(renewStarted)
			<-ctx.Done()
			return false, ctx.Err()
		},
	}
	locker := NewLeasedWorkingDirLocker(
		NewDefaultWorkingDirLocker(),
		store,
		logging.NewNoopLogger(t),
		300*time.Millisecond,
	)
	leaseLost := make(chan error, 1)
	locker.leaseLossHandler = func(err error) {
		leaseLost <- err
	}

	unlock, err := locker.TryLockPull("owner/repo", 1, command.Plan, WorkingDirLockMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	<-renewStarted
	unlock()

	select {
	case err := <-leaseLost:
		t.Fatalf("unexpected lease loss after unlock: %v", err)
	default:
	}
	select {
	case <-store.releaseCall:
	default:
		t.Fatal("expected unlock to release the lease")
	}
}
