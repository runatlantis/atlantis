// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

// startEmbeddedEtcd starts a single-node in-process etcd on loopback purely as a
// test fixture and returns a probed external backend pointed at it. The embed
// package is a test-only dependency here; the shipped binary connects to an
// external etcd cluster and does not link it.
func startEmbeddedEtcd(t *testing.T) etcd.Backend {
	t.Helper()
	cfg := embed.NewConfig()
	cfg.Dir = t.TempDir()
	cfg.LogLevel = "error"
	clientURL := mustURL(t, "http://127.0.0.1:0")
	peerURL := mustURL(t, "http://127.0.0.1:0")
	cfg.ListenClientUrls = []url.URL{clientURL}
	cfg.AdvertiseClientUrls = []url.URL{clientURL}
	cfg.ListenPeerUrls = []url.URL{peerURL}
	cfg.AdvertisePeerUrls = []url.URL{peerURL}
	cfg.InitialCluster = fmt.Sprintf("%s=%s", cfg.Name, peerURL.String())

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("starting embedded etcd: %v", err)
	}
	select {
	case <-e.Server.ReadyNotify():
	case <-time.After(30 * time.Second):
		e.Close()
		t.Fatal("embedded etcd did not become ready")
	}
	t.Cleanup(e.Close)

	// The listener chose a real port; read it back for the client endpoint.
	endpoint := e.Clients[0].Addr().String()
	cfgEtcd := &etcd.Config{
		Mode:             etcd.ModeExternal,
		DeploymentID:     "test",
		Namespace:        "/atlantis",
		RequestTimeout:   5 * time.Second,
		StartupTimeout:   30 * time.Second,
		AllowInsecureDev: true,
		Endpoints:        []string{"http://" + endpoint},
		Ownership: etcd.OwnershipConfig{
			ReplicaID:                 "test-replica",
			ReplicaAdvertiseURL:       "http://127.0.0.1:4141",
			ReplicaAdvertiseAllowlist: []string{"127.0.0.0/8"},
			TTL:                       30 * time.Second,
		},
	}
	backend, err := etcd.NewExternal(context.Background(), cfgEtcd)
	if err != nil {
		t.Fatalf("constructing external backend: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })
	return backend
}

func mustURL(t *testing.T, raw string) url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parsing url %q: %v", raw, err)
	}
	return *u
}

// newRawClient returns a second independent client against the same cluster, to
// simulate contention from a different Atlantis process.
func newRawClient(t *testing.T, backend etcd.Backend) *clientv3.Client {
	t.Helper()
	// Reuse the backend's endpoints via a fresh client so operations race as
	// they would across processes.
	eps := backend.Client().Endpoints()
	c, err := clientv3.New(clientv3.Config{Endpoints: eps, DialTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("second client: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}
