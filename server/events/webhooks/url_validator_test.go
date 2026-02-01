// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/webhooks"
)

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		allowInsecure   bool
		expectErr       error
		expectErrSubstr string
	}{
		{
			name:          "HTTP scheme not allowed in secure mode",
			url:           "http://example.com/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrInvalidScheme,
		},
		{
			name:          "HTTP scheme allowed in insecure mode",
			url:           "http://localhost:8080/webhook",
			allowInsecure: true,
			expectErr:     nil,
		},
		{
			name:          "localhost forbidden in secure mode",
			url:           "https://localhost/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "localhost allowed in insecure mode",
			url:           "http://localhost/webhook",
			allowInsecure: true,
			expectErr:     nil,
		},
		{
			name:          "127.0.0.1 forbidden in secure mode",
			url:           "https://127.0.0.1/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "127.0.0.1 allowed in insecure mode",
			url:           "http://127.0.0.1/webhook",
			allowInsecure: true,
			expectErr:     nil,
		},
		{
			name:          "127.0.0.1 with port forbidden in secure mode",
			url:           "https://127.0.0.1:8080/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "127.0.0.1 with port allowed in insecure mode",
			url:           "http://127.0.0.1:8080/webhook",
			allowInsecure: true,
			expectErr:     nil,
		},
		{
			name:          "private IP 10.0.0.1 forbidden in secure mode",
			url:           "https://10.0.0.1/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "private IP 10.0.0.1 allowed in insecure mode",
			url:           "http://10.0.0.1/webhook",
			allowInsecure: true,
			expectErr:     nil,
		},
		{
			name:          "private IP 172.16.0.1 forbidden in secure mode",
			url:           "https://172.16.0.1/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "private IP 192.168.1.1 forbidden in secure mode",
			url:           "https://192.168.1.1/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "AWS metadata service forbidden in secure mode",
			url:           "https://169.254.169.254/latest/meta-data/",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "link-local address forbidden in secure mode",
			url:           "https://169.254.0.1/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "IPv6 loopback forbidden in secure mode",
			url:           "https://[::1]/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "IPv6 loopback allowed in insecure mode",
			url:           "http://[::1]/webhook",
			allowInsecure: true,
			expectErr:     nil,
		},
		{
			name:          "IPv6 link-local forbidden in secure mode",
			url:           "https://[fe80::1]/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "IPv6 unique local forbidden in secure mode",
			url:           "https://[fc00::1]/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "IPv4-mapped IPv6 private IP forbidden in secure mode",
			url:           "https://[::ffff:192.168.1.1]/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "IPv4-mapped IPv6 loopback forbidden in secure mode",
			url:           "https://[::ffff:127.0.0.1]/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:          "IPv6 documentation address forbidden in secure mode",
			url:           "https://[2001:db8::1]/webhook",
			allowInsecure: false,
			expectErr:     webhooks.ErrPrivateIP,
		},
		{
			name:            "URL with credentials forbidden",
			url:             "https://user:pass@example.com/webhook",
			allowInsecure:   false,
			expectErrSubstr: "must not contain credentials",
		},
		{
			name:            "URL with credentials forbidden even in insecure mode",
			url:             "http://user:pass@localhost/webhook",
			allowInsecure:   true,
			expectErrSubstr: "must not contain credentials",
		},
		{
			name:            "malformed URL",
			url:             "not a url",
			allowInsecure:   false,
			expectErrSubstr: "hostname is empty",
		},
		{
			name:            "empty URL",
			url:             "",
			allowInsecure:   false,
			expectErrSubstr: "hostname is empty",
		},
		{
			name:            "URL with no hostname",
			url:             "https:///path",
			allowInsecure:   false,
			expectErrSubstr: "hostname is empty",
		},
		{
			name:            "invalid scheme in insecure mode",
			url:             "ftp://localhost/file",
			allowInsecure:   true,
			expectErrSubstr: "scheme must be http or https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := webhooks.ValidateWebhookURL(tt.url, tt.allowInsecure)
			if tt.expectErr == nil && tt.expectErrSubstr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error containing %v, got nil", tt.expectErr)
				} else if !strings.Contains(err.Error(), tt.expectErr.Error()) {
					t.Errorf("expected error containing %v, got: %v", tt.expectErr, err)
				}
			} else if tt.expectErrSubstr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectErrSubstr)
				} else if !strings.Contains(err.Error(), tt.expectErrSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.expectErrSubstr, err)
				}
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	// Note: isPrivateIP is not exported, but we test it indirectly through ValidateWebhookURL
	// This test exists to document the expected behavior
	t.Skip("isPrivateIP is an internal function tested through ValidateWebhookURL")
}
