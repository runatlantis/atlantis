// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"sort"
	"testing"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	. "github.com/runatlantis/atlantis/testing"
)

type fakeMasterScanner struct {
	masters []*redislib.Client
}

func (f fakeMasterScanner) ForEachMaster(ctx context.Context, fn func(context.Context, *redislib.Client) error) error {
	for _, master := range f.masters {
		if err := fn(ctx, master); err != nil {
			return err
		}
	}
	return nil
}

func TestScanKeysUnionsAllMasters(t *testing.T) {
	s1 := miniredis.RunT(t)
	s2 := miniredis.RunT(t)
	c1 := redislib.NewClient(&redislib.Options{Addr: s1.Addr()})
	c2 := redislib.NewClient(&redislib.Options{Addr: s2.Addr()})
	t.Cleanup(func() {
		c1.Close()
		c2.Close()
	})

	Ok(t, s1.Set("pr/owner/repo/a/default/p", "1"))
	Ok(t, s2.Set("pr/owner/repo/b/default/p", "2"))

	keys, err := scanKeys(context.Background(), fakeMasterScanner{masters: []*redislib.Client{c1, c2}}, "pr/*")
	Ok(t, err)
	sort.Strings(keys)
	Equals(t, []string{
		"pr/owner/repo/a/default/p",
		"pr/owner/repo/b/default/p",
	}, keys)
}
