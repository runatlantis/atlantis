package vcs

import (
	"fmt"
	"sync"
)

// DefaultVCSRegistry implements VCSRegistry interface
type DefaultVCSRegistry struct {
	mu      sync.RWMutex
	plugins map[string]VCSPlugin
}

// NewVCSRegistry creates a new VCS registry
func NewVCSRegistry() *DefaultVCSRegistry {
	return &DefaultVCSRegistry{
		plugins: make(map[string]VCSPlugin),
	}
}

// Register adds a VCS plugin to the registry
func (r *DefaultVCSRegistry) Register(name string, plugin VCSPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("VCS plugin name cannot be empty")
	}

	if plugin == nil {
		return fmt.Errorf("VCS plugin cannot be nil")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("VCS plugin '%s' is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Get retrieves a VCS plugin by name
func (r *DefaultVCSRegistry) Get(name string) (VCSPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("VCS plugin '%s' not found. Available plugins: %v", name, r.List())
	}

	return plugin, nil
}

// List returns all registered VCS plugin names
func (r *DefaultVCSRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// MustRegister registers a plugin or panics if registration fails
// This is useful for plugin initialization during application startup
func (r *DefaultVCSRegistry) MustRegister(name string, plugin VCSPlugin) {
	if err := r.Register(name, plugin); err != nil {
		panic(fmt.Sprintf("failed to register VCS plugin '%s': %v", name, err))
	}
}

// Unregister removes a VCS plugin from the registry
func (r *DefaultVCSRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("VCS plugin '%s' not found", name)
	}

	delete(r.plugins, name)
	return nil
}

// IsRegistered checks if a VCS plugin is registered
func (r *DefaultVCSRegistry) IsRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[name]
	return exists
} 