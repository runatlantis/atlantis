// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"sync"

	"github.com/runatlantis/atlantis/server/core/ownership"
)

// LocalClaimGuard resets replica-local pull state once for each ownership claim.
type LocalClaimGuard struct {
	mu       sync.Mutex
	prepared map[string]*localClaimState
}

type localClaimState struct {
	claimID  string
	done     chan struct{}
	complete bool
	err      error
}

// NewLocalClaimGuard creates a local ownership-generation guard.
func NewLocalClaimGuard() *LocalClaimGuard {
	return &LocalClaimGuard{prepared: make(map[string]*localClaimState)}
}

// Prepare runs reset once for the exact pull request ownership claim.
func (g *LocalClaimGuard) Prepare(key ownership.Key, claimID string, reset func() error) error {
	if claimID == "" {
		return nil
	}
	canonical, err := key.Canonical()
	if err != nil {
		return err
	}

	for {
		g.mu.Lock()
		current := g.prepared[canonical]
		if current == nil || (current.claimID != claimID && current.complete) {
			state := &localClaimState{claimID: claimID, done: make(chan struct{})}
			g.prepared[canonical] = state
			g.mu.Unlock()

			err := reset()
			g.mu.Lock()
			state.err = err
			state.complete = true
			if err != nil && g.prepared[canonical] == state {
				delete(g.prepared, canonical)
			}
			close(state.done)
			g.mu.Unlock()
			return err
		}

		if current.claimID == claimID && current.complete {
			err := current.err
			g.mu.Unlock()
			return err
		}

		done := current.done
		sameClaim := current.claimID == claimID
		g.mu.Unlock()
		<-done
		if sameClaim {
			return current.err
		}
	}
}

// Forget removes a completed preparation only when the claim still matches.
func (g *LocalClaimGuard) Forget(key ownership.Key, claimID string) error {
	if claimID == "" {
		return nil
	}
	canonical, err := key.Canonical()
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if current := g.prepared[canonical]; current != nil && current.claimID == claimID {
		delete(g.prepared, canonical)
	}
	return nil
}
