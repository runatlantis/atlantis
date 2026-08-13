// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"sync"
	"sync/atomic"

	"github.com/runatlantis/atlantis/server/core/ownership"
)

// LocalClaimGuard coordinates replica-local state across ownership generations.
type LocalClaimGuard struct {
	mu      sync.Mutex
	entries map[string]*localClaimEntry
}

type localClaimEntry struct {
	mu         sync.Mutex
	changed    chan struct{}
	generation *localClaimGeneration
	removable  atomic.Bool

	// refs is protected by LocalClaimGuard.mu. Acquirers retain their reference
	// until the returned release function runs.
	refs int
}

type localClaimGeneration struct {
	claimID         string
	active          int
	preparing       bool
	localStateReset bool
	forgetRequested bool
}

// NewLocalClaimGuard creates a local ownership-generation guard.
func NewLocalClaimGuard() *LocalClaimGuard {
	return &LocalClaimGuard{entries: make(map[string]*localClaimEntry)}
}

// Acquire admits one command into an ownership generation. A new generation
// waits for the previous generation's commands, then resets local state once.
// The returned release function is idempotent and must span the command's full
// lifetime.
func (g *LocalClaimGuard) Acquire(
	key ownership.Key,
	claimID string,
	admit func() error,
	reset func() error,
) (release func(), localStateReset bool, err error) {
	noopRelease := func() {}
	if claimID == "" {
		return noopRelease, false, nil
	}
	canonical, err := key.Canonical()
	if err != nil {
		return noopRelease, false, err
	}

	entry := g.retain(canonical)
	for {
		entry.mu.Lock()
		current := entry.generation
		if current != nil && (current.preparing || (current.claimID != claimID && current.active > 0)) {
			changed := entry.changed
			entry.mu.Unlock()
			<-changed
			continue
		}

		if admit != nil {
			if err := admit(); err != nil {
				entry.mu.Unlock()
				g.releaseReference(canonical, entry)
				return noopRelease, false, err
			}
		}

		if current != nil && current.claimID == claimID {
			current.active++
			entry.removable.Store(false)
			entry.mu.Unlock()
			return g.commandRelease(canonical, entry, current), current.localStateReset, nil
		}

		generation := &localClaimGeneration{
			claimID:         claimID,
			active:          1,
			preparing:       true,
			localStateReset: true,
		}
		entry.generation = generation
		entry.removable.Store(false)
		entry.mu.Unlock()

		var resetErr error
		if reset != nil {
			resetErr = reset()
		}

		entry.mu.Lock()
		generation.preparing = false
		if resetErr != nil && entry.generation == generation {
			entry.generation = nil
			entry.removable.Store(true)
		}
		entry.notifyLocked()
		entry.mu.Unlock()
		if resetErr != nil {
			g.releaseReference(canonical, entry)
			return noopRelease, false, resetErr
		}
		return g.commandRelease(canonical, entry, generation), true, nil
	}
}

// Prepare is retained for direct callers that only need reset-once behavior.
func (g *LocalClaimGuard) Prepare(key ownership.Key, claimID string, reset func() error) error {
	release, _, err := g.Acquire(key, claimID, nil, reset)
	if err == nil {
		release()
	}
	return err
}

// Forget removes an exact generation after all of its active commands finish.
func (g *LocalClaimGuard) Forget(key ownership.Key, claimID string) error {
	if claimID == "" {
		return nil
	}
	canonical, err := key.Canonical()
	if err != nil {
		return err
	}

	entry := g.retain(canonical)
	entry.mu.Lock()
	current := entry.generation
	if current == nil {
		entry.removable.Store(true)
	} else if current.claimID == claimID && !current.preparing {
		current.forgetRequested = true
		if current.active == 0 {
			entry.generation = nil
			entry.removable.Store(true)
			entry.notifyLocked()
		}
	}
	entry.mu.Unlock()
	g.releaseReference(canonical, entry)
	return nil
}

func (g *LocalClaimGuard) retain(canonical string) *localClaimEntry {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry := g.entries[canonical]
	if entry == nil {
		entry = &localClaimEntry{changed: make(chan struct{})}
		entry.removable.Store(true)
		g.entries[canonical] = entry
	}
	entry.refs++
	return entry
}

func (g *LocalClaimGuard) releaseReference(canonical string, entry *localClaimEntry) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry.refs--
	if entry.refs == 0 && entry.removable.Load() && g.entries[canonical] == entry {
		delete(g.entries, canonical)
	}
}

func (g *LocalClaimGuard) commandRelease(canonical string, entry *localClaimEntry, generation *localClaimGeneration) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			entry.mu.Lock()
			if entry.generation == generation && generation.active > 0 {
				generation.active--
				if generation.active == 0 {
					if generation.forgetRequested {
						entry.generation = nil
						entry.removable.Store(true)
					}
					entry.notifyLocked()
				}
			}
			entry.mu.Unlock()
			g.releaseReference(canonical, entry)
		})
	}
}

func (e *localClaimEntry) notifyLocked() {
	close(e.changed)
	e.changed = make(chan struct{})
}
