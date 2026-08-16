// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"errors"
	"fmt"
)

// DirNotExistErr is an error caused by the directory not existing.
type DirNotExistErr struct {
	RepoRelDir string
}

// Error implements the error interface.
func (d DirNotExistErr) Error() string {
	return fmt.Sprintf("dir %q does not exist", d.RepoRelDir)
}

// ErrStaleCommandHead marks a result produced against a pull head that is
// no longer current; such results must not overwrite fresher status.
var ErrStaleCommandHead = errors.New("stale command head")

// HasApplyResult returns true when any result came from an apply.
func HasApplyResult(results []ProjectResult) bool {
	for _, result := range results {
		if result.Command == Apply {
			return true
		}
	}
	return false
}
