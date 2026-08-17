// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SnapshotValidatedPlan copies sourcePath to a private, read-only file only
// when its exact contents match expectedHash.
func SnapshotValidatedPlan(sourcePath, expectedHash string) (string, error) {
	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("reading validated plan source: %w", err)
	}
	actualHash := sha256.Sum256(contents)
	if expectedHash == "" || hex.EncodeToString(actualHash[:]) != expectedHash {
		return "", errors.New("validated plan source changed before snapshot")
	}
	return writeValidatedPlanSnapshot(filepath.Dir(sourcePath), contents)
}

func writeValidatedPlanSnapshot(path string, contents []byte) (string, error) {
	snapshot, err := os.CreateTemp(path, ".atlantis-validated-plan-*")
	if err != nil {
		return "", fmt.Errorf("creating validated plan snapshot: %w", err)
	}
	snapshotPath := snapshot.Name()
	removeSnapshot := func() {
		_ = snapshot.Close()
		_ = os.Remove(snapshotPath)
	}

	if _, err := snapshot.Write(contents); err != nil {
		removeSnapshot()
		return "", fmt.Errorf("writing validated plan snapshot: %w", err)
	}
	if err := snapshot.Close(); err != nil {
		_ = os.Remove(snapshotPath)
		return "", fmt.Errorf("closing validated plan snapshot: %w", err)
	}
	if err := os.Chmod(snapshotPath, 0o400); err != nil {
		_ = os.Remove(snapshotPath)
		return "", fmt.Errorf("protecting validated plan snapshot: %w", err)
	}
	return snapshotPath, nil
}
