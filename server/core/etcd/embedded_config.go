// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file parses and validates the embedded etcd configuration file and maps
// it into an embed.Config (design §268). Atlantis parses the configuration
// without using helpers that can terminate the host process, validates it, and
// then constructs embed.Config. Unsafe native settings are hard-rejected
// (design §280): forced new cluster, disabled fsync or strict reconfiguration,
// discovery bootstrap, automatic certificates, and lifecycle-inconsistent
// cluster state.
package etcd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// EmbeddedTLS is a peer or client TLS identity for the embedded server. Peer and
// client listeners use separate identities and require client-certificate auth
// in production (design §278).
type EmbeddedTLS struct {
	CAFile         string `json:"ca_file"`
	CertFile       string `json:"cert_file"`
	KeyFile        string `json:"key_file"`
	ClientCertAuth bool   `json:"client_cert_auth"`
}

// EmbeddedFileConfig is the operator-provided embedded etcd configuration,
// parsed from JSON. It is intentionally a curated subset of embed.Config; unsafe
// native options are represented only so they can be rejected.
type EmbeddedFileConfig struct {
	Name                     string      `json:"name"`
	DataDir                  string      `json:"data_dir"`
	ListenClientURLs         []string    `json:"listen_client_urls"`
	AdvertiseClientURLs      []string    `json:"advertise_client_urls"`
	ListenPeerURLs           []string    `json:"listen_peer_urls"`
	InitialAdvertisePeerURLs []string    `json:"initial_advertise_peer_urls"`
	InitialCluster           string      `json:"initial_cluster"`
	InitialClusterToken      string      `json:"initial_cluster_token"`
	ClientTLS                EmbeddedTLS `json:"client_tls"`
	PeerTLS                  EmbeddedTLS `json:"peer_tls"`

	// Unsafe options. Any true value is rejected (design §280).
	ForceNewCluster              bool   `json:"force_new_cluster"`
	UnsafeNoFsync                bool   `json:"unsafe_no_fsync"`
	DisableStrictReconfigCheck   bool   `json:"disable_strict_reconfig_check"`
	EnableV2Discovery            string `json:"discovery"`
	AutoCompactionRetentionUnset bool   `json:"-"`
}

// ParseEmbeddedConfig reads and JSON-decodes the embedded configuration file. It
// uses no process-terminating helpers.
func ParseEmbeddedConfig(path string) (*EmbeddedFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded etcd config file: %w", err)
	}
	var cfg EmbeddedFileConfig
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding embedded etcd config file: %w", err)
	}
	return &cfg, nil
}

// Validate enforces the embedded configuration contract for a given lifecycle,
// startup purpose, and voter count. allowInsecureDev relaxes the production TLS
// and https requirements for loopback development only.
func (e *EmbeddedFileConfig) Validate(lifecycle Lifecycle, voterCount int, allowInsecureDev bool) error {
	if strings.TrimSpace(e.Name) == "" {
		return fmt.Errorf("embedded etcd config: name is required")
	}
	if strings.TrimSpace(e.DataDir) == "" {
		return fmt.Errorf("embedded etcd config: data_dir is required")
	}
	if err := e.rejectUnsafe(); err != nil {
		return err
	}
	if err := e.validateURLs(allowInsecureDev); err != nil {
		return err
	}
	return e.validateInitialCluster(lifecycle, voterCount)
}

// rejectUnsafe hard-rejects unsafe native embedded options (design §280, §819).
func (e *EmbeddedFileConfig) rejectUnsafe() error {
	switch {
	case e.ForceNewCluster:
		return fmt.Errorf("embedded etcd config: force_new_cluster is not permitted")
	case e.UnsafeNoFsync:
		return fmt.Errorf("embedded etcd config: unsafe_no_fsync is not permitted")
	case e.DisableStrictReconfigCheck:
		return fmt.Errorf("embedded etcd config: disabling strict reconfiguration check is not permitted")
	case strings.TrimSpace(e.EnableV2Discovery) != "":
		return fmt.Errorf("embedded etcd config: discovery-based bootstrap is not permitted")
	}
	return nil
}

func (e *EmbeddedFileConfig) validateURLs(allowInsecureDev bool) error {
	groups := map[string][]string{
		"listen_client_urls":          e.ListenClientURLs,
		"advertise_client_urls":       e.AdvertiseClientURLs,
		"listen_peer_urls":            e.ListenPeerURLs,
		"initial_advertise_peer_urls": e.InitialAdvertisePeerURLs,
	}
	for field, urls := range groups {
		if len(urls) == 0 {
			return fmt.Errorf("embedded etcd config: %s is required", field)
		}
		for _, raw := range urls {
			u, err := url.Parse(raw)
			if err != nil || u.Hostname() == "" {
				return fmt.Errorf("embedded etcd config: %s has invalid url %q", field, raw)
			}
			if u.Scheme == "https" {
				continue
			}
			if u.Scheme == "http" {
				if !allowInsecureDev {
					return fmt.Errorf("embedded etcd config: %s %q must use https in production", field, raw)
				}
				if !isLoopbackHost(u.Hostname()) {
					return fmt.Errorf("embedded etcd config: insecure http in %s is permitted only for loopback, got %q", field, raw)
				}
				continue
			}
			return fmt.Errorf("embedded etcd config: %s %q must use http or https", field, raw)
		}
	}
	// Production requires client-certificate authentication and complete TLS
	// material on both listeners. Asserting the files here surfaces a missing
	// cert/key/CA at config validation rather than late inside StartEtcd.
	if !allowInsecureDev {
		if err := requireCompleteTLS("peer_tls", e.PeerTLS); err != nil {
			return err
		}
		if err := requireCompleteTLS("client_tls", e.ClientTLS); err != nil {
			return err
		}
	}
	return nil
}

// requireCompleteTLS asserts a production embedded listener has client-cert auth
// enabled and a complete CA + certificate + key.
func requireCompleteTLS(field string, t EmbeddedTLS) error {
	if !t.ClientCertAuth {
		return fmt.Errorf("embedded etcd config: %s.client_cert_auth is required in production", field)
	}
	switch {
	case t.CAFile == "":
		return fmt.Errorf("embedded etcd config: %s.ca_file is required in production", field)
	case t.CertFile == "":
		return fmt.Errorf("embedded etcd config: %s.cert_file is required in production", field)
	case t.KeyFile == "":
		return fmt.Errorf("embedded etcd config: %s.key_file is required in production", field)
	}
	return nil
}

// validateInitialCluster checks the peer set and its consistency with the
// lifecycle. In bootstrap the number of unique members must equal the voter
// count (design §276).
func (e *EmbeddedFileConfig) validateInitialCluster(lifecycle Lifecycle, voterCount int) error {
	members, err := parseInitialCluster(e.InitialCluster)
	if err != nil {
		return err
	}
	if strings.TrimSpace(e.InitialClusterToken) == "" {
		return fmt.Errorf("embedded etcd config: initial_cluster_token is required")
	}
	if _, ok := members[e.Name]; !ok {
		return fmt.Errorf("embedded etcd config: this member %q is not in initial_cluster", e.Name)
	}
	if lifecycle == LifecycleBootstrap && len(members) != voterCount {
		return fmt.Errorf("embedded etcd config: bootstrap requires exactly %d unique members in initial_cluster, got %d", voterCount, len(members))
	}
	return nil
}

// parseInitialCluster parses "name0=url0,name1=url1" into a unique member map.
func parseInitialCluster(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("embedded etcd config: initial_cluster is required")
	}
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, peerURL, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(name) == "" || strings.TrimSpace(peerURL) == "" {
			return nil, fmt.Errorf("embedded etcd config: malformed initial_cluster entry %q", part)
		}
		if _, dup := out[name]; dup {
			return nil, fmt.Errorf("embedded etcd config: duplicate member name %q in initial_cluster", name)
		}
		out[name] = peerURL
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("embedded etcd config: initial_cluster has no members")
	}
	return out, nil
}

// initialClusterState returns the embed.Config cluster state implied by the
// lifecycle: "new" for bootstrap, "existing" for restart/join/restore. Atlantis
// never derives this from directory emptiness (design §286).
func initialClusterState(lifecycle Lifecycle) string {
	if lifecycle == LifecycleBootstrap {
		return "new"
	}
	return "existing"
}
