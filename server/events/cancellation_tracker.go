// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

//go:generate go tool pegomock generate --package mocks -o mocks/mock_cancellation_tracker.go CancellationTracker
import (
	"context"
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
)

type CancellationTracker interface {
	Cancel(pull models.PullRequest)
	IsCancelled(pull models.PullRequest) bool
	Clear(pull models.PullRequest)
}

type DefaultCancellationTracker struct {
	mutex          sync.RWMutex
	cancelledPulls map[string]struct{}
	waiters        map[string]map[*cancellationWaiter]context.CancelFunc
}

func NewCancellationTracker() *DefaultCancellationTracker {
	return &DefaultCancellationTracker{
		cancelledPulls: make(map[string]struct{}),
		waiters:        make(map[string]map[*cancellationWaiter]context.CancelFunc),
	}
}

// Cancel marks an entire pull request as cancelled, preventing any future operations
func (p *DefaultCancellationTracker) Cancel(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pullKeyStr := pullKey(pull)
	p.cancelledPulls[pullKeyStr] = struct{}{}
	for _, cancel := range p.waiters[pullKeyStr] {
		cancel()
	}
}

// IsCancelled checks if the entire pull request has been cancelled
func (p *DefaultCancellationTracker) IsCancelled(pull models.PullRequest) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	_, exists := p.cancelledPulls[pullKey(pull)]
	return exists
}

// Clear removes cancellation for a pull request (called when a PR is closed)
func (p *DefaultCancellationTracker) Clear(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.cancelledPulls, pullKey(pull))
}

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}

// cancellationWaiter has nonzero size so simultaneously registered pointers
// have distinct identities.
type cancellationWaiter struct{ registered bool }

// CommandContext lets bounded publication waits observe `atlantis cancel`
// without polling. The returned cleanup unregisters the completed command.
func (p *DefaultCancellationTracker) CommandContext(parent context.Context, pull models.PullRequest) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	key := pullKey(pull)
	waiter := &cancellationWaiter{registered: true}
	p.mutex.Lock()
	if p.waiters == nil {
		p.waiters = make(map[string]map[*cancellationWaiter]context.CancelFunc)
	}
	if p.waiters[key] == nil {
		p.waiters[key] = make(map[*cancellationWaiter]context.CancelFunc)
	}
	p.waiters[key][waiter] = cancel
	if _, cancelled := p.cancelledPulls[key]; cancelled {
		cancel()
	}
	p.mutex.Unlock()
	return ctx, func() {
		cancel()
		p.mutex.Lock()
		delete(p.waiters[key], waiter)
		if len(p.waiters[key]) == 0 {
			delete(p.waiters, key)
		}
		p.mutex.Unlock()
	}
}
