// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"testing"
)

func TestIsGHECloud(t *testing.T) {
	tests := []struct {
		hostname string
		want     bool
	}{
		{"github.com", false},
		{"my-enterprise.example.com", false},
		{"acme.ghe.com", true},
		{"tenant.ghe.com", true},
		{"TENANT.GHE.COM", true},
		{"ghe.com", false},
		// host:port inputs
		{"tenant.ghe.com:443", true},
		{"my-enterprise.example.com:8443", false},
		// api. prefixed inputs
		{"api.tenant.ghe.com", true},
		{"api.tenant.ghe.com:443", true},
	}
	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			if got := isGHECloud(tt.hostname); got != tt.want {
				t.Errorf("isGHECloud(%q) = %v, want %v", tt.hostname, got, tt.want)
			}
		})
	}
}

func TestResolveGithubAPIURL(t *testing.T) {
	tests := []struct {
		hostname string
		wantURL  string
	}{
		{"github.com", "https://api.github.com/"},
		{"my-enterprise.example.com", "https://my-enterprise.example.com/api/v3/"},
		{"acme.ghe.com", "https://api.acme.ghe.com/"},
		{"tenant.ghe.com", "https://api.tenant.ghe.com/"},
		// host:port — port is stripped for GHE Cloud API host
		{"tenant.ghe.com:443", "https://api.tenant.ghe.com/"},
		// api. prefix should not be doubled
		{"api.tenant.ghe.com", "https://api.tenant.ghe.com/"},
		{"api.tenant.ghe.com:443", "https://api.tenant.ghe.com/"},
	}
	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			got := resolveGithubAPIURL(tt.hostname)
			if got.String() != tt.wantURL {
				t.Errorf("resolveGithubAPIURL(%q) = %q, want %q", tt.hostname, got.String(), tt.wantURL)
			}
		})
	}
}
