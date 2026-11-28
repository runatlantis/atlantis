// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

//go:generate go tool pegomock generate --package mocks -o mocks/mock_cancellation_tracker.go CancellationTracker
import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/runatlantis/atlantis/server/events/models"
)

type CancellationTracker interface {
	Start(pull models.PullRequest) (isCancelled func() bool, finish func())
	Cancel(pull models.PullRequest)
}

type DefaultCancellationTracker struct {
	mutex sync.Mutex
	runs  map[string]map[*atomic.Bool]struct{}
}

func NewCancellationTracker() *DefaultCancellationTracker {
	return &DefaultCancellationTracker{runs: make(map[string]map[*atomic.Bool]struct{})}
}

// Start tracks a command until it finishes. A previous cancel does not affect a new command.
func (p *DefaultCancellationTracker) Start(pull models.PullRequest) (func() bool, func()) {
	key := pullKey(pull)
	cancelled := new(atomic.Bool)
	p.mutex.Lock()
	if p.runs[key] == nil {
		p.runs[key] = make(map[*atomic.Bool]struct{})
	}
	p.runs[key][cancelled] = struct{}{}
	p.mutex.Unlock()

	return cancelled.Load, func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		delete(p.runs[key], cancelled)
		if len(p.runs[key]) == 0 {
			delete(p.runs, key)
		}
	}
}

// Cancel prevents commands already running on this pull from starting more groups.
func (p *DefaultCancellationTracker) Cancel(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for run := range p.runs[pullKey(pull)] {
		run.Store(true)
	}
}

func trackProjectCmds(tracker CancellationTracker, pull models.PullRequest) (func() bool, func()) {
	if tracker == nil {
		return nil, func() {}
	}
	return tracker.Start(pull)
}

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
