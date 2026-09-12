// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// runtimeConfig builds an insecure-dev external Config pointed at the embedded
// test etcd, with ownership/internal settings populated.
func runtimeConfig(t *testing.T, backend etcd.Backend, replicaID, advertiseURL string) *etcd.Config {
	t.Helper()
	return &etcd.Config{
		Mode:             etcd.ModeExternal,
		DeploymentID:     "dep-1",
		Namespace:        "/atlantis",
		RequestTimeout:   5 * time.Second,
		StartupTimeout:   30 * time.Second,
		AllowInsecureDev: true,
		Endpoints:        backend.Client().Endpoints(),
		Ownership: etcd.OwnershipConfig{
			ReplicaID:                 replicaID,
			ReplicaAdvertiseURL:       advertiseURL,
			ReplicaAdvertiseAllowlist: []string{"127.0.0.0/8"},
			TTL:                       10 * time.Second,
		},
	}
}

// TestRuntime_AssemblesAndServesDatabase proves NewRuntime builds a working
// db.Database over the full stack and becomes ready.
func TestRuntime_AssemblesAndServesDatabase(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()

	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })

	Ok(t, rt.Ready(ctx))

	dbase := rt.Database()
	Assert(t, dbase != nil, "runtime must expose a database")

	// The database round-trips a global lock.
	_, err = dbase.LockCommand(command.Apply, time.Now())
	Ok(t, err)
	got, err := dbase.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, got != nil, "global lock should be present via the runtime database")
}

// TestRuntime_EndToEndDispatch wires two runtimes over real HTTP and proves a
// command routed at a non-owner forwards to and executes on the owner.
func TestRuntime_EndToEndDispatch(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()

	// Two runtimes share the same cluster; each mounts its internal handler.
	lazyA, lazyB := &lazyHandler{}, &lazyHandler{}
	srvA := httptest.NewServer(lazyA)
	srvB := httptest.NewServer(lazyB)
	t.Cleanup(srvA.Close)
	t.Cleanup(srvB.Close)

	rtA, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", srvA.URL))
	Ok(t, err)
	t.Cleanup(func() { _ = rtA.Close() })
	rtB, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "B", srvB.URL))
	Ok(t, err)
	t.Cleanup(func() { _ = rtB.Close() })

	execA, execB := newFakeExecutor(), newFakeExecutor()
	lazyA.set(rtA.AttachExecutor(execA))
	lazyB.set(rtB.AttachExecutor(execB))

	// A claims the pull by routing first.
	res, err := rtA.Route(ctx, cmd("d1"))
	Ok(t, err)
	Equals(t, http.StatusAccepted, res.Status)

	// B routes another command for the same pull: it must forward to A.
	res, err = rtB.Route(ctx, cmd("d2"))
	Ok(t, err)
	Equals(t, http.StatusAccepted, res.Status)
	Assert(t, execA.count("d2") == 1, "owner A must execute the forwarded command")
	Assert(t, execB.count("d2") == 0, "non-owner B must not execute locally")
}

// TestRuntime_MigratedNamespaceReady proves a runtime validates a namespace that
// was populated by the offline migrator and serves it.
func TestRuntime_MigratedNamespaceReady(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()

	m, err := etcd.BeginMigration(ctx, backend.Client().KV, keys, "dep-1", "src", 1, "")
	Ok(t, err)
	lock := projectLock("github.com", "o/r", ".", "default", 1)
	Ok(t, m.ImportProjectLock(ctx, etcd.ProjectScope{VCSHostname: "github.com", Repository: "o/r", Path: ".", Workspace: "default"}, lock))
	_, err = m.Complete(ctx)
	Ok(t, err)

	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	Ok(t, rt.Ready(ctx))

	// The migrated lock is visible through the runtime database.
	locks, err := rt.Database().List()
	Ok(t, err)
	Equals(t, 1, len(locks))
	Equals(t, "o/r", locks[0].Project.RepoFullName)
	_ = models.ProjectLock{}
}
