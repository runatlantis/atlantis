// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

func FuzzParseRepoCfgData(f *testing.F) {
	f.Add([]byte(`version: 3
projects:
- dir: .
`))
	f.Add([]byte(`version: 3
automerge: true
projects:
- dir: .
  workspace: default
  autoplan:
    when_modified: ["*.tf"]
    enabled: true
`))

	pv := config.ParserValidator{}
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = pv.ParseRepoCfgData(data, globalCfg, "github.com/test/repo", "main")
	})
}
