// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/utils"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRemoveIgnoreNonExistent_FileExists(t *testing.T) {
	file := filepath.Join(t.TempDir(), "to-remove.txt")
	Ok(t, os.WriteFile(file, []byte("content"), 0o600))

	err := utils.RemoveIgnoreNonExistent(file)
	Ok(t, err)

	_, statErr := os.Stat(file)
	Assert(t, os.IsNotExist(statErr), "expected file to be removed")
}

func TestRemoveIgnoreNonExistent_FileDoesNotExist(t *testing.T) {
	file := filepath.Join(t.TempDir(), "does-not-exist.txt")

	err := utils.RemoveIgnoreNonExistent(file)
	Ok(t, err)
}

func TestRemoveIgnoreNonExistent_RemoveError(t *testing.T) {
	nonEmptyDir := filepath.Join(t.TempDir(), "non-empty-dir")
	Ok(t, os.Mkdir(nonEmptyDir, 0o755))
	Ok(t, os.WriteFile(filepath.Join(nonEmptyDir, "child.txt"), []byte("x"), 0o600))

	err := utils.RemoveIgnoreNonExistent(nonEmptyDir)
	Assert(t, err != nil, "expected remove error for non-empty directory")
	Assert(t, !os.IsNotExist(err), "expected non-not-exist error")
}
