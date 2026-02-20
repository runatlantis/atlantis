// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"testing"

	"github.com/runatlantis/atlantis/plugin"
	ghplugin "github.com/runatlantis/atlantis/plugin/github"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGitHubPlugin_Name(t *testing.T) {
	p := &ghplugin.GitHubPlugin{}
	Equals(t, "github", p.Name())
}

func TestGitHubPlugin_Description(t *testing.T) {
	p := &ghplugin.GitHubPlugin{}
	Assert(t, p.Description() != "", "expected non-empty description")
}

func TestGitHubPlugin_Version(t *testing.T) {
	p := &ghplugin.GitHubPlugin{}
	Assert(t, p.Version() != "", "expected non-empty version")
}

func TestGitHubPlugin_ConfigKeys_RequiredPresent(t *testing.T) {
	p := &ghplugin.GitHubPlugin{}
	keys := p.ConfigKeys()
	Assert(t, len(keys) > 0, "expected at least one config key")

	required := map[string]bool{}
	for _, k := range keys {
		if k.Required {
			required[k.Flag] = true
		}
	}

	// These three flags are the minimum required to connect to GitHub.
	for _, flag := range []string{"gh-user", "gh-token", "gh-webhook-secret"} {
		Assert(t, required[flag], "expected flag %q to be marked required", flag)
	}
}

func TestGitHubPlugin_ConfigKeys_FieldsPopulated(t *testing.T) {
	p := &ghplugin.GitHubPlugin{}
	for _, k := range p.ConfigKeys() {
		Assert(t, k.Flag != "", "expected non-empty Flag for every ConfigKey")
		Assert(t, k.EnvVar != "", "expected non-empty EnvVar for every ConfigKey")
		Assert(t, k.Desc != "", "expected non-empty Desc for every ConfigKey")
	}
}

func TestGitHubPlugin_RegistersInDefaultRegistry(t *testing.T) {
	// GitHub is pre-registered in DefaultRegistry by the plugin package itself;
	// no import side-effect is required.
	p, ok := plugin.DefaultRegistry.Get("github")
	Assert(t, ok, "expected GitHub plugin to be present in DefaultRegistry by default")
	Equals(t, "github", p.Name())
}
