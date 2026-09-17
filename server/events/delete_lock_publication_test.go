// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	redisdb "github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestDeleteLocksByPull_PublicationWithoutProjects(t *testing.T) {
	for _, backend := range []string{"bolt", "redis"} {
		for _, emptyStatus := range []bool{false, true} {
			for _, publishing := range []bool{false, true} {
				t.Run(backend+"/empty="+strconv.FormatBool(emptyStatus)+"/publishing="+strconv.FormatBool(publishing), func(t *testing.T) {
					var storage db.Database
					if backend == "bolt" {
						b, err := boltdb.New(t.TempDir())
						require.NoError(t, err)
						storage = b
					} else {
						server := miniredis.RunT(t)
						port, err := strconv.Atoi(server.Port())
						require.NoError(t, err)
						b, err := redisdb.New(server.Host(), port, "", false, false, 0)
						require.NoError(t, err)
						storage = b
					}
					t.Cleanup(func() { require.NoError(t, storage.Close()) })
					pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "owner/repo"}}
					if emptyStatus {
						_, err := storage.UpdatePullWithResults(pull, nil, command.NoClaim{})
						require.NoError(t, err)
					}
					locker := locking.NewClient(storage)
					lock, err := locker.TryLock(models.NewProject(pull.BaseRepo.FullName, "path", ""), "default", pull, models.User{})
					require.NoError(t, err)
					_, err = storage.AcquirePublicationLease(context.Background(), pull, "other", time.Hour)
					require.NoError(t, err)
					fence := command.PublicationFence{Owner: "other"}
					if publishing {
						require.NoError(t, storage.BeginPublication(context.Background(), pull, fence))
					}
					workingDir := &publicationCleanupWorkingDir{}
					deleter := &events.DefaultDeleteLockCommand{Database: storage, Locker: locker, WorkingDir: workingDir}
					count, err := deleter.DeleteLocksByPull(logging.NewNoopLogger(t), pull, command.NoClaim{})
					expected := db.ErrPublicationBusy
					if publishing {
						expected = db.ErrPublicationAmbiguous
					}
					require.ErrorIs(t, err, expected)
					require.Zero(t, count)
					remaining, err := locker.GetLock(lock.LockKey)
					require.NoError(t, err)
					require.NotNil(t, remaining)
					require.Zero(t, workingDir.deleted)
					if publishing {
						require.NoError(t, storage.CompletePublication(context.Background(), pull, fence))
					}
					require.NoError(t, storage.ReleasePublicationLease(context.Background(), pull, fence))
					count, err = deleter.DeleteLocksByPull(logging.NewNoopLogger(t), pull, command.NoClaim{})
					require.NoError(t, err)
					require.Equal(t, 1, count)
					require.Equal(t, 1, workingDir.deleted)
				})
			}
		}
	}
}

type publicationCleanupWorkingDir struct {
	events.WorkingDir
	deleted int
}

func (w *publicationCleanupWorkingDir) DeletePlan(logging.SimpleLogging, models.Repo, models.PullRequest, string, string, string) error {
	w.deleted++
	return nil
}
