// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common

import "sync"

// IdentityCache shares lookups by key and retains successful, non-empty identities.
// Its zero value is ready to use.
type IdentityCache struct {
	mu      sync.Mutex
	lookups map[string]*identityLookup
}

type identityLookup struct {
	done     chan struct{}
	identity string
	err      error
}

// Get returns the cached identity or joins a lookup in progress for key.
func (c *IdentityCache) Get(key string, lookup func() (string, error)) (string, error) {
	c.mu.Lock()
	if existing := c.lookups[key]; existing != nil {
		c.mu.Unlock()
		<-existing.done
		return existing.identity, existing.err
	}
	current := &identityLookup{done: make(chan struct{})}
	if c.lookups == nil {
		c.lookups = make(map[string]*identityLookup)
	}
	c.lookups[key] = current
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		if current.err != nil || current.identity == "" {
			delete(c.lookups, key)
		}
		close(current.done)
		c.mu.Unlock()
	}()
	current.identity, current.err = lookup()
	return current.identity, current.err
}
