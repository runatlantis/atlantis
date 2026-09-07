// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package boltdb

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestPublicationRecovery_ExpiredOwnerPreservesAcceptedStatus(t *testing.T) {
	storage, err := New(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "same-head", BaseRepo: models.Repo{FullName: "owner/repo"}}
	project := command.ProjectContext{Workspace: "default", RepoRelDir: ".", ProjectName: "project"}
	_, err = storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{project}, false, command.NoClaim{})
	require.NoError(t, err)
	accepted, err := storage.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Plan, PlanGeneration: "G1", Workspace: "default", RepoRelDir: ".", ProjectName: "project",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}}, command.NoClaim{})
	require.NoError(t, err)

	// Seed a persisted crash record with a deterministic past deadline, without
	// sleeping or depending on scheduler timing to expire a live acquisition.
	crashed := db.PublicationLease{Owner: "crashed", DeadlineUnixMilli: 1, Publishing: true}
	key, err := storage.pullKey(pull)
	require.NoError(t, err)
	raw, err := json.Marshal(crashed)
	require.NoError(t, err)
	require.NoError(t, storage.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(publicationLeasesBucket)
		if err != nil {
			return err
		}
		return bucket.Put(key, raw)
	}))
	ctx := context.Background()
	inspected, err := storage.GetPublicationLease(ctx, pull)
	require.NoError(t, err)
	require.Equal(t, crashed, *inspected)
	require.ErrorIs(t, storage.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: "wrong"}), db.ErrPublicationOwnerLost)
	require.NoError(t, storage.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: "crashed"}))
	removed, err := storage.GetPublicationLease(ctx, pull)
	require.NoError(t, err)
	require.Nil(t, removed)
	_, err = storage.AcquirePublicationLease(ctx, pull, "replacement", time.Minute)
	require.NoError(t, err)
	require.ErrorIs(t, storage.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: "crashed"}), db.ErrPublicationOwnerLost)
	after, err := storage.GetPullStatus(pull)
	require.NoError(t, err)
	require.Equal(t, accepted, *after)
}
