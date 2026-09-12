// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the recovery-quarantine marker (design §852, §864). A
// snapshot cannot describe locks, plans, or external side effects created after
// its revision, so after a restore every executable Atlantis command is rejected
// across the deployment until an operator reconciles state. Clearing quarantine
// is an explicit authenticated, audited compare-and-swap operation; it is never
// timer-driven or automatic.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	quarantineKind = "recovery-quarantine"
	quarantineV1   = 1
)

// quarantineRecord marks the deployment as under recovery quarantine.
type quarantineRecord struct {
	RecoveryID string    `json:"recovery_id"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

// QuarantineStore reads and writes the single recovery-quarantine marker.
type QuarantineStore struct {
	kv   clientv3.KV
	keys Keyspace
}

// NewQuarantineStore constructs the quarantine store.
func NewQuarantineStore(backend Backend, keys Keyspace) *QuarantineStore {
	return &QuarantineStore{kv: backend.Client().KV, keys: keys}
}

// Set installs the quarantine marker (create-only) as part of recovery.
func (q *QuarantineStore) Set(ctx context.Context, recoveryID, reason string) error {
	val, err := encodeValue(quarantineKind, quarantineV1, quarantineRecord{
		RecoveryID: recoveryID, Reason: reason, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	key := q.keys.RecoveryQuarantineKey()
	resp, err := q.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(val))).
		Commit()
	if err != nil {
		return fmt.Errorf("setting recovery quarantine: %w", err)
	}
	if !resp.Succeeded {
		return errors.New("recovery quarantine is already active")
	}
	return nil
}

// IsActive reports whether recovery quarantine is in effect. Every executable
// command admission checks this and rejects while active (design §788).
func (q *QuarantineStore) IsActive(ctx context.Context) (bool, error) {
	resp, err := q.kv.Get(ctx, q.keys.RecoveryQuarantineKey())
	if err != nil {
		return false, fmt.Errorf("checking recovery quarantine: %w", err)
	}
	return len(resp.Kvs) > 0, nil
}

// Clear removes the quarantine marker. It is an explicit compare-and-swap on the
// exact recovery ID so a stale clear cannot lift a newer quarantine (design §868).
func (q *QuarantineStore) Clear(ctx context.Context, recoveryID string) error {
	key := q.keys.RecoveryQuarantineKey()
	get, err := q.kv.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("reading recovery quarantine: %w", err)
	}
	if len(get.Kvs) == 0 {
		return nil
	}
	var rec quarantineRecord
	if err := decodeValue(get.Kvs[0].Value, quarantineKind, quarantineV1, quarantineV1, &rec); err != nil {
		return err
	}
	if rec.RecoveryID != recoveryID {
		return fmt.Errorf("recovery quarantine belongs to recovery %q, not %q", rec.RecoveryID, recoveryID)
	}
	resp, err := q.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", get.Kvs[0].ModRevision)).
		Then(clientv3.OpDelete(key)).
		Commit()
	if err != nil {
		return fmt.Errorf("clearing recovery quarantine: %w", err)
	}
	if !resp.Succeeded {
		return errors.New("recovery quarantine changed before it could be cleared")
	}
	return nil
}
