// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package boltdb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	bolt "go.etcd.io/bbolt"
)

var publicationLeasesBucket = []byte("publication-leases")

// mutatePublicationLease checks cancellation inside the transaction too: a
// caller cancelled while waiting for Bolt's writer lock cannot acquire later.
func (b *BoltDB) mutatePublicationLease(ctx context.Context, pull models.PullRequest, transition func(*db.PublicationLease, time.Time) (*db.PublicationLease, error)) (db.PublicationLease, error) {
	key, err := b.pullKey(pull)
	if err != nil {
		return db.PublicationLease{}, err
	}
	if err := ctx.Err(); err != nil {
		return db.PublicationLease{}, err
	}
	type result struct {
		lease db.PublicationLease
		err   error
	}
	completed := make(chan result, 1)
	go func() {
		var next *db.PublicationLease
		err := b.db.Update(func(tx *bolt.Tx) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			bucket, err := tx.CreateBucketIfNotExists(publicationLeasesBucket)
			if err != nil {
				return err
			}
			var current *db.PublicationLease
			if raw := bucket.Get(key); raw != nil {
				current = &db.PublicationLease{}
				if err := json.Unmarshal(raw, current); err != nil {
					return err
				}
			}
			next, err = transition(current, time.Now())
			if err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if next == nil {
				return bucket.Delete(key)
			}
			raw, err := json.Marshal(next)
			if err != nil {
				return err
			}
			return bucket.Put(key, raw)
		})
		output := result{err: err}
		if next != nil {
			output.lease = *next
		}
		completed <- output
	}()
	select {
	case result := <-completed:
		return result.lease, result.err
	case <-ctx.Done():
		return db.PublicationLease{}, ctx.Err()
	}
}

func (b *BoltDB) AcquirePublicationLease(ctx context.Context, pull models.PullRequest, owner string, ttl time.Duration) (db.PublicationLease, error) {
	return b.mutatePublicationLease(ctx, pull, func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		next, err := db.AcquirePublicationLease(current, owner, now, ttl)
		return &next, err
	})
}
func (b *BoltDB) RenewPublicationLease(ctx context.Context, pull models.PullRequest, fence db.PublicationFence, ttl time.Duration) (db.PublicationLease, error) {
	return b.mutatePublicationLease(ctx, pull, func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		next, err := db.RenewPublicationLease(current, fence, now, ttl)
		return &next, err
	})
}
func (b *BoltDB) BeginPublication(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := b.mutatePublicationLease(ctx, pull, func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		next, err := db.BeginPublication(current, fence, now)
		return &next, err
	})
	return err
}
func (b *BoltDB) CompletePublication(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := b.mutatePublicationLease(ctx, pull, func(current *db.PublicationLease, _ time.Time) (*db.PublicationLease, error) {
		next, err := db.CompletePublication(current, fence)
		return &next, err
	})
	return err
}
func (b *BoltDB) ReleasePublicationLease(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := b.mutatePublicationLease(ctx, pull, func(current *db.PublicationLease, _ time.Time) (*db.PublicationLease, error) {
		return nil, db.CanReleasePublicationLease(current, fence)
	})
	return err
}

var _ db.PublicationLeaseStore = (*BoltDB)(nil)

func checkPublicationWrite(tx *bolt.Tx, key []byte, mode command.PublicationWriteMode) error {
	var current *db.PublicationLease
	if bucket := tx.Bucket(publicationLeasesBucket); bucket != nil {
		if raw := bucket.Get(key); raw != nil {
			current = &db.PublicationLease{}
			if err := json.Unmarshal(raw, current); err != nil {
				return err
			}
		}
	}
	return db.CheckPublicationWrite(current, mode, time.Now())
}

func (b *BoltDB) GetPublicationLease(ctx context.Context, pull models.PullRequest) (*db.PublicationLease, error) {
	key, err := b.pullKey(pull)
	if err != nil {
		return nil, err
	}
	type result struct {
		lease *db.PublicationLease
		err   error
	}
	completed := make(chan result, 1)
	go func() {
		var current *db.PublicationLease
		err := b.db.View(func(tx *bolt.Tx) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			bucket := tx.Bucket(publicationLeasesBucket)
			if bucket == nil {
				return nil
			}
			if raw := bucket.Get(key); raw != nil {
				current = &db.PublicationLease{}
				return json.Unmarshal(raw, current)
			}
			return nil
		})
		completed <- result{current, err}
	}()
	select {
	case result := <-completed:
		return result.lease, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *BoltDB) RecoverPublicationLease(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := b.mutatePublicationLease(ctx, pull, func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		return nil, db.CanRecoverPublicationLease(current, fence, now)
	})
	return err
}

var _ db.PublicationRecoveryStore = (*BoltDB)(nil)
