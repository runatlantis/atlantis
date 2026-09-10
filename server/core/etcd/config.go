// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package etcd implements an opt-in etcd locking and coordination backend for
// Atlantis. This phase supports the external runtime mode (connecting to an
// existing etcd cluster); the embedded in-process voter is deferred to a later
// phase. See docs/superpowers/specs/2026-09-07-dual-mode-embedded-etcd-ha-design.md
// for the architecture this package implements.
//
// This file defines the configuration contract (design §"Configuration
// contract") and its validation. Validation is a pure function over Config so
// it can be exhaustively table-tested without a live cluster.
package etcd

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Mode selects the etcd runtime. This phase supports external only; the embedded
// in-process voter (ModeEmbedded) is deferred to a later phase and rejected at
// validation.
type Mode string

const (
	ModeExternal Mode = "external"
	ModeEmbedded Mode = "embedded"
)

const (
	// MinOwnershipTTL is the floor for the ownership session TTL (design §314).
	MinOwnershipTTL = 10 * time.Second
	// DefaultOwnershipTTL is the default ownership session TTL (design §303).
	DefaultOwnershipTTL = 30 * time.Second
	// DefaultRequestTimeout bounds individual database/ownership/readiness RPCs.
	DefaultRequestTimeout = 5 * time.Second
	// DefaultStartupTimeout bounds initial connectivity and quorum formation.
	DefaultStartupTimeout = 5 * time.Minute
	// DefaultNamespace is the default Atlantis key namespace.
	DefaultNamespace = "/atlantis"
)

// TLSConfig holds a client TLS identity. Production requires a complete trusted
// CA plus a client certificate/key pair (design §164).
type TLSConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	ServerName string
}

// configured reports whether any TLS field was supplied.
func (t TLSConfig) configured() bool {
	return t.CAFile != "" || t.CertFile != "" || t.KeyFile != "" || t.ServerName != ""
}

// complete reports whether a full client identity (CA + cert + key) is present.
func (t TLSConfig) complete() bool {
	return t.CAFile != "" && t.CertFile != "" && t.KeyFile != ""
}

// OwnershipConfig holds active-active PR ownership and routing settings
// (design §292). Selecting locking-db-type=etcd always activates these.
type OwnershipConfig struct {
	ReplicaID                 string
	ReplicaAdvertiseURL       string
	ReplicaAdvertiseAllowlist []string
	InternalCommandTokenFile  string
	InternalCommandCAFile     string
	TTL                       time.Duration
}

// Config is the fully-resolved etcd backend configuration. It is built from the
// Atlantis user config (see FromUserConfig, added with the flag wiring) but is
// decoupled from it so validation can be tested in isolation.
type Config struct {
	Mode             Mode
	DeploymentID     string
	Namespace        string
	TLS              TLSConfig
	Username         string
	PasswordFile     string
	RequestTimeout   time.Duration
	StartupTimeout   time.Duration
	AllowInsecureDev bool

	// External mode.
	Endpoints []string

	// Ownership and routing.
	Ownership OwnershipConfig
}

// Validate enforces the full locked configuration contract, including
// active-active ownership and routing. It fails closed: any missing, ambiguous,
// or insecure setting is an error (design §316, §705, §1053). Use it once
// ownership is wired (Phase 2+).
func (c *Config) Validate() error {
	if err := c.ValidateDatabase(); err != nil {
		return err
	}
	return c.validateOwnership()
}

// ValidateDatabase validates only the client/database-facing settings (common +
// mode-specific), without the ownership contract. It is used to construct the
// backend and database adapter; ownership is validated and established
// separately (design §703 steps 1-5 precede the ownership session in step 6).
func (c *Config) ValidateDatabase() error {
	if err := c.validateCommon(); err != nil {
		return err
	}
	switch c.Mode {
	case ModeExternal:
		return c.validateExternal()
	case ModeEmbedded:
		return errors.New("embedded etcd mode is not available in this release; it is planned for a later phase. Use --etcd-mode=external with an external etcd cluster")
	default:
		return fmt.Errorf("etcd-mode must be %q, got %q", ModeExternal, c.Mode)
	}
}

func (c *Config) validateCommon() error {
	if strings.TrimSpace(c.DeploymentID) == "" {
		return errors.New("etcd-deployment-id is required")
	}
	if strings.TrimSpace(c.Namespace) == "" {
		return errors.New("etcd-namespace is required")
	}
	if c.RequestTimeout <= 0 {
		return errors.New("etcd-request-timeout must be positive")
	}
	if c.StartupTimeout <= 0 {
		return errors.New("etcd-startup-timeout must be positive")
	}
	// A username requires a password file; when supplied, username auth is used
	// while mTLS remains mandatory (design §168).
	if c.Username != "" && c.PasswordFile == "" {
		return errors.New("etcd-username requires etcd-password-file")
	}
	if c.PasswordFile != "" && c.Username == "" {
		return errors.New("etcd-password-file requires etcd-username")
	}
	return c.validateTLS()
}

// validateTLS enforces the production TLS requirement. Insecure (HTTP, no TLS)
// is permitted only under the explicit development flag and, per validateURL,
// only for loopback endpoints (design §176, §804).
func (c *Config) validateTLS() error {
	if c.AllowInsecureDev {
		return nil
	}
	if !c.TLS.complete() {
		return errors.New("production etcd requires a complete client TLS identity (etcd-ca-file, etcd-cert-file, etcd-key-file); set etcd-allow-insecure-dev only for loopback development")
	}
	return nil
}

func (c *Config) validateExternal() error {
	if len(c.Endpoints) == 0 {
		return errors.New("external etcd mode requires etcd-endpoints")
	}
	for _, ep := range c.Endpoints {
		if err := c.validateURL("etcd-endpoints", ep); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) validateOwnership() error {
	o := c.Ownership
	if strings.TrimSpace(o.ReplicaID) == "" {
		return errors.New("replica-id is required (defaults to pod hostname) and must resolve to a stable, unique value")
	}
	if o.ReplicaAdvertiseURL == "" {
		return errors.New("replica-advertise-url is required in etcd mode")
	}
	if err := c.validateURL("replica-advertise-url", o.ReplicaAdvertiseURL); err != nil {
		return err
	}
	if len(o.ReplicaAdvertiseAllowlist) == 0 {
		return errors.New("replica-advertise-allowlist is required in etcd mode")
	}
	if !c.AllowInsecureDev {
		if o.InternalCommandTokenFile == "" {
			return errors.New("internal-command-token-file is required in production etcd mode")
		}
		if o.InternalCommandCAFile == "" {
			return errors.New("internal-command-ca-file is required in production etcd mode")
		}
	}
	if o.TTL < MinOwnershipTTL {
		return fmt.Errorf("ownership-ttl-seconds must be at least %s, got %s", MinOwnershipTTL, o.TTL)
	}
	return nil
}

// validateURL enforces HTTPS with hostname verification in production. Under the
// development flag, HTTP is permitted only for loopback hosts (design §176).
func (c *Config) validateURL(field, raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: invalid url %q: %w", field, raw, err)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("%s: url %q has no host", field, raw)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if !c.AllowInsecureDev {
			return fmt.Errorf("%s: %q must use https in production", field, raw)
		}
		if !isLoopbackHost(u.Hostname()) {
			return fmt.Errorf("%s: insecure http is permitted only for loopback addresses, got %q", field, raw)
		}
		return nil
	default:
		return fmt.Errorf("%s: %q must use http or https, got scheme %q", field, raw, u.Scheme)
	}
}
