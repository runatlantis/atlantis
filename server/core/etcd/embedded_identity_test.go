// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/storage/datadir"
	"go.etcd.io/etcd/server/v3/storage/wal"
	"go.uber.org/zap"

	. "github.com/runatlantis/atlantis/testing"
)

// writeJSON marshals v to a temp file and returns its path.
func writeJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	Ok(t, err)
	path := filepath.Join(t.TempDir(), "artifact.json")
	Ok(t, os.WriteFile(path, data, 0o600))
	return path
}

// validIdentityManifest matches validEmbeddedFile's name, token, and peer set.
func validIdentityManifest() *IdentityManifest {
	return &IdentityManifest{
		Kind:                identityKind,
		DeploymentID:        "dep-1",
		MemberName:          "member-0",
		InitialClusterToken: "tok",
		InitialCluster:      "member-0=https://member-0:2380,member-1=https://member-1:2380,member-2=https://member-2:2380",
		Active:              true,
		ClusterID:           types.ID(0xdef).String(),
		MemberID:            types.ID(0xabc).String(),
	}
}

func validMembershipTicket() *MembershipTicket {
	return &MembershipTicket{
		Kind:         membershipTicketKind,
		DeploymentID: "dep-1",
		ClusterID:    types.ID(0xdef).String(),
		ClusterToken: "tok",
		MemberName:   "member-0",
		PeerURLs:     []string{"https://member-0:2380"},
		Nonce:        "opaque-nonce",
	}
}

func validRestoreManifest() *RestoreManifest {
	return &RestoreManifest{
		Kind:               restoreManifestKind,
		DeploymentID:       "dep-1",
		SnapshotHash:       "sha256:abc",
		RecoveryEpoch:      "11111111-1111-1111-1111-111111111111",
		RecoveryGeneration: "22222222-2222-2222-2222-222222222222",
		ClusterToken:       "tok",
		MemberName:         "member-0",
	}
}

func TestParseIdentityManifest_Valid(t *testing.T) {
	path := writeJSON(t, validIdentityManifest())
	m, err := ParseIdentityManifest(path)
	Ok(t, err)
	Equals(t, "member-0", m.MemberName)
	Equals(t, true, m.Active)
}

func TestParseIdentityManifest_MissingRequired(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*IdentityManifest)
		substr string
	}{
		{"wrong kind", func(m *IdentityManifest) { m.Kind = "nope" }, "kind must be"},
		{"no deployment id", func(m *IdentityManifest) { m.DeploymentID = "" }, "deployment_id is required"},
		{"no member name", func(m *IdentityManifest) { m.MemberName = "" }, "member_name is required"},
		{"no token", func(m *IdentityManifest) { m.InitialClusterToken = "" }, "initial_cluster_token is required"},
		{"no initial cluster", func(m *IdentityManifest) { m.InitialCluster = "" }, "initial_cluster is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validIdentityManifest()
			tc.mutate(m)
			_, err := ParseIdentityManifest(writeJSON(t, m))
			ErrContains(t, tc.substr, err)
		})
	}
}

func TestParseIdentityManifest_RejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact.json")
	Ok(t, os.WriteFile(path, []byte(`{"kind":"atlantis-etcd-identity","surprise":true}`), 0o600))
	_, err := ParseIdentityManifest(path)
	ErrContains(t, "decoding", err)
}

func TestValidateBootstrapIdentity(t *testing.T) {
	f := validEmbeddedFile()

	m := validIdentityManifest()
	m.Active = false
	Ok(t, ValidateBootstrapIdentity(m, f, "dep-1"))

	// An already-activated manifest refuses bootstrap.
	m.Active = true
	ErrContains(t, "already active", ValidateBootstrapIdentity(m, f, "dep-1"))

	// Wrong deployment refused.
	m = validIdentityManifest()
	m.Active = false
	ErrContains(t, "does not match configured deployment", ValidateBootstrapIdentity(m, f, "other"))

	// Peer set mismatch refused.
	m = validIdentityManifest()
	m.Active = false
	m.InitialCluster = "member-0=https://member-0:2380"
	ErrContains(t, "peer set does not match", ValidateBootstrapIdentity(m, f, "dep-1"))
}

func TestValidateRestartIdentity_ManifestMismatch(t *testing.T) {
	f := validEmbeddedFile()
	dataDir := t.TempDir()

	// Wrong member name.
	m := validIdentityManifest()
	m.MemberName = "member-9"
	ErrContains(t, "member_name", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Wrong token.
	m = validIdentityManifest()
	m.InitialClusterToken = "other"
	ErrContains(t, "initial_cluster_token", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Inactive manifest.
	m = validIdentityManifest()
	m.Active = false
	ErrContains(t, "requires an active identity manifest", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Missing persisted ids in manifest.
	m = validIdentityManifest()
	m.ClusterID = ""
	ErrContains(t, "cluster_id and member_id", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Consistent manifest but the member dir is absent -> refuse (empty/lost PVC).
	m = validIdentityManifest()
	ErrContains(t, "member directory", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))
}

// TestValidateRestartIdentity_PersistedIdentity exercises the real WAL-backed
// persisted-identity path: it writes a WAL carrying an etcdserverpb.Metadata and
// checks that a matching manifest passes while a mismatched one is refused.
func TestValidateRestartIdentity_PersistedIdentity(t *testing.T) {
	dataDir := t.TempDir()
	writeWALIdentity(t, dataDir, 0xabc, 0xdef) // node, cluster

	f := validEmbeddedFile()

	// Manifest matches the persisted WAL identity.
	m := validIdentityManifest()
	Ok(t, ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Leading-zero / case variants still match (numeric comparison).
	m = validIdentityManifest()
	m.MemberID = "0ABC"
	m.ClusterID = "0DEF"
	Ok(t, ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Wrong persisted member id -> refuse.
	m = validIdentityManifest()
	m.MemberID = types.ID(0x999).String()
	ErrContains(t, "member_id mismatch", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// Wrong persisted cluster id -> refuse.
	m = validIdentityManifest()
	m.ClusterID = types.ID(0x111).String()
	ErrContains(t, "cluster_id mismatch", ValidateRestartIdentity(m, f, m.DeploymentID, dataDir))

	// A deployment ID that does not match the manifest is refused by
	// ValidateRestartIdentity itself (no longer relying on a caller pre-check).
	m = validIdentityManifest()
	ErrContains(t, "deployment_id", ValidateRestartIdentity(m, f, "some-other-deployment", dataDir))
}

// writeWALIdentity creates an etcd WAL under dataDir/member/wal whose metadata
// records the given node and cluster IDs, mirroring how etcd bootstraps a member.
func writeWALIdentity(t *testing.T, dataDir string, nodeID, clusterID uint64) {
	t.Helper()
	Ok(t, os.MkdirAll(datadir.ToMemberDir(dataDir), 0o700))
	md := &etcdserverpb.Metadata{NodeID: nodeID, ClusterID: clusterID}
	data, err := md.Marshal()
	Ok(t, err)
	w, err := wal.Create(zap.NewNop(), datadir.ToWALDir(dataDir), data)
	Ok(t, err)
	Ok(t, w.Close())
}

func TestParseMembershipTicket_MissingRequired(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*MembershipTicket)
		substr string
	}{
		{"wrong kind", func(x *MembershipTicket) { x.Kind = "nope" }, "kind must be"},
		{"no token", func(x *MembershipTicket) { x.ClusterToken = "" }, "cluster_token is required"},
		{"no peer urls", func(x *MembershipTicket) { x.PeerURLs = nil }, "peer_urls is required"},
		{"no nonce", func(x *MembershipTicket) { x.Nonce = "" }, "nonce is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x := validMembershipTicket()
			tc.mutate(x)
			_, err := ParseMembershipTicket(writeJSON(t, x))
			ErrContains(t, tc.substr, err)
		})
	}
}

func TestValidateMembershipTicket_Inconsistent(t *testing.T) {
	f := validEmbeddedFile()
	now := time.Now()

	Ok(t, ValidateMembershipTicket(validMembershipTicket(), f, "dep-1", now))

	// Wrong member name.
	x := validMembershipTicket()
	x.MemberName = "member-9"
	ErrContains(t, "member_name", ValidateMembershipTicket(x, f, "dep-1", now))

	// Wrong cluster token.
	x = validMembershipTicket()
	x.ClusterToken = "other"
	ErrContains(t, "cluster_token", ValidateMembershipTicket(x, f, "dep-1", now))

	// Peer URLs do not match the embedded config.
	x = validMembershipTicket()
	x.PeerURLs = []string{"https://member-9:2380"}
	ErrContains(t, "peer_urls do not match", ValidateMembershipTicket(x, f, "dep-1", now))

	// Wrong deployment.
	x = validMembershipTicket()
	ErrContains(t, "does not match configured deployment", ValidateMembershipTicket(x, f, "other", now))

	// Expired.
	x = validMembershipTicket()
	x.ExpiresAt = now.Add(-time.Minute).Format(time.RFC3339)
	ErrContains(t, "expired", ValidateMembershipTicket(x, f, "dep-1", now))

	// Not-yet-expired ticket passes.
	x = validMembershipTicket()
	x.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
	Ok(t, ValidateMembershipTicket(x, f, "dep-1", now))
}

func TestParseRestoreManifest_MissingEpochAndFields(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*RestoreManifest)
		substr string
	}{
		{"wrong kind", func(m *RestoreManifest) { m.Kind = "nope" }, "kind must be"},
		{"no snapshot hash", func(m *RestoreManifest) { m.SnapshotHash = "" }, "snapshot_hash is required"},
		{"no epoch", func(m *RestoreManifest) { m.RecoveryEpoch = "" }, "recovery_epoch is required"},
		{"epoch not uuid", func(m *RestoreManifest) { m.RecoveryEpoch = "not-a-uuid" }, "recovery_epoch must be a UUID"},
		{"no generation", func(m *RestoreManifest) { m.RecoveryGeneration = "" }, "recovery_generation is required"},
		{"no cluster token", func(m *RestoreManifest) { m.ClusterToken = "" }, "cluster_token is required"},
		{"no member name", func(m *RestoreManifest) { m.MemberName = "" }, "member_name is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validRestoreManifest()
			tc.mutate(m)
			_, err := ParseRestoreManifest(writeJSON(t, m))
			ErrContains(t, tc.substr, err)
		})
	}
}

func TestParseRestoreManifest_Valid(t *testing.T) {
	m, err := ParseRestoreManifest(writeJSON(t, validRestoreManifest()))
	Ok(t, err)
	Ok(t, ValidateRestoreManifest(m, validEmbeddedFile(), "dep-1"))
}

func TestValidateRestoreManifest_Inconsistent(t *testing.T) {
	f := validEmbeddedFile()

	m := validRestoreManifest()
	m.MemberName = "member-9"
	ErrContains(t, "member_name", ValidateRestoreManifest(m, f, "dep-1"))

	m = validRestoreManifest()
	ErrContains(t, "does not match configured deployment", ValidateRestoreManifest(m, f, "other"))
}
