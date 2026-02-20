// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package plugin defines the plugin interface and registry for Atlantis VCS
// provider extensions. Third-party providers implement the Plugin interface,
// register themselves (typically via an init() function), and Atlantis picks
// them up through the DefaultRegistry.
package plugin

import (
	"fmt"
	"sync"
)

// ConfigKey describes a single configuration requirement for a plugin.
type ConfigKey struct {
	// Flag is the CLI flag name (e.g., "gh-token").
	Flag string
	// EnvVar is the corresponding environment variable name
	// (e.g., "ATLANTIS_GH_TOKEN").
	EnvVar string
	// Desc is a short, human-readable description of this key.
	Desc string
	// Required indicates whether this key must be supplied.
	Required bool
}

// Plugin defines the interface that every VCS provider plugin must satisfy.
type Plugin interface {
	// Name returns the unique, lowercase identifier for this plugin
	// (e.g., "github"). It is used as the argument to "atlantis plugin add".
	Name() string
	// Description returns a one-line, human-readable description.
	Description() string
	// Version returns the plugin version string (e.g., "1.0.0").
	Version() string
	// ConfigKeys returns the ordered list of configuration keys needed to
	// operate this provider with Atlantis.
	ConfigKeys() []ConfigKey
}

// Registry holds registered VCS provider plugins.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewRegistry creates a new, empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds p to the registry.
// It returns an error if a plugin with the same name is already registered.
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin %q is already registered", p.Name())
	}
	r.plugins[p.Name()] = p
	return nil
}

// Get returns the plugin registered under name, or (nil, false) if not found.
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

// List returns a snapshot of all registered plugins in unspecified order.
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result
}

// DefaultRegistry is the global plugin registry used by the CLI and server.
var DefaultRegistry = NewRegistry()
