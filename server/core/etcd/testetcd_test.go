// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// testNamespace is the fixed key namespace the integration suite uses. Every
// helper wipes it before a test runs, so a shared external etcd behaves like a
// dedicated one (tests in this package run sequentially, no t.Parallel).
const testNamespace = "/atlantis"

// testEtcdEndpointsEnv names the environment variable that points the etcd
// integration suite at a running external etcd cluster.
const testEtcdEndpointsEnv = "ATLANTIS_ETCD_TEST_ENDPOINTS"

// testEndpoints returns the configured external etcd endpoints, or skips the test
// when none are set. The suite runs against a real (external) etcd — the shipped
// binary connects to one too, and keeping the tests client-only avoids linking
// the embedded etcd server. Provide a comma-separated list, e.g.
// ATLANTIS_ETCD_TEST_ENDPOINTS=http://127.0.0.1:2379. CI runs an etcd service;
// locally, `docker run --rm -p 2379:2379 quay.io/coreos/etcd:v3.6.5 \
// /usr/local/bin/etcd --advertise-client-urls http://0.0.0.0:2379 \
// --listen-client-urls http://0.0.0.0:2379` works.
func testEndpoints(t *testing.T) []string {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(testEtcdEndpointsEnv))
	if raw == "" {
		t.Skipf("%s not set; skipping etcd integration test (point it at a running etcd, e.g. http://127.0.0.1:2379)", testEtcdEndpointsEnv)
	}
	var eps []string
	for _, p := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(p); s != "" {
			eps = append(eps, s)
		}
	}
	return eps
}

// newTestBackend returns a probed external backend against the configured etcd,
// with the test namespace wiped clean so the test sees a fresh keyspace. It skips
// the test when no external etcd is configured (see testEndpoints).
func newTestBackend(t *testing.T) etcd.Backend {
	t.Helper()
	cfg := &etcd.Config{
		Mode:             etcd.ModeExternal,
		DeploymentID:     "test",
		Namespace:        testNamespace,
		RequestTimeout:   5 * time.Second,
		StartupTimeout:   30 * time.Second,
		AllowInsecureDev: true,
		Endpoints:        testEndpoints(t),
		Ownership: etcd.OwnershipConfig{
			ReplicaID:                 "test-replica",
			ReplicaAdvertiseURL:       "http://127.0.0.1:4141",
			ReplicaAdvertiseAllowlist: []string{"127.0.0.0/8"},
			TTL:                       30 * time.Second,
		},
	}
	backend, err := etcd.NewExternal(context.Background(), cfg)
	if err != nil {
		t.Fatalf("constructing external backend: %v", err)
	}
	// Register Close first so it runs last (t.Cleanup is LIFO); wipe before the
	// test for a clean slate — the client must still be open when it runs.
	t.Cleanup(func() { _ = backend.Close() })
	wipeNamespace(t, backend)
	return backend
}

// wipeNamespace deletes every key under the test namespace so each test starts
// from an empty keyspace on the shared external cluster.
func wipeNamespace(t *testing.T, backend etcd.Backend) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := backend.Client().Delete(ctx, testNamespace, clientv3.WithPrefix()); err != nil {
		t.Fatalf("wiping test namespace %q: %v", testNamespace, err)
	}
}
