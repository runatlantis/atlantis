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

func admissionFixture(t *testing.T) (*etcd.AdmissionStore, *etcd.OwnershipStore, etcd.Claim) {
	t.Helper()
	backend := newTestBackend(t)
	keys := etcd.NewKeyspace("/atlantis")
	epoch, err := etcd.InitOrValidateNamespace(context.Background(), backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	owner, err := etcd.NewOwnershipStore(backend, keys, epoch, "replica-a", "https://a:4141", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = owner.Close() })

	claim, won, err := owner.Claim(context.Background(), ownScope)
	Ok(t, err)
	Assert(t, won, "owner should win the claim")

	store := etcd.NewAdmissionStore(backend, keys, epoch, owner.InstanceID())
	return store, owner, claim
}

var cmdID = etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: "delivery-1"}

// TestAdmission_HappyPath walks reserved -> scheduled -> running -> succeeded.
func TestAdmission_HappyPath(t *testing.T) {
	store, _, claim := admissionFixture(t)
	ctx := context.Background()

	rec, created, err := store.Reserve(ctx, cmdID, claim)
	Ok(t, err)
	Assert(t, created, "first reserve should create the record")
	Equals(t, etcd.AdmissionReserved, rec.State())

	rec, err = store.Transition(ctx, rec, etcd.AdmissionScheduled)
	Ok(t, err)
	Equals(t, etcd.AdmissionScheduled, rec.State())

	rec, err = store.Transition(ctx, rec, etcd.AdmissionRunning)
	Ok(t, err)
	Equals(t, etcd.AdmissionRunning, rec.State())

	rec, err = store.Complete(ctx, rec, true, "plan", "")
	Ok(t, err)
	Equals(t, etcd.AdmissionSucceeded, rec.State())
}

// TestAdmission_DuplicateReserveReturnsExisting proves a duplicate delivery under
// the same claim reconciles with the existing record rather than creating a
// second one (design §512, §526).
func TestAdmission_DuplicateReserveReturnsExisting(t *testing.T) {
	store, _, claim := admissionFixture(t)
	ctx := context.Background()

	first, created, err := store.Reserve(ctx, cmdID, claim)
	Ok(t, err)
	Assert(t, created, "first reserve creates")

	second, created2, err := store.Reserve(ctx, cmdID, claim)
	Ok(t, err)
	Assert(t, !created2, "duplicate reserve must not create a second record")
	Equals(t, first.Record.State, second.Record.State)
}

// TestAdmission_StaleTransitionRejected proves an out-of-date record cannot
// double-advance the state machine (design §516).
func TestAdmission_StaleTransitionRejected(t *testing.T) {
	store, _, claim := admissionFixture(t)
	ctx := context.Background()

	rec, _, err := store.Reserve(ctx, cmdID, claim)
	Ok(t, err)

	// Advance once to scheduled.
	_, err = store.Transition(ctx, rec, etcd.AdmissionScheduled)
	Ok(t, err)

	// Re-using the now-stale reserved record must fail the CAS.
	_, err = store.Transition(ctx, rec, etcd.AdmissionScheduled)
	ErrContains(t, "changed concurrently", err)
}

// TestAdmission_MarkUncertain proves a reserved record can be forced uncertain
// (owner died after 202 before execution) and never auto-advances (design §520).
func TestAdmission_MarkUncertain(t *testing.T) {
	store, _, claim := admissionFixture(t)
	ctx := context.Background()

	rec, _, err := store.Reserve(ctx, cmdID, claim)
	Ok(t, err)

	rec, err = store.MarkUncertain(ctx, rec)
	Ok(t, err)
	Equals(t, etcd.AdmissionUncertain, rec.State())

	// No transition out of uncertain is valid.
	_, err = store.Transition(ctx, rec, etcd.AdmissionScheduled)
	ErrContains(t, "invalid admission transition", err)
}

// TestAdmission_ReserveRejectsStaleClaim proves a reservation bound to a claim
// generation that no longer owns the pull is refused (design §514).
func TestAdmission_ReserveRejectsStaleClaim(t *testing.T) {
	store, owner, claim := admissionFixture(t)
	ctx := context.Background()

	// Release the claim so its create revision no longer matches the live owner.
	Ok(t, owner.Release(ctx, claim))

	_, _, err := store.Reserve(ctx, cmdID, claim)
	ErrContains(t, "stale", err)
}
