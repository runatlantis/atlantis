// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Paginated scans pin every page to the first response revision so a concurrent
// write cannot cause an entry to be seen twice or skipped (design §454).
package etcd

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// rangePageSize bounds each page so a large namespace scan respects etcd
// response limits.
const rangePageSize = 512

// mvccKV is a minimal view of an etcd key/value with its mod revision, passed to
// rangePinned visitors.
type mvccKV struct {
	Key         []byte
	Value       []byte
	ModRevision int64
}

// rangePinned scans every key under prefix, calling visit for each. All pages
// after the first are read WithRev(firstRevision) so the scan is a consistent
// snapshot. visit must not retain the mvccKV beyond the call.
func (s *scopedLockStore) rangePinned(ctx context.Context, prefix string, visit func(*mvccKV) error) error {
	return rangePinned(ctx, s.kv, prefix, visit)
}

func rangePinned(ctx context.Context, kv clientv3.KV, prefix string, visit func(*mvccKV) error) error {
	var rev int64
	key := prefix
	end := clientv3.GetPrefixRangeEnd(prefix)
	for {
		opts := []clientv3.OpOption{
			clientv3.WithRange(end),
			clientv3.WithLimit(rangePageSize),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		}
		if rev != 0 {
			opts = append(opts, clientv3.WithRev(rev))
		}
		resp, err := kv.Get(ctx, key, opts...)
		if err != nil {
			return fmt.Errorf("paginated range read: %w", err)
		}
		// Pin all subsequent pages to the header revision of the first page.
		if rev == 0 {
			rev = resp.Header.Revision
		}
		for _, kvp := range resp.Kvs {
			view := &mvccKV{Key: kvp.Key, Value: kvp.Value, ModRevision: kvp.ModRevision}
			if err := visit(view); err != nil {
				return err
			}
		}
		if !resp.More {
			return nil
		}
		// Continue after the last returned key. Append a NUL so the range start
		// is strictly greater than the last key.
		last := resp.Kvs[len(resp.Kvs)-1].Key
		key = string(append(append([]byte{}, last...), 0))
	}
}
