// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/runatlantis/atlantis/server/core/planstore"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRedisStore(t *testing.T, ttl time.Duration) (*planstore.RedisPlanStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return planstore.NewRedisPlanStore(client, ttl, logging.NewNoopLogger(t)), mr
}

func redisTestCtx() command.ProjectContext {
	return command.ProjectContext{
		BaseRepo:   models.Repo{Owner: "acme", Name: "infra"},
		Pull:       models.PullRequest{Num: 42, HeadCommit: "abc123"},
		Workspace:  "default",
		RepoRelDir: "modules/vpc",
	}
}

func writePlan(t *testing.T, contents string) string {
	t.Helper()
	planPath := filepath.Join(t.TempDir(), "default.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte(contents), 0o600))
	return planPath
}

func TestRedisSaveLoad_RoundTrip(t *testing.T) {
	store, _ := newRedisStore(t, 0)
	ctx := redisTestCtx()
	require.NoError(t, store.Save(ctx, writePlan(t, "plan-bytes")))

	dst := filepath.Join(t.TempDir(), "sub", "default.tfplan")
	require.NoError(t, store.Load(ctx, dst))
	got, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "plan-bytes", string(got))
}

func TestRedisLoad_MissingPlan(t *testing.T) {
	store, _ := newRedisStore(t, 0)
	err := store.Load(redisTestCtx(), filepath.Join(t.TempDir(), "default.tfplan"))
	assert.ErrorContains(t, err, "plan not found in redis")
}

func TestRedisLoad_StalePlanRejected(t *testing.T) {
	store, _ := newRedisStore(t, 0)
	ctx := redisTestCtx()
	require.NoError(t, store.Save(ctx, writePlan(t, "old")))

	ctx.Pull.HeadCommit = "newcommit"
	err := store.Load(ctx, filepath.Join(t.TempDir(), "default.tfplan"))
	assert.ErrorContains(t, err, "run plan again")
}

func TestRedisRestorePlans_WritesUnderPullDir(t *testing.T) {
	store, _ := newRedisStore(t, 0)
	ctx := redisTestCtx()
	require.NoError(t, store.Save(ctx, writePlan(t, "plan-bytes")))

	pullDir := t.TempDir()
	require.NoError(t, store.RestorePlans(pullDir, "acme", "infra", 42))
	got, err := os.ReadFile(filepath.Join(pullDir, "default", "modules/vpc", "default.tfplan"))
	require.NoError(t, err)
	assert.Equal(t, "plan-bytes", string(got))
}

func TestRedisRestorePlans_EmptyPullDirIsProbe(t *testing.T) {
	store, _ := newRedisStore(t, 0)
	assert.NoError(t, store.RestorePlans("", "acme", "infra", 42))
}

func TestRedisListWorkspaces_SortedAndDeduped(t *testing.T) {
	store, _ := newRedisStore(t, 0)
	for _, ws := range []string{"staging", "default", "staging"} {
		ctx := redisTestCtx()
		ctx.Workspace = ws
		require.NoError(t, store.Save(ctx, writePlan(t, "p")))
	}
	got, err := store.ListWorkspaces("acme", "infra", 42)
	require.NoError(t, err)
	assert.Equal(t, []string{"default", "staging"}, got)
}

func TestRedisDeleteForPull_RemovesPullHash(t *testing.T) {
	store, mr := newRedisStore(t, 0)
	ctx := redisTestCtx()
	require.NoError(t, store.Save(ctx, writePlan(t, "p")))

	require.NoError(t, store.DeleteForPull("acme", "infra", 42))
	ws, err := store.ListWorkspaces("acme", "infra", 42)
	require.NoError(t, err)
	assert.Empty(t, ws)
	assert.Empty(t, mr.Keys(), "the pull hash should be gone")
}

func TestRedisSave_TTLExpiryBehavior(t *testing.T) {
	withTTL, mrTTL := newRedisStore(t, time.Hour)
	require.NoError(t, withTTL.Save(redisTestCtx(), writePlan(t, "p")))
	for _, k := range mrTTL.Keys() {
		assert.NotZero(t, mrTTL.TTL(k), "key %s should expire when ttl is set", k)
	}

	noTTL, mrNo := newRedisStore(t, 0)
	require.NoError(t, noTTL.Save(redisTestCtx(), writePlan(t, "p")))
	for _, k := range mrNo.Keys() {
		assert.Zero(t, mrNo.TTL(k), "key %s must not expire when ttl is 0", k)
	}
}
