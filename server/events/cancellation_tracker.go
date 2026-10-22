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
	// Cancel makes every command on the pull request that took its mark
	// before this call skip work it has not started. Commands check
	// IsCancelled between execution order groups and before each queued
	// parallel project, so work already running finishes.
	Cancel(pull models.PullRequest)
	// Mark returns a value identifying the cancels issued so far.
	Mark() uint64
	// IsCancelled reports whether the pull request was cancelled after since
	// was returned by Mark.
	IsCancelled(pull models.PullRequest, since uint64) bool
}

type DefaultCancellationTracker struct {
	mutex sync.RWMutex
	// seq increases with every Cancel, across all pull requests.
	seq uint64
	// cancelledAt holds the seq of each pull request's latest Cancel. It keeps
	// one entry per cancelled or closed pull request until restart: dropping
	// an entry would un-cancel commands still running on that pull request.
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

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
