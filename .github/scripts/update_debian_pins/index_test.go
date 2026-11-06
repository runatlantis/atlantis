// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
	"strings"
	"testing"
)

const testIndex = `Package: curl
Version: 8.14.1-2+deb13u7
Architecture: amd64
Depends: libcurl4t64 (= 8.14.1-2+deb13u7)

Package: libcurl4t64
Version: 8.14.1-2+deb13u7

Package: curl
Version: 8.14.1-2+deb13u5
Description: command line tool for transferring data with URL syntax
 Version: 9.9.9 is a continuation line, not a field.
`

func TestReadVersions(t *testing.T) {
	got, err := readVersions(strings.NewReader(testIndex), map[string]bool{"curl": true})
	if err != nil {
		t.Fatalf("readVersions() error = %v", err)
	}

	want := map[string][]string{
		"curl": {"8.14.1-2+deb13u7", "8.14.1-2+deb13u5"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("readVersions() = %v, want %v", got, want)
	}
}
