// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Secret material uses secret-backed file inputs and is never logged or rendered
// into pod arguments (design §170).
package etcd

import (
	"fmt"
	"os"
	"strings"
)

// readSecretFile reads a secret file and trims a single trailing newline so a
// file written with a trailing newline still yields the exact secret. The
// content is never logged.
func readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading secret file: %w", err)
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}
