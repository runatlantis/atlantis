// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func newTestCluster(t *testing.T, shardCount int) ([]*miniredis.Miniredis, *redis.ClusterClient) {
	t.Helper()

	shards := make([]*miniredis.Miniredis, shardCount)
	addrs := make([]string, shardCount)
	for i := range shardCount {
		shards[i] = miniredis.RunT(t)
		addrs[i] = shards[i].Addr()
	}

	const totalSlots = 16384
	slotsPerShard := totalSlots / shardCount
	slots := make([]redis.ClusterSlot, shardCount)
	for i := range shardCount {
		start := i * slotsPerShard
		end := (i+1)*slotsPerShard - 1
		if i == shardCount-1 {
			end = totalSlots - 1
		}
		slots[i] = redis.ClusterSlot{
			Start: start,
			End:   end,
			Nodes: []redis.ClusterNode{{
				ID:   fmt.Sprintf("master-%d", i),
				Addr: addrs[i],
			}},
		}
	}

	cc := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addrs,
		ClusterSlots: func(context.Context) ([]redis.ClusterSlot, error) {
			return slots, nil
		},
	})
	t.Cleanup(func() {
		_ = cc.Close()
	})
	cc.ReloadState(context.Background())
	return shards, cc
}

func pickShardProjects(t *testing.T, cc *redis.ClusterClient, shards []*miniredis.Miniredis, prefix string, keyFn func(models.Project) string) [3]models.Project {
	t.Helper()

	var projects [3]models.Project
	var found [3]bool
	for n := 0; !(found[0] && found[1] && found[2]) && n < 200; n++ {
		proj := models.NewProject("owner/repo", fmt.Sprintf("%s-%d", prefix, n), "")
		key := keyFn(proj)
		Ok(t, cc.Set(context.Background(), key, "probe", 0).Err())
		for i, shard := range shards {
			if shard.Exists(key) {
				if !found[i] {
					projects[i] = proj
					found[i] = true
				}
				break
			}
		}
		Ok(t, cc.Del(context.Background(), key).Err())
	}
	Assert(t, found[0] && found[1] && found[2], "expected %s keys spanning all 3 cluster primaries", prefix)
	return projects
}

func TestCluster_ListUnlockByPullAndMigrate(t *testing.T) {
	shards, cc := newTestCluster(t, 3)
	rdb := &RedisDB{client: cc}

	// Pick projects whose lock keys route to each of the 3 primaries.
	projects := pickShardProjects(t, cc, shards, "dir", func(p models.Project) string {
		return rdb.lockKey(p, "default")
	})

	for i, proj := range projects {
		acquired, _, err := rdb.TryLock(models.ProjectLock{
			Project:   proj,
			Workspace: "default",
			Pull:      models.PullRequest{Num: 42},
			User:      models.User{Username: "alice"},
			Time:      time.Now(),
		})
		Ok(t, err)
		Assert(t, acquired, "expected lock %d to be acquired", i)
		Assert(t, len(shards[i].Keys()) == 1, "expected shard %d to hold 1 lock key", i)
	}

	// Also create a lock belonging to another PR.
	otherProj := models.NewProject("owner/repo", "other-dir", "")
	acquired, _, err := rdb.TryLock(models.ProjectLock{
		Project:   otherProj,
		Workspace: "default",
		Pull:      models.PullRequest{Num: 99},
		User:      models.User{Username: "bob"},
		Time:      time.Now(),
	})
	Ok(t, err)
	Assert(t, acquired, "expected lock for other PR to be acquired")

	listed, err := rdb.List()
	Ok(t, err)
	Equals(t, 4, len(listed))

	unlocked, err := rdb.UnlockByPull("owner/repo", 42)
	Ok(t, err)
	Equals(t, 3, len(unlocked))

	remaining, err := rdb.List()
	Ok(t, err)
	Equals(t, 1, len(remaining))
	Equals(t, 99, remaining[0].Pull.Num)

	// Clean up remaining lock and verify legacy migration spans all primaries too.
	_, err = rdb.Unlock(otherProj, "default")
	Ok(t, err)

	legacyProjects := pickShardProjects(t, cc, shards, "legacy", func(p models.Project) string {
		return fmt.Sprintf("pr/%s/%s/default", p.RepoFullName, p.Path)
	})
	for i, proj := range legacyProjects {
		oldKey := fmt.Sprintf("pr/%s/%s/default", proj.RepoFullName, proj.Path)
		payload, err := json.Marshal(models.ProjectLock{
			Project:   proj,
			Workspace: "default",
			Pull:      models.PullRequest{Num: 77},
			User:      models.User{Username: "carol"},
			Time:      time.Now(),
		})
		Ok(t, err)
		Ok(t, cc.Set(context.Background(), oldKey, string(payload), 0).Err())
		Assert(t, shards[i].Exists(oldKey), "expected legacy key %q on shard %d", oldKey, i)
	}

	Ok(t, migrateOldLockKeys(context.Background(), cc))

	migrated, err := rdb.List()
	Ok(t, err)
	Equals(t, 3, len(migrated))
	for i, proj := range legacyProjects {
		oldKey := fmt.Sprintf("pr/%s/%s/default", proj.RepoFullName, proj.Path)
		Assert(t, !shards[i].Exists(oldKey), "expected old key %q on shard %d to be removed", oldKey, i)
		got, err := rdb.GetLock(proj, "default")
		Ok(t, err)
		Assert(t, got != nil, "expected migrated lock for %s", proj.Path)
		Equals(t, 77, got.Pull.Num)
	}
}

type hijackOnEvalHook struct {
	mr       *miniredis.Miniredis
	key      string
	newValue string
	swapped  atomic.Bool
}

func (h *hijackOnEvalHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *hijackOnEvalHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if strings.EqualFold(cmd.Name(), "eval") && !h.swapped.Swap(true) {
			_ = h.mr.Set(h.key, h.newValue)
		}
		return next(ctx, cmd)
	}
}

func (h *hijackOnEvalHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestUnlockByPull_SkipsReassignedLock(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	t.Cleanup(func() {
		_ = client.Close()
	})
	rdb := &RedisDB{client: client}

	proj := models.NewProject("owner/repo", "app", "")
	_, _, err := rdb.TryLock(models.ProjectLock{
		Project:   proj,
		Workspace: "default",
		Pull:      models.PullRequest{Num: 42},
		User:      models.User{Username: "alice"},
		Time:      time.Now(),
	})
	Ok(t, err)

	reassigned, err := json.Marshal(models.ProjectLock{
		Project:   proj,
		Workspace: "default",
		Pull:      models.PullRequest{Num: 99},
		User:      models.User{Username: "bob"},
		Time:      time.Now(),
	})
	Ok(t, err)

	client.AddHook(&hijackOnEvalHook{
		mr:       s,
		key:      rdb.lockKey(proj, "default"),
		newValue: string(reassigned),
	})

	unlocked, err := rdb.UnlockByPull("owner/repo", 42)
	Ok(t, err)
	Equals(t, 0, len(unlocked))

	stillLocked, err := rdb.GetLock(proj, "default")
	Ok(t, err)
	Assert(t, stillLocked != nil, "expected reassigned lock not to be deleted")
	Equals(t, 99, stillLocked.Pull.Num)
}
