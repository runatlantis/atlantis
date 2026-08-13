// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// RemoveIgnoreNonExistent removes a file, ignoring if it doesn't exist.
func RemoveIgnoreNonExistent(file string) error {
	err := os.Remove(file)
	if err == nil || os.IsNotExist(err) {
		return nil
	}

	return err
}
