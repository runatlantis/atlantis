// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the embedded etcd runtime (design §"Embedded runtime").
// It validates the embedded configuration and data-directory state for the
// selected lifecycle, starts embed.Etcd, waits for server readiness with a
// bounded timeout, constructs the shared client, probes the cluster, and returns
// a Backend that owns both the client and the embedded server.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

// NewEmbedded validates the embedded configuration and lifecycle, starts the
// embedded etcd server, and returns a Backend. It fails startup before serving
// public traffic on any inconsistency (design §785).
func NewEmbedded(ctx context.Context, cfg *Config) (Backend, error) {
	if cfg.Mode != ModeEmbedded {
		return nil, fmt.Errorf("NewEmbedded called with mode %q", cfg.Mode)
	}
	if err := cfg.ValidateDatabase(); err != nil {
		return nil, err
	}

	fileCfg, err := ParseEmbeddedConfig(cfg.Embedded.ConfigFile)
	if err != nil {
		return nil, err
	}
	if err := fileCfg.Validate(cfg.Embedded.Lifecycle, cfg.Embedded.VoterCount, cfg.AllowInsecureDev); err != nil {
		return nil, err
	}
	if err := validateDataDirState(cfg.Embedded.Lifecycle, fileCfg.DataDir); err != nil {
		return nil, err
	}
	// Validate the declared identity manifest, membership ticket, or restore
	// manifest for the lifecycle before starting the server, cross-checking the
	// data directory's persisted identity on restart (design §716 step 1, §785).
	if err := validateLifecycleIdentity(cfg, fileCfg); err != nil {
		return nil, err
	}

	embedCfg, err := buildEmbedConfig(fileCfg, cfg.Embedded.Lifecycle)
	if err != nil {
		return nil, err
	}

	e, err := embed.StartEtcd(embedCfg)
	if err != nil {
		return nil, fmt.Errorf("starting embedded etcd: %w", err)
	}

	// Wait for readiness, a fatal error, cancellation, or the startup timeout.
	select {
	case <-e.Server.ReadyNotify():
	case err := <-e.Err():
		e.Close()
		return nil, fmt.Errorf("embedded etcd failed during startup: %w", err)
	case <-ctx.Done():
		e.Close()
		return nil, ctx.Err()
	case <-time.After(cfg.StartupTimeout):
		e.Close()
		return nil, fmt.Errorf("embedded etcd did not become ready within %s", cfg.StartupTimeout)
	}

	client, err := newEmbeddedClient(cfg, fileCfg, e)
	if err != nil {
		e.Close()
		return nil, err
	}

	b := &backend{
		client:         client,
		requestTimeout: cfg.RequestTimeout,
		probeKey:       NewKeyspace(cfg.Namespace).Root(),
		embeddedClose: func() error {
			e.Close()
			select {
			case <-e.Server.StopNotify():
			case <-time.After(30 * time.Second):
				return errors.New("embedded etcd did not stop within 30s")
			}
			return nil
		},
	}

	pctx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	if err := b.Ready(pctx); err != nil {
		_ = b.Close()
		return nil, err
	}
	return b, nil
}

// validateDataDirState enforces the lifecycle's data-directory precondition
// (design §221-234). Atlantis never derives the lifecycle from directory
// emptiness; it validates the directory against the declared lifecycle.
func validateDataDirState(lifecycle Lifecycle, dataDir string) error {
	empty, err := dirEmpty(dataDir)
	if err != nil {
		return err
	}
	switch lifecycle {
	case LifecycleBootstrap, LifecycleJoinExisting:
		if !empty {
			return fmt.Errorf("%s mode requires an empty data directory, but %q is not empty", lifecycle, dataDir)
		}
	case LifecycleRestart, LifecycleRestore:
		if empty {
			return fmt.Errorf("%s mode requires a non-empty data directory, but %q is empty or missing", lifecycle, dataDir)
		}
	}
	return nil
}

// dirEmpty reports whether path is missing or contains no entries.
func dirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("inspecting data directory: %w", err)
	}
	return len(entries) == 0, nil
}

// buildEmbedConfig maps the validated file config into an embed.Config, forcing
// safe defaults: strict reconfiguration checking on, initial corruption check on,
// fsync on, and the cluster state derived from the lifecycle (design §282).
func buildEmbedConfig(f *EmbeddedFileConfig, lifecycle Lifecycle) (*embed.Config, error) {
	c := embed.NewConfig()
	c.Name = f.Name
	c.Dir = f.DataDir
	c.LogLevel = "warn"

	var err error
	if c.ListenClientUrls, err = parseURLs(f.ListenClientURLs); err != nil {
		return nil, err
	}
	if c.AdvertiseClientUrls, err = parseURLs(f.AdvertiseClientURLs); err != nil {
		return nil, err
	}
	if c.ListenPeerUrls, err = parseURLs(f.ListenPeerURLs); err != nil {
		return nil, err
	}
	if c.AdvertisePeerUrls, err = parseURLs(f.InitialAdvertisePeerURLs); err != nil {
		return nil, err
	}
	c.InitialCluster = f.InitialCluster
	c.InitialClusterToken = f.InitialClusterToken
	c.ClusterState = initialClusterState(lifecycle)

	// Safe, non-negotiable defaults (design §282).
	c.StrictReconfigCheck = true
	c.ExperimentalInitialCorruptCheck = true
	c.ForceNewCluster = false

	c.ClientTLSInfo = tlsInfo(f.ClientTLS)
	c.PeerTLSInfo = tlsInfo(f.PeerTLS)
	return c, nil
}

func tlsInfo(t EmbeddedTLS) transport.TLSInfo {
	return transport.TLSInfo{
		TrustedCAFile:  t.CAFile,
		CertFile:       t.CertFile,
		KeyFile:        t.KeyFile,
		ClientCertAuth: t.ClientCertAuth,
	}
}

func parseURLs(raw []string) ([]url.URL, error) {
	out := make([]url.URL, 0, len(raw))
	for _, r := range raw {
		u, err := url.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("parsing url %q: %w", r, err)
		}
		out = append(out, *u)
	}
	return out, nil
}

// newEmbeddedClient builds the shared client against the embedded server's bound
// client listeners.
func newEmbeddedClient(cfg *Config, f *EmbeddedFileConfig, e *embed.Etcd) (*clientv3.Client, error) {
	scheme := "http"
	if f.ClientTLS.CertFile != "" {
		scheme = "https"
	}
	endpoints := make([]string, 0, len(e.Clients))
	for _, l := range e.Clients {
		endpoints = append(endpoints, scheme+"://"+l.Addr().String())
	}
	if len(endpoints) == 0 {
		return nil, errors.New("embedded etcd exposed no client listeners")
	}

	clientCfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: cfg.RequestTimeout,
	}
	if f.ClientTLS.CertFile != "" {
		info := transport.TLSInfo{
			TrustedCAFile: f.ClientTLS.CAFile,
			CertFile:      cfg.TLS.CertFile,
			KeyFile:       cfg.TLS.KeyFile,
			ServerName:    cfg.TLS.ServerName,
		}
		tc, err := info.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("building embedded client TLS: %w", err)
		}
		clientCfg.TLS = tc
	}
	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("constructing embedded etcd client: %w", err)
	}
	return client, nil
}
