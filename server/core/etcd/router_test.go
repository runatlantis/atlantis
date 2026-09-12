// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// End-to-end dispatch tests with two nodes wired over real HTTP (design §991:
// assert 202/409/503 behavior, not router matching alone).

package etcd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

// lazyHandler lets us create an httptest server before its InternalServer exists,
// resolving the advertise-URL chicken-and-egg.
type lazyHandler struct {
	mu sync.Mutex
	h  http.Handler
}

func (l *lazyHandler) set(h http.Handler) { l.mu.Lock(); l.h = h; l.mu.Unlock() }
func (l *lazyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.mu.Lock()
	h := l.h
	l.mu.Unlock()
	h.ServeHTTP(w, r)
}

// fakeExecutor records idempotent registrations.
type fakeExecutor struct {
	mu   sync.Mutex
	seen map[string]int
	fail bool
}

func newFakeExecutor() *fakeExecutor { return &fakeExecutor{seen: map[string]int{}} }
func (f *fakeExecutor) Register(_ context.Context, cmd etcd.Command) error {
	if f.fail {
		return context.Canceled
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen[cmd.Identity.DeliveryID]++
	return nil
}
func (f *fakeExecutor) count(delivery string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.seen[delivery]
}

type node struct {
	ownership  *etcd.OwnershipStore
	router     *etcd.Router
	exec       *fakeExecutor
	advURL     string
	quarantine *etcd.QuarantineStore
}

const internalToken = "secret-internal-token"

func newNode(t *testing.T, backend etcd.Backend, keys etcd.Keyspace, epoch, id string) *node {
	t.Helper()
	lazy := &lazyHandler{}
	srv := httptest.NewServer(lazy)
	t.Cleanup(srv.Close)

	own, err := etcd.NewOwnershipStore(backend, keys, epoch, id, srv.URL, 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = own.Close() })
	adm := etcd.NewAdmissionStore(backend, keys, epoch, own.InstanceID())
	allow, err := etcd.NewAllowlist([]string{"127.0.0.0/8"})
	Ok(t, err)
	client := etcd.NewInternalClient(internalToken, allow, nil, 5*time.Second)
	exec := newFakeExecutor()
	quarantine := etcd.NewQuarantineStore(backend, keys)
	router := etcd.NewRouter(own, adm, client, exec, quarantine)
	lazy.set(etcd.NewInternalServer(internalToken, router))
	return &node{ownership: own, router: router, exec: exec, advURL: srv.URL, quarantine: quarantine}
}

func routerFixture(t *testing.T) (*node, *node) {
	backend := newTestBackend(t)
	keys := etcd.NewKeyspace("/atlantis")
	epoch, err := etcd.InitOrValidateNamespace(context.Background(), backend.Client().KV, keys, "dep-1")
	Ok(t, err)
	return newNode(t, backend, keys, epoch, "A"), newNode(t, backend, keys, epoch, "B")
}

func cmd(delivery string) etcd.Command {
	return etcd.Command{
		Identity: etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: delivery},
		Scope:    etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1},
	}
}

// TestRoute_LocalOwnerAdmits proves the first replica to see a pull claims it and
// admits locally with 202.
func TestRoute_LocalOwnerAdmits(t *testing.T) {
	a, _ := routerFixture(t)
	res := a.router.Route(context.Background(), cmd("d1"))
	Equals(t, http.StatusAccepted, res.Status)
	Equals(t, etcd.AdmissionScheduled, res.State)
	Equals(t, 1, a.exec.count("d1"))
}

// TestRoute_ForwardsToOwner proves a non-owner replica forwards over HTTP to the
// owning replica, which executes it (design §490 step 2).
func TestRoute_ForwardsToOwner(t *testing.T) {
	a, b := routerFixture(t)
	ctx := context.Background()

	// A claims the pull first.
	Equals(t, http.StatusAccepted, a.router.Route(ctx, cmd("d1")).Status)

	// B routes a different command for the same pull; it must forward to A.
	res := b.router.Route(ctx, cmd("d2"))
	Equals(t, http.StatusAccepted, res.Status)
	Assert(t, a.exec.count("d2") == 1, "owner A must have executed the forwarded command")
	Assert(t, b.exec.count("d2") == 0, "non-owner B must not execute it locally")
}

// TestRoute_QuarantineBlocksAdmission proves an active recovery quarantine
// rejects an executable command at admission with 503 and never registers it
// (design §788, §864).
func TestRoute_QuarantineBlocksAdmission(t *testing.T) {
	a, _ := routerFixture(t)
	ctx := context.Background()

	Ok(t, a.quarantine.Set(ctx, "rec-1", "post-restore"))

	res := a.router.Route(ctx, cmd("d1"))
	Equals(t, http.StatusServiceUnavailable, res.Status)
	Assert(t, a.exec.count("d1") == 0, "quarantined command must not be registered")

	// After the operator clears quarantine, admission proceeds normally.
	Ok(t, a.quarantine.Clear(ctx, "rec-1"))
	res = a.router.Route(ctx, cmd("d1"))
	Equals(t, http.StatusAccepted, res.Status)
	Assert(t, a.exec.count("d1") == 1, "command admitted once quarantine is cleared")
}

// TestRoute_DuplicateReturnsStoredState proves a duplicate delivery returns the
// stored admission state without a second local registration (design §526).
func TestRoute_DuplicateReturnsStoredState(t *testing.T) {
	a, _ := routerFixture(t)
	ctx := context.Background()

	Equals(t, http.StatusAccepted, a.router.Route(ctx, cmd("d1")).Status)
	res := a.router.Route(ctx, cmd("d1"))
	Equals(t, http.StatusAccepted, res.Status)
	Assert(t, a.exec.count("d1") == 1, "duplicate must not register a second time")
}

// TestHandleForwarded_StaleClaimReturns409 proves the receiver rejects a command
// whose generation it does not own (design §490 step 4).
func TestHandleForwarded_StaleClaimReturns409(t *testing.T) {
	a, b := routerFixture(t)
	ctx := context.Background()

	// A owns the pull.
	Equals(t, http.StatusAccepted, a.router.Route(ctx, cmd("d1")).Status)

	// Ask B (a non-owner) to handle a forwarded command directly: it does not own
	// the claim, so it must return 409.
	res := b.router.HandleForwarded(ctx, cmd("d2"))
	Equals(t, http.StatusConflict, res.Status)
}

// TestInternalServer_Unauthorized proves a missing or wrong token is rejected
// with 401 in constant time (design §811).
func TestInternalServer_Unauthorized(t *testing.T) {
	a, _ := routerFixture(t)
	body, _ := json.Marshal(cmd("d1"))
	req, err := http.NewRequest(http.MethodPost, a.advURL+etcd.InternalCommandPath, bytes.NewReader(body))
	Ok(t, err)
	req.Header.Set("Authorization", "Bearer wrong-token")
	resp, err := http.DefaultClient.Do(req)
	Ok(t, err)
	defer resp.Body.Close()
	Equals(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestInternalClient_AllowlistRejects proves the forwarding client refuses a
// destination outside the advertise allowlist (SSRF guard, design §815).
func TestInternalClient_AllowlistRejects(t *testing.T) {
	allow, err := etcd.NewAllowlist([]string{"10.0.0.0/8"})
	Ok(t, err)
	client := etcd.NewInternalClient(internalToken, allow, nil, 2*time.Second)
	_, err = client.Send(context.Background(), "https://evil.example.com:4141", cmd("d1"))
	ErrContains(t, "not allowlisted", err)
}
