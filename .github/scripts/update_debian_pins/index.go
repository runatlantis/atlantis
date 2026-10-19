// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// readVersions scans a Debian Packages index and returns every version listed
// for the wanted packages. It only reads the Package and Version fields, so the
// rest of each stanza is skipped. Continuation lines start with whitespace and
// are never mistaken for fields.
func readVersions(r io.Reader, wanted map[string]bool) (map[string][]string, error) {
	versions := make(map[string][]string)
	scanner := bufio.NewScanner(r)
	// Some stanzas carry long single-line fields (for example Provides).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	pkg := ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			pkg = ""
		case strings.HasPrefix(line, "Package: "):
			pkg = strings.TrimSpace(strings.TrimPrefix(line, "Package: "))
		case strings.HasPrefix(line, "Version: ") && wanted[pkg]:
			versions[pkg] = append(versions[pkg], strings.TrimSpace(strings.TrimPrefix(line, "Version: ")))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading package index: %w", err)
	}
	return versions, nil
}
