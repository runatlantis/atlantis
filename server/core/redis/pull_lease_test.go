// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestPullLeaseLifecycle(t *testing.T) {
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	ctx := context.Background()

	acquired, current, err := r.TryAcquirePullLease(ctx, "owner/repo", 12, "owner-a", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !acquired || current != "owner-a" {
		t.Fatalf("expected owner-a to acquire lease, acquired=%v current=%q", acquired, current)
	}

	acquired, current, err = r.TryAcquirePullLease(ctx, "owner/repo", 12, "owner-b", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if acquired || current != "owner-a" {
		t.Fatalf("expected owner-a to retain lease, acquired=%v current=%q", acquired, current)
	}

	released, err := r.ReleasePullLease(ctx, "owner/repo", 12, "owner-b")
	if err != nil {
		t.Fatal(err)
	}
	if released {
		t.Fatal("expected a non-owner release to leave the lease intact")
	}

	renewed, err := r.RenewPullLease(ctx, "owner/repo", 12, "owner-a", 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !renewed {
		t.Fatal("expected owner-a to renew its lease")
	}

	released, err = r.ReleasePullLease(ctx, "owner/repo", 12, "owner-a")
	if err != nil {
		t.Fatal(err)
	}
	if !released {
		t.Fatal("expected owner-a to release its lease")
	}

	acquired, _, err = r.TryAcquirePullLease(ctx, "owner/repo", 12, "owner-b", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("expected owner-b to acquire the released lease")
	}
}

func TestPullLeaseExpiresAfterOwnerStopsRenewing(t *testing.T) {
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	ctx := context.Background()

	acquired, _, err := r.TryAcquirePullLease(ctx, "owner/repo", 12, "owner-a", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("expected owner-a to acquire lease")
	}

	s.FastForward(time.Minute + time.Second)

	acquired, _, err = r.TryAcquirePullLease(ctx, "owner/repo", 12, "owner-b", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("expected owner-b to acquire the expired lease")
	}
}

func TestPullLeaseDoesNotPolluteProjectLockListing(t *testing.T) {
	s := miniredis.RunT(t)
	r := newTestRedis(s)

	acquired, _, err := r.TryAcquirePullLease(context.Background(), "owner/repo", 12, "owner-a", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("expected lease acquisition")
	}

	locks, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(locks) != 0 {
		t.Fatalf("expected no project locks, got %d", len(locks))
	}
}
