// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package etcd implements an opt-in etcd locking and coordination backend for
// Atlantis with two runtime modes, external and embedded. See
// docs/superpowers/specs/2026-09-07-dual-mode-embedded-etcd-ha-design.md for
// the architecture this package implements.
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

// Mode selects the etcd runtime: an external cluster or an embedded voter.
type Mode string

const (
	ModeExternal Mode = "external"
	ModeEmbedded Mode = "embedded"
)

// Lifecycle is the mandatory embedded startup lifecycle (design §215). Atlantis
// never derives it from network reachability or directory emptiness.
type Lifecycle string

const (
	LifecycleBootstrap    Lifecycle = "bootstrap"
	LifecycleRestart      Lifecycle = "restart"
	LifecycleJoinExisting Lifecycle = "join-existing"
	LifecycleRestore      Lifecycle = "restore"
)

// StartupPurpose gates whether a started embedded process serves Atlantis
// traffic or only exposes the cluster to operator tooling (design §251).
type StartupPurpose string

const (
	PurposeServe       StartupPurpose = "serve"
	PurposeMaintenance StartupPurpose = "maintenance"
)

// supportedVoterCounts is the closed set of allowed embedded voter counts
// (design §206). Even values and values above 7 are rejected.
var supportedVoterCounts = map[int]struct{}{3: {}, 5: {}, 7: {}}

const (
	// MinOwnershipTTL is the floor for the ownership session TTL (design §314).
	MinOwnershipTTL = 10 * time.Second
	// DefaultOwnershipTTL is the default ownership session TTL (design §303).
	DefaultOwnershipTTL = 30 * time.Second
	// DefaultRequestTimeout bounds individual database/ownership/readiness RPCs.
	DefaultRequestTimeout = 5 * time.Second
	// DefaultStartupTimeout bounds initial connectivity and quorum formation.
	DefaultStartupTimeout = 5 * time.Minute
	// DefaultVoterCount is the default embedded voter count (design §208).
	DefaultVoterCount = 3
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

// EmbeddedConfig holds embedded-only settings (design §192).
type EmbeddedConfig struct {
	ConfigFile           string
	VoterCount           int
	Lifecycle            Lifecycle
	StartupPurpose       StartupPurpose
	IdentityFile         string
	JoinEndpoints        []string // join-existing only
	MembershipTicketFile string   // join-existing only
	RestoreManifestFile  string   // restore only
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

	// Embedded mode.
	Embedded EmbeddedConfig

	// Ownership and routing (both modes).
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
		return c.validateEmbedded()
	default:
		return fmt.Errorf("etcd-mode must be %q or %q, got %q", ModeExternal, ModeEmbedded, c.Mode)
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
	if err := c.rejectEmbeddedFields(); err != nil {
		return err
	}
	for _, ep := range c.Endpoints {
		if err := c.validateURL("etcd-endpoints", ep); err != nil {
			return err
		}
	}
	return nil
}

// rejectEmbeddedFields refuses embedded-only settings in external mode (design §186).
func (c *Config) rejectEmbeddedFields() error {
	e := c.Embedded
	switch {
	case e.ConfigFile != "":
		return errors.New("etcd-embedded-config-file is not valid in external mode")
	case e.Lifecycle != "":
		return errors.New("etcd-embedded-lifecycle is not valid in external mode")
	case e.IdentityFile != "":
		return errors.New("etcd-embedded-identity-file is not valid in external mode")
	case len(e.JoinEndpoints) > 0:
		return errors.New("etcd-embedded-join-endpoints is not valid in external mode")
	case e.MembershipTicketFile != "":
		return errors.New("etcd-embedded-membership-ticket-file is not valid in external mode")
	case e.RestoreManifestFile != "":
		return errors.New("etcd-embedded-restore-manifest-file is not valid in external mode")
	}
	return nil
}

func (c *Config) validateEmbedded() error {
	if len(c.Endpoints) > 0 {
		return errors.New("etcd-endpoints is not valid in embedded mode")
	}
	e := c.Embedded
	if e.ConfigFile == "" {
		return errors.New("embedded etcd mode requires etcd-embedded-config-file")
	}
	if e.IdentityFile == "" {
		return errors.New("embedded etcd mode requires etcd-embedded-identity-file")
	}
	if _, ok := supportedVoterCounts[e.VoterCount]; !ok {
		return fmt.Errorf("etcd-embedded-voter-count must be 3, 5, or 7, got %d", e.VoterCount)
	}
	if err := validateLifecyclePurpose(e.Lifecycle, e.StartupPurpose); err != nil {
		return err
	}
	return validateLifecycleFields(e)
}

// validateLifecyclePurpose enforces the production lifecycle/purpose matrix
// (design §251): bootstrap/maintenance, restore/maintenance, restart/serve,
// join-existing/serve.
func validateLifecyclePurpose(l Lifecycle, p StartupPurpose) error {
	switch p {
	case PurposeServe, PurposeMaintenance:
	default:
		return fmt.Errorf("etcd-embedded-startup-purpose must be %q or %q, got %q", PurposeServe, PurposeMaintenance, p)
	}
	var want StartupPurpose
	switch l {
	case LifecycleBootstrap, LifecycleRestore:
		want = PurposeMaintenance
	case LifecycleRestart, LifecycleJoinExisting:
		want = PurposeServe
	default:
		return fmt.Errorf("etcd-embedded-lifecycle must be one of bootstrap|restart|join-existing|restore, got %q", l)
	}
	if p != want {
		return fmt.Errorf("lifecycle %q requires startup-purpose %q, got %q", l, want, p)
	}
	return nil
}

// validateLifecycleFields enforces that join/restore-only fields appear only in
// their own lifecycle and are otherwise rejected (design §236).
func validateLifecycleFields(e EmbeddedConfig) error {
	joinFieldsSet := len(e.JoinEndpoints) > 0 || e.MembershipTicketFile != ""
	if e.Lifecycle == LifecycleJoinExisting {
		if len(e.JoinEndpoints) == 0 {
			return errors.New("join-existing mode requires etcd-embedded-join-endpoints")
		}
		if e.MembershipTicketFile == "" {
			return errors.New("join-existing mode requires etcd-embedded-membership-ticket-file")
		}
	} else if joinFieldsSet {
		return fmt.Errorf("etcd-embedded-join-endpoints and etcd-embedded-membership-ticket-file are only valid in join-existing mode, not %q", e.Lifecycle)
	}

	if e.Lifecycle == LifecycleRestore {
		if e.RestoreManifestFile == "" {
			return errors.New("restore mode requires etcd-embedded-restore-manifest-file")
		}
	} else if e.RestoreManifestFile != "" {
		return fmt.Errorf("etcd-embedded-restore-manifest-file is only valid in restore mode, not %q", e.Lifecycle)
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
