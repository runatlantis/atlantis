// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// Both callers observe the same initial record before either may commit.
// Subsequent reads are not blocked, so the losing CAS must retry against the
// winner's state instead of overwriting it.
type simultaneousPullReads struct {
	goredis.Cmdable
	reads atomic.Int32
	ready chan struct{}
}

func (c *simultaneousPullReads) Get(ctx context.Context, key string) *goredis.StringCmd {
	result := c.Cmdable.Get(ctx, key)
	n := c.reads.Add(1)
	if n == 2 {
		close(c.ready)
	}
	if n <= 2 {
		select {
		case <-c.ready:
		case <-ctx.Done():
			return goredis.NewStringResult("", ctx.Err())
		}
	}
	return result
}

func TestPlanGeneration_ConcurrentAdmissionRetriesCAS(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { Ok(t, client.Close()) })
	reads := &simultaneousPullReads{Cmdable: client, ready: make(chan struct{})}
	database := &RedisDB{client: reads}
	pull := models.PullRequest{Num: 1, HeadCommit: "same-head", BaseRepo: models.Repo{FullName: "owner/repo"}}
	results := make(chan error, 2)
	for _, name := range []string{"a", "b"} {
		go func() {
			_, err := database.BeginPlanGeneration(pull, name, []command.ProjectContext{{Workspace: "default", RepoRelDir: ".", ProjectName: name}}, false)
			results <- err
		}()
	}
	Ok(t, <-results)
	Ok(t, <-results)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, 2, len(status.Projects))
	generations := map[string]string{}
	for _, project := range status.Projects {
		generations[project.ProjectName] = project.PlanGeneration
	}
	Equals(t, map[string]string{"a": "a", "b": "b"}, generations)
	Assert(t, reads.reads.Load() >= 4, "one admission must reread after losing CAS")
}
