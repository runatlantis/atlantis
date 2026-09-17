// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file builds a resolved Config from raw string settings supplied by the
// Atlantis server flags. It lives in this package (rather than importing
// server.UserConfig) to avoid an import cycle, since server imports this
// package.
package etcd

import (
	"fmt"
	"strings"
	"time"
)

// Settings is the raw, string-oriented view of the etcd flags. The Atlantis
// server maps its UserConfig into this struct; BuildConfig parses and defaults
// it into a validated-shape Config.
type Settings struct {
	Mode             string
	DeploymentID     string
	Namespace        string
	CAFile           string
	CertFile         string
	KeyFile          string
	ServerName       string
	Username         string
	PasswordFile     string
	RequestTimeout   string // duration, e.g. "5s"
	StartupTimeout   string // duration, e.g. "5m"
	AllowInsecureDev bool
	Endpoints        string // comma-separated

	// Ownership and routing.
	ReplicaID                 string
	ReplicaAdvertiseURL       string
	ReplicaAdvertiseAllowlist string // comma-separated
	InternalCommandTokenFile  string
	InternalCommandCAFile     string
	OwnershipTTLSeconds       int
}

// BuildConfig parses and defaults Settings into a Config. It does not validate;
// callers invoke Config.ValidateDatabase or Config.Validate.
func BuildConfig(s Settings) (*Config, error) {
	namespace := s.Namespace
	if strings.TrimSpace(namespace) == "" {
		namespace = DefaultNamespace
	}

	reqTimeout, err := parseDurationDefault(s.RequestTimeout, DefaultRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("etcd-request-timeout: %w", err)
	}
	startTimeout, err := parseDurationDefault(s.StartupTimeout, DefaultStartupTimeout)
	if err != nil {
		return nil, fmt.Errorf("etcd-startup-timeout: %w", err)
	}

	ttl := DefaultOwnershipTTL
	if s.OwnershipTTLSeconds > 0 {
		ttl = time.Duration(s.OwnershipTTLSeconds) * time.Second
	}

	return &Config{
		Mode:             Mode(s.Mode),
		DeploymentID:     s.DeploymentID,
		Namespace:        namespace,
		TLS:              TLSConfig{CAFile: s.CAFile, CertFile: s.CertFile, KeyFile: s.KeyFile, ServerName: s.ServerName},
		Username:         s.Username,
		PasswordFile:     s.PasswordFile,
		RequestTimeout:   reqTimeout,
		StartupTimeout:   startTimeout,
		AllowInsecureDev: s.AllowInsecureDev,
		Endpoints:        splitCSV(s.Endpoints),
		Ownership: OwnershipConfig{
			ReplicaID:                 s.ReplicaID,
			ReplicaAdvertiseURL:       s.ReplicaAdvertiseURL,
			ReplicaAdvertiseAllowlist: splitCSV(s.ReplicaAdvertiseAllowlist),
			InternalCommandTokenFile:  s.InternalCommandTokenFile,
			InternalCommandCAFile:     s.InternalCommandCAFile,
			TTL:                       ttl,
		},
	}, nil
}

func parseDurationDefault(raw string, def time.Duration) (time.Duration, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	return time.ParseDuration(raw)
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
