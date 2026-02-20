// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package plugin defines the plugin interface and registry for Atlantis VCS
// provider extensions. Third-party providers implement the Plugin interface,
// register themselves (typically via an init() function), and Atlantis picks
// them up through the DefaultRegistry.
//
// DefaultRegistry is pre-populated with the built-in GitHub provider so that
// callers importing only this package always find GitHub available.
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
// It is pre-populated with the built-in GitHub plugin so that it is available
// without any additional import.
var DefaultRegistry = newDefaultRegistry()

// newDefaultRegistry returns a Registry pre-populated with built-in plugins.
// Built-in plugins are registered here rather than via init() in a sub-package
// to avoid requiring callers to use blank imports for side-effects.
func newDefaultRegistry() *Registry {
	r := NewRegistry()
	_ = r.Register(&githubPlugin{})
	return r
}

// githubPlugin is the built-in GitHub VCS provider pre-registered in
// DefaultRegistry. It mirrors the exported GitHubPlugin in the
// plugin/github sub-package; the sub-package's type is kept for users
// who need to reference it directly or instantiate it in tests.
//
// NOTE: The ConfigKeys list below is intentionally duplicated from
// plugin/github.GitHubPlugin to avoid an import cycle (plugin/github imports
// plugin). If you add or change a config key in plugin/github/github.go,
// you must update this list too.
type githubPlugin struct{}

func (g *githubPlugin) Name() string { return "github" }
func (g *githubPlugin) Description() string {
	return "GitHub VCS provider (github.com and GitHub Enterprise)"
}
func (g *githubPlugin) Version() string { return "1.0.0" }
func (g *githubPlugin) ConfigKeys() []ConfigKey {
	return []ConfigKey{
		{
			Flag:     "gh-user",
			EnvVar:   "ATLANTIS_GH_USER",
			Desc:     "GitHub username of the API user",
			Required: true,
		},
		{
			Flag:     "gh-token",
			EnvVar:   "ATLANTIS_GH_TOKEN",
			Desc:     "GitHub personal access token of the API user",
			Required: true,
		},
		{
			Flag:     "gh-webhook-secret",
			EnvVar:   "ATLANTIS_GH_WEBHOOK_SECRET",
			Desc:     "Secret used to validate GitHub webhooks",
			Required: true,
		},
		{
			Flag:     "gh-hostname",
			EnvVar:   "ATLANTIS_GH_HOSTNAME",
			Desc:     "Hostname of your GitHub Enterprise instance (default: github.com)",
			Required: false,
		},
		{
			Flag:     "gh-app-id",
			EnvVar:   "ATLANTIS_GH_APP_ID",
			Desc:     "GitHub App ID; use instead of gh-user/gh-token when using a GitHub App",
			Required: false,
		},
		{
			Flag:     "gh-app-key-file",
			EnvVar:   "ATLANTIS_GH_APP_KEY_FILE",
			Desc:     "Path to the GitHub App private key file",
			Required: false,
		},
		{
			Flag:     "gh-app-slug",
			EnvVar:   "ATLANTIS_GH_APP_SLUG",
			Desc:     "GitHub App slug (the URL-friendly name of the app)",
			Required: false,
		},
		{
			Flag:     "gh-org",
			EnvVar:   "ATLANTIS_GH_ORG",
			Desc:     "GitHub organization to restrict Atlantis access to",
			Required: false,
		},
		{
			Flag:     "gh-team-allowlist",
			EnvVar:   "ATLANTIS_GH_TEAM_ALLOWLIST",
			Desc:     "Comma-separated list of GitHub team slugs permitted to approve plans",
			Required: false,
		},
		{
			Flag:     "gh-allow-mergeable-bypass-apply",
			EnvVar:   "ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY",
			Desc:     "Allow mergeability check to be bypassed when applying",
			Required: false,
		},
	}
}
