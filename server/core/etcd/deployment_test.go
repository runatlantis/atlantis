// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

func TestInitOrValidateNamespace_FreshThenValidate(t *testing.T) {
	backend := newTestBackend(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	epoch1, err := etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-1")
	Ok(t, err)
	Assert(t, epoch1 != "", "fresh init must produce a coordination epoch")

	// A second call validates the existing namespace and returns the same epoch.
	epoch2, err := etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-1")
	Ok(t, err)
	Equals(t, epoch1, epoch2)
}

func TestInitOrValidateNamespace_WrongDeploymentID(t *testing.T) {
	backend := newTestBackend(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	_, err := etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-1")
	Ok(t, err)
	_, err = etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-OTHER")
	ErrContains(t, "belongs to deployment", err)
}

// TestInitOrValidateNamespace_DataWithoutSchema proves startup refuses a
// namespace holding data but no schema marker (design §635).
func TestInitOrValidateNamespace_DataWithoutSchema(t *testing.T) {
	backend := newTestBackend(t)
	kv := backend.Client().KV
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	// Write a stray data key under the namespace with no schema marker.
	_, err := kv.Put(ctx, keys.Root()+"/db/stray", "x")
	Ok(t, err)

	_, err = etcd.InitOrValidateNamespace(ctx, kv, keys, "dep-1")
	ErrContains(t, "no Atlantis schema marker", err)
}
