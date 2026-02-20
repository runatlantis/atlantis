// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package github provides the GitHubPlugin type for the built-in GitHub VCS
// provider. The plugin is pre-registered in plugin.DefaultRegistry by the
// plugin package itself, so no import side-effects are required.
//
// This package is useful when you need to reference or instantiate the
// exported GitHubPlugin type directly (e.g. in tests or custom integrations).
package github

import "github.com/runatlantis/atlantis/plugin"

// GitHubPlugin is the built-in Atlantis plugin for the GitHub VCS provider.
type GitHubPlugin struct{}

// Name returns the plugin identifier used with "atlantis plugin add github".
func (g *GitHubPlugin) Name() string { return "github" }

// Description returns a short description of the plugin.
func (g *GitHubPlugin) Description() string {
	return "GitHub VCS provider (github.com and GitHub Enterprise)"
}

// Version returns the plugin version.
func (g *GitHubPlugin) Version() string { return "1.0.0" }

// SourceURL returns the canonical source repository for the GitHub plugin.
// It delegates to plugin.LookupSource so that the URL is defined exactly once
// in the plugin catalog (plugin/plugin.go), preventing drift.
//
// When the plugin is extracted to its own repository, only the catalog entry
// in plugin/plugin.go needs updating; this method automatically reflects the
// new value.
func (g *GitHubPlugin) SourceURL() string {
	url, _ := plugin.LookupSource("github")
	return url
}

// ConfigKeys returns the configuration keys required and accepted by the
// GitHub provider.
//
// NOTE: This list is intentionally mirrored in plugin.githubPlugin.ConfigKeys
// (plugin/plugin.go) to avoid an import cycle. Keep both in sync.
func (g *GitHubPlugin) ConfigKeys() []plugin.ConfigKey {
	return []plugin.ConfigKey{
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
