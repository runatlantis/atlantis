// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

// writeFile creates a small file in dir so it is non-empty.
func writeFile(dir string) error {
	return os.WriteFile(filepath.Join(dir, "member"), []byte("x"), 0o600)
}

// validEmbeddedFile returns a production-valid three-member bootstrap config that
// tests mutate to exercise one failure at a time.
func validEmbeddedFile() *EmbeddedFileConfig {
	return &EmbeddedFileConfig{
		Name:                     "member-0",
		DataDir:                  "/data/member-0",
		ListenClientURLs:         []string{"https://member-0:2379"},
		AdvertiseClientURLs:      []string{"https://member-0:2379"},
		ListenPeerURLs:           []string{"https://member-0:2380"},
		InitialAdvertisePeerURLs: []string{"https://member-0:2380"},
		InitialCluster:           "member-0=https://member-0:2380,member-1=https://member-1:2380,member-2=https://member-2:2380",
		InitialClusterToken:      "tok",
		ClientTLS:                EmbeddedTLS{CAFile: "ca", CertFile: "c", KeyFile: "k", ClientCertAuth: true},
		PeerTLS:                  EmbeddedTLS{CAFile: "ca", CertFile: "c", KeyFile: "k", ClientCertAuth: true},
	}
}

func TestEmbeddedConfig_ValidBootstrap(t *testing.T) {
	Ok(t, validEmbeddedFile().Validate(LifecycleBootstrap, 3, false))
}

func TestEmbeddedConfig_RejectsUnsafe(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*EmbeddedFileConfig)
		substr string
	}{
		{"force new cluster", func(e *EmbeddedFileConfig) { e.ForceNewCluster = true }, "force_new_cluster"},
		{"no fsync", func(e *EmbeddedFileConfig) { e.UnsafeNoFsync = true }, "unsafe_no_fsync"},
		{"disable strict reconfig", func(e *EmbeddedFileConfig) { e.DisableStrictReconfigCheck = true }, "strict reconfiguration"},
		{"discovery", func(e *EmbeddedFileConfig) { e.EnableV2Discovery = "https://discovery" }, "discovery"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := validEmbeddedFile()
			tc.mutate(e)
			ErrContains(t, tc.substr, e.Validate(LifecycleBootstrap, 3, false))
		})
	}
}

func TestEmbeddedConfig_BootstrapMemberCount(t *testing.T) {
	e := validEmbeddedFile()
	// Bootstrap voter count 5 but only 3 members -> reject (design §276).
	ErrContains(t, "exactly 5 unique members", e.Validate(LifecycleBootstrap, 5, false))
}

func TestEmbeddedConfig_MemberNotInCluster(t *testing.T) {
	e := validEmbeddedFile()
	e.Name = "member-9"
	ErrContains(t, "not in initial_cluster", e.Validate(LifecycleBootstrap, 3, false))
}

func TestEmbeddedConfig_DuplicateMember(t *testing.T) {
	e := validEmbeddedFile()
	e.InitialCluster = "member-0=https://member-0:2380,member-0=https://x:2380,member-2=https://member-2:2380"
	ErrContains(t, "duplicate member name", e.Validate(LifecycleBootstrap, 3, false))
}

func TestEmbeddedConfig_ProductionRequiresHTTPS(t *testing.T) {
	e := validEmbeddedFile()
	e.ListenClientURLs = []string{"http://member-0:2379"}
	ErrContains(t, "must use https", e.Validate(LifecycleBootstrap, 3, false))
}

func TestEmbeddedConfig_ProductionRequiresClientCertAuth(t *testing.T) {
	e := validEmbeddedFile()
	e.PeerTLS.ClientCertAuth = false
	ErrContains(t, "peer_tls.client_cert_auth", e.Validate(LifecycleBootstrap, 3, false))
}

// TestEmbeddedConfig_ProductionRequiresTLSFiles proves production validation
// rejects a listener with client-cert-auth enabled but empty CA/cert/key paths,
// surfacing the error at config validation instead of late inside StartEtcd.
func TestEmbeddedConfig_ProductionRequiresTLSFiles(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*EmbeddedFileConfig)
		substr string
	}{
		{"peer ca missing", func(e *EmbeddedFileConfig) { e.PeerTLS.CAFile = "" }, "peer_tls.ca_file"},
		{"peer cert missing", func(e *EmbeddedFileConfig) { e.PeerTLS.CertFile = "" }, "peer_tls.cert_file"},
		{"peer key missing", func(e *EmbeddedFileConfig) { e.PeerTLS.KeyFile = "" }, "peer_tls.key_file"},
		{"client ca missing", func(e *EmbeddedFileConfig) { e.ClientTLS.CAFile = "" }, "client_tls.ca_file"},
		{"client cert missing", func(e *EmbeddedFileConfig) { e.ClientTLS.CertFile = "" }, "client_tls.cert_file"},
		{"client key missing", func(e *EmbeddedFileConfig) { e.ClientTLS.KeyFile = "" }, "client_tls.key_file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := validEmbeddedFile()
			tc.mutate(e)
			ErrContains(t, tc.substr, e.Validate(LifecycleBootstrap, 3, false))
		})
	}
}

func TestEmbeddedConfig_InsecureDevLoopbackOnly(t *testing.T) {
	e := validEmbeddedFile()
	e.ListenClientURLs = []string{"http://127.0.0.1:2379"}
	e.AdvertiseClientURLs = []string{"http://127.0.0.1:2379"}
	e.ListenPeerURLs = []string{"http://127.0.0.1:2380"}
	e.InitialAdvertisePeerURLs = []string{"http://127.0.0.1:2380"}
	e.InitialCluster = "member-0=http://127.0.0.1:2380,member-1=http://127.0.0.2:2380,member-2=http://127.0.0.3:2380"
	e.ClientTLS = EmbeddedTLS{}
	e.PeerTLS = EmbeddedTLS{}
	Ok(t, e.Validate(LifecycleBootstrap, 3, true))

	e.ListenClientURLs = []string{"http://192.168.1.1:2379"}
	ErrContains(t, "loopback", e.Validate(LifecycleBootstrap, 3, true))
}

func TestValidateDataDirState(t *testing.T) {
	emptyDir := t.TempDir()
	// bootstrap/join require empty
	Ok(t, validateDataDirState(LifecycleBootstrap, emptyDir))
	Ok(t, validateDataDirState(LifecycleJoinExisting, emptyDir))
	// restart/restore require non-empty
	ErrContains(t, "non-empty data directory", validateDataDirState(LifecycleRestart, emptyDir))

	nonEmpty := t.TempDir()
	Ok(t, writeFile(nonEmpty))
	Ok(t, validateDataDirState(LifecycleRestart, nonEmpty))
	ErrContains(t, "empty data directory", validateDataDirState(LifecycleBootstrap, nonEmpty))
}
