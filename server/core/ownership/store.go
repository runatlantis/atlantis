// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package ownership defines pull-request ownership for multi-replica Atlantis.
package ownership

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// ErrDraining indicates that a replica is no longer accepting routed work.
var ErrDraining = errors.New("replica ownership store is draining")

// Key identifies ownership for one pull request.
type Key struct {
	VCSHostname  string `json:"vcs_hostname"`
	RepoFullName string `json:"repo_full_name"`
	PullNum      int    `json:"pull_num"`
}

// NewKey creates an ownership key from a repository and pull request number.
func NewKey(repo models.Repo, pullNum int) Key {
	return Key{
		VCSHostname:  strings.ToLower(strings.TrimSpace(repo.VCSHost.Hostname)),
		RepoFullName: strings.TrimSpace(repo.FullName),
		PullNum:      pullNum,
	}
}

// Canonical returns the normalized string used to derive backend keys.
func (k Key) Canonical() (string, error) {
	hostname := strings.ToLower(strings.TrimSpace(k.VCSHostname))
	repoFullName := strings.TrimSpace(k.RepoFullName)
	if hostname == "" || repoFullName == "" || k.PullNum <= 0 {
		return "", fmt.Errorf("invalid ownership key for host %q repo %q pull %d", hostname, repoFullName, k.PullNum)
	}
	return fmt.Sprintf("%s\x00%s\x00%d", hostname, repoFullName, k.PullNum), nil
}

// Record describes the replica process that currently owns a pull request.
type Record struct {
	SchemaVersion int       `json:"schema_version"`
	ReplicaID     string    `json:"replica_id"`
	InstanceID    string    `json:"instance_id"`
	AdvertiseURL  string    `json:"advertise_url"`
	ClaimID       string    `json:"claim_id"`
	ClaimedAt     time.Time `json:"claimed_at"`
}

// Store coordinates pull request ownership across Atlantis replicas.
type Store interface {
	Claim(ctx context.Context, key Key) (Record, error)
	Admit(ctx context.Context, key Key, claimID string) (bool, error)
	Current(ctx context.Context, key Key) (Record, bool, error)
	Owns(key Key, claimID string) bool
	Release(ctx context.Context, key Key, claimID string) error
	BeginDrain()
	Ready(ctx context.Context) error
	Close() error
}
