// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

var (
	// ErrInvalidScheme is returned when the URL scheme is not HTTPS.
	ErrInvalidScheme = errors.New("webhook URL must use HTTPS scheme")
	// ErrPrivateIP is returned when the URL resolves to a private IP address.
	ErrPrivateIP = errors.New("webhook URL must not resolve to private IP addresses")
	// ErrInvalidURL is returned when the URL is malformed.
	ErrInvalidURL = errors.New("webhook URL is invalid")
)

// isPrivateIP checks if an IP address is in a private range.
// This includes loopback, link-local, private networks, and cloud metadata services.
func isPrivateIP(ip net.IP) bool {
	// Check for loopback (127.0.0.0/8 for IPv4, ::1 for IPv6)
	if ip.IsLoopback() {
		return true
	}

	// Check for link-local addresses (169.254.0.0/16 for IPv4, fe80::/10 for IPv6)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Define private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",        // RFC1918
		"172.16.0.0/12",     // RFC1918
		"192.168.0.0/16",    // RFC1918
		"fc00::/7",          // IPv6 Unique Local Addresses
		"169.254.169.254/32", // AWS, GCP, Azure metadata service
	}

	for _, cidr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// ValidateWebhookURL validates that a webhook URL is safe to use.
// It checks:
// 1. The URL is well-formed
// 2. The URL uses HTTPS scheme
// 3. The hostname resolves to a public IP address (not private/internal)
func ValidateWebhookURL(urlStr string) error {
	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// Ensure HTTPS scheme
	if parsedURL.Scheme != "https" {
		return ErrInvalidScheme
	}

	// Get the hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("%w: hostname is empty", ErrInvalidURL)
	}

	// Resolve the hostname to IP addresses
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("%w: failed to resolve hostname: %v", ErrInvalidURL, err)
	}

	// Check that at least one IP is not private
	// and ensure no resolved IPs are private
	allPrivate := true
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("%w: %s resolves to private IP %s", ErrPrivateIP, hostname, ip.String())
		}
		if !isPrivateIP(ip) {
			allPrivate = false
		}
	}

	if allPrivate {
		return ErrPrivateIP
	}

	return nil
}
