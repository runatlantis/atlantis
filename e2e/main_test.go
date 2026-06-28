// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateCleanPath(t *testing.T) {
	tempDir := os.TempDir()
	cwd, _ := os.Getwd()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		skip    string
	}{
		// Allowed paths
		{
			name:    "default workspace under /tmp",
			path:    "/tmp/atlantis-tests",
			wantErr: false,
		},
		{
			name:    "nested workspace under /tmp",
			path:    "/tmp/atlantis-tests/deep/clone",
			wantErr: false,
		},
		{
			name:    "child of os.TempDir",
			path:    filepath.Join(tempDir, "e2e-workspace"),
			wantErr: false,
		},

		// Rejected: empty/whitespace
		{
			name:    "empty string",
			path:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			path:    "   ",
			wantErr: true,
		},

		// Rejected: dangerous roots
		{
			name:    "filesystem root",
			path:    "/",
			wantErr: true,
		},
		{
			name:    "dot (current dir)",
			path:    ".",
			wantErr: true,
		},
		{
			name:    "dotdot (parent dir)",
			path:    "..",
			wantErr: true,
		},
		{
			name:    "double dotdot",
			path:    "../..",
			wantErr: true,
		},

		// Rejected: temp roots themselves
		{
			name:    "/tmp itself",
			path:    "/tmp",
			wantErr: true,
		},
		{
			name:    "/var/tmp itself",
			path:    "/var/tmp",
			wantErr: true,
		},
		{
			name:    "os.TempDir itself",
			path:    tempDir,
			wantErr: true,
		},

		// Rejected: cwd and parent
		{
			name:    "current working directory",
			path:    cwd,
			wantErr: true,
		},
		{
			name:    "parent of cwd (repo root)",
			path:    filepath.Dir(cwd),
			wantErr: true,
		},

		// Rejected: home directory
		{
			name:    "home directory",
			path:    mustHomeDir(),
			wantErr: true,
		},

		// Rejected: arbitrary paths outside temp roots
		{
			name:    "arbitrary absolute path",
			path:    "/opt/some-dir",
			wantErr: true,
		},
		{
			name:    "user-relative path outside temp",
			path:    filepath.Join(mustHomeDir(), "foo"),
			wantErr: true,
		},

		// macOS symlink: /private/tmp
		{
			name:    "/private/tmp itself",
			path:    "/private/tmp",
			wantErr: true,
			skip:    skipUnlessDarwin(),
		},
		{
			name:    "child of /private/tmp is allowed",
			path:    "/private/tmp/atlantis-tests",
			wantErr: false,
			skip:    skipUnlessDarwin(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			_, err := validateCleanPath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("validateCleanPath(%q) = nil error, want rejection", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateCleanPath(%q) = %v, want allowed", tt.path, err)
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
