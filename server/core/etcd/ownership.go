// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the PR ownership store (design §"Ownership store"). Each
// process holds one renewable session lease; every PR ownership key it wins is
// attached to that lease, so a crashed process's claims expire. Claiming is a
// single key-absent transaction. The fencing generation is the key's etcd
// creation revision plus the coordination epoch. Reusing a replica ID never
// adopts an older process's claim because the record carries a unique per-process
// instance ID. Watches are wake-up only; every admission uses a linearizable read
// or transaction.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	ownershipKind = "pr-ownership"
	ownershipV1   = 1
)

// Generation is the fencing identity of a claim: the coordination epoch and the
// ownership key's etcd creation revision. A stale generation starts no work.
type Generation struct {
	Epoch          string
	CreateRevision int64
}

// ownershipRecord is the stored value of a PR ownership claim (design §469).
type ownershipRecord struct {
	ReplicaID    string    `json:"replica_id"`
	InstanceID   string    `json:"instance_id"`
	AdvertiseURL string    `json:"advertise_url"`
	ClaimID      string    `json:"claim_id"`
	ClaimedAt    time.Time `json:"claimed_at"`
	Epoch        string    `json:"epoch"`
}

// Claim is a resolved PR ownership claim.
type Claim struct {
	Scope          PullScope
	Record         ownershipRecord
	CreateRevision int64
}

// Generation returns the fencing generation of this claim.
func (c Claim) Generation() Generation {
	return Generation{Epoch: c.Record.Epoch, CreateRevision: c.CreateRevision}
}

// AdvertiseURL is the internal address of the owning process, used for
// forwarding.
func (c Claim) AdvertiseURL() string { return c.Record.AdvertiseURL }

// ClaimFromGeneration reconstructs a minimal Claim from a pull scope and fencing
// generation. It carries only the fields the execution-barrier path compares —
// the scope, the coordination epoch, and the ownership key's creation revision —
// and is used by the owner-side executor, which receives the generation in the
// forwarded command envelope rather than a full ownership record. The barrier
// transaction re-validates the live ownership key against this create revision,
// so a stale generation cannot start a step (barrier.go StartStep, design §546).
func ClaimFromGeneration(scope PullScope, gen Generation) Claim {
	return Claim{
		Scope:          scope,
		Record:         ownershipRecord{Epoch: gen.Epoch},
		CreateRevision: gen.CreateRevision,
	}
}

// OwnershipStore manages this process's PR ownership claims over one session
// lease.
type OwnershipStore struct {
	kv             clientv3.KV
	keys           Keyspace
	session        *concurrency.Session
	replicaID      string
	instanceID     string
	advertiseURL   string
	epoch          string
	requestTimeout time.Duration
}

// NewOwnershipStore establishes the process ownership session for a validated
// coordination epoch. The session lease TTL is the ownership TTL; the concurrency
// session keeps it alive until Close.
func NewOwnershipStore(backend Backend, keys Keyspace, epoch, replicaID, advertiseURL string, ttl, requestTimeout time.Duration) (*OwnershipStore, error) {
	if epoch == "" {
		return nil, errors.New("ownership store requires a non-empty coordination epoch")
	}
	ttlSeconds := int(ttl.Seconds())
	if ttlSeconds < int(MinOwnershipTTL.Seconds()) {
		return nil, fmt.Errorf("ownership TTL %s is below the minimum %s", ttl, MinOwnershipTTL)
	}
	session, err := concurrency.NewSession(backend.Client(), concurrency.WithTTL(ttlSeconds))
	if err != nil {
		return nil, fmt.Errorf("establishing ownership session: %w", err)
	}
	return &OwnershipStore{
		kv:             backend.Client().KV,
		keys:           keys,
		session:        session,
		replicaID:      replicaID,
		instanceID:     uuid.NewString(),
		advertiseURL:   advertiseURL,
		epoch:          epoch,
		requestTimeout: requestTimeout,
	}, nil
}

// InstanceID is the unique per-process identity, distinguishing a restarted
// process from its predecessor even under the same replica ID.
func (o *OwnershipStore) InstanceID() string { return o.instanceID }

// Done is closed when the session lease is lost. Callers become unready and
// cancel work when this fires (design §776).
func (o *OwnershipStore) Done() <-chan struct{} { return o.session.Done() }

// Claim attempts to win ownership of a pull. On success it returns the local
// claim; if another process already owns it, it returns that owner's claim.
// Exactly one process wins because the transaction commits only when the key is
// absent.
func (o *OwnershipStore) Claim(ctx context.Context, scope PullScope) (Claim, bool, error) {
	key := o.keys.OwnershipKey(scope)
	record := ownershipRecord{
		ReplicaID:    o.replicaID,
		InstanceID:   o.instanceID,
		AdvertiseURL: o.advertiseURL,
		ClaimID:      uuid.NewString(),
		ClaimedAt:    time.Now().UTC(),
		Epoch:        o.epoch,
	}
	val, err := encodeValue(ownershipKind, ownershipV1, record)
	if err != nil {
		return Claim{}, false, err
	}

	resp, err := o.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(val), clientv3.WithLease(o.session.Lease())), clientv3.OpGet(key)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return Claim{}, false, fmt.Errorf("claiming pull ownership: %w", err)
	}
	if resp.Succeeded {
		claim, err := claimFromKVSlice(scope, resp.Responses[1].GetResponseRange().Kvs)
		return claim, true, err
	}
	claim, err := claimFromKVSlice(scope, resp.Responses[0].GetResponseRange().Kvs)
	return claim, false, err
}

// Get performs a linearizable read of the current owner, or nil if unowned.
func (o *OwnershipStore) Get(ctx context.Context, scope PullScope) (*Claim, error) {
	resp, err := o.kv.Get(ctx, o.keys.OwnershipKey(scope))
	if err != nil {
		return nil, fmt.Errorf("reading pull ownership: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	claim, err := claimFromKVs(scope, resp.Kvs[0])
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

// OwnedLocally reports whether this exact process currently owns the claim at the
// given generation. It uses a linearizable read and matches the instance ID and
// creation revision, so a reused replica ID or a superseded generation is not
// treated as ownership (design §483, §486).
func (o *OwnershipStore) OwnedLocally(ctx context.Context, scope PullScope, gen Generation) (bool, error) {
	claim, err := o.Get(ctx, scope)
	if err != nil {
		return false, err
	}
	if claim == nil {
		return false, nil
	}
	return claim.Record.InstanceID == o.instanceID &&
		claim.CreateRevision == gen.CreateRevision &&
		claim.Record.Epoch == gen.Epoch, nil
}

// Release deletes an ownership claim this process holds, comparing the exact
// creation revision so a newer claim is never removed (design §482).
func (o *OwnershipStore) Release(ctx context.Context, claim Claim) error {
	key := o.keys.OwnershipKey(claim.Scope)
	resp, err := o.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", claim.CreateRevision)).
		Then(clientv3.OpDelete(key)).
		Commit()
	if err != nil {
		return fmt.Errorf("releasing pull ownership: %w", err)
	}
	if !resp.Succeeded {
		// The claim was already superseded or expired; nothing to release.
		return nil
	}
	return nil
}

// Close revokes the session lease, which deletes every ownership key this
// process holds.
func (o *OwnershipStore) Close() error {
	return o.session.Close()
}

func claimFromKVSlice(scope PullScope, kvs []*mvccpb.KeyValue) (Claim, error) {
	if len(kvs) == 0 {
		return Claim{}, errors.New("ownership claim missing after write")
	}
	return claimFromKVs(scope, kvs[0])
}

func claimFromKVs(scope PullScope, kv *mvccpb.KeyValue) (Claim, error) {
	var record ownershipRecord
	if err := decodeValue(kv.Value, ownershipKind, ownershipV1, ownershipV1, &record); err != nil {
		return Claim{}, err
	}
	return Claim{Scope: scope, Record: record, CreateRevision: kv.CreateRevision}, nil
}
