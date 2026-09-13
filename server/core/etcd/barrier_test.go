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

func barrierFixture(t *testing.T) (*etcd.ExecutionBarrierStore, *etcd.OwnershipStore, etcd.Claim) {
	t.Helper()
	backend := newTestBackend(t)
	keys := etcd.NewKeyspace("/atlantis")
	epoch, err := etcd.InitOrValidateNamespace(context.Background(), backend.Client().KV, keys, "dep-1")
	Ok(t, err)
	owner, err := etcd.NewOwnershipStore(backend, keys, epoch, "A", "https://a:4141", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = owner.Close() })
	claim, won, err := owner.Claim(context.Background(), ownScope)
	Ok(t, err)
	Assert(t, won, "owner should win")
	return etcd.NewExecutionBarrierStore(backend, keys, epoch, owner.InstanceID()), owner, claim
}

// TestBarrier_StartComplete proves a barrier can be created and cleared.
func TestBarrier_StartComplete(t *testing.T) {
	store, _, claim := barrierFixture(t)
	ctx := context.Background()

	b, err := store.StartStep(ctx, claim, "exec-1")
	Ok(t, err)
	Equals(t, "exec-1", b.ExecutionID())
	Ok(t, store.CompleteStep(ctx, b))
}

// TestBarrier_ParallelSameGeneration proves parallel executions under one valid
// owner generation both proceed (design §546).
func TestBarrier_ParallelSameGeneration(t *testing.T) {
	store, _, claim := barrierFixture(t)
	ctx := context.Background()

	b1, err := store.StartStep(ctx, claim, "exec-1")
	Ok(t, err)
	b2, err := store.StartStep(ctx, claim, "exec-2")
	Ok(t, err)
	Ok(t, store.CompleteStep(ctx, b1))
	Ok(t, store.CompleteStep(ctx, b2))
}

// TestBarrier_OlderGenerationBlocks proves an unresolved barrier from an older
// generation blocks a new owner generation from starting work for that pull,
// while a cleared barrier does not (design §558).
func TestBarrier_OlderGenerationBlocks(t *testing.T) {
	store, owner, claim := barrierFixture(t)
	ctx := context.Background()

	// Old generation starts a step and does NOT complete it (owner dies).
	oldBarrier, err := store.StartStep(ctx, claim, "exec-1")
	Ok(t, err)

	// The claim is released and reacquired: a new fencing generation.
	Ok(t, owner.Release(ctx, claim))
	newClaim, won, err := owner.Claim(ctx, ownScope)
	Ok(t, err)
	Assert(t, won, "owner should reclaim")
	Assert(t, newClaim.Generation().CreateRevision != claim.Generation().CreateRevision, "new generation expected")

	// The new generation is blocked by the old generation's unresolved barrier.
	_, err = store.StartStep(ctx, newClaim, "exec-2")
	ErrContains(t, "older owner generation", err)

	// Once the old barrier is cleared, the new generation may proceed.
	Ok(t, store.CompleteStep(ctx, oldBarrier))
	b, err := store.StartStep(ctx, newClaim, "exec-2")
	Ok(t, err)
	Ok(t, store.CompleteStep(ctx, b))
}

// TestBarrier_ClaimLostFailsStart proves a barrier cannot be created once the
// owning claim generation has changed (design §560).
func TestBarrier_ClaimLostFailsStart(t *testing.T) {
	store, owner, claim := barrierFixture(t)
	ctx := context.Background()

	Ok(t, owner.Release(ctx, claim))
	// The stale claim no longer matches the live ownership key.
	_, err := store.StartStep(ctx, claim, "exec-1")
	ErrContains(t, "claim lost", err)
}

// TestBarrier_UncertainBlocks proves an uncertain barrier keeps blocking a new
// generation until explicit resolution (design §558).
func TestBarrier_UncertainBlocks(t *testing.T) {
	store, owner, claim := barrierFixture(t)
	ctx := context.Background()

	oldBarrier, err := store.StartStep(ctx, claim, "exec-1")
	Ok(t, err)
	Ok(t, store.MarkUncertain(ctx, oldBarrier))

	Ok(t, owner.Release(ctx, claim))
	newClaim, _, err := owner.Claim(ctx, ownScope)
	Ok(t, err)

	_, err = store.StartStep(ctx, newClaim, "exec-2")
	ErrContains(t, "older owner generation", err)
}
