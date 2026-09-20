// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the embedded-etcd lifecycle safety validation for the
// declared identity, membership ticket, and restore manifest (design §"Embedded
// mode", §"Voter-count lifecycle", §"Embedded runtime"). The operator-facing
// config file (embedded_config.go) is validated for shape; the sidecar identity
// artifacts validated here bind that config to the persisted cluster so a
// stale, cloned, or wrong PVC cannot silently start and diverge the cluster.
//
// Every function fails closed: any missing required field, any inconsistency
// between a manifest/ticket and the embedded file config, and any inability to
// confirm the data directory's persisted identity is an error so startup
// refuses before serving public traffic (design §785, failure-semantics rows
// "Restart mode finds an empty/lost PVC" and "Embedded bootstrap configuration
// is inconsistent"). Parsing uses no process-terminating helpers, matching
// ParseEmbeddedConfig.
package etcd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/storage/datadir"
	"go.etcd.io/etcd/server/v3/storage/wal"
	"go.etcd.io/etcd/server/v3/storage/wal/walpb"
	"go.uber.org/zap"
)

// Manifest/ticket kind discriminators. A file whose kind does not match the
// expected artifact is rejected rather than misread as another type.
const (
	identityKind         = "atlantis-etcd-identity"
	membershipTicketKind = "atlantis-etcd-membership-ticket"
	restoreManifestKind  = "atlantis-etcd-restore-manifest"
)

// IdentityManifest is the durable, PVC-external identity artifact for an
// embedded member (design §256: "Identity manifests are retained outside
// individual member PVCs so accidental PVC loss cannot be mistaken for a new
// cluster"). Bootstrap declares it in the new (inactive) state; deployment
// tooling activates it after quorum forms and records the observed cluster and
// member IDs. Restart requires an active manifest whose IDs match the data
// directory (design §221-224).
type IdentityManifest struct {
	Kind         string `json:"kind"`
	DeploymentID string `json:"deployment_id"`

	// MemberName and InitialClusterToken must equal the embedded file config's
	// name and token; InitialCluster is the complete original member set in the
	// same "name=peerURL,..." form as the file config.
	MemberName          string `json:"member_name"`
	InitialClusterToken string `json:"initial_cluster_token"`
	InitialCluster      string `json:"initial_cluster"`

	// Active marks the manifest as activated. Bootstrap requires it to be false
	// (bootstrap is refused after activation); restart requires it to be true.
	Active bool `json:"active"`

	// BootstrapGeneration identifies the bootstrap generation (design §219).
	BootstrapGeneration string `json:"bootstrap_generation,omitempty"`

	// ClusterID and MemberID are the base-16 etcd identities observed after the
	// cluster formed. They are absent in a new (pre-activation) bootstrap
	// manifest and required in an active manifest so restart can cross-check the
	// data directory's persisted WAL metadata.
	ClusterID string `json:"cluster_id,omitempty"`
	MemberID  string `json:"member_id,omitempty"`
}

// MembershipTicket is the one-time ticket produced by an explicit learner-add
// operation for join-existing (design §230, §240). It binds the target cluster,
// the assigned member name, and the peer URLs to join, plus an opaque nonce the
// runtime consumes exactly once against the live cluster.
type MembershipTicket struct {
	Kind         string `json:"kind"`
	DeploymentID string `json:"deployment_id"`

	ClusterID    string `json:"cluster_id"`
	ClusterToken string `json:"cluster_token"`
	MemberName   string `json:"member_name"`

	// PeerURLs are the advertised peer URLs assigned to the joining member; they
	// must match the embedded file config's initial_advertise_peer_urls.
	PeerURLs []string `json:"peer_urls"`

	// Nonce is the opaque one-time secret whose hash the membership-ticket
	// record binds; it is consumed against the live cluster before start.
	Nonce string `json:"nonce"`

	// TransitionGeneration ties the ticket to a durable membership-transition
	// record (design §"Changing voter count").
	TransitionGeneration string `json:"transition_generation,omitempty"`

	// ExpiresAt, when present, is an RFC3339 instant after which the ticket is
	// refused even if otherwise consistent.
	ExpiresAt string `json:"expires_at,omitempty"`
}

// RestoreManifest is the pending recovery manifest emitted by the offline
// snapshot-restore tool and kept outside the member PVC (design §231, §846-859).
// It binds the snapshot identity to a fresh coordination epoch created outside
// the snapshot, so restoring the same snapshot cannot reuse a discarded
// generation (design §610, §629).
type RestoreManifest struct {
	Kind         string `json:"kind"`
	DeploymentID string `json:"deployment_id"`

	// SnapshotHash identifies the snapshot the data directory was restored from.
	SnapshotHash string `json:"snapshot_hash"`
	// BumpedRevision is the documented offline revision bump (design §846).
	BumpedRevision int64 `json:"bumped_revision,omitempty"`

	// RecoveryEpoch is the fresh coordination-epoch UUID created outside the
	// snapshot (design §610). It is mandatory: restore must rotate the epoch.
	RecoveryEpoch string `json:"recovery_epoch"`
	// RecoveryGeneration is the fresh recovery-generation UUID (design §850).
	RecoveryGeneration string `json:"recovery_generation"`

	// ClusterToken and MemberName must equal the embedded file config. The
	// restore tool creates new cluster/member identities, so those IDs are not
	// cross-checked against pre-restore data here.
	ClusterToken   string `json:"cluster_token"`
	MemberName     string `json:"member_name"`
	InitialCluster string `json:"initial_cluster,omitempty"`
}

// parseJSONFile reads and strictly JSON-decodes a sidecar identity artifact,
// using no process-terminating helpers (mirrors ParseEmbeddedConfig).
func parseJSONFile(path, what string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", what, err)
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decoding %s: %w", what, err)
	}
	return nil
}

// ParseIdentityManifest reads, decodes, and structurally validates the identity
// manifest. It checks the kind discriminator and that the always-required fields
// are present; lifecycle-specific checks live in the Validate* functions.
func ParseIdentityManifest(path string) (*IdentityManifest, error) {
	var m IdentityManifest
	if err := parseJSONFile(path, "embedded etcd identity manifest", &m); err != nil {
		return nil, err
	}
	if m.Kind != identityKind {
		return nil, fmt.Errorf("identity manifest: kind must be %q, got %q", identityKind, m.Kind)
	}
	if strings.TrimSpace(m.DeploymentID) == "" {
		return nil, fmt.Errorf("identity manifest: deployment_id is required")
	}
	if strings.TrimSpace(m.MemberName) == "" {
		return nil, fmt.Errorf("identity manifest: member_name is required")
	}
	if strings.TrimSpace(m.InitialClusterToken) == "" {
		return nil, fmt.Errorf("identity manifest: initial_cluster_token is required")
	}
	if strings.TrimSpace(m.InitialCluster) == "" {
		return nil, fmt.Errorf("identity manifest: initial_cluster is required")
	}
	return &m, nil
}

// ParseMembershipTicket reads, decodes, and structurally validates a membership
// ticket. Consistency with the file config and expiry are checked in
// ValidateMembershipTicket.
func ParseMembershipTicket(path string) (*MembershipTicket, error) {
	var t MembershipTicket
	if err := parseJSONFile(path, "embedded etcd membership ticket", &t); err != nil {
		return nil, err
	}
	if t.Kind != membershipTicketKind {
		return nil, fmt.Errorf("membership ticket: kind must be %q, got %q", membershipTicketKind, t.Kind)
	}
	if strings.TrimSpace(t.DeploymentID) == "" {
		return nil, fmt.Errorf("membership ticket: deployment_id is required")
	}
	if strings.TrimSpace(t.ClusterToken) == "" {
		return nil, fmt.Errorf("membership ticket: cluster_token is required")
	}
	if strings.TrimSpace(t.MemberName) == "" {
		return nil, fmt.Errorf("membership ticket: member_name is required")
	}
	if len(t.PeerURLs) == 0 {
		return nil, fmt.Errorf("membership ticket: peer_urls is required")
	}
	if strings.TrimSpace(t.Nonce) == "" {
		return nil, fmt.Errorf("membership ticket: nonce is required")
	}
	return &t, nil
}

// ParseRestoreManifest reads, decodes, and structurally validates a restore
// manifest. It requires the fresh recovery epoch (a valid UUID) because restore
// must rotate the coordination epoch (design §610).
func ParseRestoreManifest(path string) (*RestoreManifest, error) {
	var m RestoreManifest
	if err := parseJSONFile(path, "embedded etcd restore manifest", &m); err != nil {
		return nil, err
	}
	if m.Kind != restoreManifestKind {
		return nil, fmt.Errorf("restore manifest: kind must be %q, got %q", restoreManifestKind, m.Kind)
	}
	if strings.TrimSpace(m.DeploymentID) == "" {
		return nil, fmt.Errorf("restore manifest: deployment_id is required")
	}
	if strings.TrimSpace(m.SnapshotHash) == "" {
		return nil, fmt.Errorf("restore manifest: snapshot_hash is required")
	}
	if strings.TrimSpace(m.RecoveryEpoch) == "" {
		return nil, fmt.Errorf("restore manifest: recovery_epoch is required (restore must rotate the coordination epoch)")
	}
	if _, err := uuid.Parse(strings.TrimSpace(m.RecoveryEpoch)); err != nil {
		return nil, fmt.Errorf("restore manifest: recovery_epoch must be a UUID: %w", err)
	}
	if strings.TrimSpace(m.RecoveryGeneration) == "" {
		return nil, fmt.Errorf("restore manifest: recovery_generation is required")
	}
	if strings.TrimSpace(m.ClusterToken) == "" {
		return nil, fmt.Errorf("restore manifest: cluster_token is required")
	}
	if strings.TrimSpace(m.MemberName) == "" {
		return nil, fmt.Errorf("restore manifest: member_name is required")
	}
	return &m, nil
}

// ValidateBootstrapIdentity checks the identity manifest for the bootstrap
// lifecycle: it must belong to this deployment, agree with the file config's
// member name / token / peer set, and NOT be activated yet — bootstrap is
// refused after the manifest has been activated (design §218).
func ValidateBootstrapIdentity(m *IdentityManifest, fileCfg *EmbeddedFileConfig, deploymentID string) error {
	if err := checkManifestConsistency(m, fileCfg, deploymentID); err != nil {
		return err
	}
	if m.Active {
		return fmt.Errorf("identity manifest: bootstrap is refused because the manifest is already active; use restart")
	}
	return nil
}

// ValidateJoinIdentity checks the identity manifest for join-existing: it is the
// active cluster identity the joiner attaches to, so it must belong to this
// deployment and agree with the file config's token and peer set.
func ValidateJoinIdentity(m *IdentityManifest, fileCfg *EmbeddedFileConfig, deploymentID string) error {
	if err := checkManifestConsistency(m, fileCfg, deploymentID); err != nil {
		return err
	}
	if !m.Active {
		return fmt.Errorf("identity manifest: join-existing requires an active cluster identity manifest")
	}
	return nil
}

// ValidateRestartIdentity is the steady-state safety check. It cross-checks the
// declared identity manifest against the embedded file config AND against the
// identity persisted in the data directory's WAL. It fails closed: if the data
// directory's persisted cluster/member identity cannot be confirmed to match the
// manifest, startup refuses rather than risking a diverged cluster from a stale,
// cloned, or wrong PVC (design §222, failure-semantics "Restart mode finds an
// empty/lost PVC").
func ValidateRestartIdentity(manifest *IdentityManifest, fileCfg *EmbeddedFileConfig, deploymentID, dataDir string) error {
	if err := checkManifestConsistency(manifest, fileCfg, deploymentID); err != nil {
		return err
	}
	if !manifest.Active {
		return fmt.Errorf("identity manifest: restart requires an active identity manifest")
	}
	if strings.TrimSpace(manifest.ClusterID) == "" || strings.TrimSpace(manifest.MemberID) == "" {
		return fmt.Errorf("identity manifest: restart requires cluster_id and member_id to cross-check the data directory")
	}

	// The etcd member directory must exist under the data dir.
	memberDir := datadir.ToMemberDir(dataDir)
	if fi, err := os.Stat(memberDir); err != nil || !fi.IsDir() {
		return fmt.Errorf("restart: member directory %q is missing or not a directory; refuse startup and use the controlled replacement-member procedure", memberDir)
	}

	// Read the identity persisted in the WAL and require it to match the manifest.
	clusterID, memberID, err := readPersistedIdentity(dataDir)
	if err != nil {
		return fmt.Errorf("restart: cannot confirm persisted identity in %q: %w", dataDir, err)
	}
	if err := equalID("cluster_id", manifest.ClusterID, clusterID); err != nil {
		return fmt.Errorf("restart: data directory does not match the identity manifest: %w", err)
	}
	if err := equalID("member_id", manifest.MemberID, memberID); err != nil {
		return fmt.Errorf("restart: data directory does not match the identity manifest: %w", err)
	}
	return nil
}

// ValidateMembershipTicket checks a join-existing ticket for consistency with
// the file config and, if the ticket carries an expiry, that it has not lapsed.
// The atomic single-use consumption of the ticket against the live cluster is a
// separate step performed with the shared client before start (design §240,
// §719); see the documented gap in NewEmbedded's caller.
func ValidateMembershipTicket(t *MembershipTicket, fileCfg *EmbeddedFileConfig, deploymentID string, now time.Time) error {
	if t.DeploymentID != deploymentID {
		return fmt.Errorf("membership ticket: deployment_id %q does not match configured deployment %q", t.DeploymentID, deploymentID)
	}
	if t.MemberName != fileCfg.Name {
		return fmt.Errorf("membership ticket: member_name %q does not match embedded config name %q", t.MemberName, fileCfg.Name)
	}
	if t.ClusterToken != fileCfg.InitialClusterToken {
		return fmt.Errorf("membership ticket: cluster_token does not match embedded config initial_cluster_token")
	}
	if !sameURLSet(t.PeerURLs, fileCfg.InitialAdvertisePeerURLs) {
		return fmt.Errorf("membership ticket: peer_urls do not match embedded config initial_advertise_peer_urls")
	}
	if strings.TrimSpace(t.ExpiresAt) != "" {
		exp, err := time.Parse(time.RFC3339, t.ExpiresAt)
		if err != nil {
			return fmt.Errorf("membership ticket: expires_at is not RFC3339: %w", err)
		}
		if !now.Before(exp) {
			return fmt.Errorf("membership ticket: expired at %s", t.ExpiresAt)
		}
	}
	return nil
}

// ValidateRestoreManifest checks a restore manifest for consistency with the
// file config. The heavy binding (snapshot hash, revision-bump receipts, new
// cluster/member IDs) is verified by recovery tooling after quorum forms
// (design §853); here we bind the manifest to this deployment and file config so
// an ordinary pre-restore data directory or a foreign manifest is refused.
func ValidateRestoreManifest(m *RestoreManifest, fileCfg *EmbeddedFileConfig, deploymentID string) error {
	if m.DeploymentID != deploymentID {
		return fmt.Errorf("restore manifest: deployment_id %q does not match configured deployment %q", m.DeploymentID, deploymentID)
	}
	if m.MemberName != fileCfg.Name {
		return fmt.Errorf("restore manifest: member_name %q does not match embedded config name %q", m.MemberName, fileCfg.Name)
	}
	if m.ClusterToken != fileCfg.InitialClusterToken {
		return fmt.Errorf("restore manifest: cluster_token does not match embedded config initial_cluster_token")
	}
	return nil
}

// checkManifestConsistency enforces the invariants shared by every lifecycle
// that uses an identity manifest: deployment ownership, member name, cluster
// token, and the complete peer set all agree with the embedded file config.
func checkManifestConsistency(m *IdentityManifest, fileCfg *EmbeddedFileConfig, deploymentID string) error {
	if m.DeploymentID != deploymentID {
		return fmt.Errorf("identity manifest: deployment_id %q does not match configured deployment %q", m.DeploymentID, deploymentID)
	}
	if m.MemberName != fileCfg.Name {
		return fmt.Errorf("identity manifest: member_name %q does not match embedded config name %q", m.MemberName, fileCfg.Name)
	}
	if m.InitialClusterToken != fileCfg.InitialClusterToken {
		return fmt.Errorf("identity manifest: initial_cluster_token does not match embedded config")
	}
	manifestMembers, err := parseInitialCluster(m.InitialCluster)
	if err != nil {
		return fmt.Errorf("identity manifest: %w", err)
	}
	cfgMembers, err := parseInitialCluster(fileCfg.InitialCluster)
	if err != nil {
		return err
	}
	if !sameMemberSet(manifestMembers, cfgMembers) {
		return fmt.Errorf("identity manifest: initial_cluster peer set does not match embedded config")
	}
	return nil
}

// readPersistedIdentity reads the member and cluster IDs etcd persisted in the
// WAL metadata at bootstrap (etcdserverpb.Metadata{NodeID, ClusterID}; see
// go.etcd.io/etcd/server/v3/etcdserver bootstrap). It opens the WAL read-only —
// this runs before embed.StartEtcd, so nothing else holds the directory — and
// uses the generated proto Unmarshal (no panicking pbutil helper). Both IDs are
// returned as base-16 strings, matching types.ID.String().
//
// Residual gap: this confirms the persisted member/cluster identity, which is
// the split-brain-relevant part. It does not additionally verify the on-disk
// backend db checksum or the snap consistent-index; etcd's own forced initial
// corruption check (buildEmbedConfig sets ExperimentalInitialCorruptCheck) and
// strict reconfiguration cover on-disk integrity at start.
func readPersistedIdentity(dataDir string) (clusterID, memberID string, err error) {
	walDir := datadir.ToWALDir(dataDir)
	w, err := wal.OpenForRead(zap.NewNop(), walDir, walpb.Snapshot{})
	if err != nil {
		return "", "", fmt.Errorf("opening WAL for read: %w", err)
	}
	defer func() { _ = w.Close() }()

	metadata, _, _, err := w.ReadAll()
	if err != nil {
		return "", "", fmt.Errorf("reading WAL metadata: %w", err)
	}
	var md etcdserverpb.Metadata
	if err := md.Unmarshal(metadata); err != nil {
		return "", "", fmt.Errorf("decoding WAL metadata: %w", err)
	}
	if md.ClusterID == 0 || md.NodeID == 0 {
		return "", "", fmt.Errorf("WAL metadata has empty cluster/member id")
	}
	return types.ID(md.ClusterID).String(), types.ID(md.NodeID).String(), nil
}

// equalID compares two base-16 etcd identities numerically so leading-zero and
// case differences do not cause a false mismatch.
func equalID(field, want, got string) error {
	w, err := types.IDFromString(strings.TrimSpace(want))
	if err != nil {
		return fmt.Errorf("manifest %s %q is not a base-16 id: %w", field, want, err)
	}
	g, err := types.IDFromString(strings.TrimSpace(got))
	if err != nil {
		return fmt.Errorf("persisted %s %q is not a base-16 id: %w", field, got, err)
	}
	if w != g {
		return fmt.Errorf("%s mismatch: manifest %s, data directory %s", field, w.String(), g.String())
	}
	return nil
}

// sameMemberSet reports whether two parsed initial-cluster maps have identical
// member names and peer URLs.
func sameMemberSet(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for name, peer := range a {
		if b[name] != peer {
			return false
		}
	}
	return true
}

// sameURLSet reports whether two URL lists contain the same set of entries,
// ignoring order and duplicates.
func sameURLSet(a, b []string) bool {
	set := func(in []string) map[string]struct{} {
		out := make(map[string]struct{}, len(in))
		for _, v := range in {
			out[strings.TrimSpace(v)] = struct{}{}
		}
		return out
	}
	sa, sb := set(a), set(b)
	if len(sa) != len(sb) {
		return false
	}
	for k := range sa {
		if _, ok := sb[k]; !ok {
			return false
		}
	}
	return true
}

// validateLifecycleIdentity dispatches to the identity/ticket/manifest
// validation required for the configured lifecycle. It runs after the file
// config and data-directory-state checks and before embed.StartEtcd, so any
// inconsistency refuses startup before serving public traffic (design §716
// step 1, §785).
func validateLifecycleIdentity(cfg *Config, fileCfg *EmbeddedFileConfig) error {
	e := cfg.Embedded
	switch e.Lifecycle {
	case LifecycleBootstrap:
		m, err := ParseIdentityManifest(e.IdentityFile)
		if err != nil {
			return err
		}
		return ValidateBootstrapIdentity(m, fileCfg, cfg.DeploymentID)

	case LifecycleRestart:
		m, err := ParseIdentityManifest(e.IdentityFile)
		if err != nil {
			return err
		}
		// ValidateRestartIdentity now binds the deployment ID itself via
		// checkManifestConsistency, so no separate pre-check is needed here.
		return ValidateRestartIdentity(m, fileCfg, cfg.DeploymentID, fileCfg.DataDir)

	case LifecycleJoinExisting:
		m, err := ParseIdentityManifest(e.IdentityFile)
		if err != nil {
			return err
		}
		if err := ValidateJoinIdentity(m, fileCfg, cfg.DeploymentID); err != nil {
			return err
		}
		t, err := ParseMembershipTicket(e.MembershipTicketFile)
		if err != nil {
			return err
		}
		// Residual gap: the ticket is validated for consistency and expiry here,
		// but its atomic single-use consumption against the live cluster (design
		// §240, §719 step 2) requires the shared client and namespace and is
		// performed separately in the join startup path.
		return ValidateMembershipTicket(t, fileCfg, cfg.DeploymentID, time.Now())

	case LifecycleRestore:
		m, err := ParseRestoreManifest(e.RestoreManifestFile)
		if err != nil {
			return err
		}
		return ValidateRestoreManifest(m, fileCfg, cfg.DeploymentID)

	default:
		return fmt.Errorf("embedded etcd: unknown lifecycle %q", e.Lifecycle)
	}
}
