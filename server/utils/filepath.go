// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// EnsureSubPath returns an error if path is not contained within base.
// This prevents path traversal attacks where user-controlled data could
// escape the intended directory.
//
// Note: validation is lexical (filepath.Clean) and does not resolve symlinks.
// This is intentional since the paths are constructed, not traversed through
// the filesystem, so symlink resolution is not required.
func EnsureSubPath(base, path string) error {
	cleanBase := filepath.Clean(base)
	cleanPath := filepath.Clean(path)
	// A path is within the base if it equals the base or starts with base + separator.
	if cleanPath != cleanBase && !strings.HasPrefix(cleanPath, cleanBase+string(filepath.Separator)) {
		return fmt.Errorf("path %q escapes base directory %q", cleanPath, cleanBase)
	}
	return nil
}
