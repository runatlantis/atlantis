// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

//go:generate go tool pegomock generate --package mocks -o mocks/mock_cancellation_tracker.go CancellationTracker
import (
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
)

// CancellationTracker records `atlantis cancel` per pull request. Several
// commands can run on one pull request at once, so a cancel is scoped to the
// commands that started before it: each command takes a Mark when it starts
// and passes it to IsCancelled.
type CancellationTracker interface {
	// Cancel cancels every command on the pull request that took its mark
	// before this call.
	Cancel(pull models.PullRequest)
	// Mark returns a value identifying the cancels issued so far.
	Mark() uint64
	// IsCancelled reports whether the pull request was cancelled after since
	// was returned by Mark.
	IsCancelled(pull models.PullRequest, since uint64) bool
	// Clear forgets cancels for the pull request (called when a PR is closed).
	Clear(pull models.PullRequest)
}

type DefaultCancellationTracker struct {
	mutex sync.RWMutex
	// seq increases with every Cancel, across all pull requests.
	seq uint64
	// cancelledAt holds the seq of each pull request's latest Cancel.
	cancelledAt map[string]uint64
}

func NewCancellationTracker() *DefaultCancellationTracker {
	return &DefaultCancellationTracker{
		cancelledAt: make(map[string]uint64),
	}
}

// Cancel marks the pull request's running commands as cancelled, preventing
// their future operations
func (p *DefaultCancellationTracker) Cancel(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.seq++
	p.cancelledAt[pullKey(pull)] = p.seq
}

// Mark returns the current cancel sequence
func (p *DefaultCancellationTracker) Mark() uint64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.seq
}

// IsCancelled checks if the pull request has been cancelled since the mark was taken
func (p *DefaultCancellationTracker) IsCancelled(pull models.PullRequest, since uint64) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.cancelledAt[pullKey(pull)] > since
}

// Clear removes cancellation for a pull request (called when a PR is closed)
func (p *DefaultCancellationTracker) Clear(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.cancelledAt, pullKey(pull))
}

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
