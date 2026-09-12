// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements EtcdDatabase, the db.Database adapter (design §"Etcd
// database adapter"). Status, global-lock, health, and lifecycle behavior match
// BoltDB semantics without changing BoltDB. Project locking is delegated to the
// host-aware ScopedProjectLockStore; the legacy single-project methods here
// exist for db.Database compatibility and are host-exact only in single-host
// deployments (etcd call sites use the scoped contract directly — design §82).
//
// All operations use bounded contexts internally because the common db.Database
// interface does not carry caller contexts (design §432).
package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdDatabase implements the common db.Database contract.
var _ db.Database = (*EtcdDatabase)(nil)

const (
	pullStatusKind = "pull-status"
	pullStatusV1   = 1

	globalLockKind = "global-lock"
	globalLockV1   = 1

	// casRetries bounds the compare-and-swap retry loop for status merges. It
	// must exceed the realistic number of concurrent writers to one pull so the
	// last writer in a contended burst still commits.
	casRetries = 64
)

// EtcdDatabase implements db.Database over etcd. It borrows the backend's client
// and delegates its Close to the backend (design §99).
type EtcdDatabase struct {
	backend        Backend
	kv             clientv3.KV
	keys           Keyspace
	locks          *scopedLockStore
	requestTimeout time.Duration
}

// NewDatabase builds an EtcdDatabase over an already-probed backend.
func NewDatabase(backend Backend, namespace string, requestTimeout time.Duration) *EtcdDatabase {
	keys := NewKeyspace(namespace)
	kv := backend.Client().KV
	return &EtcdDatabase{
		backend:        backend,
		kv:             kv,
		keys:           keys,
		locks:          &scopedLockStore{kv: kv, lease: backend.Client().Lease, keys: keys},
		requestTimeout: requestTimeout,
	}
}

// Scoped exposes the host-aware project-lock store used by etcd call sites.
func (d *EtcdDatabase) Scoped() ScopedProjectLockStore { return d.locks }

func (d *EtcdDatabase) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d.requestTimeout)
}

// projectScope derives a ProjectScope from a full ProjectLock, which carries the
// VCS hostname via Pull.BaseRepo.
func projectScopeFromLock(l models.ProjectLock) ProjectScope {
	return ProjectScope{
		VCSHostname: l.Pull.BaseRepo.VCSHost.Hostname,
		Repository:  l.Project.RepoFullName,
		Path:        l.Project.Path,
		Project:     l.Project.ProjectName,
		Workspace:   l.Workspace,
	}
}

// --- project locks (db.Database compat; scoped path is authoritative) ---

func (d *EtcdDatabase) TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	acquired, current, err := d.locks.AcquireProjectLock(ctx, projectScopeFromLock(lock), lock)
	if errors.Is(err, errPullCleaning) {
		// Fail closed: a lock cannot be acquired while the pull is cleaning.
		return false, models.ProjectLock{}, err
	}
	return acquired, current, err
}

func (d *EtcdDatabase) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	scope, ok, err := d.findScopeByProject(ctx, project, workspace)
	if err != nil || !ok {
		return nil, err
	}
	return d.locks.UnlockProjectScope(ctx, scope)
}

func (d *EtcdDatabase) UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	scope, ok, err := d.findScopeByProject(ctx, project, workspace)
	if err != nil || !ok {
		return nil, err
	}
	return d.locks.UnlockIfOwnedByPull(ctx, scope, pullNum)
}

func (d *EtcdDatabase) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	scope, ok, err := d.findScopeByProject(ctx, project, workspace)
	if err != nil || !ok {
		return nil, err
	}
	return d.locks.GetProjectLock(ctx, scope)
}

func (d *EtcdDatabase) List() ([]models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	return d.locks.ListProjectLocks(ctx)
}

func (d *EtcdDatabase) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	// The legacy signature lacks a VCS hostname, so resolve the scope by
	// scanning for a lock matching the repo/pull. Etcd call sites use
	// UnlockByPullScope directly with an exact PullScope.
	scope, ok, err := d.findPullScope(ctx, repoFullName, pullNum)
	if err != nil || !ok {
		return nil, err
	}
	return d.locks.UnlockByPullScope(ctx, scope, false)
}

// UnlockByPullForClose deletes every project lock for a closed pull and advances
// the pull's lifecycle to closed. Unlike the legacy UnlockByPull it takes the
// exact VCS hostname, so there is no scan and no cross-host ambiguity, and it
// closes the lifecycle generation rather than leaving it open (design §442,
// §450). The pull-close executor calls this in etcd mode; the other, host-less
// UnlockByPull callers are non-close unlocks that must not close the generation.
func (d *EtcdDatabase) UnlockByPullForClose(repoFullName, vcsHostname string, pullNum int) ([]models.ProjectLock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	scope := PullScope{VCSHostname: vcsHostname, Repository: repoFullName, PullNum: pullNum}
	return d.locks.UnlockByPullScope(ctx, scope, true)
}

// errAmbiguousHost is returned when a host-less legacy lookup matches locks on
// more than one VCS host, so no host can be resolved without risking acting on
// the wrong host's lock. Etcd call sites use the scoped API with an exact
// hostname and never hit this; the legacy path fails closed rather than guess
// (design §419: two hosts with identical repo names must stay isolated).
var errAmbiguousHost = errors.New("host-less lookup matched locks on multiple VCS hosts; use the host-aware scoped API")

// findScopeByProject resolves a ProjectScope from a host-less models.Project by
// scanning existing locks. It returns the single match; if matches span more
// than one VCS host it fails closed rather than returning an arbitrary one.
func (d *EtcdDatabase) findScopeByProject(ctx context.Context, project models.Project, workspace string) (ProjectScope, bool, error) {
	locks, err := d.locks.ListProjectLocks(ctx)
	if err != nil {
		return ProjectScope{}, false, err
	}
	var match *models.ProjectLock
	for i := range locks {
		l := &locks[i]
		if l.Project.RepoFullName == project.RepoFullName &&
			l.Project.Path == project.Path &&
			l.Project.ProjectName == project.ProjectName &&
			l.Workspace == workspace {
			if match != nil && match.Pull.BaseRepo.VCSHost.Hostname != l.Pull.BaseRepo.VCSHost.Hostname {
				return ProjectScope{}, false, errAmbiguousHost
			}
			match = l
		}
	}
	if match == nil {
		return ProjectScope{}, false, nil
	}
	return projectScopeFromLock(*match), true, nil
}

func (d *EtcdDatabase) findPullScope(ctx context.Context, repoFullName string, pullNum int) (PullScope, bool, error) {
	locks, err := d.locks.ListProjectLocks(ctx)
	if err != nil {
		return PullScope{}, false, err
	}
	var host string
	found := false
	for _, l := range locks {
		if l.Project.RepoFullName == repoFullName && l.Pull.Num == pullNum {
			if found && host != l.Pull.BaseRepo.VCSHost.Hostname {
				return PullScope{}, false, errAmbiguousHost
			}
			host = l.Pull.BaseRepo.VCSHost.Hostname
			found = true
		}
	}
	if !found {
		return PullScope{}, false, nil
	}
	return PullScope{VCSHostname: host, Repository: repoFullName, PullNum: pullNum}, true, nil
}

// --- pull/project status (CAS merge, design §438) ---

func (d *EtcdDatabase) pullScope(pull models.PullRequest) PullScope {
	return PullScope{
		VCSHostname: pull.BaseRepo.VCSHost.Hostname,
		Repository:  pull.BaseRepo.FullName,
		PullNum:     pull.Num,
	}
}

func (d *EtcdDatabase) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	status, _, err := d.getPullStatus(ctx, d.keys.PullStatusKey(d.pullScope(pull)))
	return status, err
}

func (d *EtcdDatabase) getPullStatus(ctx context.Context, key string) (*models.PullStatus, int64, error) {
	resp, err := d.kv.Get(ctx, key)
	if err != nil {
		return nil, 0, fmt.Errorf("reading pull status: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, 0, nil
	}
	var status models.PullStatus
	if err := decodeValue(resp.Kvs[0].Value, pullStatusKind, pullStatusV1, pullStatusV1, &status); err != nil {
		// Tolerate an unreadable prior entry, matching BoltDB (design parity).
		return nil, resp.Kvs[0].ModRevision, nil
	}
	return &status, resp.Kvs[0].ModRevision, nil
}

func (d *EtcdDatabase) DeletePullStatus(pull models.PullRequest) error {
	ctx, cancel := d.ctx()
	defer cancel()
	if _, err := d.kv.Delete(ctx, d.keys.PullStatusKey(d.pullScope(pull))); err != nil {
		return fmt.Errorf("deleting pull status: %w", err)
	}
	return nil
}

// UpdatePullWithResults merges results into the stored pull status under a
// revision-based compare-and-swap loop so concurrent updates are not lost
// (design §438).
func (d *EtcdDatabase) UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error) {
	key := d.keys.PullStatusKey(d.pullScope(pull))
	for attempt := 0; attempt < casRetries; attempt++ {
		ctx, cancel := d.ctx()
		curr, rev, err := d.getPullStatus(ctx, key)
		if err != nil {
			cancel()
			return models.PullStatus{}, err
		}
		merged := mergePullStatus(curr, pull, newResults)
		val, err := encodeValue(pullStatusKind, pullStatusV1, merged)
		if err != nil {
			cancel()
			return models.PullStatus{}, err
		}
		ok, err := d.casPut(ctx, key, string(val), rev)
		cancel()
		if err != nil {
			return models.PullStatus{}, err
		}
		if ok {
			return merged, nil
		}
		casBackoff(attempt)
	}
	return models.PullStatus{}, errors.New("pull status update lost too many compare-and-swap races")
}

// UpdateProjectStatus updates a single project's plan status under a CAS loop.
func (d *EtcdDatabase) UpdateProjectStatus(pull models.PullRequest, workspace, repoRelDir string, newStatus models.ProjectPlanStatus) error {
	key := d.keys.PullStatusKey(d.pullScope(pull))
	for attempt := 0; attempt < casRetries; attempt++ {
		ctx, cancel := d.ctx()
		curr, rev, err := d.getPullStatus(ctx, key)
		if err != nil {
			cancel()
			return err
		}
		if curr == nil {
			cancel()
			return nil
		}
		updated := *curr
		for i := range updated.Projects {
			proj := &updated.Projects[i]
			if proj.Workspace == workspace && proj.RepoRelDir == repoRelDir {
				proj.Status = newStatus
				break
			}
		}
		val, err := encodeValue(pullStatusKind, pullStatusV1, updated)
		if err != nil {
			cancel()
			return err
		}
		ok, err := d.casPut(ctx, key, string(val), rev)
		cancel()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		casBackoff(attempt)
	}
	return errors.New("project status update lost too many compare-and-swap races")
}

// casBackoff sleeps a small, growing, jittered interval between compare-and-swap
// attempts to reduce livelock under contention.
func casBackoff(attempt int) {
	d := time.Duration(attempt+1) * time.Millisecond
	if d > 25*time.Millisecond {
		d = 25 * time.Millisecond
	}
	// Cheap jitter from the wall clock to desynchronize contending writers.
	d += time.Duration(time.Now().UnixNano()%int64(time.Millisecond)) % time.Millisecond
	time.Sleep(d)
}

// casPut writes val only if the key is still at expectedRev (0 = absent),
// returning whether the swap committed.
func (d *EtcdDatabase) casPut(ctx context.Context, key, val string, expectedRev int64) (bool, error) {
	resp, err := d.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", expectedRev)).
		Then(clientv3.OpPut(key, val)).
		Commit()
	if err != nil {
		return false, fmt.Errorf("compare-and-swap put: %w", err)
	}
	return resp.Succeeded, nil
}

// --- global command locks (design §434) ---

func (d *EtcdDatabase) LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	lock := command.Lock{
		CommandName:  cmdName,
		LockMetadata: command.LockMetadata{UnixTime: lockTime.Unix()},
	}
	val, err := encodeValue(globalLockKind, globalLockV1, lock)
	if err != nil {
		return nil, err
	}
	key := d.keys.GlobalLockKey(cmdName.String())
	resp, err := d.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(val))).
		Commit()
	if err != nil {
		return nil, fmt.Errorf("acquiring global lock: %w", err)
	}
	if !resp.Succeeded {
		return nil, errors.New("lock already exists")
	}
	return &lock, nil
}

func (d *EtcdDatabase) UnlockCommand(cmdName command.Name) error {
	ctx, cancel := d.ctx()
	defer cancel()
	key := d.keys.GlobalLockKey(cmdName.String())
	resp, err := d.kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "!=", 0)).
		Then(clientv3.OpDelete(key)).
		Commit()
	if err != nil {
		return fmt.Errorf("releasing global lock: %w", err)
	}
	if !resp.Succeeded {
		return errors.New("no lock exists")
	}
	return nil
}

func (d *EtcdDatabase) CheckCommandLock(cmdName command.Name) (*command.Lock, error) {
	ctx, cancel := d.ctx()
	defer cancel()
	resp, err := d.kv.Get(ctx, d.keys.GlobalLockKey(cmdName.String()))
	if err != nil {
		return nil, fmt.Errorf("checking global lock: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	var lock command.Lock
	if err := decodeValue(resp.Kvs[0].Value, globalLockKind, globalLockV1, globalLockV1, &lock); err != nil {
		return nil, err
	}
	return &lock, nil
}

// --- health / lifecycle ---

// Ping performs a bounded linearizable range operation (design §456).
func (d *EtcdDatabase) Ping() error {
	ctx, cancel := d.ctx()
	defer cancel()
	return d.backend.Ready(ctx)
}

// Close is the designated idempotent delegate to Backend.Close (design §99).
func (d *EtcdDatabase) Close() error {
	return d.backend.Close()
}

// mergePullStatus reproduces BoltDB's UpdatePullWithResults merge semantics
// (server/core/boltdb/boltdb.go): replace on outdated/absent status preserving
// prior policy status, otherwise merge per-project results in place.
func mergePullStatus(curr *models.PullStatus, pull models.PullRequest, newResults []command.ProjectResult) models.PullStatus {
	if curr == nil || pullStatusOutdatedForPull(curr.Pull, pull) {
		statuses := make([]models.ProjectStatus, 0, len(newResults))
		for _, r := range newResults {
			statuses = append(statuses, projectResultToProjectStatus(r))
		}
		if curr != nil {
			for i := range statuses {
				for _, old := range curr.Projects {
					if statuses[i].Workspace == old.Workspace &&
						statuses[i].RepoRelDir == old.RepoRelDir &&
						statuses[i].ProjectName == old.ProjectName &&
						len(old.PolicyStatus) > 0 {
						statuses[i].PolicyStatus = old.PolicyStatus
						break
					}
				}
			}
		}
		return models.PullStatus{Pull: pull, Projects: statuses}
	}

	merged := *curr
	for _, res := range newResults {
		updatedExisting := false
		for i := range merged.Projects {
			proj := &merged.Projects[i]
			if res.Workspace == proj.Workspace && res.RepoRelDir == proj.RepoRelDir && res.ProjectName == proj.ProjectName {
				proj.Status = res.PlanStatus()
				if len(proj.PolicyStatus) > 0 {
					for i, oldPS := range proj.PolicyStatus {
						for _, newPS := range res.PolicyStatus() {
							if oldPS.PolicySetName == newPS.PolicySetName {
								proj.PolicyStatus[i] = newPS
							}
						}
					}
				} else {
					proj.PolicyStatus = res.PolicyStatus()
				}
				updatedExisting = true
				break
			}
		}
		if !updatedExisting {
			merged.Projects = append(merged.Projects, projectResultToProjectStatus(res))
		}
	}
	return merged
}

// pullStatusOutdatedForPull mirrors the BoltDB predicate of the same name.
func pullStatusOutdatedForPull(statusPull, pull models.PullRequest) bool {
	if statusPull.HeadCommit != pull.HeadCommit {
		return true
	}
	if pull.BaseBranch == "" {
		return false
	}
	return statusPull.BaseBranch == "" || statusPull.BaseBranch != pull.BaseBranch
}

func projectResultToProjectStatus(p command.ProjectResult) models.ProjectStatus {
	return models.ProjectStatus{
		Workspace:    p.Workspace,
		RepoRelDir:   p.RepoRelDir,
		ProjectName:  p.ProjectName,
		PolicyStatus: p.PolicyStatus(),
		Status:       p.PlanStatus(),
	}
}
