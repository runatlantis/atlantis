// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestValidateAPIAllowVarPattern(t *testing.T) {
	cases := []struct {
		pattern string
		expErr  string
	}{
		{pattern: "*"},
		{pattern: "foo_bar_*"},
		{pattern: "foo_bar_baz"},
		{pattern: "_foo-bar"},
		{pattern: "", expErr: `invalid api_allow_vars pattern "": must be a variable name, optionally ending in '*'`},
		{pattern: "**", expErr: `invalid api_allow_vars pattern "**": '*' is only allowed at the end`},
		{pattern: "foo_*_baz", expErr: `invalid api_allow_vars pattern "foo_*_baz": '*' is only allowed at the end`},
		{pattern: "*_baz", expErr: `invalid api_allow_vars pattern "*_baz": '*' is only allowed at the end`},
		{pattern: "1foo", expErr: `invalid api_allow_vars pattern "1foo": must be a variable name, optionally ending in '*'`},
		{pattern: "foo.bar*", expErr: `invalid api_allow_vars pattern "foo.bar*": must be a variable name, optionally ending in '*'`},
	}
	for _, c := range cases {
		t.Run(c.pattern, func(t *testing.T) {
			err := valid.ValidateAPIAllowVarPattern(c.pattern)
			if c.expErr == "" {
				Ok(t, err)
			} else {
				ErrEquals(t, c.expErr, err)
			}
		})
	}
}

func TestAPIVarAllowed(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string
		varName  string
		exp      bool
	}{
		{name: "no patterns", patterns: nil, varName: "foo", exp: false},
		{name: "any", patterns: []string{"*"}, varName: "foo", exp: true},
		{name: "exact match", patterns: []string{"foo_bar_baz"}, varName: "foo_bar_baz", exp: true},
		{name: "exact pattern does not match a prefix", patterns: []string{"foo_bar"}, varName: "foo_bar_baz", exp: false},
		{name: "prefix match", patterns: []string{"foo_bar_*"}, varName: "foo_bar_baz", exp: true},
		{name: "prefix matches the bare prefix", patterns: []string{"foo_bar_*"}, varName: "foo_bar_", exp: true},
		{name: "prefix mismatch", patterns: []string{"foo_bar_*"}, varName: "foo_baz", exp: false},
		{name: "case sensitive", patterns: []string{"foo_*"}, varName: "FOO_bar", exp: false},
		{name: "second pattern matches", patterns: []string{"baz", "foo_*"}, varName: "foo_bar", exp: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			Equals(t, c.exp, valid.APIVarAllowed(c.patterns, c.varName))
		})
	}
}

func TestGlobalCfg_APIAllowVars(t *testing.T) {
	const repoID = "github.com/owner/repo"
	cases := []struct {
		name        string
		flag        []string
		serverCfg   string
		projectVars []string
		exp         []string
	}{
		{
			name: "nothing configured",
			exp:  nil,
		},
		{
			name: "server flag",
			flag: []string{"foo_*"},
			exp:  []string{"foo_*"},
		},
		{
			name: "server-side repo config replaces the flag",
			flag: []string{"foo_*"},
			serverCfg: `
repos:
- id: /.*/
  api_allow_vars: [bar_*]
`,
			exp: []string{"bar_*"},
		},
		{
			name: "last matching server-side repo wins",
			serverCfg: `
repos:
- id: /.*/
  api_allow_vars: [bar_*]
- id: github.com/owner/repo
  api_allow_vars: [baz]
- id: github.com/owner/other
  api_allow_vars: [qux]
`,
			exp: []string{"baz"},
		},
		{
			name: "project replaces server-side config when allowed",
			flag: []string{"foo_*"},
			serverCfg: `
repos:
- id: /.*/
  allowed_overrides: [api_allow_vars]
  api_allow_vars: [bar_*]
`,
			projectVars: []string{"baz"},
			exp:         []string{"baz"},
		},
		{
			name: "empty project list allows nothing",
			flag: []string{"foo_*"},
			serverCfg: `
repos:
- id: /.*/
  allowed_overrides: [api_allow_vars]
`,
			projectVars: []string{},
			exp:         []string{},
		},
		{
			name: "project without the key inherits",
			flag: []string{"foo_*"},
			serverCfg: `
repos:
- id: /.*/
  allowed_overrides: [api_allow_vars]
`,
			exp: []string{"foo_*"},
		},
		{
			name: "project ignored without allowed_overrides",
			flag: []string{"foo_*"},
			serverCfg: `
repos:
- id: /.*/
  allowed_overrides: [workflow]
`,
			projectVars: []string{"*"},
			exp:         []string{"foo_*"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			global := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{APIAllowVars: c.flag})
			if c.serverCfg != "" {
				path := filepath.Join(t.TempDir(), "repos.yaml")
				Ok(t, os.WriteFile(path, []byte(c.serverCfg), 0600))
				var err error
				global, err = (&config.ParserValidator{}).ParseGlobalCfg(path, global)
				Ok(t, err)
			}
			log := logging.NewNoopLogger(t)
			proj := valid.Project{Dir: ".", Workspace: "default", APIAllowVars: c.projectVars}

			merged := global.MergeProjectCfg(log, repoID, proj, valid.RepoCfg{})
			Equals(t, c.exp, merged.APIAllowVars)

			if c.projectVars == nil {
				Equals(t, c.exp, global.DefaultProjCfg(log, repoID, ".", "default").APIAllowVars)
			}
		})
	}
}
