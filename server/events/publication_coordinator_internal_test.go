// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/google/go-github/v88/github"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type publicationTestTimer struct {
	at    time.Time
	delay time.Duration
	ch    chan time.Time
}
type publicationTestClock struct {
	mu         sync.Mutex
	now        time.Time
	timers     []publicationTestTimer
	registered chan struct{}
}

func newPublicationTestClock() *publicationTestClock {
	return &publicationTestClock{now: time.Unix(1000, 0), registered: make(chan struct{}, 100)}
}
func (c *publicationTestClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *publicationTestClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan time.Time, 1)
	c.timers = append(c.timers, publicationTestTimer{c.now.Add(d), d, ch})
	c.registered <- struct{}{}
	return ch
}
func (c *publicationTestClock) waitFor(d time.Duration) {
	for {
		c.mu.Lock()
		found := false
		for _, timer := range c.timers {
			if timer.delay == d {
				found = true
				break
			}
		}
		c.mu.Unlock()
		if found {
			return
		}
		<-c.registered
	}
}
func (c *publicationTestClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	var pending []publicationTestTimer
	for _, timer := range c.timers {
		if !timer.at.After(c.now) {
			timer.ch <- c.now
		} else {
			pending = append(pending, timer)
		}
	}
	c.timers = pending
}

type publicationTestStore struct {
	mu      sync.Mutex
	clock   *publicationTestClock
	lease   *db.PublicationLease
	renewed chan struct{}
}

func (s *publicationTestStore) AcquirePublicationLease(ctx context.Context, _ models.PullRequest, owner string, ttl time.Duration) (db.PublicationLease, error) {
	if err := ctx.Err(); err != nil {
		return db.PublicationLease{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := db.AcquirePublicationLease(s.lease, owner, s.clock.Now(), ttl)
	if err == nil {
		s.lease = &next
	}
	return next, err
}
func (s *publicationTestStore) RenewPublicationLease(ctx context.Context, _ models.PullRequest, fence db.PublicationFence, ttl time.Duration) (db.PublicationLease, error) {
	if err := ctx.Err(); err != nil {
		return db.PublicationLease{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := db.RenewPublicationLease(s.lease, fence, s.clock.Now(), ttl)
	if err == nil {
		s.lease = &next
		s.renewed <- struct{}{}
	}
	return next, err
}
func (s *publicationTestStore) BeginPublication(_ context.Context, _ models.PullRequest, fence db.PublicationFence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := db.BeginPublication(s.lease, fence, s.clock.Now())
	if err == nil {
		s.lease = &next
	}
	return err
}
func (s *publicationTestStore) CompletePublication(_ context.Context, _ models.PullRequest, fence db.PublicationFence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := db.CompletePublication(s.lease, fence)
	if err == nil {
		s.lease = &next
	}
	return err
}
func (s *publicationTestStore) ReleasePublicationLease(_ context.Context, _ models.PullRequest, fence db.PublicationFence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := db.CanReleasePublicationLease(s.lease, fence)
	if err == nil {
		s.lease = nil
	}
	return err
}
func testPublicationCoordinator() (*PublicationCoordinator, *publicationTestStore, *publicationTestClock) {
	clock := newPublicationTestClock()
	store := &publicationTestStore{clock: clock, renewed: make(chan struct{}, 10)}
	p := NewPublicationCoordinator(store, context.Background())
	p.Clock = clock
	return p, store, clock
}

func TestPublicationCoordinator_BoundedCancellableWait(t *testing.T) {
	for _, reason := range []string{"cancel", "shutdown", "maximum wait"} {
		t.Run(reason, func(t *testing.T) {
			p, store, clock := testPublicationCoordinator()
			pull := models.PullRequest{Num: 1}
			_, err := store.AcquirePublicationLease(context.Background(), pull, "other", time.Minute)
			require.NoError(t, err)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			shutdown, stop := context.WithCancel(context.Background())
			defer stop()
			p.Shutdown = shutdown
			result := make(chan error, 1)
			go func() {
				result <- p.Run(ctx, pull, func(*HeldPublicationLease) error { return errors.New("busy command must not execute") })
			}()
			clock.waitFor(p.RetryInterval)
			expected := context.Canceled
			switch reason {
			case "cancel":
				cancel()
			case "shutdown":
				stop()
				expected = ErrPublicationShutdown
			case "maximum wait":
				clock.advance(p.MaxWait)
				expected = ErrPublicationWaitLimit
			}
			require.ErrorIs(t, <-result, expected)
			require.NoError(t, store.ReleasePublicationLease(context.Background(), pull, db.PublicationFence{Owner: "other"}))
		})
	}
}

func TestPublicationCoordinator_PreExecutionFailuresAndPanicRelease(t *testing.T) {
	for _, name := range []string{"pending status", "discard reviews", "hide comments", "validation", "lock", "panic"} {
		t.Run(name, func(t *testing.T) {
			p, store, _ := testPublicationCoordinator()
			pull := models.PullRequest{Num: 1}
			failure := errors.New(name)
			if name == "panic" {
				require.Panics(t, func() {
					_ = p.Run(context.Background(), pull, func(*HeldPublicationLease) error { panic("callback panic") })
				})
			} else {
				require.ErrorIs(t, p.Run(context.Background(), pull, func(*HeldPublicationLease) error { return failure }), failure)
			}
			_, err := store.AcquirePublicationLease(context.Background(), pull, "replacement", time.Minute)
			require.NoError(t, err)
		})
	}
}

func TestPublicationCoordinator_TerminalAmbiguityAndHeartbeat(t *testing.T) {
	t.Run("ambiguous request", func(t *testing.T) {
		p, store, clock := testPublicationCoordinator()
		pull := models.PullRequest{Num: 1}
		err := p.Run(context.Background(), pull, func(lease *HeldPublicationLease) error {
			err := lease.Publish(func() error { return errors.New("remote request outcome unknown") })
			require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
			require.ErrorIs(t, lease.Publish(func() error { t.Error("must not issue another mutation after an ambiguous request"); return nil }), db.ErrPublicationAmbiguous)
			return err
		})
		require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
		clock.advance(2 * p.LeaseTTL)
		_, err = store.AcquirePublicationLease(context.Background(), pull, "replacement", time.Minute)
		require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
	})
	t.Run("heartbeat only inside critical section", func(t *testing.T) {
		p, store, clock := testPublicationCoordinator()
		pull := models.PullRequest{Num: 1}
		finish := make(chan struct{})
		result := make(chan error, 1)
		go func() {
			result <- p.Run(context.Background(), pull, func(*HeldPublicationLease) error { <-finish; return nil })
		}()
		for range 4 {
			clock.waitFor(p.LeaseTTL / 3)
			clock.advance(p.LeaseTTL / 3)
			<-store.renewed
		}
		_, err := store.AcquirePublicationLease(context.Background(), pull, "other", time.Minute)
		require.ErrorIs(t, err, db.ErrPublicationBusy)
		close(finish)
		require.NoError(t, <-result)
		// An arbitrarily long execution interval outside Run owns no lease.
		clock.advance(10 * p.LeaseTTL)
		_, err = store.AcquirePublicationLease(context.Background(), pull, "other", time.Minute)
		require.NoError(t, err)
	})
}

func TestPublicationCoordinator_KnownRejectionReleases(t *testing.T) {
	response := func(code int) *http.Response {
		return &http.Response{StatusCode: code, Request: &http.Request{Method: http.MethodPost, URL: &url.URL{Scheme: "https", Host: "vcs.example", Path: "/status"}}}
	}
	for _, tc := range []struct {
		name      string
		err       error
		ambiguous bool
	}{
		{"github forbidden", &github.ErrorResponse{Response: response(403)}, false},
		{"github rate limit", &github.RateLimitError{Response: response(429)}, false},
		{"github abuse limit", &github.AbuseRateLimitError{Response: response(403)}, false},
		{"gitlab forbidden wrapped", fmt.Errorf("setting status: %w", &gitlab.ErrorResponse{Response: response(403)}), false},
		{"github server error", &github.ErrorResponse{Response: response(503)}, true},
		{"gitlab timeout", &gitlab.ErrorResponse{Response: response(408)}, true},
		{"no response", &github.ErrorResponse{}, true},
		{"joined unknown", errors.Join(&github.ErrorResponse{Response: response(403)}, errors.New("unknown outcome")), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, store, clock := testPublicationCoordinator()
			pull := models.PullRequest{Num: 1}
			err := p.Run(context.Background(), pull, func(lease *HeldPublicationLease) error {
				return lease.Publish(func() error { return tc.err })
			})
			require.ErrorIs(t, err, tc.err)
			require.Equal(t, tc.ambiguous, errors.Is(err, db.ErrPublicationAmbiguous))
			clock.advance(2 * p.LeaseTTL)
			_, err = store.AcquirePublicationLease(context.Background(), pull, "replacement", time.Minute)
			if tc.ambiguous {
				require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPublicationCoordinator_CommandCancellationInterruptsWait(t *testing.T) {
	p, store, clock := testPublicationCoordinator()
	tracker := NewCancellationTracker()
	p.CancellationTracker = tracker
	pull := models.PullRequest{Num: 1}
	_, err := store.AcquirePublicationLease(context.Background(), pull, "other", time.Minute)
	require.NoError(t, err)
	result := make(chan error, 1)
	go func() {
		result <- p.RunCommand(&command.Context{Pull: pull}, func() error {
			t.Error("cancelled command must not execute")
			return nil
		})
	}()
	clock.waitFor(p.RetryInterval)
	tracker.Cancel(pull)
	require.ErrorIs(t, <-result, context.Canceled)
	require.Empty(t, tracker.waiters, "completed wait must unregister cancellation callback")
	require.NoError(t, store.ReleasePublicationLease(context.Background(), pull, db.PublicationFence{Owner: "other"}))
	tracker.Clear(pull)
	require.NoError(t, p.RunCommand(&command.Context{Pull: pull}, func() error { return nil }))
}

func TestCancellationTracker_IndependentWaitersAndClear(t *testing.T) {
	tracker := NewCancellationTracker()
	pull := models.PullRequest{Num: 1}
	first, stopFirst := tracker.CommandContext(context.Background(), pull)
	defer stopFirst()
	second, stopSecond := tracker.CommandContext(context.Background(), pull)
	defer stopSecond()
	other, stopOther := tracker.CommandContext(context.Background(), models.PullRequest{Num: 2})
	defer stopOther()
	tracker.Cancel(pull)
	require.ErrorIs(t, first.Err(), context.Canceled)
	require.ErrorIs(t, second.Err(), context.Canceled)
	require.NoError(t, other.Err())
	alreadyCancelled, stopCancelled := tracker.CommandContext(context.Background(), pull)
	defer stopCancelled()
	require.ErrorIs(t, alreadyCancelled.Err(), context.Canceled)
	tracker.Clear(pull)
	next, stopNext := tracker.CommandContext(context.Background(), pull)
	defer stopNext()
	require.NoError(t, next.Err())
	require.ErrorIs(t, first.Err(), context.Canceled, "clearing cancellation cannot resurrect an old waiter")
}

func TestPublicationCoordinator_AtomicStatusDeletionEndsOwnership(t *testing.T) {
	p, store, clock := testPublicationCoordinator()
	pull := models.PullRequest{Num: 1}
	require.NoError(t, p.Run(context.Background(), pull, func(held *HeldPublicationLease) error {
		// The backend parity test proves atomic status-plus-lease deletion.
		// Here the lease store simulates the resulting absent lease.
		require.NoError(t, held.DeleteStatus(func() error {
			return store.ReleasePublicationLease(context.Background(), pull, held.Fence())
		}))
		_, err := store.AcquirePublicationLease(context.Background(), pull, "replacement", time.Minute)
		require.NoError(t, err)
		clock.advance(p.LeaseTTL / 3)
		require.ErrorIs(t, held.Publish(func() error {
			t.Error("publication after deleting status must not execute")
			return nil
		}), db.ErrPublicationOwnerLost)
		return nil
	}))
	// Final release/heartbeat must not interfere with the replacement owner.
	require.NoError(t, store.ReleasePublicationLease(context.Background(), pull, db.PublicationFence{Owner: "replacement"}))
}

func TestPublicationCoordinator_PreExecutionFailureCannotOverwriteNewGeneration(t *testing.T) {
	for _, superseded := range []bool{false, true} {
		t.Run(fmt.Sprint(superseded), func(t *testing.T) {
			store, err := boltdb.New(t.TempDir())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, store.Close()) })
			pull := models.PullRequest{Num: 1, HeadCommit: "same-head", BaseRepo: models.Repo{FullName: "owner/repo"}}
			projects := []command.ProjectContext{{Workspace: "default", RepoRelDir: "."}}
			_, err = store.BeginPlanGeneration(pull, "G1", projects, false, command.NoClaim{})
			require.NoError(t, err)
			observed, err := store.GetPullStatus(pull)
			require.NoError(t, err)
			if superseded {
				_, err = store.BeginPlanGeneration(pull, "G2", projects, false, command.NoClaim{})
				require.NoError(t, err)
			}
			p := NewPublicationCoordinator(store, context.Background())
			published := false
			err = p.RunWithObservedStatus(&command.Context{Pull: pull, PullStatus: observed}, store, nil, func() error {
				published = true
				return nil
			})
			if superseded {
				require.ErrorIs(t, err, db.ErrPlanGenerationSuperseded)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, !superseded, published)
			_, err = store.AcquirePublicationLease(context.Background(), pull, "next", time.Minute)
			require.NoError(t, err)
		})
	}
}

func TestPublicationCoordinator_RejectsPublicationOutsideOwnedSection(t *testing.T) {
	p, _, _ := testPublicationCoordinator()
	ctx := &command.Context{Pull: models.PullRequest{Num: 1}}
	require.NoError(t, p.RunCommand(ctx, func() error {
		return publishTerminal(ctx, func() error { return nil })
	}))
	require.ErrorIs(t, publishTerminal(ctx, func() error { t.Error("must not publish terminal state after release"); return nil }), db.ErrPublicationOwnerLost)
}
