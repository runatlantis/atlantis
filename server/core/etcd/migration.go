// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements offline migration into an etcd namespace (design
// §"Migration and rollback"). There is no live mixed BoltDB/etcd fleet and no
// dual-write mode. Import uses create-only transactions associated with a
// migration ID; conflicting records abort the migration. Completion verifies an
// exact count and a deterministic checksum, then in one transaction writes the
// schema and deployment/epoch records and marks the migration complete. An
// interrupted import leaves the manifest in progress and normal startup refuses
// the namespace until it is resumed or discarded.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	migrationKind = "migration"
	migrationV1   = 1

	migrationInProgress = "in-progress"
	migrationComplete   = "complete"
)

// migrationManifest is the durable record of a migration (design §895).
type migrationManifest struct {
	MigrationID    string `json:"migration_id"`
	DeploymentID   string `json:"deployment_id"`
	SourceIdentity string `json:"source_identity"`
	ExpectedCount  int    `json:"expected_count"`
	Checksum       string `json:"checksum"`
	Epoch          string `json:"epoch"`
	State          string `json:"state"`
}

// Migrator imports records into a fresh etcd namespace under one migration ID.
type Migrator struct {
	kv        clientv3.KV
	keys      Keyspace
	manifest  migrationManifest
	count     int
	checksum  uint64
	committed bool
}

// BeginMigration validates that the target namespace is genuinely empty and
// writes an in-progress manifest, returning the Migrator. It generates a fresh
// coordination epoch created outside any snapshot (design §895 step 6).
// expectedChecksum, when non-empty, is the operator-supplied deterministic
// checksum of the source records; it is pinned in the in-progress manifest and
// verified at Complete against the checksum computed over what was actually
// imported (design §897). An empty value opts into count-only verification.
func BeginMigration(ctx context.Context, kv clientv3.KV, keys Keyspace, deploymentID, sourceIdentity string, expectedCount int, expectedChecksum string) (*Migrator, error) {
	empty, err := namespaceEmpty(ctx, kv, keys)
	if err != nil {
		return nil, err
	}
	if !empty {
		return nil, errors.New("migration target namespace is not empty")
	}
	m := &Migrator{
		kv:   kv,
		keys: keys,
		manifest: migrationManifest{
			MigrationID:    uuid.NewString(),
			DeploymentID:   deploymentID,
			SourceIdentity: sourceIdentity,
			ExpectedCount:  expectedCount,
			Checksum:       expectedChecksum, // pinned expected; verified at Complete
			Epoch:          uuid.NewString(),
			State:          migrationInProgress,
		},
	}
	val, err := encodeValue(migrationKind, migrationV1, m.manifest)
	if err != nil {
		return nil, err
	}
	key := keys.MigrationKey(m.manifest.MigrationID)
	resp, err := kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(val))).
		Commit()
	if err != nil {
		return nil, fmt.Errorf("writing migration manifest: %w", err)
	}
	if !resp.Succeeded {
		return nil, errors.New("migration manifest already exists")
	}
	return m, nil
}

// MigrationID returns the manifest's migration ID.
func (m *Migrator) MigrationID() string { return m.manifest.MigrationID }

// ImportProjectLock imports one project lock with a create-only transaction. A
// conflicting existing record aborts the migration.
func (m *Migrator) ImportProjectLock(ctx context.Context, scope ProjectScope, lock models.ProjectLock) error {
	val, err := encodeValue(lockRecordKind, lockRecordV1, lock)
	if err != nil {
		return err
	}
	return m.createOnly(ctx, m.keys.ProjectLockKey(scope), val)
}

// ImportPullStatus imports one pull status record.
func (m *Migrator) ImportPullStatus(ctx context.Context, scope PullScope, status models.PullStatus) error {
	val, err := encodeValue(pullStatusKind, pullStatusV1, status)
	if err != nil {
		return err
	}
	return m.createOnly(ctx, m.keys.PullStatusKey(scope), val)
}

// ImportGlobalLock imports one global command lock.
func (m *Migrator) ImportGlobalLock(ctx context.Context, name string, lock command.Lock) error {
	val, err := encodeValue(globalLockKind, globalLockV1, lock)
	if err != nil {
		return err
	}
	return m.createOnly(ctx, m.keys.GlobalLockKey(name), val)
}

// createOnly writes a record only if absent, aborting on any conflict. It updates
// the running count and order-independent checksum.
func (m *Migrator) createOnly(ctx context.Context, key string, val []byte) error {
	resp, err := m.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(val))).
		Commit()
	if err != nil {
		return fmt.Errorf("importing record: %w", err)
	}
	if !resp.Succeeded {
		return fmt.Errorf("migration aborted: record %q already exists", key)
	}
	m.count++
	m.checksum += recordHash(key, val)
	return nil
}

// recordHash is an order-independent per-record contribution to the migration
// checksum.
func recordHash(key string, val []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write(val)
	return h.Sum64()
}

// Checksum returns the deterministic checksum accumulated so far.
func (m *Migrator) Checksum() string { return fmt.Sprintf("%016x", m.checksum) }

// Count returns the number of imported records.
func (m *Migrator) Count() int { return m.count }

// Complete verifies the imported count against the manifest, then atomically
// writes the schema and deployment/epoch records and marks the migration
// complete, all guarded by the in-progress manifest revision (design §895 step 9).
func (m *Migrator) Complete(ctx context.Context) (string, error) {
	if m.committed {
		return m.manifest.Epoch, nil
	}
	if m.count != m.manifest.ExpectedCount {
		return "", fmt.Errorf("migration count mismatch: imported %d, expected %d", m.count, m.manifest.ExpectedCount)
	}
	// Verify content integrity against the pinned expected checksum when supplied
	// (design §897 step 8). An empty pinned checksum means the operator opted into
	// count-only verification.
	computed := m.Checksum()
	if m.manifest.Checksum != "" && computed != m.manifest.Checksum {
		return "", fmt.Errorf("migration checksum mismatch: computed %s, expected %s", computed, m.manifest.Checksum)
	}

	migKey := m.keys.MigrationKey(m.manifest.MigrationID)
	// Read the in-progress manifest to pin its revision.
	get, err := m.kv.Get(ctx, migKey)
	if err != nil {
		return "", fmt.Errorf("reading migration manifest: %w", err)
	}
	if len(get.Kvs) == 0 {
		return "", errors.New("migration manifest disappeared")
	}
	manifestRev := get.Kvs[0].ModRevision

	completed := m.manifest
	completed.State = migrationComplete
	completed.Checksum = computed
	migVal, err := encodeValue(migrationKind, migrationV1, completed)
	if err != nil {
		return "", err
	}
	schemaVal, err := encodeValue(schemaKind, schemaV1, schemaRecord{Product: product, Version: SchemaVersion1})
	if err != nil {
		return "", err
	}
	deployVal, err := encodeValue(deploymentKind, deploymentV1, deploymentRecord{DeploymentID: m.manifest.DeploymentID, Epoch: m.manifest.Epoch})
	if err != nil {
		return "", err
	}

	resp, err := m.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(migKey), "=", manifestRev)).
		Then(
			clientv3.OpPut(migKey, string(migVal)),
			clientv3.OpPut(m.keys.SchemaKey(), string(schemaVal)),
			clientv3.OpPut(m.keys.DeploymentKey(), string(deployVal)),
		).
		Commit()
	if err != nil {
		return "", fmt.Errorf("completing migration: %w", err)
	}
	if !resp.Succeeded {
		return "", errors.New("migration manifest changed before completion")
	}
	m.committed = true
	return m.manifest.Epoch, nil
}
