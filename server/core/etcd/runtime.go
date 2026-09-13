// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file assembles the full etcd runtime (design §"Startup and readiness").
// It constructs the shared backend (external or embedded), initializes or
// validates the namespace and coordination epoch, and builds the database,
// ownership, admission, barrier, and quarantine adapters over one client. The
// owner-routing dispatch layer is attached once the command executor exists.
package etcd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"go.etcd.io/etcd/client/pkg/v3/transport"
)

// Runtime is the assembled etcd backend for one Atlantis process. It owns the
// coordination stores and drives readiness and shutdown.
type Runtime struct {
	cfg     *Config
	keys    Keyspace
	backend Backend

	// maintenance is true for an embedded maintenance-purpose start: the server
	// is up for operator tooling but no adapters exist and readiness stays false.
	maintenance bool

	database  *EtcdDatabase
	ownership *OwnershipStore
	admission *AdmissionStore
	barriers  *ExecutionBarrierStore
	quarant   *QuarantineStore
	epoch     string

	internalToken string
	internalTLS   *tls.Config
	allowlist     *Allowlist

	router         *Router
	internalServer *InternalServer

	// cleanerCancel stops the background admission dedup-window cleaner; cleanerDone
	// is closed when that goroutine has exited.
	cleanerCancel context.CancelFunc
	cleanerDone   chan struct{}
}

// admissionCleanupInterval is how often the dedup-window cleaner sweeps terminal
// command-admission records. The retention window itself is DedupWindow (24h);
// sweeping hourly keeps the keyspace bounded without frequent full scans.
const admissionCleanupInterval = time.Hour

// NewRuntime builds the backend and, for a serving process, the full
// coordination stack. In embedded maintenance mode it returns a runtime that
// exposes only the backend and never becomes ready (design §731 step 6).
func NewRuntime(ctx context.Context, cfg *Config) (*Runtime, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	keys := NewKeyspace(cfg.Namespace)

	backend, err := buildBackend(ctx, cfg)
	if err != nil {
		return nil, err
	}

	rt := &Runtime{cfg: cfg, keys: keys, backend: backend}

	if cfg.Mode == ModeEmbedded && cfg.Embedded.StartupPurpose == PurposeMaintenance {
		rt.maintenance = true
		return rt, nil
	}

	if err := rt.buildAdapters(ctx); err != nil {
		_ = backend.Close()
		return nil, err
	}
	return rt, nil
}

func buildBackend(ctx context.Context, cfg *Config) (Backend, error) {
	switch cfg.Mode {
	case ModeExternal:
		return NewExternal(ctx, cfg)
	case ModeEmbedded:
		return NewEmbedded(ctx, cfg)
	default:
		return nil, fmt.Errorf("unknown etcd mode %q", cfg.Mode)
	}
}

// buildAdapters initializes the namespace and constructs the coordination stores
// and the process ownership session for the validated epoch.
func (rt *Runtime) buildAdapters(ctx context.Context) error {
	rctx, cancel := context.WithTimeout(ctx, rt.cfg.StartupTimeout)
	defer cancel()

	epoch, err := InitOrValidateNamespace(rctx, rt.backend.Client().KV, rt.keys, rt.cfg.DeploymentID)
	if err != nil {
		return err
	}
	rt.epoch = epoch

	rt.database = NewDatabase(rt.backend, rt.cfg.Namespace, rt.cfg.RequestTimeout)
	rt.quarant = NewQuarantineStore(rt.backend, rt.keys)

	ownership, err := NewOwnershipStore(rt.backend, rt.keys, epoch,
		rt.cfg.Ownership.ReplicaID, rt.cfg.Ownership.ReplicaAdvertiseURL,
		rt.cfg.Ownership.TTL, rt.cfg.RequestTimeout)
	if err != nil {
		return err
	}
	rt.ownership = ownership
	rt.admission = NewAdmissionStore(rt.backend, rt.keys, epoch, ownership.InstanceID())
	rt.barriers = NewExecutionBarrierStore(rt.backend, rt.keys, epoch, ownership.InstanceID())

	allowlist, err := NewAllowlist(rt.cfg.Ownership.ReplicaAdvertiseAllowlist)
	if err != nil {
		return err
	}
	rt.allowlist = allowlist

	if err := rt.loadInternalTransportSecurity(); err != nil {
		return err
	}

	rt.startAdmissionCleaner()
	return nil
}

// startAdmissionCleaner launches the background goroutine that periodically
// compare-and-swap deletes terminal command-admission records older than the
// dedup window. It is stopped by Close.
func (rt *Runtime) startAdmissionCleaner() {
	ctx, cancel := context.WithCancel(context.Background())
	rt.cleanerCancel = cancel
	rt.cleanerDone = make(chan struct{})
	go func() {
		defer close(rt.cleanerDone)
		ticker := time.NewTicker(admissionCleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCtx, runCancel := context.WithTimeout(ctx, 2*time.Minute)
				// Best-effort: the package has no logger, and a failed sweep is
				// retried on the next tick. Records are only ever deleted under an
				// exact-revision guard, so a partial sweep is always safe.
				_, _ = rt.admission.CleanupExpired(runCtx, time.Now())
				runCancel()
			}
		}
	}()
}

// loadInternalTransportSecurity reads the internal command token and internal CA.
// In insecure development the token may be empty and TLS is nil.
func (rt *Runtime) loadInternalTransportSecurity() error {
	o := rt.cfg.Ownership
	if o.InternalCommandTokenFile != "" {
		tok, err := readSecretFile(o.InternalCommandTokenFile)
		if err != nil {
			return fmt.Errorf("reading internal-command-token-file: %w", err)
		}
		rt.internalToken = tok
	} else if !rt.cfg.AllowInsecureDev {
		return errors.New("internal-command-token-file is required in production etcd mode")
	}
	if o.InternalCommandCAFile != "" {
		info := transport.TLSInfo{TrustedCAFile: o.InternalCommandCAFile}
		tc, err := info.ClientConfig()
		if err != nil {
			return fmt.Errorf("building internal transport TLS: %w", err)
		}
		rt.internalTLS = tc
	}
	return nil
}

// AttachExecutor wires the owner-routing dispatch layer once the local command
// executor exists, returning the internal command HTTP handler to mount. It must
// be called before Route or InternalHandler are used.
func (rt *Runtime) AttachExecutor(executor Executor) http.Handler {
	client := NewInternalClient(rt.internalToken, rt.allowlist, rt.internalTLS, rt.cfg.RequestTimeout)
	rt.router = NewRouter(rt.ownership, rt.admission, client, executor, rt.quarant)
	rt.internalServer = NewInternalServer(rt.internalToken, rt.router)
	return rt.internalServer
}

// Route dispatches an ingress command through owner resolution. AttachExecutor
// must have been called.
func (rt *Runtime) Route(ctx context.Context, cmd Command) (Result, error) {
	if rt.router == nil {
		return Result{}, errors.New("etcd runtime router not attached")
	}
	return rt.router.Route(ctx, cmd), nil
}

// Database returns the db.Database adapter, or nil in maintenance mode.
func (rt *Runtime) Database() db.Database {
	if rt.database == nil {
		return nil
	}
	return rt.database
}

// Ownership exposes the ownership store for lease-loss monitoring.
func (rt *Runtime) Ownership() *OwnershipStore { return rt.ownership }

// Barriers exposes the execution-barrier store for the runner fencing path.
func (rt *Runtime) Barriers() *ExecutionBarrierStore { return rt.barriers }

// Quarantine exposes the recovery-quarantine store for command admission checks.
func (rt *Runtime) Quarantine() *QuarantineStore { return rt.quarant }

// Admission exposes the command-admission store so the owner-side executor can
// advance a record through its running and terminal states after the router has
// reserved and scheduled it (admission.go, design §512).
func (rt *Runtime) Admission() *AdmissionStore { return rt.admission }

// RequestTimeout is the bound on individual coordination RPCs, reused by the
// executor for its fence and admission-lifecycle transactions.
func (rt *Runtime) RequestTimeout() time.Duration { return rt.cfg.RequestTimeout }

// Ready reports readiness: backend authority plus a live ownership session. In
// maintenance mode it is never ready (design §714, §731).
func (rt *Runtime) Ready(ctx context.Context) error {
	if rt.maintenance {
		return errors.New("embedded etcd is in maintenance mode; not serving")
	}
	if err := rt.backend.Ready(ctx); err != nil {
		return err
	}
	select {
	case <-rt.ownership.Done():
		return errors.New("ownership session lease is lost")
	default:
		return nil
	}
}

// Close performs the fixed shutdown order: release ownership while the client is
// available, then close the database, which idempotently closes the backend. In
// maintenance or partial-startup paths it closes the backend directly (design §751).
func (rt *Runtime) Close() error {
	var errs []error
	// Stop the background cleaner before closing the shared client it uses.
	if rt.cleanerCancel != nil {
		rt.cleanerCancel()
		<-rt.cleanerDone
	}
	if rt.ownership != nil {
		if err := rt.ownership.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing ownership session: %w", err))
		}
	}
	if rt.database != nil {
		if err := rt.database.Close(); err != nil {
			errs = append(errs, err)
		}
	} else if rt.backend != nil {
		if err := rt.backend.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
