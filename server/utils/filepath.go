// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrPathEscapesBase is returned when a path is outside the expected base directory.
var ErrPathEscapesBase = errors.New("path escapes base directory")

// EnsureSubPath returns an error if path is not contained within base.
// This prevents path traversal attacks where user-controlled data could
// escape the intended directory.
func EnsureSubPath(base, path string) error {
	cleanBase := filepath.Clean(base)
	cleanPath := filepath.Clean(path)
	// A path is within the base if it equals the base or starts with base + separator.
	if cleanPath != cleanBase && !strings.HasPrefix(cleanPath, cleanBase+string(os.PathSeparator)) {
		return ErrPathEscapesBase
	}
	return nil
}

// EnsureSubPathResolved returns an error if path is not contained within base
// once symlinks in both are resolved. EnsureSubPath compares the lexical
// paths only, so a symlink inside base whose target lies outside it passes
// that check; this function follows the links and re-checks containment.
//
// It returns ErrPathEscapesBase when the lexical or the resolved path escapes
// base, and the underlying filesystem error (wrapping fs.ErrNotExist for a
// missing or dangling path) when either path cannot be resolved.
func EnsureSubPathResolved(base, path string) error {
	if err := EnsureSubPath(base, path); err != nil {
		return err
	}
	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return fmt.Errorf("resolving base directory: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	return EnsureSubPath(resolvedBase, resolvedPath)
}
