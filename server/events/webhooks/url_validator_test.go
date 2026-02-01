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
		expectErr       error
		expectErrSubstr string
	}{
		{
			name:      "HTTP scheme not allowed",
			url:       "http://example.com/webhook",
			expectErr: webhooks.ErrInvalidScheme,
		},
		{
			name:      "localhost forbidden",
			url:       "https://localhost/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "127.0.0.1 forbidden",
			url:       "https://127.0.0.1/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "127.0.0.1 with port forbidden",
			url:       "https://127.0.0.1:8080/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "private IP 10.0.0.1 forbidden",
			url:       "https://10.0.0.1/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "private IP 172.16.0.1 forbidden",
			url:       "https://172.16.0.1/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "private IP 192.168.1.1 forbidden",
			url:       "https://192.168.1.1/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "AWS metadata service forbidden",
			url:       "https://169.254.169.254/latest/meta-data/",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "link-local address forbidden",
			url:       "https://169.254.0.1/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "IPv6 loopback forbidden",
			url:       "https://[::1]/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "IPv6 link-local forbidden",
			url:       "https://[fe80::1]/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "IPv6 unique local forbidden",
			url:       "https://[fc00::1]/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "IPv4-mapped IPv6 private IP forbidden",
			url:       "https://[::ffff:192.168.1.1]/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "IPv4-mapped IPv6 loopback forbidden",
			url:       "https://[::ffff:127.0.0.1]/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:      "IPv6 documentation address forbidden",
			url:       "https://[2001:db8::1]/webhook",
			expectErr: webhooks.ErrPrivateIP,
		},
		{
			name:            "URL with credentials forbidden",
			url:             "https://user:pass@example.com/webhook",
			expectErrSubstr: "must not contain credentials",
		},
		{
			name:      "malformed URL",
			url:       "not a url",
			expectErr: webhooks.ErrInvalidScheme,
		},
		{
			name:      "empty URL",
			url:       "",
			expectErr: webhooks.ErrInvalidScheme,
		},
		{
			name:            "URL with no hostname",
			url:             "https:///path",
			expectErrSubstr: "hostname is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := webhooks.ValidateWebhookURL(tt.url)
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
