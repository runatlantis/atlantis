// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/plugin"
	. "github.com/runatlantis/atlantis/testing"
)

// fakePlugin is a minimal Plugin implementation used in tests.
type fakePlugin struct {
	name    string
	version string
}

func (f *fakePlugin) Name() string                   { return f.name }
func (f *fakePlugin) Description() string            { return "fake plugin for testing" }
func (f *fakePlugin) Version() string                { return f.version }
func (f *fakePlugin) ConfigKeys() []plugin.ConfigKey { return nil }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := plugin.NewRegistry()
	p := &fakePlugin{name: "testprovider", version: "0.1.0"}

	err := r.Register(p)
	Ok(t, err)

	got, ok := r.Get("testprovider")
	Assert(t, ok, "expected plugin to be found after registration")
	Equals(t, p, got)
}

func TestRegistry_Register_DuplicateReturnsError(t *testing.T) {
	r := plugin.NewRegistry()
	p := &fakePlugin{name: "dup", version: "1.0.0"}

	err := r.Register(p)
	Ok(t, err)

	err = r.Register(p)
	Assert(t, err != nil, "expected error when registering duplicate plugin")
	Assert(t, strings.Contains(err.Error(), "already registered"), "expected 'already registered' in error message, got: %q", err.Error())
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := plugin.NewRegistry()
	_, ok := r.Get("nonexistent")
	Assert(t, !ok, "expected not-found for unregistered plugin")
}

func TestRegistry_List(t *testing.T) {
	r := plugin.NewRegistry()

	Equals(t, 0, len(r.List()))

	_ = r.Register(&fakePlugin{name: "alpha", version: "1.0.0"})
	_ = r.Register(&fakePlugin{name: "beta", version: "2.0.0"})

	list := r.List()
	Equals(t, 2, len(list))

	names := make(map[string]bool)
	for _, p := range list {
		names[p.Name()] = true
	}
	Assert(t, names["alpha"], "expected 'alpha' in list")
	Assert(t, names["beta"], "expected 'beta' in list")
}

func TestRegistry_List_Empty(t *testing.T) {
	r := plugin.NewRegistry()
	Equals(t, 0, len(r.List()))
}

func TestDefaultRegistry_ContainsGitHub(t *testing.T) {
	// DefaultRegistry must have GitHub pre-populated without any blank import.
	// This test imports only the plugin package, proving no side-effect import
	// is required by callers.
	p, ok := plugin.DefaultRegistry.Get("github")
	Assert(t, ok, "expected github to be present in DefaultRegistry by default")
	Equals(t, "github", p.Name())
	Assert(t, len(p.ConfigKeys()) > 0, "expected github plugin to have config keys")
}
