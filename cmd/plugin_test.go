// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/plugin"
	. "github.com/runatlantis/atlantis/testing"
)

// stubPlugin is a minimal plugin used in CLI tests.
type stubPlugin struct {
	name     string
	desc     string
	version  string
	required []plugin.ConfigKey
	optional []plugin.ConfigKey
}

func (s *stubPlugin) Name() string        { return s.name }
func (s *stubPlugin) Description() string { return s.desc }
func (s *stubPlugin) Version() string     { return s.version }
func (s *stubPlugin) ConfigKeys() []plugin.ConfigKey {
	return append(s.required, s.optional...)
}

func newTestRegistry(t *testing.T, plugins ...plugin.Plugin) *plugin.Registry {
	t.Helper()
	r := plugin.NewRegistry()
	for _, p := range plugins {
		err := r.Register(p)
		Ok(t, err)
	}
	return r
}

// captureOutput executes the cobra command and returns stdout as a string.
func captureOutput(t *testing.T, pc *cmd.PluginCmd, args ...string) (string, error) {
	t.Helper()
	root := pc.Init()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// ---------- plugin list ----------

func TestPluginList_Empty(t *testing.T) {
	pc := &cmd.PluginCmd{Registry: plugin.NewRegistry()}
	out, err := captureOutput(t, pc, "list")
	Ok(t, err)
	Assert(t, strings.Contains(out, "No plugins registered."), "expected empty-registry message, got: %q", out)
}

func TestPluginList_ShowsRegisteredPlugins(t *testing.T) {
	r := newTestRegistry(t,
		&stubPlugin{name: "alpha", desc: "Alpha provider", version: "1.0.0"},
		&stubPlugin{name: "beta", desc: "Beta provider", version: "2.1.0"},
	)
	pc := &cmd.PluginCmd{Registry: r}
	out, err := captureOutput(t, pc, "list")
	Ok(t, err)
	Assert(t, strings.Contains(out, "alpha"), "expected 'alpha' in output, got: %q", out)
	Assert(t, strings.Contains(out, "beta"), "expected 'beta' in output, got: %q", out)
	Assert(t, strings.Contains(out, "1.0.0"), "expected '1.0.0' in output, got: %q", out)
}

func TestPluginList_SortedAlphabetically(t *testing.T) {
	r := newTestRegistry(t,
		&stubPlugin{name: "zzz", desc: "Z provider", version: "1.0.0"},
		&stubPlugin{name: "aaa", desc: "A provider", version: "1.0.0"},
	)
	pc := &cmd.PluginCmd{Registry: r}
	out, err := captureOutput(t, pc, "list")
	Ok(t, err)
	idxAAA := strings.Index(out, "aaa")
	idxZZZ := strings.Index(out, "zzz")
	Assert(t, idxAAA < idxZZZ, "expected 'aaa' to appear before 'zzz' in output")
}

// ---------- plugin add ----------

func TestPluginAdd_UnknownPlugin(t *testing.T) {
	pc := &cmd.PluginCmd{Registry: plugin.NewRegistry()}
	_, err := captureOutput(t, pc, "add", "unknown")
	Assert(t, err != nil, "expected error for unknown plugin")
	Assert(t, strings.Contains(err.Error(), "unknown"), "expected plugin name in error, got: %q", err.Error())
}

func TestPluginAdd_ShowsPluginInfo(t *testing.T) {
	p := &stubPlugin{
		name:    "myprovider",
		desc:    "My custom VCS provider",
		version: "3.0.0",
		required: []plugin.ConfigKey{
			{Flag: "my-token", EnvVar: "ATLANTIS_MY_TOKEN", Desc: "API token", Required: true},
		},
		optional: []plugin.ConfigKey{
			{Flag: "my-hostname", EnvVar: "ATLANTIS_MY_HOSTNAME", Desc: "Custom hostname", Required: false},
		},
	}
	r := newTestRegistry(t, p)
	pc := &cmd.PluginCmd{Registry: r}
	out, err := captureOutput(t, pc, "add", "myprovider")
	Ok(t, err)
	Assert(t, strings.Contains(out, "myprovider"), "expected plugin name in output, got: %q", out)
	Assert(t, strings.Contains(out, "3.0.0"), "expected version in output, got: %q", out)
	Assert(t, strings.Contains(out, "My custom VCS provider"), "expected description in output, got: %q", out)
	Assert(t, strings.Contains(out, "Required configuration:"), "expected required-config section, got: %q", out)
	Assert(t, strings.Contains(out, "my-token"), "expected required flag in output, got: %q", out)
	Assert(t, strings.Contains(out, "Optional configuration:"), "expected optional-config section, got: %q", out)
	Assert(t, strings.Contains(out, "my-hostname"), "expected optional flag in output, got: %q", out)
}

func TestPluginAdd_NoConfigKeys(t *testing.T) {
	p := &stubPlugin{name: "minimal", desc: "Minimal provider", version: "0.1.0"}
	r := newTestRegistry(t, p)
	pc := &cmd.PluginCmd{Registry: r}
	out, err := captureOutput(t, pc, "add", "minimal")
	Ok(t, err)
	Assert(t, strings.Contains(out, "minimal"), "expected plugin name in output, got: %q", out)
	Assert(t, !strings.Contains(out, "Required configuration:"), "unexpected required-config section in output: %q", out)
	Assert(t, !strings.Contains(out, "Optional configuration:"), "unexpected optional-config section in output: %q", out)
}

func TestPluginAdd_MissingArgument(t *testing.T) {
	pc := &cmd.PluginCmd{Registry: plugin.NewRegistry()}
	_, err := captureOutput(t, pc, "add")
	Assert(t, err != nil, "expected error when no plugin name is given")
}
