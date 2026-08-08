// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/stretchr/testify/require"
)

func TestLocalClaimGuard_PreparesClaimOnceConcurrently(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12}
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	reset := func() error {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return nil
	}

	const workers = 16
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- guard.Prepare(key, "claim-1", reset)
		}()
	}

	<-started
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, int32(1), calls.Load())
}

func TestLocalClaimGuard_RetriesFailedReset(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12}
	var calls int
	reset := func() error {
		calls++
		if calls == 1 {
			return errors.New("disk busy")
		}
		return nil
	}

	require.EqualError(t, guard.Prepare(key, "claim-1", reset), "disk busy")
	require.NoError(t, guard.Prepare(key, "claim-1", reset))
	require.Equal(t, 2, calls)
}

func TestLocalClaimGuard_NewClaimResetsLocalStateAgain(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12}
	var calls int
	reset := func() error {
		calls++
		return nil
	}

	require.NoError(t, guard.Prepare(key, "claim-1", reset))
	require.NoError(t, guard.Prepare(key, "claim-1", reset))
	require.NoError(t, guard.Prepare(key, "claim-2", reset))
	require.Equal(t, 2, calls)
}

func TestLocalClaimGuard_EmptyClaimLeavesLocalStateUntouched(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12}
	called := false

	require.NoError(t, guard.Prepare(key, "", func() error {
		called = true
		return nil
	}))
	require.False(t, called)
}

func TestLocalClaimGuard_ForgetOnlyRemovesExactClaim(t *testing.T) {
	guard := events.NewLocalClaimGuard()
	key := ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12}
	var calls int
	reset := func() error {
		calls++
		return nil
	}

	require.NoError(t, guard.Prepare(key, "claim-2", reset))
	require.NoError(t, guard.Forget(key, "claim-1"))
	require.NoError(t, guard.Prepare(key, "claim-2", reset))
	require.Equal(t, 1, calls)

	require.NoError(t, guard.Forget(key, "claim-2"))
	require.NoError(t, guard.Prepare(key, "claim-2", reset))
	require.Equal(t, 2, calls)
}
