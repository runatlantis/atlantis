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

// TestAdmissionCleanup_DeletesOnlyExpiredTerminal proves the dedup-window cleaner
// removes terminal records past the window, keeps terminal records within the
// window, and never removes a non-terminal record.
func TestAdmissionCleanup_DeletesOnlyExpiredTerminal(t *testing.T) {
	backend := newTestBackend(t)
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()
	epoch, err := etcd.InitOrValidateNamespace(ctx, backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	own, err := etcd.NewOwnershipStore(backend, keys, epoch, "A", "http://127.0.0.1:4142", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = own.Close() })
	adm := etcd.NewAdmissionStore(backend, keys, epoch, own.InstanceID())

	scope := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1}
	claim, won, err := own.Claim(ctx, scope)
	Ok(t, err)
	Assert(t, won, "expected to win the claim")

	// A terminal (succeeded) record.
	termID := etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: "terminal"}
	rec, _, err := adm.Reserve(ctx, termID, claim)
	Ok(t, err)
	rec, err = adm.Transition(ctx, rec, etcd.AdmissionScheduled)
	Ok(t, err)
	rec, err = adm.Transition(ctx, rec, etcd.AdmissionRunning)
	Ok(t, err)
	_, err = adm.Complete(ctx, rec, true, "", "")
	Ok(t, err)

	// A non-terminal (scheduled) record.
	liveID := etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: "live"}
	live, _, err := adm.Reserve(ctx, liveID, claim)
	Ok(t, err)
	_, err = adm.Transition(ctx, live, etcd.AdmissionScheduled)
	Ok(t, err)

	// Within the window: nothing is deleted.
	deleted, err := adm.CleanupExpired(ctx, time.Now())
	Ok(t, err)
	Equals(t, 0, deleted)

	// Past the window: only the terminal record is deleted.
	deleted, err = adm.CleanupExpired(ctx, time.Now().Add(etcd.DedupWindow+time.Hour))
	Ok(t, err)
	Equals(t, 1, deleted)

	gotTerm, err := adm.Get(ctx, termID)
	Ok(t, err)
	Assert(t, gotTerm == nil, "expired terminal record should be deleted")

	gotLive, err := adm.Get(ctx, liveID)
	Ok(t, err)
	Assert(t, gotLive != nil, "non-terminal record must never be cleaned")
}
