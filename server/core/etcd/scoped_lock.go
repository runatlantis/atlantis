// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the host-aware scoped project-lock store (design
// §"Etcd database adapter"). The legacy db.Database project-lock methods omit
// the VCS hostname, so the etcd HA path uses this supplemental contract, which
// carries an exact ProjectScope/PullScope through every operation. Etcd mode
// never invokes an ambiguous legacy project-lock operation.
package etcd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// leaseRevokeTimeout bounds a best-effort cleaning-lease revoke so a partitioned
// or hung etcd cannot block the calling goroutine indefinitely. A failed revoke
// is harmless: the cleaning marker is lease-bound and self-heals on TTL expiry
// (cleaningLeaseTTLSeconds).
const leaseRevokeTimeout = 5 * time.Second

// revokeCleaningLease revokes a cleaning lease under a bounded context.
func (s *scopedLockStore) revokeCleaningLease(leaseID clientv3.LeaseID) {
	ctx, cancel := context.WithTimeout(context.Background(), leaseRevokeTimeout)
	defer cancel()
	_, _ = s.lease.Revoke(ctx, leaseID)
}

const (
	lockRecordKind = "project-lock"
	lockRecordV1   = 1

	lifecycleRecordKind = "pull-lifecycle"
	lifecycleRecordV1   = 1

	// uiLockIDVersion prefixes the opaque UI lock ID so its decoder can never
	// fall back to the legacy models.GenerateLockKey parser (design §428).
	uiLockIDVersion = "e1"

	// cleaningLeaseTTLSeconds bounds how long a pull's cleaning marker can survive
	// after the cleaning process dies without finishing. The marker is attached to
	// this lease, so a crashed or hung cleanup self-heals instead of wedging the
	// pull's lock acquisition forever. A single pull's project-lock set is small,
	// so this is far longer than any real cleanup; the in-process error path drops
	// the marker immediately via lease revoke and does not wait for expiry.
	cleaningLeaseTTLSeconds = 300
)

// LifecycleState is the state of a pull-scoped lock lifecycle record (design §442).
type LifecycleState string

const (
	LifecycleOpen     LifecycleState = "open"
	LifecycleCleaning LifecycleState = "cleaning"
	LifecycleClosed   LifecycleState = "closed"
)

// lifecycleRecord tracks whether a pull is accepting new project locks. A
// generation advances on close/reopen so delayed deliveries cannot resurrect a
// stale state (design §450).
type lifecycleRecord struct {
	State      LifecycleState `json:"state"`
	Generation int64          `json:"generation"`
}

// ScopedProjectLockStore is the host-aware project-lock contract used by all
// etcd project-lock call sites (design §80).
type ScopedProjectLockStore interface {
	AcquireProjectLock(ctx context.Context, scope ProjectScope, lock models.ProjectLock) (acquired bool, current models.ProjectLock, err error)
	GetProjectLock(ctx context.Context, scope ProjectScope) (*models.ProjectLock, error)
	UnlockProjectScope(ctx context.Context, scope ProjectScope) (*models.ProjectLock, error)
	UnlockIfOwnedByPull(ctx context.Context, scope ProjectScope, pullNum int) (*models.ProjectLock, error)
	ListProjectLocks(ctx context.Context) ([]models.ProjectLock, error)
	UnlockByPullScope(ctx context.Context, scope PullScope, closeGen bool) ([]models.ProjectLock, error)
	ReopenProjectPull(ctx context.Context, scope PullScope) error
	UILockID(scope ProjectScope) string
	DecodeUILockID(id string) (ProjectScope, error)
}

// scopedLockStore is the etcd implementation of ScopedProjectLockStore.
type scopedLockStore struct {
	kv    clientv3.KV
	lease clientv3.Lease
	keys  Keyspace
}

// acquireLifecycleRetries bounds how many times AcquireProjectLock re-reads the
// pull lifecycle when a concurrent close/reopen invalidates its acquire
// transaction. A handful is ample: each retry follows a committed lifecycle
// transition, which is rare relative to lock acquisition.
const acquireLifecycleRetries = 4

// AcquireProjectLock acquires a project lock with a create-only transaction
// (design §434: CreateRevision(key) == 0). It refuses acquisition while the
// pull's lifecycle record is in the cleaning state, and refuses it outright when
// the pull's lifecycle is closed, so a delayed command cannot lock a project on a
// closed pull until the pull is reopened (design §442, §450). The acquire
// transaction is pinned to the lifecycle record's revision so a close committing
// concurrently invalidates the acquire rather than racing it (TOCTOU-safe).
func (s *scopedLockStore) AcquireProjectLock(ctx context.Context, scope ProjectScope, lock models.ProjectLock) (bool, models.ProjectLock, error) {
	key := s.keys.ProjectLockKey(scope)
	pull := PullScope{VCSHostname: scope.VCSHostname, Repository: scope.Repository, PullNum: lock.Pull.Num}
	cleaningKey := s.keys.PullCleaningKey(pull)
	lifecycleKey := s.keys.PullLifecycleKey(pull)

	val, err := encodeValue(lockRecordKind, lockRecordV1, lock)
	if err != nil {
		return false, models.ProjectLock{}, err
	}

	for attempt := 0; attempt < acquireLifecycleRetries; attempt++ {
		_, state, lrev, err := s.readLifecycle(ctx, lifecycleKey)
		if err != nil {
			return false, models.ProjectLock{}, err
		}
		if state == LifecycleClosed {
			return false, models.ProjectLock{}, errPullClosed
		}

		// The transaction commits only when the lock key does not yet exist, no
		// cleaning is in progress, AND the lifecycle record is unchanged since it
		// was read. Absence is expressed as CreateRevision==0 (reliable in a txn,
		// unlike a Value comparison on a possibly-absent key); the lifecycle
		// ModRevision guard closes the race against a concurrent close/reopen.
		resp, err := s.kv.Txn(ctx).
			If(
				clientv3.Compare(clientv3.CreateRevision(key), "=", 0),
				clientv3.Compare(clientv3.CreateRevision(cleaningKey), "=", 0),
				clientv3.Compare(clientv3.ModRevision(lifecycleKey), "=", lrev),
			).
			Then(clientv3.OpPut(key, string(val))).
			Else(clientv3.OpGet(key), clientv3.OpGet(cleaningKey)).
			Commit()
		if err != nil {
			return false, models.ProjectLock{}, fmt.Errorf("acquiring project lock: %w", err)
		}
		if resp.Succeeded {
			return true, lock, nil
		}

		// Classify the failure. If a lock exists, return the current holder; if a
		// cleaning is in progress, fail closed; otherwise the lifecycle changed
		// under us and we retry (the next read sees closed → refuse, or a new open
		// revision → retry the acquire).
		lockKvs := resp.Responses[0].GetResponseRange().Kvs
		if len(lockKvs) > 0 {
			var current models.ProjectLock
			if err := decodeValue(lockKvs[0].Value, lockRecordKind, lockRecordV1, lockRecordV1, &current); err != nil {
				return false, models.ProjectLock{}, err
			}
			return false, current, nil
		}
		if len(resp.Responses[1].GetResponseRange().Kvs) > 0 {
			return false, models.ProjectLock{}, errPullCleaning
		}
	}
	return false, models.ProjectLock{}, errors.New("project lock acquisition lost too many lifecycle races")
}

// errPullCleaning is returned when a project lock cannot be acquired because the
// pull is being cleaned up. Callers treat it as fail-closed.
var errPullCleaning = errors.New("pull is being cleaned up; project lock acquisition is refused")

// errPullClosed is returned when a project lock cannot be acquired because the
// pull's lifecycle is closed. It is cleared when the pull is reopened.
var errPullClosed = errors.New("pull is closed; project lock acquisition is refused until it is reopened")

// readLifecycle reads a pull's lifecycle record. An absent record is treated as
// open with revision 0 (so the acquire transaction's ModRevision==0 guard means
// "still absent"). It returns the record, its state, and its mod revision.
func (s *scopedLockStore) readLifecycle(ctx context.Context, lifecycleKey string) (lifecycleRecord, LifecycleState, int64, error) {
	resp, err := s.kv.Get(ctx, lifecycleKey)
	if err != nil {
		return lifecycleRecord{}, "", 0, fmt.Errorf("reading pull lifecycle: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return lifecycleRecord{State: LifecycleOpen}, LifecycleOpen, 0, nil
	}
	var rec lifecycleRecord
	if err := decodeValue(resp.Kvs[0].Value, lifecycleRecordKind, lifecycleRecordV1, lifecycleRecordV1, &rec); err != nil {
		return lifecycleRecord{}, "", 0, err
	}
	return rec, rec.State, resp.Kvs[0].ModRevision, nil
}

// ReopenProjectPull advances a closed pull's lifecycle back to open, bumping its
// generation, so a reopened pull request accepts new project locks again. It is a
// no-op (and cheap) when the pull is not closed (design §450). The transition is
// compare-and-swap on the lifecycle record's revision.
func (s *scopedLockStore) ReopenProjectPull(ctx context.Context, scope PullScope) error {
	lifecycleKey := s.keys.PullLifecycleKey(scope)
	rec, state, rev, err := s.readLifecycle(ctx, lifecycleKey)
	if err != nil {
		return err
	}
	if state != LifecycleClosed {
		return nil
	}
	next := lifecycleRecord{State: LifecycleOpen, Generation: rec.Generation + 1}
	val, err := encodeValue(lifecycleRecordKind, lifecycleRecordV1, next)
	if err != nil {
		return err
	}
	resp, err := s.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(lifecycleKey), "=", rev)).
		Then(clientv3.OpPut(lifecycleKey, string(val))).
		Commit()
	if err != nil {
		return fmt.Errorf("reopening pull lifecycle: %w", err)
	}
	// A lost CAS means another writer transitioned it concurrently; that writer's
	// state stands (open on reopen, or a fresh close). Either way, no error.
	_ = resp
	return nil
}

// GetProjectLock returns the lock at scope, or nil if absent.
func (s *scopedLockStore) GetProjectLock(ctx context.Context, scope ProjectScope) (*models.ProjectLock, error) {
	resp, err := s.kv.Get(ctx, s.keys.ProjectLockKey(scope))
	if err != nil {
		return nil, fmt.Errorf("getting project lock: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	lock, err := decodeLock(resp.Kvs[0].Value)
	if err != nil {
		return nil, err
	}
	return &lock, nil
}

// UnlockProjectScope deletes the lock at scope and returns it, or nil if absent.
// Deletion is conditional on the exact observed revision so a concurrent
// re-lock is never clobbered (design §436).
func (s *scopedLockStore) UnlockProjectScope(ctx context.Context, scope ProjectScope) (*models.ProjectLock, error) {
	key := s.keys.ProjectLockKey(scope)
	get, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("reading project lock for unlock: %w", err)
	}
	if len(get.Kvs) == 0 {
		return nil, nil
	}
	lock, err := decodeLock(get.Kvs[0].Value)
	if err != nil {
		return nil, err
	}
	rev := get.Kvs[0].ModRevision
	resp, err := s.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", rev)).
		Then(clientv3.OpDelete(key)).
		Commit()
	if err != nil {
		return nil, fmt.Errorf("deleting project lock: %w", err)
	}
	if !resp.Succeeded {
		// The lock changed under us; treat as no-op rather than deleting a newer
		// lock.
		return nil, nil
	}
	return &lock, nil
}

// UnlockIfOwnedByPull deletes the lock only when it is still held by pullNum,
// comparing the exact stored owner and revision (design §436).
func (s *scopedLockStore) UnlockIfOwnedByPull(ctx context.Context, scope ProjectScope, pullNum int) (*models.ProjectLock, error) {
	key := s.keys.ProjectLockKey(scope)
	get, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("reading project lock: %w", err)
	}
	if len(get.Kvs) == 0 {
		return nil, nil
	}
	lock, err := decodeLock(get.Kvs[0].Value)
	if err != nil {
		return nil, err
	}
	if lock.Pull.Num != pullNum {
		return nil, nil
	}
	rev := get.Kvs[0].ModRevision
	resp, err := s.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", rev)).
		Then(clientv3.OpDelete(key)).
		Commit()
	if err != nil {
		return nil, fmt.Errorf("deleting project lock: %w", err)
	}
	if !resp.Succeeded {
		return nil, nil
	}
	return &lock, nil
}

// ListProjectLocks returns every project lock, paginating with all pages pinned
// to the first response revision so a concurrent write cannot make an entry
// appear twice or be skipped (design §454).
func (s *scopedLockStore) ListProjectLocks(ctx context.Context) ([]models.ProjectLock, error) {
	var out []models.ProjectLock
	err := s.rangePinned(ctx, s.keys.ProjectLockPrefix(), func(kv *mvccKV) error {
		lock, err := decodeLock(kv.Value)
		if err != nil {
			return err
		}
		out = append(out, lock)
		return nil
	})
	return out, err
}

// UnlockByPullScope removes every project lock for a pull. It first publishes a
// cleaning lifecycle record that acquisition transactions observe, then ranges
// the project-lock namespace pinned to one revision, conditionally deleting each
// lock owned by the pull. It finishes by advancing the lifecycle to open (manual
// cleanup) or closed (pull-close cleanup) (design §442).
func (s *scopedLockStore) UnlockByPullScope(ctx context.Context, scope PullScope, closeGen bool) ([]models.ProjectLock, error) {
	leaseID, err := s.beginCleaning(ctx, scope)
	if err != nil {
		return nil, err
	}
	// Revoke the cleaning lease on every exit. On the success path finishCleaning
	// has already deleted the marker in a transaction, so this frees the now-empty
	// lease; on any error path it deletes the marker immediately rather than
	// waiting for the lease TTL, so a failed cleanup never wedges the pull.
	defer s.revokeCleaningLease(leaseID)

	// Capture the lifecycle revision at the start of cleanup. finishCleaning uses it
	// to detect a reopen that commits while this cleanup runs and, for a pull-close,
	// abandon the close rather than clobbering the newer open generation (design §450).
	_, _, startRev, err := s.readLifecycle(ctx, s.keys.PullLifecycleKey(scope))
	if err != nil {
		return nil, err
	}

	var removed []models.ProjectLock
	err = s.rangePinned(ctx, s.keys.ProjectLockPrefix(), func(kv *mvccKV) error {
		lock, err := decodeLock(kv.Value)
		if err != nil {
			return err
		}
		if lock.Pull.Num != scope.PullNum ||
			lock.Project.RepoFullName != scope.Repository ||
			lock.Pull.BaseRepo.VCSHost.Hostname != scope.VCSHostname {
			return nil
		}
		// Conditionally delete only this exact revision.
		resp, derr := s.kv.Txn(ctx).
			If(clientv3.Compare(clientv3.ModRevision(string(kv.Key)), "=", kv.ModRevision)).
			Then(clientv3.OpDelete(string(kv.Key))).
			Commit()
		if derr != nil {
			return fmt.Errorf("deleting project lock during pull cleanup: %w", derr)
		}
		if resp.Succeeded {
			removed = append(removed, lock)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.finishCleaning(ctx, scope, closeGen, startRev); err != nil {
		return removed, err
	}
	return removed, nil
}

// beginCleaning marks a pull as being cleaned up by creating a cleaning key that
// acquisition transactions observe as absence-of-cleaning failing. The key is
// attached to a bounded lease so an interrupted cleanup (client deadline, etcd
// blip, or a process crash) cannot leave the marker set forever and permanently
// refuse the pull's lock acquisition. It returns the lease ID so the caller can
// revoke it on completion.
func (s *scopedLockStore) beginCleaning(ctx context.Context, scope PullScope) (clientv3.LeaseID, error) {
	grant, err := s.lease.Grant(ctx, cleaningLeaseTTLSeconds)
	if err != nil {
		return 0, fmt.Errorf("granting cleaning lease: %w", err)
	}
	if _, err := s.kv.Put(ctx, s.keys.PullCleaningKey(scope), "1", clientv3.WithLease(grant.ID)); err != nil {
		s.revokeCleaningLease(grant.ID)
		return 0, fmt.Errorf("marking pull cleaning: %w", err)
	}
	return grant.ID, nil
}

// lifecycleWriteRetries bounds how many times finishCleaning re-reads and
// re-attempts its compare-and-swap when a concurrent lifecycle write intervenes.
const lifecycleWriteRetries = 4

// finishCleaning records the persistent lifecycle generation and clears the
// cleaning marker: open for manual cleanup, or a closed generation for pull-close
// cleanup (design §442). The write is a compare-and-swap on the lifecycle
// revision — mirroring ReopenProjectPull — so it advances the generation
// monotonically and never clobbers a concurrent reopen or cleanup (design §976).
// For a pull-close, if a reopen commits while this cleanup runs (the lifecycle
// moved to a newer open generation since cleaning began), the close is abandoned
// rather than overwriting that newer generation; the cleaning marker is still
// cleared so the reopened pull can acquire locks (design §450).
//
// Note: a stale close that arrives after a reopen has already fully completed
// cannot be distinguished here without authoritative VCS state; that revalidation
// belongs to the owner-routed close handler (§450). This CAS keeps the storage
// layer consistent and self-healing meanwhile.
func (s *scopedLockStore) finishCleaning(ctx context.Context, scope PullScope, closeGen bool, startRev int64) error {
	lifecycleKey := s.keys.PullLifecycleKey(scope)
	cleaningKey := s.keys.PullCleaningKey(scope)

	for attempt := 0; attempt < lifecycleWriteRetries; attempt++ {
		cur, curState, curRev, err := s.readLifecycle(ctx, lifecycleKey)
		if err != nil {
			return err
		}

		// A reopen committed during this pull-close cleanup: abandon the close so
		// the newer open generation stands, and just drop the cleaning marker.
		if closeGen && curState == LifecycleOpen && curRev != startRev {
			if _, err := s.kv.Delete(ctx, cleaningKey); err != nil {
				return fmt.Errorf("clearing cleaning marker after superseding reopen: %w", err)
			}
			return nil
		}

		state := LifecycleOpen
		if closeGen {
			state = LifecycleClosed
		}
		val, err := encodeValue(lifecycleRecordKind, lifecycleRecordV1, lifecycleRecord{State: state, Generation: cur.Generation + 1})
		if err != nil {
			return err
		}
		resp, err := s.kv.Txn(ctx).
			If(clientv3.Compare(clientv3.ModRevision(lifecycleKey), "=", curRev)).
			Then(
				clientv3.OpPut(lifecycleKey, string(val)),
				clientv3.OpDelete(cleaningKey),
			).Commit()
		if err != nil {
			return fmt.Errorf("finishing pull cleanup: %w", err)
		}
		if resp.Succeeded {
			return nil
		}
		// The lifecycle changed under us; re-read and retry.
	}
	return errors.New("finishing pull cleanup lost too many lifecycle races")
}

// UILockID returns a versioned opaque encoding of the scope for the lock UI. It
// includes the VCS hostname and cannot be parsed by the legacy decoder.
func (s *scopedLockStore) UILockID(scope ProjectScope) string {
	return uiLockIDVersion + "." + encodeProjectScope(scope)
}

// DecodeUILockID reverses UILockID. It refuses any value not carrying the etcd
// version prefix; there is no legacy fallback (design §428).
func (s *scopedLockStore) DecodeUILockID(id string) (ProjectScope, error) {
	prefix := uiLockIDVersion + "."
	if !strings.HasPrefix(id, prefix) {
		return ProjectScope{}, fmt.Errorf("not an etcd UI lock id: %q", id)
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(id, prefix))
	if err != nil {
		return ProjectScope{}, fmt.Errorf("decoding UI lock id: %w", err)
	}
	fields, err := canonicalDecode(raw, 5)
	if err != nil {
		return ProjectScope{}, err
	}
	return ProjectScope{
		VCSHostname: fields[0], Repository: fields[1], Path: fields[2],
		Project: fields[3], Workspace: fields[4],
	}, nil
}

func decodeLock(data []byte) (models.ProjectLock, error) {
	var lock models.ProjectLock
	if err := decodeValue(data, lockRecordKind, lockRecordV1, lockRecordV1, &lock); err != nil {
		return models.ProjectLock{}, err
	}
	lock.Time = lock.Time.Local()
	return lock, nil
}

// canonicalDecode reverses canonicalEncode's length-prefixed serialization,
// requiring exactly want fields.
func canonicalDecode(raw []byte, want int) ([]string, error) {
	s := string(raw)
	fields := make([]string, 0, want)
	for len(s) > 0 {
		colon := strings.IndexByte(s, ':')
		if colon < 0 {
			return nil, errors.New("malformed canonical encoding: missing length delimiter")
		}
		n, err := strconv.Atoi(s[:colon])
		if err != nil || n < 0 {
			return nil, fmt.Errorf("malformed canonical encoding: bad length %q", s[:colon])
		}
		s = s[colon+1:]
		if len(s) < n {
			return nil, errors.New("malformed canonical encoding: truncated field")
		}
		fields = append(fields, s[:n])
		s = s[n:]
	}
	if len(fields) != want {
		return nil, fmt.Errorf("expected %d fields, got %d", want, len(fields))
	}
	return fields, nil
}
