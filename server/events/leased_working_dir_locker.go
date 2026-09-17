// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
)

const pullLeaseOperationTimeout = 5 * time.Second

type pullLeaseStore interface {
	TryAcquirePullLease(context.Context, string, int, string, time.Duration) (bool, string, error)
	RenewPullLease(context.Context, string, int, string, time.Duration) (bool, error)
	ReleasePullLease(context.Context, string, int, string) (bool, error)
}

type pullLeaseValue struct {
	Token string `json:"token"`
	workingDirLock
}

type heldPullLease struct {
	value string
	ctx   context.Context
	stop  context.CancelFunc
	done  chan struct{}
	once  sync.Once
}

// LeasedWorkingDirLocker adds cross-process pull-level exclusion to an
// existing WorkingDirLocker while leaving its project-level parallelism intact.
type LeasedWorkingDirLocker struct {
	local            WorkingDirLocker
	store            pullLeaseStore
	logger           logging.SimpleLogging
	ttl              time.Duration
	leaseLossHandler func(error)

	mutex sync.Mutex
	held  map[string]*heldPullLease
}

func NewLeasedWorkingDirLocker(local WorkingDirLocker, store pullLeaseStore, logger logging.SimpleLogging, ttl time.Duration) *LeasedWorkingDirLocker {
	return &LeasedWorkingDirLocker{
		local:  local,
		store:  store,
		logger: logger,
		ttl:    ttl,
		held:   make(map[string]*heldPullLease),
		leaseLossHandler: func(err error) {
			// Atlantis cannot cancel an in-flight command yet. Continuing after
			// lease loss could corrupt a shared working directory, so fail-stop
			// the process before another replica can acquire the lease.
			panic(err)
		},
	}
}

func (l *LeasedWorkingDirLocker) TryLockPull(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	unlockLocal, err := l.local.TryLockPull(repoFullName, pullNum, cmdName, metadata)
	if err != nil {
		return func() {}, err
	}

	record := pullLeaseValue{
		Token: uuid.NewString(),
		workingDirLock: workingDirLock{
			Command:                cmdName,
			WorkingDirLockMetadata: metadata,
		},
	}
	serialized, err := json.Marshal(record)
	if err != nil {
		unlockLocal()
		return func() {}, fmt.Errorf("serializing pull lease: %w", err)
	}

	acquiredAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), pullLeaseOperationTimeout)
	acquired, current, err := l.store.TryAcquirePullLease(ctx, repoFullName, pullNum, string(serialized), l.ttl)
	cancel()
	if err != nil {
		unlockLocal()
		return func() {}, err
	}
	if !acquired {
		unlockLocal()
		var owner pullLeaseValue
		if err := json.Unmarshal([]byte(current), &owner); err != nil {
			return func() {}, fmt.Errorf("pull request %d is currently locked by another Atlantis process", pullNum)
		}
		return func() {}, newRemotePullLockError(pullNum, cmdName, owner.workingDirLock)
	}

	key := l.pullKey(repoFullName, pullNum)
	renewalCtx, stopRenewal := context.WithCancel(context.Background())
	lease := &heldPullLease{
		value: string(serialized),
		ctx:   renewalCtx,
		stop:  stopRenewal,
		done:  make(chan struct{}),
	}
	l.mutex.Lock()
	l.held[key] = lease
	l.mutex.Unlock()
	go l.renew(repoFullName, pullNum, lease, acquiredAt.Add(l.ttl))

	return func() {
		lease.once.Do(func() {
			lease.stop()
			<-lease.done

			ctx, cancel := context.WithTimeout(context.Background(), pullLeaseOperationTimeout)
			released, err := l.store.ReleasePullLease(ctx, repoFullName, pullNum, lease.value)
			cancel()
			if err != nil {
				l.logger.Err("releasing working directory lease for pull request %d: %s", pullNum, err)
			} else if !released {
				l.logger.Warn("working directory lease for pull request %d was no longer owned by this process", pullNum)
			}

			l.mutex.Lock()
			if l.held[key] == lease {
				delete(l.held, key)
			}
			l.mutex.Unlock()
			unlockLocal()
		})
	}, nil
}

func (l *LeasedWorkingDirLocker) TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	return l.local.TryLock(repoFullName, pullNum, workspace, path, projectName, cmdName, metadata)
}

func (l *LeasedWorkingDirLocker) HasCommandLock(repoFullName string, pullNum int, cmdName command.Name) bool {
	return l.local.HasCommandLock(repoFullName, pullNum, cmdName)
}

func (l *LeasedWorkingDirLocker) UnlockByPull(repoFullName string, pullNum int) {
	l.local.UnlockByPull(repoFullName, pullNum)
}

func (l *LeasedWorkingDirLocker) renew(repoFullName string, pullNum int, lease *heldPullLease, expiresAt time.Time) {
	defer close(lease.done)

	renewInterval := l.ttl / 3
	retryInterval := l.ttl / 12
	safetyMargin := l.ttl / 6
	timer := time.NewTimer(renewInterval)
	defer timer.Stop()
	retrying := false

	for {
		select {
		case <-lease.ctx.Done():
			return
		case <-timer.C:
			failStopAt := expiresAt.Add(-safetyMargin)
			if !time.Now().Before(failStopAt) {
				if lease.ctx.Err() != nil {
					return
				}
				l.leaseLossHandler(fmt.Errorf("unable to confirm ownership of working directory lease for %s pull request %d before expiry", repoFullName, pullNum))
				return
			}
			startedAt := time.Now()
			operationDeadline := startedAt.Add(pullLeaseOperationTimeout)
			if failStopAt.Before(operationDeadline) {
				operationDeadline = failStopAt
			}
			ctx, cancel := context.WithDeadline(lease.ctx, operationDeadline)
			renewed, err := l.store.RenewPullLease(ctx, repoFullName, pullNum, lease.value, l.ttl)
			cancel()
			if lease.ctx.Err() != nil {
				return
			}
			if err == nil && renewed {
				if retrying {
					l.logger.Info("Renewed working directory lease for %s pull request %d after a transient Redis error", repoFullName, pullNum)
				}
				expiresAt = startedAt.Add(l.ttl)
				retrying = false
				timer.Reset(renewInterval)
				continue
			}

			if err == nil {
				l.leaseLossHandler(fmt.Errorf("working directory lease for %s pull request %d is no longer owned by this process", repoFullName, pullNum))
				return
			}

			if !retrying {
				l.logger.Warn("Unable to renew working directory lease for %s pull request %d; retrying before the lease expires: %s", repoFullName, pullNum, err)
				retrying = true
			}
			if !time.Now().Before(failStopAt) {
				l.leaseLossHandler(fmt.Errorf("unable to confirm ownership of working directory lease for %s pull request %d before expiry: %w", repoFullName, pullNum, err))
				return
			}
			timer.Reset(min(retryInterval, time.Until(failStopAt)))
		}
	}
}

func (l *LeasedWorkingDirLocker) pullKey(repoFullName string, pullNum int) string {
	return fmt.Sprintf("%d:%s/%d", len(repoFullName), repoFullName, pullNum)
}

func newRemotePullLockError(pullNum int, cmdName command.Name, currentLock workingDirLock) error {
	return &workingDirLockError{
		message: fmt.Sprintf("cannot run %q: pull request %d is currently locked by %q%s.\n"+
			"Wait until the previous command is complete and try again", cmdName, pullNum, currentLock.Command, formatCommitSuffix(currentLock)),
		metadata: currentLock.WorkingDirLockMetadata,
	}
}
