// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateCleanPath(t *testing.T) {
	cwd, _ := os.Getwd()
	home := mustHomeDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		skip    string
	}{
		// Allowed
		{name: "default workspace", path: "/tmp/atlantis-tests", wantErr: false},
		{name: "nested under /tmp", path: "/tmp/atlantis-tests/deep/clone", wantErr: false},
		{name: "dotdot-prefixed name under /tmp", path: "/tmp/..foo", wantErr: false},

		// Rejected: empty/whitespace
		{name: "empty", path: "", wantErr: true},
		{name: "whitespace", path: "   ", wantErr: true},

		// Rejected: roots and traversal
		{name: "filesystem root", path: "/", wantErr: true},
		{name: "dot", path: ".", wantErr: true},
		{name: "dotdot", path: "..", wantErr: true},
		{name: "double dotdot", path: "../..", wantErr: true},

		// Rejected: temp roots themselves
		{name: "/tmp itself", path: "/tmp", wantErr: true},
		{name: "/var/tmp itself", path: "/var/tmp", wantErr: true},

		// Rejected: protected runtime paths
		{name: "cwd", path: cwd, wantErr: true},
		{name: "parent of cwd", path: filepath.Dir(cwd), wantErr: true},
		{name: "home directory", path: home, wantErr: true},

		// Rejected: arbitrary paths outside temp roots
		{name: "arbitrary /opt path", path: "/opt/some-dir", wantErr: true},
		{name: "child of home", path: filepath.Join(home, "foo"), wantErr: true},

		// macOS /private/tmp
		{name: "/private/tmp itself", path: "/private/tmp", wantErr: true, skip: skipUnlessDarwin()},
		{name: "child of /private/tmp", path: "/private/tmp/atlantis-tests", wantErr: false, skip: skipUnlessDarwin()},
		{name: "/private/var/tmp itself", path: "/private/var/tmp", wantErr: true, skip: skipUnlessDarwin()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			_, err := validateCleanPath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("validateCleanPath(%q) = nil, want error", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateCleanPath(%q) = %v, want nil", tt.path, err)
			}
		})
	}
}

func TestValidateCleanPath_UnsafeTMPDIR(t *testing.T) {
	home := mustHomeDir()

	tests := []struct {
		name   string
		tmpdir string
		path   string
	}{
		{name: "TMPDIR=home", tmpdir: home, path: filepath.Join(home, "workspace")},
		{name: "TMPDIR=/opt/tmp", tmpdir: "/opt/tmp", path: "/opt/tmp/workspace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TMPDIR", tt.tmpdir)
			_, err := validateCleanPath(tt.path)
			if err == nil {
				t.Errorf("validateCleanPath(%q) with TMPDIR=%q should be rejected", tt.path, tt.tmpdir)
			}
		})
	}
}

func TestValidateCleanPath_NestedTMPDIRRoot(t *testing.T) {
	// Regression: TMPDIR=/tmp/session should not allow CLONE_DIR=/tmp/session
	// even though /tmp/session is a child of approved root /tmp. The nested
	// root itself must be rejected by equality check before child acceptance.
	nestedRoot, err := os.MkdirTemp("/tmp", "atlantis-e2e-tmpdir-root-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(nestedRoot) }) //nolint:errcheck

	t.Setenv("TMPDIR", nestedRoot)

	// The nested root itself must be rejected.
	_, err = validateCleanPath(nestedRoot)
	if err == nil {
		t.Errorf("validateCleanPath(%q) with TMPDIR=%q should reject the root itself", nestedRoot, nestedRoot)
	}

	// A child of the nested root must be allowed.
	child := filepath.Join(nestedRoot, "workspace")
	_, err = validateCleanPath(child)
	if err != nil {
		t.Errorf("validateCleanPath(%q) with TMPDIR=%q should allow child: %v", child, nestedRoot, err)
	}
}

func TestValidateCleanPath_CheckoutUnderTmp(t *testing.T) {
	// Create test dirs explicitly under /tmp so they are under approved roots.
	tmpRoot, err := os.MkdirTemp("/tmp", "atlantis-e2e-validate-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpRoot) }) //nolint:errcheck

	fakeRepo := filepath.Join(tmpRoot, "atlantis")
	fakeE2E := filepath.Join(fakeRepo, "e2e")
	if err := os.MkdirAll(fakeE2E, 0700); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(fakeE2E); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) }) //nolint:errcheck

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "dot from checkout under tmp", path: ".", wantErr: true},
		{name: "dotdot from checkout under tmp", path: "..", wantErr: true},
		{name: "fake repo root", path: fakeRepo, wantErr: true},
		{name: "fake e2e dir", path: fakeE2E, wantErr: true},
		{name: "child inside fake repo", path: filepath.Join(fakeRepo, "subdir"), wantErr: true},
		{name: "sibling of fake repo is allowed", path: filepath.Join(tmpRoot, "other-workspace"), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateCleanPath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("validateCleanPath(%q) = nil, want error (checkout under /tmp)", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateCleanPath(%q) = %v, want nil", tt.path, err)
			}
		})
	}
}

func TestIsPathBelow(t *testing.T) {
	tests := []struct {
		base      string
		candidate string
		want      bool
	}{
		{"/tmp", "/tmp/child", true},
		{"/tmp", "/tmp/deep/nested", true},
		{"/tmp", "/tmp/..foo", true},
		{"/tmp", "/tmp", false},
		{"/tmp", "/tmp/../etc", false},
		{"/tmp", "/other", false},
		{"/tmp", "/", false},
	}
	for _, tt := range tests {
		t.Run(tt.candidate, func(t *testing.T) {
			got := isPathBelow(tt.base, tt.candidate)
			if got != tt.want {
				t.Errorf("isPathBelow(%q, %q) = %v, want %v", tt.base, tt.candidate, got, tt.want)
			}
		})
	}
}

func mustHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/nonexistent-home"
	}
	return home
}

func skipUnlessDarwin() string {
	if runtime.GOOS != "darwin" {
		return "macOS-specific path test"
	}
	return ""
}
