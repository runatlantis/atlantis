// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_working_dir_locker.go WorkingDirLocker

// WorkingDirLocker is used to prevent multiple commands from executing
// at the same time for a single repo, pull, and workspace. We need to prevent
// this from happening because a specific repo/pull/workspace has a single workspace
// on disk and we haven't written Atlantis (yet) to handle concurrent execution
// within this workspace.
type WorkingDirLockMetadata struct {
	HeadCommit string
	CommitURL  string
	JobURL     string
}

type workingDirLockError struct {
	message      string
	metadata     WorkingDirLockMetadata
	multipleJobs bool
}

func (e *workingDirLockError) Error() string {
	return e.message
}

type workingDirLock struct {
	owner *pullLockWaiter
	// shared counts the holders of a pull-level lock taken with
	// TryLockPullShared. It is zero for exclusive locks.
	shared int
	// autoplan marks a pull-level lock taken by LockPullAfterStalePlan.
	autoplan bool
	Command  command.Name
	WorkingDirLockMetadata
}

var (
	// ErrPullLockSuperseded is returned by LockPullAfterStalePlan when the
	// caller stopped waiting because it is no longer the newest commit
	// waiting for this pull request, or because a plan for its commit is
	// already running.
	ErrPullLockSuperseded = errors.New("pull lock wait superseded")
	// ErrPullLockCancelled is returned by LockPullAfterStalePlan when the
	// caller stopped waiting because the pull request's locks were released
	// with UnlockByPull (pull request closed or command cancelled).
	ErrPullLockCancelled = errors.New("pull lock wait cancelled")
)

// pullLockWaiter is the single queued pull-level lock request for a pull
// request until it is authorized to start. reason is set when it is cancelled or superseded.
type pullLockWaiter struct {
	reason error
}

type PullLock struct {
	Start          func() bool
	Unlock         func()
	UpdateMetadata func(WorkingDirLockMetadata) bool
}

type WorkingDirLocker interface {
	// TryLockPull tries to acquire a pull-level lock for a command. It is used to
	// coordinate commands that can make pull-wide plan state inconsistent.
	TryLockPull(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error)
	// TryLockPullShared acquires the pull-level lock like TryLockPull, but
	// several holders with the same command name and head commit can hold it
	// at once. It is used by targeted applies, which only touch the projects
	// they name; TryLock still serializes them per project. It conflicts with
	// any exclusive pull-level lock and with running plans.
	TryLockPullShared(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error)
	// LockPullAfterStalePlan acquires the pull-level lock like TryLockPull,
	// except when the pull request is held by a plan for a different head
	// commit. In that case it calls onWait with the holder's metadata and
	// blocks until that plan releases its locks, then takes the lock for
	// metadata.HeadCommit.
	//
	// Only the newest waiter per pull request is kept. An older waiter
	// returns ErrPullLockSuperseded, and UnlockByPull makes waiters return
	// ErrPullLockCancelled. Any other conflict returns the same error as
	// TryLockPull without waiting.
	LockPullAfterStalePlan(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata, onWait func(holder WorkingDirLockMetadata)) (PullLock, error)
	CancelAutoplanWaiter(repoFullName string, pullNum int)
	// TryLock tries to acquire a lock for this repo, pull, workspace, and path.
	// It returns a function that should be used to unlock the workspace and
	// an error if the workspace is already locked. The error is expected to
	// be printed to the pull request.
	TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error)
	// HasCommandLock reports whether this pull request has an active lock for cmdName.
	HasCommandLock(repoFullName string, pullNum int, cmdName command.Name) bool
	// UnlockByPull unlocks all workspaces for a specific pull request
	UnlockByPull(repoFullName string, pullNum int)
}

// DefaultWorkingDirLocker implements WorkingDirLocker.
type DefaultWorkingDirLocker struct {
	// mutex prevents against multiple threads calling functions on this struct
	// concurrently. It's only used for entry/exit to each function.
	mutex sync.Mutex
	// locks is a map of workspaces and their owning command metadata.
	locks map[string]workingDirLock
	// waiters holds the queued or refreshing LockPullAfterStalePlan request per pull
	// request, keyed by pullLockPrefix.
	waiters map[string]*pullLockWaiter
	// released is signalled whenever locks or waiters change so that
	// LockPullAfterStalePlan callers can re-check. It uses mutex.
	released *sync.Cond
}

// NewDefaultWorkingDirLocker is a constructor.
func NewDefaultWorkingDirLocker() *DefaultWorkingDirLocker {
	d := &DefaultWorkingDirLocker{
		locks:   make(map[string]workingDirLock),
		waiters: make(map[string]*pullLockWaiter),
	}
	d.released = sync.NewCond(&d.mutex)
	return d
}

func (d *DefaultWorkingDirLocker) TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	workspaceKey := d.workspaceKey(repoFullName, pullNum, workspace, path, projectName)
	if currentLock, exists := d.locks[workspaceKey]; exists {
		return func() {}, &workingDirLockError{
			message: fmt.Sprintf("cannot run %q: the %s workspace at path %s is currently locked for this pull request by %q%s.\n"+
				"Wait until the previous command is complete and try again", cmdName, workspace, path, currentLock.Command, formatCommitSuffix(currentLock)),
			metadata: currentLock.WorkingDirLockMetadata,
		}
	}
	owner := &pullLockWaiter{}
	d.locks[workspaceKey] = workingDirLock{owner: owner, Command: cmdName, WorkingDirLockMetadata: metadata}
	return func() {
		d.unlock(workspaceKey, d.pullLockPrefix(repoFullName, pullNum), owner)
	}, nil
}

func (d *DefaultWorkingDirLocker) TryLockPull(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if currentLock, blocked := d.blockingPullLock(repoFullName, pullNum, metadata); blocked {
		return func() {}, d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
	}
	return d.lockPull(repoFullName, pullNum, cmdName, metadata, &pullLockWaiter{}, false).Unlock, nil
}

func (d *DefaultWorkingDirLocker) TryLockPullShared(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	key := d.pullKey(repoFullName, pullNum)
	if current, exists := d.locks[key]; exists && current.shared > 0 && current.Command == cmdName && current.HeadCommit == metadata.HeadCommit {
		current.shared++
		d.locks[key] = current
		return d.unlockShared(key, current.owner), nil
	}
	if currentLock, blocked := d.blockingPullLock(repoFullName, pullNum, metadata); blocked {
		return func() {}, d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
	}
	owner := &pullLockWaiter{}
	d.locks[key] = workingDirLock{owner: owner, shared: 1, Command: cmdName, WorkingDirLockMetadata: metadata}
	return d.unlockShared(key, owner), nil
}

// unlockShared returns a function that releases one hold on a shared
// pull-level lock. Calling it more than once has no further effect.
func (d *DefaultWorkingDirLocker) unlockShared(key string, owner *pullLockWaiter) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			d.mutex.Lock()
			defer d.mutex.Unlock()
			lock, ok := d.locks[key]
			if !ok || lock.owner != owner {
				return
			}
			lock.shared--
			if lock.shared > 0 {
				d.locks[key] = lock
				return
			}
			delete(d.locks, key)
			d.broadcast()
		})
	}
}

func (d *DefaultWorkingDirLocker) LockPullAfterStalePlan(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata, onWait func(holder WorkingDirLockMetadata)) (PullLock, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	prefix := d.pullLockPrefix(repoFullName, pullNum)
	var waiter *pullLockWaiter
	acquired := false
	defer func() {
		if !acquired && waiter != nil && d.waiters[prefix] == waiter {
			delete(d.waiters, prefix)
		}
	}()

	for {
		if waiter != nil && d.waiters[prefix] != waiter {
			return PullLock{}, waiter.reason
		}
		currentLock, blocked := d.blockingPullLock(repoFullName, pullNum, metadata)
		if !blocked {
			if previous := d.waiters[prefix]; previous != nil && previous != waiter {
				previous.reason = ErrPullLockSuperseded
				delete(d.waiters, prefix)
				d.broadcast()
			}
			if waiter == nil {
				waiter = &pullLockWaiter{}
				d.waiters[prefix] = waiter
			}
			acquired = true
			return d.lockPull(repoFullName, pullNum, cmdName, metadata, waiter, true), nil
		}
		if !isStalePlanLock(currentLock, metadata) {
			if currentLock.Command == command.Plan && currentLock.HeadCommit == metadata.HeadCommit && (waiter != nil || currentLock.autoplan) {
				// A plan for our commit is already running: either another
				// plan got the lock while we waited, or an autoplan already
				// adopted this commit as the live head (a late or redelivered
				// webhook). Nothing is left for this autoplan to do.
				return PullLock{}, ErrPullLockSuperseded
			}
			return PullLock{}, d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
		}
		if waiter == nil {
			if previous := d.waiters[prefix]; previous != nil {
				previous.reason = ErrPullLockSuperseded
			}
			waiter = &pullLockWaiter{}
			d.waiters[prefix] = waiter
			d.released.Broadcast()
			if onWait != nil {
				onWait(currentLock.WorkingDirLockMetadata)
			}
		}
		d.released.Wait()
	}
}

func (d *DefaultWorkingDirLocker) CancelAutoplanWaiter(repoFullName string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	if waiter := d.waiters[prefix]; waiter != nil {
		waiter.reason = ErrPullLockCancelled
		delete(d.waiters, prefix)
		d.broadcast()
	}
}

// isStalePlanLock reports whether lock belongs to a plan for a different
// commit than the one described by metadata.
func isStalePlanLock(lock workingDirLock, metadata WorkingDirLockMetadata) bool {
	return lock.Command == command.Plan &&
		lock.HeadCommit != "" &&
		metadata.HeadCommit != "" &&
		lock.HeadCommit != metadata.HeadCommit
}

// blockingPullLock returns the lock that prevents taking the pull-level lock
// for metadata, if any. d.mutex must be held.
func (d *DefaultWorkingDirLocker) blockingPullLock(repoFullName string, pullNum int, metadata WorkingDirLockMetadata) (workingDirLock, bool) {
	if currentLock, exists := d.locks[d.pullKey(repoFullName, pullNum)]; exists {
		return currentLock, true
	}
	currentLock, exists := d.findCommandLock(repoFullName, pullNum, command.Plan, metadata.HeadCommit)
	if !exists && metadata.HeadCommit != "" {
		currentLock, exists = d.findCommandLock(repoFullName, pullNum, command.Plan, "")
	}
	return currentLock, exists
}

// lockPull records the pull-level lock. d.mutex must be held.
func (d *DefaultWorkingDirLocker) lockPull(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata, owner *pullLockWaiter, autoplan bool) PullLock {
	key := d.pullKey(repoFullName, pullNum)
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	d.locks[key] = workingDirLock{owner: owner, autoplan: autoplan, Command: cmdName, WorkingDirLockMetadata: metadata}
	return PullLock{
		Start: func() bool {
			d.mutex.Lock()
			defer d.mutex.Unlock()
			lock, ok := d.locks[key]
			if !ok || lock.owner != owner || owner.reason != nil || d.waiters[prefix] != owner {
				return false
			}
			delete(d.waiters, prefix)
			return true
		},
		Unlock: func() { d.unlock(key, prefix, owner) },
		UpdateMetadata: func(metadata WorkingDirLockMetadata) bool {
			d.mutex.Lock()
			defer d.mutex.Unlock()
			lock, ok := d.locks[key]
			if !ok || lock.owner != owner || owner.reason != nil {
				return false
			}
			lock.WorkingDirLockMetadata = metadata
			d.locks[key] = lock
			d.broadcast()
			return true
		},
	}
}

func (d *DefaultWorkingDirLocker) HasCommandLock(repoFullName string, pullNum int, cmdName command.Name) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, exists := d.findCommandLock(repoFullName, pullNum, cmdName, "")
	return exists
}

func (d *DefaultWorkingDirLocker) findCommandLock(repoFullName string, pullNum int, cmdName command.Name, headCommit string) (workingDirLock, bool) {
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	var match workingDirLock
	var matchKey string
	for key, currentLock := range d.locks {
		if strings.HasPrefix(key, prefix) && currentLock.Command == cmdName && (headCommit == "" || currentLock.HeadCommit == headCommit) && (matchKey == "" || key < matchKey) {
			match = currentLock
			matchKey = key
		}
	}
	return match, matchKey != ""
}

func (d *DefaultWorkingDirLocker) pullLockError(repoFullName string, pullNum int, cmdName command.Name, currentLock workingDirLock) error {
	metadata := currentLock.WorkingDirLockMetadata
	metadata.JobURL = ""
	jobURLs := make(map[string]struct{})
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	pullKey := d.pullKey(repoFullName, pullNum)
	for key, lock := range d.locks {
		if key != pullKey && strings.HasPrefix(key, prefix) && lock.Command == currentLock.Command && lock.HeadCommit == currentLock.HeadCommit && lock.JobURL != "" {
			jobURLs[lock.JobURL] = struct{}{}
		}
	}

	multipleJobs := len(jobURLs) > 1
	if len(jobURLs) == 1 {
		for jobURL := range jobURLs {
			metadata.JobURL = jobURL
		}
	}
	message := fmt.Sprintf("cannot run %q: pull request %d is currently locked by %q%s.\n"+
		"Wait until the previous command is complete and try again", cmdName, pullNum, currentLock.Command, formatCommitSuffix(currentLock))
	return &workingDirLockError{message: message, metadata: metadata, multipleJobs: multipleJobs}
}

func formatCommitSuffix(lock workingDirLock) string {
	if lock.HeadCommit == "" {
		return ""
	}
	return " for commit " + shortSHA(lock.HeadCommit)
}

func WorkingDirLockMetadataForPull(pull models.PullRequest) WorkingDirLockMetadata {
	metadata := WorkingDirLockMetadata{HeadCommit: pull.HeadCommit}
	if pull.HeadCommit == "" || pull.URL == "" {
		return metadata
	}

	pullURL := strings.TrimRight(pull.URL, "/")
	switch pull.BaseRepo.VCSHost.Type {
	case models.Github, models.Gitea:
		metadata.CommitURL = pullURL + "/commits/" + pull.HeadCommit
	case models.Gitlab:
		metadata.CommitURL = pullURL + "/diffs?commit_id=" + pull.HeadCommit
	}
	return metadata
}

func WorkingDirLockMetadataForProject(ctx command.ProjectContext, jobURL string) WorkingDirLockMetadata {
	metadata := WorkingDirLockMetadataForPull(ctx.Pull)
	metadata.JobURL = jobURL
	return metadata
}

// UnlockByPull unlocks all workspaces for a specific pull request
func (d *DefaultWorkingDirLocker) UnlockByPull(repoFullName string, pullNum int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Find and remove all locks for this pull request
	prefix := d.pullLockPrefix(repoFullName, pullNum)
	for key := range d.locks {
		if strings.HasPrefix(key, prefix) {
			delete(d.locks, key)
		}
	}
	if waiter := d.waiters[prefix]; waiter != nil {
		waiter.reason = ErrPullLockCancelled
		delete(d.waiters, prefix)
	}
	d.broadcast()
}

// broadcast wakes LockPullAfterStalePlan callers. d.mutex must be held.
func (d *DefaultWorkingDirLocker) broadcast() {
	if d.released != nil {
		d.released.Broadcast()
	}
}

func (d *DefaultWorkingDirLocker) unlock(key, prefix string, owner *pullLockWaiter) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.waiters[prefix] == owner {
		delete(d.waiters, prefix)
	}
	if lock, ok := d.locks[key]; ok && lock.owner == owner {
		delete(d.locks, key)
		d.broadcast()
	}
}

func (d *DefaultWorkingDirLocker) workspaceKey(repo string, pull int, workspace string, path string, projectName string) string {
	return strings.TrimRight(fmt.Sprintf("%s%s/%s/%s", d.pullLockPrefix(repo, pull), workspace, path, projectName), "/")
}

func (d *DefaultWorkingDirLocker) pullKey(repo string, pull int) string {
	return d.pullLockPrefix(repo, pull) + ".pull"
}

func (d *DefaultWorkingDirLocker) pullLockPrefix(repo string, pull int) string {
	return fmt.Sprintf("%d:%s/%d/", len(repo), repo, pull)
}
