// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/stretchr/testify/require"
)

func TestLocalClaimGuard_AcquirePreparesClaimOnceConcurrently(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := localClaimTestKey()
	started := make(chan struct{})
	continueReset := make(chan struct{})
	var resetCalls atomic.Int32
	var admitCalls atomic.Int32
	reset := func() error {
		if resetCalls.Add(1) == 1 {
			close(started)
		}
		<-continueReset
		return nil
	}

	const workers = 16
	releases := make(chan func(), workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, reset, err := guard.Acquire(key, "claim-1", func() error {
				admitCalls.Add(1)
				return nil
			}, reset)
			if err != nil {
				errs <- err
				return
			}
			if !reset {
				errs <- errors.New("expected reset generation")
				return
			}
			releases <- release
			errs <- nil
		}()
	}

	<-started
	close(continueReset)
	wg.Wait()
	close(errs)
	close(releases)
	for err := range errs {
		require.NoError(t, err)
	}
	for release := range releases {
		release()
	}
	require.Equal(t, int32(1), resetCalls.Load())
	require.Equal(t, int32(workers), admitCalls.Load())
}

func TestLocalClaimGuard_NewClaimWaitsForActiveGeneration(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := localClaimTestKey()
	releaseOld, _, err := guard.Acquire(key, "claim-1", func() error { return nil }, func() error { return nil })
	require.NoError(t, err)

	newAdmitStarted := make(chan struct{})
	acquired := make(chan error, 1)
	go func() {
		release, _, err := guard.Acquire(key, "claim-2", func() error {
			close(newAdmitStarted)
			return nil
		}, func() error { return nil })
		if err == nil {
			release()
		}
		acquired <- err
	}()

	select {
	case <-newAdmitStarted:
		t.Fatal("new claim was admitted while the old generation was active")
	case <-time.After(100 * time.Millisecond):
	}

	releaseOld()
	select {
	case <-newAdmitStarted:
	case <-time.After(time.Second):
		t.Fatal("new claim did not proceed after the old generation was released")
	}
	require.NoError(t, <-acquired)
}

func TestLocalClaimGuard_RetriesFailedReset(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := localClaimTestKey()
	var calls int
	reset := func() error {
		calls++
		if calls == 1 {
			return errors.New("disk busy")
		}
		return nil
	}

	_, _, err := guard.Acquire(key, "claim-1", func() error { return nil }, reset)
	require.EqualError(t, err, "disk busy")
	release, _, err := guard.Acquire(key, "claim-1", func() error { return nil }, reset)
	require.NoError(t, err)
	release()
	require.Equal(t, 2, calls)
}

func TestLocalClaimGuard_StaleClaimCannotReplaceNewerGeneration(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := localClaimTestKey()
	var resetCalls int
	reset := func() error {
		resetCalls++
		return nil
	}

	releaseOld, _, err := guard.Acquire(key, "claim-1", func() error { return nil }, reset)
	require.NoError(t, err)
	releaseOld()
	releaseNew, _, err := guard.Acquire(key, "claim-2", func() error { return nil }, reset)
	require.NoError(t, err)
	releaseNew()

	_, _, err = guard.Acquire(key, "claim-1", func() error { return events.ErrOwnershipChanged }, reset)
	require.ErrorIs(t, err, events.ErrOwnershipChanged)
	require.Equal(t, 2, resetCalls)

	releaseCurrent, _, err := guard.Acquire(key, "claim-2", func() error { return nil }, reset)
	require.NoError(t, err)
	releaseCurrent()
	require.Equal(t, 2, resetCalls)
}

func TestLocalClaimGuard_ReleaseIsIdempotent(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := localClaimTestKey()
	release, _, err := guard.Acquire(key, "claim-1", func() error { return nil }, func() error { return nil })
	require.NoError(t, err)
	release()
	release()

	releaseNext, _, err := guard.Acquire(key, "claim-2", func() error { return nil }, func() error { return nil })
	require.NoError(t, err)
	releaseNext()
}

func TestLocalClaimGuard_EmptyClaimLeavesLocalStateUntouched(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	called := false

	release, reset, err := guard.Acquire(localClaimTestKey(), "", func() error {
		called = true
		return nil
	}, func() error {
		called = true
		return nil
	})
	require.NoError(t, err)
	require.False(t, reset)
	release()
	require.False(t, called)
}

func TestLocalClaimGuard_ForgetOnlyRemovesExactInactiveClaim(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := localClaimTestKey()
	var calls int
	reset := func() error {
		calls++
		return nil
	}

	release, _, err := guard.Acquire(key, "claim-2", func() error { return nil }, reset)
	require.NoError(t, err)
	release()
	require.NoError(t, guard.Forget(key, "claim-1"))
	release, _, err = guard.Acquire(key, "claim-2", func() error { return nil }, reset)
	require.NoError(t, err)
	release()
	require.Equal(t, 1, calls)

	require.NoError(t, guard.Forget(key, "claim-2"))
	release, _, err = guard.Acquire(key, "claim-2", func() error { return nil }, reset)
	require.NoError(t, err)
	release()
	require.Equal(t, 2, calls)
}

func localClaimTestKey() ownership.Key {
	return ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12}
}
