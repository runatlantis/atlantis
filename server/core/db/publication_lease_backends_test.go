// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	redisdb "github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
)

func TestPublicationLease_BackendParity(t *testing.T) {
	backends := map[string]func(*testing.T) db.PublicationLeaseStore{
		"bolt": func(t *testing.T) db.PublicationLeaseStore {
			storage, err := boltdb.New(t.TempDir())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, storage.Close()) })
			return storage
		},
		"redis": func(t *testing.T) db.PublicationLeaseStore {
			server := miniredis.RunT(t)
			port, err := strconv.Atoi(server.Port())
			require.NoError(t, err)
			storage, err := redisdb.New(server.Host(), port, "", false, false, 0)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, storage.Close()) })
			return storage
		},
	}
	for name, create := range backends {
		t.Run(name, func(t *testing.T) {
			store := create(t)
			ctx := context.Background()
			pull := models.PullRequest{Num: 1, BaseRepo: models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Hostname: "github.com"}}}
			database := store.(db.Database)
			initial, err := database.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "")}, command.NoClaim{})
			require.NoError(t, err)
			fence := db.PublicationFence{Owner: "owner"}
			lease, err := store.AcquirePublicationLease(ctx, pull, fence.Owner, time.Minute)
			require.NoError(t, err)
			require.Equal(t, fence.Owner, lease.Owner)
			for _, mode := range []command.PublicationWriteMode{command.NoClaim{}, command.PublicationFence{Owner: "wrong"}} {
				expected := db.ErrPublicationBusy
				if _, fenced := mode.(command.PublicationFence); fenced {
					expected = db.ErrPublicationOwnerLost
				}
				_, err := database.BeginPlanGeneration(pull, "G1", []command.ProjectContext{generationProject("a")}, false, mode)
				require.ErrorIs(t, err, expected)
				_, err = database.UpdatePullWithResults(pull, nil, mode)
				require.ErrorIs(t, err, expected)
				require.ErrorIs(t, database.UpdateProjectStatus(pull, "default", ".", models.DiscardedPlanStatus, mode), expected)
				_, err = database.DiscardPlanStatus(pull, initial.Projects[0], mode)
				require.ErrorIs(t, err, expected)
				require.ErrorIs(t, database.DeletePullStatus(pull, mode), expected)
			}
			unchanged, err := database.GetPullStatus(pull)
			require.NoError(t, err)
			require.Equal(t, initial, *unchanged)
			_, err = database.BeginPlanGeneration(pull, "G1", []command.ProjectContext{generationProject("a")}, false, fence)
			require.NoError(t, err)
			accepted, err := database.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G1")}, fence)
			require.NoError(t, err)
			require.Equal(t, "G1", accepted.Projects[0].AcceptedPlanGeneration)

			_, err = store.AcquirePublicationLease(ctx, pull, "other", time.Minute)
			require.ErrorIs(t, err, db.ErrPublicationBusy)
			require.ErrorIs(t, store.ReleasePublicationLease(ctx, pull, db.PublicationFence{Owner: "other"}), db.ErrPublicationOwnerLost)
			_, err = store.RenewPublicationLease(ctx, pull, fence, time.Minute)
			require.NoError(t, err)
			require.NoError(t, store.BeginPublication(ctx, pull, fence))
			recovery := store.(db.PublicationRecoveryStore)
			inspected, err := recovery.GetPublicationLease(ctx, pull)
			require.NoError(t, err)
			require.Equal(t, fence.Owner, inspected.Owner)
			require.True(t, inspected.Publishing)
			require.ErrorIs(t, recovery.RecoverPublicationLease(ctx, pull, fence), db.ErrPublicationBusy)
			require.ErrorIs(t, recovery.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: "wrong"}), db.ErrPublicationOwnerLost)
			require.ErrorIs(t, store.ReleasePublicationLease(ctx, pull, fence), db.ErrPublicationAmbiguous)
			_, err = database.UpdatePullWithResults(pull, nil, fence)
			require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
			require.ErrorIs(t, database.DeletePullStatus(pull, fence), db.ErrPublicationAmbiguous)
			require.NoError(t, store.CompletePublication(ctx, pull, fence))
			// Resolving publication never changes accepted durable status.
			afterPublication, err := database.GetPullStatus(pull)
			require.NoError(t, err)
			require.Equal(t, accepted, *afterPublication)
			// Status and lease disappear together, with owner checks enforced.
			require.NoError(t, database.DeletePullStatus(pull, fence))
			removed, err := database.GetPullStatus(pull)
			require.NoError(t, err)
			require.Nil(t, removed)

			_, err = store.AcquirePublicationLease(ctx, pull, "replacement", time.Minute)
			require.NoError(t, err)
			require.ErrorIs(t, store.ReleasePublicationLease(ctx, pull, fence), db.ErrPublicationOwnerLost)
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			_, err = store.AcquirePublicationLease(cancelled, pull, "cancelled", time.Minute)
			require.ErrorIs(t, err, context.Canceled)
		})
	}
}

func TestPublicationLease_RedisExpiryUsesServerClock(t *testing.T) {
	server := miniredis.RunT(t)
	now := time.Unix(1000, 0)
	server.SetTime(now)
	port, err := strconv.Atoi(server.Port())
	require.NoError(t, err)
	store, err := redisdb.New(server.Host(), port, "", false, false, 0)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })
	ctx := context.Background()
	pull := models.PullRequest{Num: 1, BaseRepo: models.Repo{FullName: "owner/repo"}}
	initial, err := store.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "")}, command.NoClaim{})
	require.NoError(t, err)
	lease, err := store.AcquirePublicationLease(ctx, pull, "owner", time.Minute)
	require.NoError(t, err)
	require.Equal(t, now.Add(time.Minute).UnixMilli(), lease.DeadlineUnixMilli)
	server.SetTime(now.Add(time.Minute))
	_, err = store.RenewPublicationLease(ctx, pull, db.PublicationFence{Owner: "owner"}, time.Minute)
	require.ErrorIs(t, err, db.ErrPublicationOwnerLost)
	_, err = store.AcquirePublicationLease(ctx, pull, "replacement", time.Minute)
	require.NoError(t, err)
	require.ErrorIs(t, store.ReleasePublicationLease(ctx, pull, db.PublicationFence{Owner: "owner"}), db.ErrPublicationOwnerLost)
	require.NoError(t, store.BeginPublication(ctx, pull, db.PublicationFence{Owner: "replacement"}))
	server.SetTime(now.Add(2 * time.Minute))
	_, err = store.AcquirePublicationLease(ctx, pull, "third", time.Minute)
	require.ErrorIs(t, err, db.ErrPublicationAmbiguous)
	require.NoError(t, store.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: "replacement"}))
	status, err := store.GetPullStatus(pull)
	require.NoError(t, err)
	require.Equal(t, initial, *status, "recovery must not change accepted status")
	_, err = store.AcquirePublicationLease(ctx, pull, "third", time.Minute)
	require.NoError(t, err)
	require.ErrorIs(t, store.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: "replacement"}), db.ErrPublicationOwnerLost)
}

func TestPublicationLease_RecoveryRequiresExactExpiredPublishingOwner(t *testing.T) {
	now := time.Unix(1000, 0)
	fence := db.PublicationFence{Owner: "original"}
	lease := db.PublicationLease{Owner: fence.Owner, DeadlineUnixMilli: now.UnixMilli(), Publishing: true}
	require.NoError(t, db.CanRecoverPublicationLease(&lease, fence, now))
	require.ErrorIs(t, db.CanRecoverPublicationLease(&lease, db.PublicationFence{Owner: "other"}, now), db.ErrPublicationOwnerLost)
	require.ErrorIs(t, db.CanRecoverPublicationLease(&lease, fence, now.Add(-time.Second)), db.ErrPublicationBusy)
	lease.Publishing = false
	require.ErrorIs(t, db.CanRecoverPublicationLease(&lease, fence, now), db.ErrPublicationRecoveryNotNeeded)
}
