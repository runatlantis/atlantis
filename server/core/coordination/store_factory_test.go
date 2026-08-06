// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package coordination_test

import (
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/backends"
	"github.com/runatlantis/atlantis/server/core/coordination"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestNewStore_BoltDBBackend(t *testing.T) {
	db, err := bolt.Open(filepath.Join(t.TempDir(), "atlantis.db"), 0600, nil)
	assert.NoError(t, err)
	store, err := coordination.NewStore(logging.NewNoopLogger(t), &backends.BoltDBBackend{DB: db})
	assert.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	assert.NoError(t, store.Ping())
}

func TestNewStore_UnsupportedBackend(t *testing.T) {
	_, err := coordination.NewStore(logging.NewNoopLogger(t), &backends.S3Backend{})
	assert.ErrorContains(t, err, "the coordination store has no s3 driver")
}
