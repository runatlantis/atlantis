// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

// ownershipFixture starts etcd, initializes the namespace, and returns two
// ownership stores that simulate two independent Atlantis processes (distinct
// session leases and instance IDs) against the same cluster.
func ownershipFixture(t *testing.T) (*etcd.OwnershipStore, *etcd.OwnershipStore) {
	t.Helper()
	backend := newTestBackend(t)
	keys := etcd.NewKeyspace("/atlantis")
	epoch, err := etcd.InitOrValidateNamespace(context.Background(), backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	a, err := etcd.NewOwnershipStore(backend, keys, epoch, "replica-a", "https://a:4141", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = a.Close() })
	b, err := etcd.NewOwnershipStore(backend, keys, epoch, "replica-b", "https://b:4141", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = b.Close() })
	return a, b
}

var ownScope = etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1}

// TestOwnership_SingleWinner proves two processes racing for the same pull yield
// exactly one owner, and the loser sees the winner's advertise URL (design §979).
func TestOwnership_SingleWinner(t *testing.T) {
	a, b := ownershipFixture(t)
	ctx := context.Background()

	claimA, wonA, err := a.Claim(ctx, ownScope)
	Ok(t, err)
	claimB, wonB, err := b.Claim(ctx, ownScope)
	Ok(t, err)

	Assert(t, wonA != wonB, "exactly one process must win")
	if wonA {
		Equals(t, "https://a:4141", claimA.AdvertiseURL())
		Equals(t, "https://a:4141", claimB.AdvertiseURL()) // loser sees the winner
	} else {
		Equals(t, "https://b:4141", claimB.AdvertiseURL())
		Equals(t, "https://b:4141", claimA.AdvertiseURL())
	}
}

// TestOwnership_OwnedLocally proves the owner recognizes its own generation and a
// different process does not (design §483).
func TestOwnership_OwnedLocally(t *testing.T) {
	a, b := ownershipFixture(t)
	ctx := context.Background()

	claim, won, err := a.Claim(ctx, ownScope)
	Ok(t, err)
	Assert(t, won, "a should win the uncontested claim")

	ownedByA, err := a.OwnedLocally(ctx, ownScope, claim.Generation())
	Ok(t, err)
	Assert(t, ownedByA, "owner must recognize its own claim generation")

	ownedByB, err := b.OwnedLocally(ctx, ownScope, claim.Generation())
	Ok(t, err)
	Assert(t, !ownedByB, "a different process must not consider itself the owner")
}

// TestOwnership_ReleaseThenReclaim proves release frees the claim and a new claim
// gets a distinct fencing generation (design §482).
func TestOwnership_ReleaseThenReclaim(t *testing.T) {
	a, b := ownershipFixture(t)
	ctx := context.Background()

	claimA, won, err := a.Claim(ctx, ownScope)
	Ok(t, err)
	Assert(t, won, "a should win")

	Ok(t, a.Release(ctx, claimA))

	claimB, wonB, err := b.Claim(ctx, ownScope)
	Ok(t, err)
	Assert(t, wonB, "b should win after a releases")
	Assert(t, claimB.Generation().CreateRevision != claimA.Generation().CreateRevision,
		"a reclaimed key must have a new fencing generation")
}

// TestOwnership_SessionCloseRevokesClaims proves a process's claims disappear
// when its session lease is revoked, so another process can take over (design
// §776 lease loss).
func TestOwnership_SessionCloseRevokesClaims(t *testing.T) {
	a, b := ownershipFixture(t)
	ctx := context.Background()

	_, won, err := a.Claim(ctx, ownScope)
	Ok(t, err)
	Assert(t, won, "a should win")

	// Simulate process a going away.
	Ok(t, a.Close())

	// The claim key is lease-bound; after revocation b can claim.
	got, err := b.Get(ctx, ownScope)
	Ok(t, err)
	Assert(t, got == nil, "a's lease-bound claim must be gone after session close")

	_, wonB, err := b.Claim(ctx, ownScope)
	Ok(t, err)
	Assert(t, wonB, "b should win after a's session is revoked")
}
