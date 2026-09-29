// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

// goReleaserSignConfig represents the signs section of .goreleaser.yml.
type goReleaserSignConfig struct {
	Signs []struct {
		Artifacts string   `yaml:"artifacts"`
		Args      []string `yaml:"args"`
	} `yaml:"signs"`
}

// TestGoReleaserSigningConfig validates that .goreleaser.yml contains the
// expected GPG signing configuration. This guards against accidental removal
// or corruption of the signing setup that would break release signatures.
func TestGoReleaserSigningConfig(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		t.Fatalf("reading .goreleaser.yml: %v", err)
	}

	var cfg goReleaserSignConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing .goreleaser.yml: %v", err)
	}

	if len(cfg.Signs) == 0 {
		t.Fatal("signs section is missing from .goreleaser.yml; GPG signing will not work")
	}

	sign := cfg.Signs[0]

	if sign.Artifacts != "checksum" {
		t.Errorf("signs[0].artifacts: got %q, want %q", sign.Artifacts, "checksum")
	}

	// These args are required for the gpg detached-signature command to work
	// correctly in CI using the fingerprint exported by ghaction-import-gpg.
	requiredArgs := []string{
		"--batch",
		"-u",
		"{{ .Env.GPG_FINGERPRINT }}",
		"--output",
		"${signature}",
		"--detach-sign",
		"${artifact}",
	}
	for _, want := range requiredArgs {
		if !slices.Contains(sign.Args, want) {
			t.Errorf("signs[0].args is missing required argument %q (got: %v)", want, sign.Args)
		}
	}
}
