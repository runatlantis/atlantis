// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the durable command-admission record (design §"Command
// dispatch and local fencing", §512). Internal forwarding may be duplicated or
// lost, so the receiving owner atomically validates its exact claim and creates
// a persistent record keyed by the immutable command identity, then advances it
// reserved -> scheduled -> running -> terminal. The record is an audit and
// idempotency record, not a distributed queue: no other replica consumes it.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	admissionKind = "command-admission"
	admissionV1   = 1

	// DedupWindow is the fixed retention for completed command records before
	// they are compare-and-swap deleted (design §647).
	DedupWindow = 24 * time.Hour
)

// AdmissionState is the lifecycle state of a command-admission record.
type AdmissionState string

const (
	AdmissionReserved  AdmissionState = "reserved"
	AdmissionScheduled AdmissionState = "scheduled"
	AdmissionRunning   AdmissionState = "running"
	AdmissionSucceeded AdmissionState = "succeeded"
	AdmissionFailed    AdmissionState = "failed"
	// AdmissionUncertain marks a command whose outcome cannot be positively
	// determined. It persists until explicit resolution and is never auto-replayed
	// (design §531).
	AdmissionUncertain AdmissionState = "uncertain"
)

// terminalStates are the states eligible for dedup-window cleanup.
func (s AdmissionState) terminal() bool {
	return s == AdmissionSucceeded || s == AdmissionFailed
}

// CommandIdentity is the immutable canonical identity of a command (design §509):
// source kind, normalized VCS hostname, and source delivery or request ID.
type CommandIdentity struct {
	SourceKind  string // e.g. "webhook", "api", "autoplan"
	VCSHostname string
	DeliveryID  string
}

func (c CommandIdentity) encode() string {
	return canonicalEncode(c.SourceKind, c.VCSHostname, c.DeliveryID)
}

// admissionRecord is the stored value.
type admissionRecord struct {
	Identity    CommandIdentity `json:"identity"`
	State       AdmissionState  `json:"state"`
	Epoch       string          `json:"epoch"`
	OwnerRev    int64           `json:"owner_rev"`   // owning claim's create revision
	InstanceID  string          `json:"instance_id"` // process that reserved it
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	ResultKind  string          `json:"result_kind,omitempty"`
	ResultError string          `json:"result_error,omitempty"`
}

// AdmissionRecord is the resolved record returned to callers.
type AdmissionRecord struct {
	Record      admissionRecord
	ModRevision int64
}

func (r AdmissionRecord) State() AdmissionState { return r.Record.State }

// AdmissionStore persists command-admission records under one coordination epoch.
type AdmissionStore struct {
	kv         clientv3.KV
	keys       Keyspace
	epoch      string
	instanceID string
}

// NewAdmissionStore constructs an admission store bound to a coordination epoch
// and the owning process instance.
func NewAdmissionStore(backend Backend, keys Keyspace, epoch, instanceID string) *AdmissionStore {
	return &AdmissionStore{kv: backend.Client().KV, keys: keys, epoch: epoch, instanceID: instanceID}
}

func (a *AdmissionStore) key(id CommandIdentity) string {
	return a.keys.CommandKey(a.epoch, id.encode())
}

// Reserve atomically validates the exact owning claim and creates a persistent
// reserved record. It commits only when no record exists AND the ownership key is
// still at the claim's creation revision, binding the reservation to that owner
// generation (design §514). If a record already exists it is returned for
// idempotent reconciliation.
func (a *AdmissionStore) Reserve(ctx context.Context, id CommandIdentity, claim Claim) (AdmissionRecord, bool, error) {
	key := a.key(id)
	ownerKey := a.keys.OwnershipKey(claim.Scope)
	now := time.Now().UTC()
	rec := admissionRecord{
		Identity:   id,
		State:      AdmissionReserved,
		Epoch:      a.epoch,
		OwnerRev:   claim.CreateRevision,
		InstanceID: a.instanceID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	val, err := encodeValue(admissionKind, admissionV1, rec)
	if err != nil {
		return AdmissionRecord{}, false, err
	}
	resp, err := a.kv.Txn(ctx).
		If(
			clientv3.Compare(clientv3.CreateRevision(key), "=", 0),
			clientv3.Compare(clientv3.CreateRevision(ownerKey), "=", claim.CreateRevision),
		).
		Then(clientv3.OpPut(key, string(val)), clientv3.OpGet(key)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return AdmissionRecord{}, false, fmt.Errorf("reserving command admission: %w", err)
	}
	if resp.Succeeded {
		got := resp.Responses[1].GetResponseRange()
		out, derr := decodeAdmission(got.Kvs[0].Value, got.Kvs[0].ModRevision)
		return out, true, derr
	}
	// Either the record already exists, or the owner generation no longer matches.
	got := resp.Responses[0].GetResponseRange()
	if len(got.Kvs) == 0 {
		return AdmissionRecord{}, false, errStaleClaim
	}
	out, derr := decodeAdmission(got.Kvs[0].Value, got.Kvs[0].ModRevision)
	return out, false, derr
}

// errStaleClaim indicates the owning claim changed between resolution and
// reservation; the ingress replica must refresh ownership and reroute (design
// §501).
var errStaleClaim = errors.New("ownership claim is stale; refresh and reroute")

// Get returns the admission record for id, or nil if absent, using a
// linearizable read.
func (a *AdmissionStore) Get(ctx context.Context, id CommandIdentity) (*AdmissionRecord, error) {
	resp, err := a.kv.Get(ctx, a.key(id))
	if err != nil {
		return nil, fmt.Errorf("reading command admission: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	out, err := decodeAdmission(resp.Kvs[0].Value, resp.Kvs[0].ModRevision)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Transition compare-and-swap advances a record from an expected state to a new
// state. It fails if the record moved under it, so a duplicate delivery cannot
// double-advance the state machine (design §516).
func (a *AdmissionStore) Transition(ctx context.Context, cur AdmissionRecord, to AdmissionState) (AdmissionRecord, error) {
	if !validTransition(cur.Record.State, to) {
		return AdmissionRecord{}, fmt.Errorf("invalid admission transition %s -> %s", cur.Record.State, to)
	}
	// Only the reserving instance may advance a record it owns.
	if cur.Record.InstanceID != a.instanceID {
		return AdmissionRecord{}, fmt.Errorf("admission record owned by instance %q, not this process", cur.Record.InstanceID)
	}
	next := cur.Record
	next.State = to
	next.UpdatedAt = time.Now().UTC()
	return a.casWrite(ctx, cur.ModRevision, next)
}

// Complete writes a terminal result (succeeded or failed) with the outcome.
func (a *AdmissionStore) Complete(ctx context.Context, cur AdmissionRecord, success bool, resultKind, resultErr string) (AdmissionRecord, error) {
	to := AdmissionSucceeded
	if !success {
		to = AdmissionFailed
	}
	if !validTransition(cur.Record.State, to) {
		return AdmissionRecord{}, fmt.Errorf("invalid admission transition %s -> %s", cur.Record.State, to)
	}
	next := cur.Record
	next.State = to
	next.ResultKind = resultKind
	next.ResultError = resultErr
	next.UpdatedAt = time.Now().UTC()
	return a.casWrite(ctx, cur.ModRevision, next)
}

// MarkUncertain forces a record to the uncertain terminal state from any
// non-terminal state. It is used when an owner dies after 202 but before
// completion, or when a duplicate cannot be reconciled (design §520, §531).
func (a *AdmissionStore) MarkUncertain(ctx context.Context, cur AdmissionRecord) (AdmissionRecord, error) {
	next := cur.Record
	next.State = AdmissionUncertain
	next.UpdatedAt = time.Now().UTC()
	return a.casWrite(ctx, cur.ModRevision, next)
}

// casWrite writes next only if the record is still at expectedRev, so a
// concurrent writer cannot be clobbered.
func (a *AdmissionStore) casWrite(ctx context.Context, expectedRev int64, next admissionRecord) (AdmissionRecord, error) {
	key := a.key(next.Identity)
	val, err := encodeValue(admissionKind, admissionV1, next)
	if err != nil {
		return AdmissionRecord{}, err
	}
	resp, err := a.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", expectedRev)).
		Then(clientv3.OpPut(key, string(val)), clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return AdmissionRecord{}, fmt.Errorf("advancing command admission: %w", err)
	}
	if !resp.Succeeded {
		return AdmissionRecord{}, errAdmissionConflict
	}
	got := resp.Responses[1].GetResponseRange()
	return decodeAdmission(got.Kvs[0].Value, got.Kvs[0].ModRevision)
}

// errAdmissionConflict indicates a concurrent writer advanced the record; the
// caller re-reads and reconciles rather than blindly retrying.
var errAdmissionConflict = errors.New("command admission record changed concurrently")

// CleanupExpired compare-and-swap deletes every terminal (succeeded/failed)
// command-admission record whose last update is older than DedupWindow, bounding
// keyspace growth (design §647). Uncertain and non-terminal records are never
// cleaned: uncertain persists until explicit resolution, and in-flight records
// are still authoritative. Each delete is guarded by the record's mod revision so
// a record advanced concurrently is left for the next pass. It returns the number
// of records deleted.
func (a *AdmissionStore) CleanupExpired(ctx context.Context, now time.Time) (int, error) {
	cutoff := now.Add(-DedupWindow)
	deleted := 0
	err := rangePinned(ctx, a.kv, a.keys.CommandPrefix(a.epoch), func(kv *mvccKV) error {
		rec, derr := decodeAdmission(kv.Value, kv.ModRevision)
		if derr != nil {
			// Skip records this binary cannot decode rather than deleting them.
			return nil //nolint:nilerr
		}
		if !rec.Record.State.terminal() || !rec.Record.UpdatedAt.Before(cutoff) {
			return nil
		}
		key := a.key(rec.Record.Identity)
		resp, txErr := a.kv.Txn(ctx).
			If(clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)).
			Then(clientv3.OpDelete(key)).
			Commit()
		if txErr != nil {
			return fmt.Errorf("deleting expired command admission: %w", txErr)
		}
		if resp.Succeeded {
			deleted++
		}
		return nil
	})
	return deleted, err
}

// validTransition enforces the state machine. Terminal and uncertain states are
// sinks; the dedup cleaner deletes terminal records after the window.
func validTransition(from, to AdmissionState) bool {
	switch from {
	case AdmissionReserved:
		return to == AdmissionScheduled || to == AdmissionUncertain
	case AdmissionScheduled:
		return to == AdmissionRunning || to == AdmissionUncertain
	case AdmissionRunning:
		return to == AdmissionSucceeded || to == AdmissionFailed || to == AdmissionUncertain
	default:
		return false
	}
}

func decodeAdmission(data []byte, rev int64) (AdmissionRecord, error) {
	var rec admissionRecord
	if err := decodeValue(data, admissionKind, admissionV1, admissionV1, &rec); err != nil {
		return AdmissionRecord{}, err
	}
	return AdmissionRecord{Record: rec, ModRevision: rev}, nil
}
