// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements persistent execution barriers (design §"Execution
// admission boundaries", §546). Before starting every workflow subprocess or
// side-effecting step, the owner creates a barrier bound to its exact claim.
// Barriers are keyed by pull scope, owner generation, and execution ID, so
// parallel projects within one valid owner generation are supported. A new owner
// may claim after lease expiry but cannot start any step while an unresolved
// barrier from an older generation exists; that barrier favors same-PR safety
// over automatic takeover.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	barrierKind = "execution-barrier"
	barrierV1   = 1
)

// BarrierState is the state of an execution barrier.
type BarrierState string

const (
	BarrierActive    BarrierState = "active"
	BarrierUncertain BarrierState = "uncertain"
)

// barrierRecord is the stored value of an execution barrier.
type barrierRecord struct {
	Scope       PullScope    `json:"scope"`
	Epoch       string       `json:"epoch"`
	OwnerRev    int64        `json:"owner_rev"`
	ExecutionID string       `json:"execution_id"`
	State       BarrierState `json:"state"`
	InstanceID  string       `json:"instance_id"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Barrier is a handle to a created execution barrier.
type Barrier struct {
	key    string
	record barrierRecord
	rev    int64
}

// ExecutionID returns the barrier's execution ID.
func (b Barrier) ExecutionID() string { return b.record.ExecutionID }

// ExecutionBarrierStore manages execution barriers for a coordination epoch and
// process instance.
type ExecutionBarrierStore struct {
	kv         clientv3.KV
	keys       Keyspace
	epoch      string
	instanceID string
}

// NewExecutionBarrierStore constructs the barrier store.
func NewExecutionBarrierStore(backend Backend, keys Keyspace, epoch, instanceID string) *ExecutionBarrierStore {
	return &ExecutionBarrierStore{kv: backend.Client().KV, keys: keys, epoch: epoch, instanceID: instanceID}
}

// generationString encodes the fencing generation for the barrier key.
func generationString(gen Generation) string {
	return gen.Epoch + "-" + strconv.FormatInt(gen.CreateRevision, 10)
}

// errBlockedByOlderGeneration is returned when a new owner cannot start work
// because an unresolved barrier from an older generation still exists. The PR is
// blocked until authenticated completion or explicit resolution (design §558).
var errBlockedByOlderGeneration = errors.New("execution blocked by an unresolved barrier from an older owner generation")

// errClaimLostAtBarrier is returned when the barrier transaction fails because
// the owning claim is no longer at the expected generation (design §560).
var errClaimLostAtBarrier = errors.New("claim lost before execution barrier could be established")

// ErrBlockedByOlderGeneration exposes the sentinel for callers that must map it
// to the uncertain-outcome policy.
func ErrBlockedByOlderGeneration() error { return errBlockedByOlderGeneration }

// StartStep establishes an execution barrier for a step. It first refuses to
// proceed when any unresolved older-generation barrier exists for the pull, then
// creates the barrier in one transaction that compares the exact owning claim
// (design §546, §558).
func (s *ExecutionBarrierStore) StartStep(ctx context.Context, claim Claim, executionID string) (Barrier, error) {
	gen := claim.Generation()

	blocked, err := s.blockedByOlderGeneration(ctx, claim.Scope, gen)
	if err != nil {
		return Barrier{}, err
	}
	if blocked {
		return Barrier{}, errBlockedByOlderGeneration
	}

	key := s.keys.ExecutionBarrierKey(claim.Scope, generationString(gen), executionID)
	ownerKey := s.keys.OwnershipKey(claim.Scope)
	rec := barrierRecord{
		Scope:       claim.Scope,
		Epoch:       gen.Epoch,
		OwnerRev:    gen.CreateRevision,
		ExecutionID: executionID,
		State:       BarrierActive,
		InstanceID:  s.instanceID,
		CreatedAt:   time.Now().UTC(),
	}
	val, err := encodeValue(barrierKind, barrierV1, rec)
	if err != nil {
		return Barrier{}, err
	}
	resp, err := s.kv.Txn(ctx).
		If(
			clientv3.Compare(clientv3.CreateRevision(ownerKey), "=", gen.CreateRevision),
			clientv3.Compare(clientv3.CreateRevision(key), "=", 0),
		).
		Then(clientv3.OpPut(key, string(val)), clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return Barrier{}, fmt.Errorf("creating execution barrier: %w", err)
	}
	if !resp.Succeeded {
		return Barrier{}, errClaimLostAtBarrier
	}
	got := resp.Responses[1].GetResponseRange()
	return Barrier{key: key, record: rec, rev: got.Kvs[0].ModRevision}, nil
}

// CompleteStep clears a barrier after the child process has exited and the step
// result is recorded. It deletes only the exact barrier revision (design §551).
func (s *ExecutionBarrierStore) CompleteStep(ctx context.Context, b Barrier) error {
	resp, err := s.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(b.key), "=", b.rev)).
		Then(clientv3.OpDelete(b.key)).
		Commit()
	if err != nil {
		return fmt.Errorf("clearing execution barrier: %w", err)
	}
	if !resp.Succeeded {
		return errors.New("execution barrier changed before it could be cleared")
	}
	return nil
}

// MarkUncertain converts an active barrier to a persistent uncertain barrier that
// blocks the PR until explicit operator resolution (design §558).
func (s *ExecutionBarrierStore) MarkUncertain(ctx context.Context, b Barrier) error {
	next := b.record
	next.State = BarrierUncertain
	val, err := encodeValue(barrierKind, barrierV1, next)
	if err != nil {
		return err
	}
	resp, err := s.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(b.key), "=", b.rev)).
		Then(clientv3.OpPut(b.key, string(val))).
		Commit()
	if err != nil {
		return fmt.Errorf("marking execution barrier uncertain: %w", err)
	}
	if !resp.Succeeded {
		return errors.New("execution barrier changed before it could be marked uncertain")
	}
	return nil
}

// blockedByOlderGeneration reports whether any barrier for the pull belongs to a
// different (older) generation and is still active or uncertain. Same-generation
// barriers never block, so parallel projects under one valid owner proceed.
func (s *ExecutionBarrierStore) blockedByOlderGeneration(ctx context.Context, scope PullScope, current Generation) (bool, error) {
	prefix := s.keys.ExecutionBarrierPrefix(scope)
	blocked := false
	err := rangePinned(ctx, s.kv, prefix, func(kv *mvccKV) error {
		var rec barrierRecord
		if derr := decodeValue(kv.Value, barrierKind, barrierV1, barrierV1, &rec); derr != nil {
			return derr
		}
		sameGen := rec.Epoch == current.Epoch && rec.OwnerRev == current.CreateRevision
		if !sameGen && (rec.State == BarrierActive || rec.State == BarrierUncertain) {
			blocked = true
		}
		return nil
	})
	return blocked, err
}
