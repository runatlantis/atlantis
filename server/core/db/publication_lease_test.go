// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/stretchr/testify/require"
)

func TestPublicationLease_ExpiryRenewalAndOwnership(t *testing.T) {
	now := time.Unix(1000, 0)
	first, err := db.AcquirePublicationLease(nil, "first", now, time.Minute)
	require.NoError(t, err)
	_, err = db.AcquirePublicationLease(&first, "second", now, time.Minute)
	require.ErrorIs(t, err, db.ErrPublicationBusy)
	require.ErrorIs(t, db.CanReleasePublicationLease(&first, db.PublicationFence{Owner: "wrong"}), db.ErrPublicationOwnerLost)
	renewed, err := db.RenewPublicationLease(&first, db.PublicationFence{Owner: "first"}, now.Add(30*time.Second), time.Minute)
	require.NoError(t, err)
	_, err = db.AcquirePublicationLease(&renewed, "second", now.Add(time.Minute), time.Minute)
	require.ErrorIs(t, err, db.ErrPublicationBusy)
	replacement, err := db.AcquirePublicationLease(&renewed, "second", time.UnixMilli(renewed.DeadlineUnixMilli), time.Minute)
	require.NoError(t, err)
	require.Equal(t, "second", replacement.Owner)
	require.ErrorIs(t, db.CanReleasePublicationLease(&replacement, db.PublicationFence{Owner: "first"}), db.ErrPublicationOwnerLost)
	_, err = db.RenewPublicationLease(&renewed, db.PublicationFence{Owner: "first"}, time.UnixMilli(renewed.DeadlineUnixMilli), time.Minute)
	require.ErrorIs(t, err, db.ErrPublicationOwnerLost)
	require.NoError(t, db.CanReleasePublicationLease(&replacement, db.PublicationFence{Owner: "second"}))
	require.ErrorIs(t, db.CheckPublicationFence(&replacement, db.PublicationFence{}, now), db.ErrPublicationOwnerLost)
}

func TestPublicationLease_AmbiguousExternalRequestCannotBeFencedByExpiry(t *testing.T) {
	now := time.Unix(1000, 0)
	fence := db.PublicationFence{Owner: "owner"}
	held, err := db.AcquirePublicationLease(nil, fence.Owner, now, time.Minute)
	require.NoError(t, err)
	publishing, err := db.BeginPublication(&held, fence, now)
	require.NoError(t, err)
	require.ErrorIs(t, db.CanReleasePublicationLease(&publishing, fence), db.ErrPublicationAmbiguous)
	_, err = db.AcquirePublicationLease(&publishing, "replacement", time.UnixMilli(held.DeadlineUnixMilli), time.Minute)
	require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
	_, err = db.CompletePublication(&publishing, db.PublicationFence{Owner: "wrong"})
	require.ErrorIs(t, err, db.ErrPublicationOwnerLost)
	resolved, err := db.CompletePublication(&publishing, fence)
	require.NoError(t, err)
	// Knowing the response permits cleanup, not further calls under an expired owner.
	require.NoError(t, db.CanReleasePublicationLease(&resolved, fence))
	_, err = db.BeginPublication(&resolved, fence, time.UnixMilli(held.DeadlineUnixMilli))
	require.ErrorIs(t, err, db.ErrPublicationOwnerLost)
	replacement, err := db.AcquirePublicationLease(&resolved, "replacement", time.UnixMilli(held.DeadlineUnixMilli), time.Minute)
	require.NoError(t, err)
	require.Equal(t, "replacement", replacement.Owner)
}

func TestPublicationLease_PreExecutionFailureCanReleaseAfterExpiry(t *testing.T) {
	now := time.Unix(1000, 0)
	held, err := db.AcquirePublicationLease(nil, "owner", now, time.Minute)
	require.NoError(t, err)
	// Pending-status and validation failures never mark terminal publication.
	require.NoError(t, db.CanReleasePublicationLease(&held, db.PublicationFence{Owner: "owner"}))
	replacement, err := db.AcquirePublicationLease(&held, "replacement", time.UnixMilli(held.DeadlineUnixMilli).Add(time.Second), time.Minute)
	require.NoError(t, err)
	require.ErrorIs(t, db.CanReleasePublicationLease(&replacement, db.PublicationFence{Owner: "owner"}), db.ErrPublicationOwnerLost)
}

func TestPublicationLease_DurableWritesRequireExplicitMode(t *testing.T) {
	now := time.Unix(1000, 0)
	lease, err := db.AcquirePublicationLease(nil, "owner", now, time.Minute)
	require.NoError(t, err)
	require.ErrorIs(t, db.CheckPublicationWrite(&lease, nil, now), db.ErrPublicationOwnerLost)
	require.ErrorIs(t, db.CheckPublicationWrite(&lease, command.NoClaim{}, now), db.ErrPublicationBusy)
	require.NoError(t, db.CheckPublicationWrite(&lease, command.PublicationFence{Owner: "owner"}, now))
	require.ErrorIs(t, db.CheckPublicationWrite(&lease, command.PublicationFence{Owner: "wrong"}, now), db.ErrPublicationOwnerLost)
	expired := time.UnixMilli(lease.DeadlineUnixMilli)
	require.NoError(t, db.CheckPublicationWrite(&lease, command.NoClaim{}, expired))
	lease.Publishing = true
	require.ErrorIs(t, db.CheckPublicationWrite(&lease, command.NoClaim{}, expired), db.ErrPublicationAmbiguous)
}
