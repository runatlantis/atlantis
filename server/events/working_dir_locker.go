// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
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
	Command command.Name
	WorkingDirLockMetadata
	// owner identifies the holder, so only its own release deletes the lock.
	owner uint64
}

type WorkingDirLocker interface {
	// TryLockPull tries to acquire an exclusive pull-level lock for a command. It
	// is used to coordinate commands that can make pull-wide plan state
	// inconsistent. It conflicts with every other pull-level lock and with
	// in-flight plans.
	TryLockPull(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error)
	// TryLockPullShared tries to acquire a pull-level lock that other shared
	// holders can hold at the same time. Targeted apply uses it so applies of
	// different projects on one pull request can run in parallel. It conflicts
	// with TryLockPull holders and in-flight plans.
	TryLockPullShared(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error)
	// TryLockProjectApply claims a project for an apply command until the
	// command has recorded its results, so applies sharing the pull lock cannot
	// apply the same project. Callers must already hold a pull-level lock. It
	// does not conflict with TryLock.
	TryLockProjectApply(repoFullName string, pullNum int, workspace string, path string, projectName string, metadata WorkingDirLockMetadata) (func(), error)
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
	// ownerSeq gives each lock holder its own owner, so a release after
	// UnlockByPull cannot drop a newer holder's lock.
	ownerSeq uint64
}

// NewDefaultWorkingDirLocker is a constructor.
func NewDefaultWorkingDirLocker() *DefaultWorkingDirLocker {
	return &DefaultWorkingDirLocker{locks: make(map[string]workingDirLock)}
}

func (d *DefaultWorkingDirLocker) TryLock(repoFullName string, pullNum int, workspace string, path string, projectName string, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	workspaceKey := d.workspaceKey(repoFullName, pullNum, workspace, path, projectName)
	if currentLock, exists := d.locks[workspaceKey]; exists {
		return func() {}, workspaceLockError(cmdName, workspace, path, currentLock)
	}
	return d.lockOwned(workspaceKey, d.nextOwner(), cmdName, metadata), nil
}

func (d *DefaultWorkingDirLocker) TryLockPull(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.pullLockConflict(repoFullName, pullNum, cmdName, metadata); err != nil {
		return func() {}, err
	}
	if currentLock, exists := d.findSharedPullLock(repoFullName, pullNum); exists {
		return func() {}, d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
	}
	return d.lockOwned(d.pullKey(repoFullName, pullNum), d.nextOwner(), cmdName, metadata), nil
}

func (d *DefaultWorkingDirLocker) TryLockPullShared(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.pullLockConflict(repoFullName, pullNum, cmdName, metadata); err != nil {
		return func() {}, err
	}
	owner := d.nextOwner()
	sharedKey := fmt.Sprintf("%s%d", d.sharedPullKeyPrefix(repoFullName, pullNum), owner)
	if currentLock, exists := d.locks[sharedKey]; exists {
		// Unreachable while reserved keys stay outside the workspace key space.
		return func() {}, d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
	}
	return d.lockOwned(sharedKey, owner, cmdName, metadata), nil
}

func (d *DefaultWorkingDirLocker) TryLockProjectApply(repoFullName string, pullNum int, workspace string, path string, projectName string, metadata WorkingDirLockMetadata) (func(), error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	claimKey := d.applyClaimKey(repoFullName, pullNum, workspace, path, projectName)
	if currentLock, exists := d.locks[claimKey]; exists {
		// The claim holder has no job URL. While it applies the project, the
		// project runner's lock carries the job URL.
		if runningLock, running := d.locks[d.workspaceKey(repoFullName, pullNum, workspace, path, projectName)]; running && runningLock.JobURL != "" {
			currentLock.JobURL = runningLock.JobURL
		}
		return func() {}, workspaceLockError(command.Apply, workspace, path, currentLock)
	}
	return d.lockOwned(claimKey, d.nextOwner(), command.Apply, metadata), nil
}

// nextOwner returns a new lock owner. The caller must hold d.mutex.
func (d *DefaultWorkingDirLocker) nextOwner() uint64 {
	d.ownerSeq++
	return d.ownerSeq
}

// lockOwned stores a lock for owner and returns a release that deletes the
// lock only while owner still holds it. The caller must hold d.mutex.
func (d *DefaultWorkingDirLocker) lockOwned(key string, owner uint64, cmdName command.Name, metadata WorkingDirLockMetadata) func() {
	d.locks[key] = workingDirLock{Command: cmdName, WorkingDirLockMetadata: metadata, owner: owner}
	return func() {
		d.unlockOwned(key, owner)
	}
}

// pullLockConflict returns an error when an exclusive pull lock or an
// in-flight plan blocks a new pull-level lock.
func (d *DefaultWorkingDirLocker) pullLockConflict(repoFullName string, pullNum int, cmdName command.Name, metadata WorkingDirLockMetadata) error {
	if currentLock, exists := d.locks[d.pullKey(repoFullName, pullNum)]; exists {
		return d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
	}
	currentLock, exists := d.findCommandLock(repoFullName, pullNum, command.Plan, metadata.HeadCommit)
	if !exists && metadata.HeadCommit != "" {
		currentLock, exists = d.findCommandLock(repoFullName, pullNum, command.Plan, "")
	}
	if exists {
		return d.pullLockError(repoFullName, pullNum, cmdName, currentLock)
	}
	return nil
}

func (d *DefaultWorkingDirLocker) findSharedPullLock(repoFullName string, pullNum int) (workingDirLock, bool) {
	prefix := d.sharedPullKeyPrefix(repoFullName, pullNum)
	var match workingDirLock
	var matchKey string
	for key, currentLock := range d.locks {
		if strings.HasPrefix(key, prefix) && (matchKey == "" || key < matchKey) {
			match = currentLock
			matchKey = key
		}
	}
	return match, matchKey != ""
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

func workspaceLockError(cmdName command.Name, workspace string, path string, currentLock workingDirLock) error {
	return &workingDirLockError{
		message: fmt.Sprintf("cannot run %q: the %s workspace at path %s is currently locked for this pull request by %q%s.\n"+
			"Wait until the previous command is complete and try again", cmdName, workspace, path, currentLock.Command, formatCommitSuffix(currentLock)),
		metadata: currentLock.WorkingDirLockMetadata,
	}
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
}

func (d *DefaultWorkingDirLocker) unlockOwned(key string, owner uint64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if currentLock, exists := d.locks[key]; exists && currentLock.owner == owner {
		delete(d.locks, key)
	}
}

func (d *DefaultWorkingDirLocker) workspaceKey(repo string, pull int, workspace string, path string, projectName string) string {
	return strings.TrimRight(fmt.Sprintf("%s%s/%s/%s", d.pullLockPrefix(repo, pull), workspace, path, projectName), "/")
}

// Pull-level locks and apply claims use reserved keys that start with '$'.
// valid.ValidateWorkspaceName rejects '$', and a workspace key starts with the
// workspace name, so no workspace key can collide with a reserved key.
const reservedKeyMarker = "$"

func (d *DefaultWorkingDirLocker) pullKey(repo string, pull int) string {
	return d.pullLockPrefix(repo, pull) + reservedKeyMarker + "pull"
}

func (d *DefaultWorkingDirLocker) sharedPullKeyPrefix(repo string, pull int) string {
	return d.pullLockPrefix(repo, pull) + reservedKeyMarker + "shared-pull/"
}

func (d *DefaultWorkingDirLocker) applyClaimKey(repo string, pull int, workspace string, path string, projectName string) string {
	return fmt.Sprintf("%s%sapply-claim/%s/%s/%s", d.pullLockPrefix(repo, pull), reservedKeyMarker, workspace, path, projectName)
}

func (d *DefaultWorkingDirLocker) pullLockPrefix(repo string, pull int) string {
	return fmt.Sprintf("%d:%s/%d/", len(repo), repo, pull)
}
