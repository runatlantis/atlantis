// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

// TestCoordinator_BeginAPIFence_FencesAndReleases proves a positive-PR API
// execution on the owner is fenced by an execution barrier and released cleanly,
// so a subsequent fence for the same pull is permitted again (design §534).
func TestCoordinator_BeginAPIFence_FencesAndReleases(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	// This replica must own the pull (ResolveOwner claims it).
	local, _, err := coord.ResolveOwner(ctx, "github.com", "o/r", 1)
	Ok(t, err)
	Assert(t, local, "owner should be local")

	release, err := coord.BeginAPIFence(ctx, "github.com", "o/r", 1)
	Ok(t, err)
	Assert(t, release != nil, "a fenced owner execution returns a release func")
	release()

	// After release the barrier is cleared, so a new fence is permitted.
	release2, err := coord.BeginAPIFence(ctx, "github.com", "o/r", 1)
	Ok(t, err)
	release2()
}

// TestCoordinator_BeginAPIFence_BlockedByOlderGeneration proves the API path is
// fenced: it refuses to run while an unresolved barrier from an older owner
// generation exists for the pull (design §558).
func TestCoordinator_BeginAPIFence_BlockedByOlderGeneration(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()
	epoch, err := etcd.InitOrValidateNamespace(ctx, backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	scope := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1}

	// An older generation leaves an active barrier, then its owner dies (claim
	// released); barriers are persistent, so the fence remains.
	old, err := etcd.NewOwnershipStore(backend, keys, epoch, "old", "http://127.0.0.1:4999", 10*time.Second, 5*time.Second)
	Ok(t, err)
	oldClaim, won, err := old.Claim(ctx, scope)
	Ok(t, err)
	Assert(t, won, "old generation should win the claim")
	barriers := etcd.NewExecutionBarrierStore(backend, keys, epoch, old.InstanceID())
	_, err = barriers.StartStep(ctx, oldClaim, "old-exec")
	Ok(t, err)
	Ok(t, old.Close()) // release the claim; the barrier persists

	// A new owner claims the now-free pull and attempts a fenced API execution.
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)
	local, _, err := coord.ResolveOwner(ctx, "github.com", "o/r", 1)
	Ok(t, err)
	Assert(t, local, "new owner should be local")

	_, err = coord.BeginAPIFence(ctx, "github.com", "o/r", 1)
	ErrContains(t, "previous server generation", err)
}

// TestCoordinator_BeginAPIFence_OwnershipMoved proves the API fence fails closed
// when this replica no longer owns the pull (design §490 step 4).
func TestCoordinator_BeginAPIFence_OwnershipMoved(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()
	epoch, err := etcd.InitOrValidateNamespace(ctx, backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	// Another replica owns the pull.
	other, err := etcd.NewOwnershipStore(backend, keys, epoch, "B", "http://127.0.0.1:4999", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = other.Close() })
	_, won, err := other.Claim(ctx, etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1})
	Ok(t, err)
	Assert(t, won, "other replica should own the pull")

	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	_, err = coord.BeginAPIFence(ctx, "github.com", "o/r", 1)
	ErrContains(t, "moved to another replica", err)
}

// TestCoordinator_ResolveOwner proves the first replica to resolve a pull claims
// it (local), and a second replica resolves to the first's advertise URL.
func TestCoordinator_ResolveOwner(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()

	rtA, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "https://replica-a:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rtA.Close() })
	rtB, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "B", "https://replica-b:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rtB.Close() })

	coordA := etcd.NewRuntimeCoordinator(rtA)
	coordB := etcd.NewRuntimeCoordinator(rtB)

	local, advertise, err := coordA.ResolveOwner(ctx, "github.com", "o/r", 1)
	Ok(t, err)
	Assert(t, local, "A should own the pull it resolved first")
	Equals(t, "", advertise)

	local, advertise, err = coordB.ResolveOwner(ctx, "github.com", "o/r", 1)
	Ok(t, err)
	Assert(t, !local, "B must not own A's pull")
	Equals(t, "https://replica-a:4142", advertise)
}

// TestCoordinator_ForwardAPIRequest proves the forwarder posts the body to
// advertiseURL+path with the API token and loop-guard header, and returns the
// owner's status and body.
func TestCoordinator_ForwardAPIRequest(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()

	var gotToken, gotProxied, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Atlantis-Token")
		gotProxied = r.Header.Get(etcd.InternalProxiedHeader)
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	status, body, err := coord.ForwardAPIRequest(ctx, srv.URL, "/api/plan", "sekret", []byte(`{"Repository":"o/r"}`))
	Ok(t, err)
	Equals(t, http.StatusOK, status)
	Equals(t, `{"ok":true}`, string(body))
	Equals(t, "sekret", gotToken)
	Equals(t, "1", gotProxied)
	Equals(t, "/api/plan", gotPath)
	Equals(t, `{"Repository":"o/r"}`, gotBody)
}

// TestCoordinator_ForwardAPIRequest_AllowlistRejects proves a non-allowlisted
// destination host is refused before dialing.
func TestCoordinator_ForwardAPIRequest_AllowlistRejects(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	_, _, err = coord.ForwardAPIRequest(ctx, "https://evil.example.com:4142", "/api/plan", "sekret", []byte(`{}`))
	Assert(t, err != nil, "a non-allowlisted destination must be refused")
}
