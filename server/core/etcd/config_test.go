// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

// validExternal returns a production-valid external-mode config that individual
// tests mutate to exercise one failure at a time.
func validExternal() *etcd.Config {
	return &etcd.Config{
		Mode:           etcd.ModeExternal,
		DeploymentID:   "dep-1",
		Namespace:      "/atlantis",
		TLS:            etcd.TLSConfig{CAFile: "ca.pem", CertFile: "c.pem", KeyFile: "k.pem", ServerName: "etcd"},
		RequestTimeout: 5 * time.Second,
		StartupTimeout: 5 * time.Minute,
		Endpoints:      []string{"https://member-0:2379"},
		Ownership: etcd.OwnershipConfig{
			ReplicaID:                 "replica-0",
			ReplicaAdvertiseURL:       "https://replica-0:4141",
			ReplicaAdvertiseAllowlist: []string{"10.0.0.0/8"},
			InternalCommandTokenFile:  "token",
			InternalCommandCAFile:     "internal-ca.pem",
			TTL:                       30 * time.Second,
		},
	}
}

func TestValidate_External_OK(t *testing.T) {
	Ok(t, validExternal().Validate())
}

// TestValidate_Embedded_Deferred proves embedded mode is rejected in this phase
// with a message pointing operators at external mode.
func TestValidate_Embedded_Deferred(t *testing.T) {
	c := validExternal()
	c.Mode = etcd.ModeEmbedded
	ErrContains(t, "not available in this release", c.Validate())
}

func TestValidate_CommonFailures(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*etcd.Config)
		substr string
	}{
		{"no deployment id", func(c *etcd.Config) { c.DeploymentID = "" }, "etcd-deployment-id"},
		{"no namespace", func(c *etcd.Config) { c.Namespace = "" }, "etcd-namespace"},
		{"bad mode", func(c *etcd.Config) { c.Mode = "weird" }, "etcd-mode"},
		{"nonpositive request timeout", func(c *etcd.Config) { c.RequestTimeout = 0 }, "etcd-request-timeout"},
		{"nonpositive startup timeout", func(c *etcd.Config) { c.StartupTimeout = 0 }, "etcd-startup-timeout"},
		{"username without password", func(c *etcd.Config) { c.Username = "u" }, "etcd-password-file"},
		{"password without username", func(c *etcd.Config) { c.PasswordFile = "p" }, "etcd-username"},
		{"incomplete tls in prod", func(c *etcd.Config) { c.TLS = etcd.TLSConfig{CAFile: "ca"} }, "complete client TLS identity"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := validExternal()
			tc.mutate(c)
			ErrContains(t, tc.substr, c.Validate())
		})
	}
}

func TestValidate_ExternalFailures(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*etcd.Config)
		substr string
	}{
		{"no endpoints", func(c *etcd.Config) { c.Endpoints = nil }, "requires etcd-endpoints"},
		{"http endpoint in prod", func(c *etcd.Config) { c.Endpoints = []string{"http://member-0:2379"} }, "must use https"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := validExternal()
			tc.mutate(c)
			ErrContains(t, tc.substr, c.Validate())
		})
	}
}

func TestValidate_OwnershipFailures(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*etcd.Config)
		substr string
	}{
		{"no replica id", func(c *etcd.Config) { c.Ownership.ReplicaID = "" }, "replica-id"},
		{"no advertise url", func(c *etcd.Config) { c.Ownership.ReplicaAdvertiseURL = "" }, "replica-advertise-url"},
		{"http advertise url in prod", func(c *etcd.Config) { c.Ownership.ReplicaAdvertiseURL = "http://replica-0:4141" }, "must use https"},
		{"no allowlist", func(c *etcd.Config) { c.Ownership.ReplicaAdvertiseAllowlist = nil }, "replica-advertise-allowlist"},
		{"no internal token in prod", func(c *etcd.Config) { c.Ownership.InternalCommandTokenFile = "" }, "internal-command-token-file"},
		{"no internal ca in prod", func(c *etcd.Config) { c.Ownership.InternalCommandCAFile = "" }, "internal-command-ca-file"},
		{"ttl below floor", func(c *etcd.Config) { c.Ownership.TTL = 5 * time.Second }, "ownership-ttl-seconds must be at least"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := validExternal()
			tc.mutate(c)
			ErrContains(t, tc.substr, c.Validate())
		})
	}
}

// TestValidate_InsecureDev_LoopbackOnly proves the development flag permits HTTP
// only for loopback and still rejects a non-loopback insecure endpoint.
func TestValidate_InsecureDev_LoopbackOnly(t *testing.T) {
	c := validExternal()
	c.AllowInsecureDev = true
	c.TLS = etcd.TLSConfig{}
	c.Endpoints = []string{"http://127.0.0.1:2379"}
	c.Ownership.ReplicaAdvertiseURL = "http://localhost:4141"
	c.Ownership.InternalCommandTokenFile = ""
	c.Ownership.InternalCommandCAFile = ""
	Ok(t, c.Validate())

	c.Endpoints = []string{"http://member-0:2379"}
	ErrContains(t, "loopback", c.Validate())
}
