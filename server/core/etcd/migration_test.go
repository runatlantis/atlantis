// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/core/etcd"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// TestMigration_ImportCompleteThenStartupValidates proves an offline migration
// imports records, completes atomically with schema+deployment, and that normal
// namespace validation then succeeds against the migrated namespace (design §895).
func TestMigration_ImportCompleteThenStartupValidates(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	m, err := etcd.BeginMigration(ctx, kv, keys, "dep-1", "boltdb:/data/atlantis.db", 2, "")
	Ok(t, err)

	// While a migration is in progress (no schema yet), normal startup refuses.
	_, err = etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-1")
	ErrContains(t, "no Atlantis schema marker", err)

	lock := projectLock("github.com", "o/r", ".", "default", 1)
	Ok(t, m.ImportProjectLock(ctx, etcd.ProjectScope{VCSHostname: "github.com", Repository: "o/r", Path: ".", Workspace: "default"}, lock))
	Ok(t, m.ImportPullStatus(ctx, etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1},
		models.PullStatus{Pull: lock.Pull}))
	Equals(t, 2, m.Count())

	epoch, err := m.Complete(ctx)
	Ok(t, err)
	Assert(t, epoch != "", "completion yields a coordination epoch")

	// After completion, normal startup validates and returns the same epoch.
	got, err := etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-1")
	Ok(t, err)
	Equals(t, epoch, got)
}

// TestMigration_CountMismatchRefusesComplete proves completion refuses when the
// imported count does not match the manifest (design §895 step 8).
func TestMigration_CountMismatchRefusesComplete(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	m, err := etcd.BeginMigration(ctx, kv, keys, "dep-1", "src", 5, "")
	Ok(t, err)
	lock := projectLock("github.com", "o/r", ".", "default", 1)
	Ok(t, m.ImportProjectLock(ctx, etcd.ProjectScope{VCSHostname: "github.com", Repository: "o/r", Path: ".", Workspace: "default"}, lock))

	_, err = m.Complete(ctx)
	ErrContains(t, "count mismatch", err)
}

// TestMigration_NonEmptyTargetRefused proves migration refuses a non-empty target.
func TestMigration_NonEmptyTargetRefused(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	_, err := kv.Put(ctx, keys.Root()+"/db/stray", "x")
	Ok(t, err)
	_, err = etcd.BeginMigration(ctx, kv, keys, "dep-1", "src", 0, "")
	ErrContains(t, "not empty", err)
}

// TestMigration_ChecksumMismatchRefusesComplete proves a supplied source
// checksum is an integrity gate: completion refuses when the imported content's
// checksum does not match the pinned expected value even if the count matches
// (design §897 step 8).
func TestMigration_ChecksumMismatchRefusesComplete(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	m, err := etcd.BeginMigration(ctx, kv, keys, "dep-1", "src", 1, "deadbeefdeadbeef")
	Ok(t, err)
	lock := projectLock("github.com", "o/r", ".", "default", 1)
	Ok(t, m.ImportProjectLock(ctx, etcd.ProjectScope{VCSHostname: "github.com", Repository: "o/r", Path: ".", Workspace: "default"}, lock))
	Equals(t, 1, m.Count()) // count matches...

	_, err = m.Complete(ctx)
	ErrContains(t, "checksum mismatch", err) // ...but the checksum gate still refuses
}

// TestMigration_SecondMigrationRefused proves a second migration cannot begin
// while one is already in progress: the migration-active sentinel (and the
// non-empty namespace) serialize migrations, closing the concurrent
// empty-namespace TOCTOU (design §895).
func TestMigration_SecondMigrationRefused(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	_, err := etcd.BeginMigration(ctx, kv, keys, "dep-1", "src", 1, "")
	Ok(t, err)

	// A second BeginMigration must be refused rather than writing a parallel manifest.
	_, err = etcd.BeginMigration(ctx, kv, keys, "dep-1", "src", 1, "")
	Assert(t, err != nil, "a second concurrent migration must be refused")
}

// TestQuarantine_SetBlocksThenClear proves quarantine is set once, reported
// active, and cleared only with the exact recovery ID (design §864).
func TestQuarantine_SetBlocksThenClear(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	keys := etcd.NewKeyspace("/atlantis")
	q := etcd.NewQuarantineStore(backend, keys)
	ctx := context.Background()

	active, err := q.IsActive(ctx)
	Ok(t, err)
	Assert(t, !active, "no quarantine initially")

	Ok(t, q.Set(ctx, "rec-1", "post-restore"))
	active, err = q.IsActive(ctx)
	Ok(t, err)
	Assert(t, active, "quarantine active after set")

	// A wrong recovery ID cannot clear it.
	ErrContains(t, "belongs to recovery", q.Clear(ctx, "rec-WRONG"))

	Ok(t, q.Clear(ctx, "rec-1"))
	active, err = q.IsActive(ctx)
	Ok(t, err)
	Assert(t, !active, "quarantine cleared")
}
