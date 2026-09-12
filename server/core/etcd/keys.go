// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the etcd key layout (design §"Data model"). All keys
// live below a normalized namespace and a version prefix. Tuple identities are
// canonically serialized and encoded with unpadded URL-safe base64; code must
// never depend on path escaping, substring matching, or glob semantics.
package etcd

import (
	"encoding/base64"
	"net"
	"strconv"
	"strings"
)

// SchemaVersion is the current key-layout version. It appears in the key prefix
// as v1 and is asserted by the stored schema marker.
const SchemaVersion = "v1"

// ProjectScope identifies a project lock. Unlike the legacy lock key
// (models.GenerateLockKey), it carries the VCS hostname so two hosts with
// identical repository names never collide (design §419).
type ProjectScope struct {
	VCSHostname string
	Repository  string // repo full name, e.g. "owner/repo"
	Path        string // repo-relative dir; "." at root
	Project     string // configured project name, may be empty
	Workspace   string
}

// PullScope identifies a pull request for locks, ownership, and cleanup.
type PullScope struct {
	VCSHostname string
	Repository  string
	PullNum     int
}

// Keyspace builds fully-qualified etcd keys for one normalized namespace.
type Keyspace struct {
	// root is the normalized namespace joined with the schema version, e.g.
	// "/atlantis/v1". It never has a trailing slash.
	root string
}

// NewKeyspace normalizes the configured namespace and binds it to the schema
// version. Normalization is done once (design §162); all keys derive from root.
func NewKeyspace(namespace string) Keyspace {
	return Keyspace{root: normalizeNamespace(namespace) + "/" + SchemaVersion}
}

// normalizeNamespace produces a canonical "/a/b" form: a single leading slash,
// no trailing slash, no empty or dot segments, no duplicate slashes.
func normalizeNamespace(ns string) string {
	segs := strings.Split(ns, "/")
	out := make([]string, 0, len(segs))
	for _, s := range segs {
		if s == "" || s == "." {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return "/atlantis"
	}
	return "/" + strings.Join(out, "/")
}

// Root returns the normalized namespace+version prefix. A range over Root+"/"
// covers exactly this deployment's keys.
func (k Keyspace) Root() string { return k.root }

// --- meta keys (design §601) ---

func (k Keyspace) SchemaKey() string     { return k.root + "/meta/schema" }
func (k Keyspace) DeploymentKey() string { return k.root + "/meta/deployment" }
func (k Keyspace) RecoveryQuarantineKey() string {
	return k.root + "/meta/recovery-quarantine"
}
func (k Keyspace) MigrationKey(id string) string {
	return k.root + "/meta/migrations/" + encodeSegment(id)
}

// MigrationActiveKey is a single well-known sentinel that serializes migrations:
// a create-only write of it admits exactly one concurrent migrator, closing the
// empty-namespace TOCTOU between BeginMigration and the manifest write.
func (k Keyspace) MigrationActiveKey() string {
	return k.root + "/meta/migration-active"
}
func (k Keyspace) RecoveryKey(id string) string {
	return k.root + "/meta/recoveries/" + encodeSegment(id)
}
func (k Keyspace) CapabilityKey(instanceID string) string {
	return k.root + "/meta/capabilities/" + encodeSegment(instanceID)
}
func (k Keyspace) MembershipTransitionKey(id string) string {
	return k.root + "/meta/membership-transitions/" + encodeSegment(id)
}
func (k Keyspace) MembershipTicketKey(id string) string {
	return k.root + "/meta/membership-tickets/" + encodeSegment(id)
}

// --- db keys (design §611) ---

// DBPrefix is the range prefix for the project-lock namespace. UnlockByPullScope
// ranges over exactly this prefix (design §442).
func (k Keyspace) ProjectLockPrefix() string { return k.root + "/db/project-locks/" }

func (k Keyspace) ProjectLockKey(s ProjectScope) string {
	return k.ProjectLockPrefix() + encodeProjectScope(s)
}
func (k Keyspace) PullLifecycleKey(s PullScope) string {
	return k.root + "/db/pull-lifecycle/" + encodePullScope(s)
}

// PullCleaningKey exists only while a pull is being cleaned up. Acquisition
// asserts its absence with CreateRevision==0, which (unlike a Value comparison
// on a possibly-absent key) is reliable in an etcd transaction.
func (k Keyspace) PullCleaningKey(s PullScope) string {
	return k.root + "/db/pull-cleaning/" + encodePullScope(s)
}
func (k Keyspace) PullStatusKey(s PullScope) string {
	return k.root + "/db/pulls/" + encodePullScope(s)
}
func (k Keyspace) GlobalLockKey(name string) string {
	return k.root + "/db/global-locks/" + encodeSegment(name)
}

// --- ownership / commands / execution keys (design §615) ---

func (k Keyspace) OwnershipKey(s PullScope) string {
	return k.root + "/ownership/pulls/" + encodePullScope(s)
}

// CommandPrefix ranges over every command-admission record for one coordination
// epoch, used by the dedup-window cleaner.
func (k Keyspace) CommandPrefix(epoch string) string {
	return k.root + "/commands/" + encodeSegment(epoch) + "/"
}

func (k Keyspace) CommandKey(epoch, deliveryID string) string {
	return k.CommandPrefix(epoch) + encodeSegment(deliveryID)
}

// ExecutionBarrierPrefix ranges over every barrier for one pull, across all
// generations, so a new owner can detect an unresolved older-generation barrier.
func (k Keyspace) ExecutionBarrierPrefix(s PullScope) string {
	return k.root + "/execution/pulls/" + encodePullScope(s) + "/"
}

func (k Keyspace) ExecutionBarrierKey(s PullScope, generation, executionID string) string {
	return k.ExecutionBarrierPrefix(s) + encodeSegment(generation) + "/" + encodeSegment(executionID)
}

// --- canonical encoding ---

// encodeProjectScope canonically serializes a ProjectScope and base64-encodes
// it. The serialization is injective (length-prefixed), so distinct scopes never
// produce the same key and no field delimiter can be spoofed.
func encodeProjectScope(s ProjectScope) string {
	return canonicalEncode(s.VCSHostname, s.Repository, s.Path, s.Project, s.Workspace)
}

func encodePullScope(s PullScope) string {
	return canonicalEncode(s.VCSHostname, s.Repository, strconv.Itoa(s.PullNum))
}

// encodeSegment encodes an opaque single-field identity (lock name, id) so it is
// safe as one key segment without depending on escaping.
func encodeSegment(v string) string {
	return canonicalEncode(v)
}

// canonicalEncode length-prefixes each field, concatenates, and applies unpadded
// URL-safe base64. Length prefixing guarantees the concatenation is a bijection
// of the field tuple regardless of field contents.
func canonicalEncode(fields ...string) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.WriteString(f)
	}
	return base64.RawURLEncoding.EncodeToString([]byte(b.String()))
}

// isLoopbackHost reports whether host is a loopback address or "localhost". Used
// to gate the insecure development mode (design §176).
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
