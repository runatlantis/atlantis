// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the runtime boundary between Atlantis and etcd (design
// §"Etcd runtime/backend"). The Backend owns the long-lived client and, in
// embedded mode, the embedded server. Consumers borrow the client and never
// close it directly; the database is the designated close delegate.
package etcd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Backend is the runtime boundary. The backend owns the client and embedded
// runtime; the database is its designated close delegate and the ownership
// adapter borrows the client (design §394).
type Backend interface {
	// Client returns the shared long-lived client. Callers must not close it.
	Client() *clientv3.Client
	// Ready performs a bounded linearizable probe proving cluster authority.
	Ready(context.Context) error
	// Close is idempotent; it closes the client and, in embedded mode, the
	// embedded server.
	Close() error
}

// backend is the shared object returned by both external and embedded runtimes.
// A single close guard makes every shutdown and partial-startup path mutually
// idempotent (design §99, §751).
type backend struct {
	client         *clientv3.Client
	requestTimeout time.Duration
	probeKey       string

	// embeddedClose, when set, stops the embedded server after the client has
	// closed. It is nil in external mode. Set by the embedded runtime (Phase 5).
	embeddedClose func() error

	closeOnce sync.Once
	closeErr  error
}

// Client implements Backend.
func (b *backend) Client() *clientv3.Client { return b.client }

// Ready performs a bounded linearizable range read. clientv3.New is nonblocking,
// so successful client construction is not connectivity proof; a real
// linearizable operation is required (design §187, §303 "Ping").
func (b *backend) Ready(ctx context.Context) error {
	rctx, cancel := context.WithTimeout(ctx, b.requestTimeout)
	defer cancel()
	// A linearizable Get (the clientv3 default, no WithSerializable) forces a
	// quorum read and fails closed if quorum is unavailable.
	if _, err := b.client.Get(rctx, b.probeKey); err != nil {
		return fmt.Errorf("etcd linearizable readiness probe failed: %w", err)
	}
	return nil
}

// Close implements Backend. It is safe to call from multiple shutdown and
// partial-startup paths; only the first call performs the work.
func (b *backend) Close() error {
	b.closeOnce.Do(func() {
		var errs []error
		if b.client != nil {
			if err := b.client.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing etcd client: %w", err))
			}
		}
		if b.embeddedClose != nil {
			if err := b.embeddedClose(); err != nil {
				errs = append(errs, fmt.Errorf("closing embedded etcd server: %w", err))
			}
		}
		b.closeErr = errors.Join(errs...)
	})
	return b.closeErr
}

// NewExternal constructs the external-mode backend: it builds one client over
// all configured endpoints and performs a bounded linearizable probe before
// returning. A failed probe closes the client so no resource leaks (design
// §703 "External runtime").
func NewExternal(ctx context.Context, cfg *Config) (Backend, error) {
	if cfg.Mode != ModeExternal {
		return nil, fmt.Errorf("NewExternal called with mode %q", cfg.Mode)
	}
	if err := cfg.ValidateDatabase(); err != nil {
		return nil, err
	}
	client, err := newClient(cfg, cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	b := &backend{
		client:         client,
		requestTimeout: cfg.RequestTimeout,
		probeKey:       NewKeyspace(cfg.Namespace).Root(),
	}
	pctx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	if err := b.Ready(pctx); err != nil {
		_ = b.Close()
		return nil, err
	}
	return b, nil
}

// newClient constructs a single long-lived clientv3.Client from the resolved
// config. TLS is mandatory in production; the insecure path is reachable only
// under AllowInsecureDev, which Validate has already restricted to loopback.
func newClient(cfg *Config, endpoints []string) (*clientv3.Client, error) {
	clientCfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: cfg.RequestTimeout,
		// Fail closed rather than blocking indefinitely on a dead cluster.
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: cfg.RequestTimeout,
	}

	tlsCfg, err := buildTLS(cfg)
	if err != nil {
		return nil, err
	}
	clientCfg.TLS = tlsCfg

	if cfg.Username != "" {
		pw, err := readSecretFile(cfg.PasswordFile)
		if err != nil {
			return nil, fmt.Errorf("reading etcd-password-file: %w", err)
		}
		clientCfg.Username = cfg.Username
		clientCfg.Password = pw
	}

	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("constructing etcd client: %w", err)
	}
	return client, nil
}

// buildTLS returns the client TLS config, or nil when insecure development mode
// is active and no TLS material was supplied. There is no InsecureSkipVerify
// path in production (design §804).
func buildTLS(cfg *Config) (*tls.Config, error) {
	if !cfg.TLS.configured() {
		if cfg.AllowInsecureDev {
			return nil, nil
		}
		return nil, errors.New("production etcd requires client TLS")
	}
	info := transport.TLSInfo{
		TrustedCAFile: cfg.TLS.CAFile,
		CertFile:      cfg.TLS.CertFile,
		KeyFile:       cfg.TLS.KeyFile,
		ServerName:    cfg.TLS.ServerName,
	}
	tlsCfg, err := info.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("building etcd client TLS: %w", err)
	}
	return tlsCfg, nil
}
