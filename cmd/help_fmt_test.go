// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"strings"
	"testing"
)

func TestUsageTmplIncludesInt64Flags(t *testing.T) {
	out := usageTmpl(stringFlags, intFlags, boolFlags, int64Flags)
	for _, name := range []string{GHAppIDFlag, GHAppInstallationIDFlag} {
		if !strings.Contains(out, "--"+name+"=<value>") {
			t.Errorf("usage template does not include int64 flag %q", name)
		}
	}
}

func TestUsageTmplExcludesHiddenInt64Flags(t *testing.T) {
	hidden := map[string]int64Flag{
		"hidden-int64-flag": {
			description: "hidden flag",
			hidden:      true,
		},
	}
	out := usageTmpl(map[string]stringFlag{}, map[string]intFlag{}, map[string]boolFlag{}, hidden)
	if strings.Contains(out, "--hidden-int64-flag") {
		t.Error("usage template includes hidden int64 flag")
	}
}
