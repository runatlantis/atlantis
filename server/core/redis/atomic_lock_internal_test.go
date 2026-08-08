// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
)

type getBarrierClient struct {
	redislib.Cmdable
	getCalls   atomic.Int32
	setNXCalls atomic.Int32
	ready      chan struct{}
}

func (c *getBarrierClient) Get(ctx context.Context, key string) *redislib.StringCmd {
	cmd := c.Cmdable.Get(ctx, key)
	if c.setNXCalls.Load() != 0 {
		return cmd
	}
	if c.getCalls.Add(1) == 2 {
		close(c.ready)
	}
	<-c.ready
	return cmd
}

func (c *getBarrierClient) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redislib.BoolCmd {
	c.setNXCalls.Add(1)
	return c.Cmdable.SetNX(ctx, key, value, expiration)
}

type lockResult struct {
	acquired bool
	err      error
}

func TestRedisDB_TryLockHasOneConcurrentWinner(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	database := &RedisDB{client: &getBarrierClient{Cmdable: client, ready: make(chan struct{})}}
	lock := projectLockForPull(41)

	start := make(chan struct{})
	results := make(chan lockResult, 2)
	for range 2 {
		go func() {
			<-start
			acquired, _, err := database.TryLock(lock)
			results <- lockResult{acquired: acquired, err: err}
		}()
	}
	close(start)

	winners := 0
	for range 2 {
		result := <-results
		require.NoError(t, result.err)
		if result.acquired {
			winners++
		}
	}
	require.Equal(t, 1, winners)
}

func TestRedisDB_LockCommandHasOneConcurrentWinner(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	database := &RedisDB{client: &getBarrierClient{Cmdable: client, ready: make(chan struct{})}}

	start := make(chan struct{})
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := database.LockCommand(command.Apply, time.Unix(1700000000, 0))
			errs <- err
		}()
	}
	close(start)

	successes := 0
	conflicts := 0
	for range 2 {
		err := <-errs
		switch {
		case err == nil:
			successes++
		case err.Error() == "db transaction failed: lock already exists":
			conflicts++
		default:
			t.Fatalf("unexpected LockCommand error: %v", err)
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
}

type replaceAfterGetClient struct {
	redislib.Cmdable
	test        *testing.T
	once        sync.Once
	replacement []byte
}

type multiMasterClient struct {
	redislib.Cmdable
	masters []*redislib.Client
}

func (c *multiMasterClient) ForEachMaster(ctx context.Context, fn func(context.Context, *redislib.Client) error) error {
	for _, master := range c.masters {
		if err := fn(ctx, master); err != nil {
			return err
		}
	}
	return nil
}

func (c *multiMasterClient) Get(ctx context.Context, key string) *redislib.StringCmd {
	for _, master := range c.masters {
		cmd := master.Get(ctx, key)
		if cmd.Err() != redislib.Nil {
			return cmd
		}
	}
	return c.Cmdable.Get(ctx, key)
}

func (c *multiMasterClient) Eval(ctx context.Context, script string, keys []string, args ...any) *redislib.Cmd {
	if len(keys) > 0 {
		for _, master := range c.masters {
			if master.Exists(ctx, keys[0]).Val() > 0 {
				return master.Eval(ctx, script, keys, args...)
			}
		}
	}
	return c.Cmdable.Eval(ctx, script, keys, args...)
}

func (c *replaceAfterGetClient) Get(ctx context.Context, key string) *redislib.StringCmd {
	cmd := c.Cmdable.Get(ctx, key)
	c.once.Do(func() {
		require.NoError(c.test, c.Cmdable.Set(ctx, key, c.replacement, 0).Err())
	})
	return cmd
}

func projectLockForPull(pullNum int) models.ProjectLock {
	return models.ProjectLock{
		Project:   models.NewProject("owner/repo", "env/prod", "prod"),
		Workspace: "default",
		Pull:      models.PullRequest{Num: pullNum},
	}
}

func TestRedisDB_UnlockByPullDoesNotDeleteReplacementLock(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	oldLock := projectLockForPull(41)
	replacement := projectLockForPull(42)
	key := (&RedisDB{}).lockKey(oldLock.Project, oldLock.Workspace)
	oldJSON, err := json.Marshal(oldLock)
	require.NoError(t, err)
	replacementJSON, err := json.Marshal(replacement)
	require.NoError(t, err)
	require.NoError(t, client.Set(context.Background(), key, oldJSON, 0).Err())
	database := &RedisDB{client: &replaceAfterGetClient{
		Cmdable: client, test: t, replacement: replacementJSON,
	}}

	removed, err := database.UnlockByPull(oldLock.Project.RepoFullName, oldLock.Pull.Num)
	require.NoError(t, err)
	require.Empty(t, removed)
	current, err := database.GetLock(replacement.Project, replacement.Workspace)
	require.NoError(t, err)
	require.NotNil(t, current)
	require.Equal(t, replacement.Pull.Num, current.Pull.Num)
}

func TestRedisDB_UnlockByPullScansEveryClusterMaster(t *testing.T) {
	firstRedis := miniredis.RunT(t)
	secondRedis := miniredis.RunT(t)
	first := redislib.NewClient(&redislib.Options{Addr: firstRedis.Addr()})
	second := redislib.NewClient(&redislib.Options{Addr: secondRedis.Addr()})
	t.Cleanup(func() {
		require.NoError(t, first.Close())
		require.NoError(t, second.Close())
	})

	database := &RedisDB{client: &multiMasterClient{
		Cmdable: first,
		masters: []*redislib.Client{first, second},
	}}
	locks := []models.ProjectLock{
		projectLockForPull(41),
		{
			Project:   models.NewProject("owner/repo", "env/stage", "stage"),
			Workspace: "default",
			Pull:      models.PullRequest{Num: 41},
		},
	}
	for i, lock := range locks {
		serialized, err := json.Marshal(lock)
		require.NoError(t, err)
		require.NoError(t, database.client.(*multiMasterClient).masters[i].Set(
			context.Background(), database.lockKey(lock.Project, lock.Workspace), serialized, 0,
		).Err())
	}

	removed, err := database.UnlockByPull("owner/repo", 41)
	require.NoError(t, err)
	require.ElementsMatch(t, locks, removed)
	require.Equal(t, int64(0), first.DBSize(context.Background()).Val())
	require.Equal(t, int64(0), second.DBSize(context.Background()).Val())
}
