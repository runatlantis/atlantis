// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/utils"
	. "github.com/runatlantis/atlantis/testing"
)

func TestEnsureSubPath(t *testing.T) {
	cases := []struct {
		base      string
		path      string
		expectErr bool
		desc      string
	}{
		{
			desc:      "path equals base",
			base:      "/data/repos",
			path:      "/data/repos",
			expectErr: false,
		},
		{
			desc:      "path within base",
			base:      "/data/repos",
			path:      "/data/repos/owner/repo/1/default",
			expectErr: false,
		},
		{
			desc:      "path traversal with .. staying within base",
			base:      "/data/repos",
			path:      "/data/repos/owner/repo/../../etc/passwd",
			expectErr: false, // resolves to /data/repos/etc/passwd which is within base
		},
		{
			desc:      "path traversal escaping base with ..",
			base:      "/data/repos",
			path:      "/data/repos/owner/../../../../etc/passwd",
			expectErr: true, // resolves to /etc/passwd which escapes base
		},
		{
			desc:      "path traversal escaping base",
			base:      "/data/repos",
			path:      "/etc/passwd",
			expectErr: true,
		},
		{
			desc:      "path with trailing separator in base",
			base:      "/data/repos/",
			path:      "/data/repos/owner/repo",
			expectErr: false,
		},
		{
			desc:      "base prefix but not subpath",
			base:      "/data/repos",
			path:      "/data/repos-evil/something",
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := utils.EnsureSubPath(tc.base, tc.path)
			if tc.expectErr {
				Assert(t, err != nil, "expected error but got nil")
				Assert(t, errors.Is(err, utils.ErrPathEscapesBase), "expected ErrPathEscapesBase but got %v", err)
				cleanBase := filepath.Clean(tc.base)
				cleanPath := filepath.Clean(tc.path)
				Assert(t, !strings.Contains(err.Error(), tc.base), "error should not expose base path: %q", err.Error())
				Assert(t, !strings.Contains(err.Error(), cleanBase), "error should not expose clean base path: %q", err.Error())
				Assert(t, !strings.Contains(err.Error(), tc.path), "error should not expose candidate path: %q", err.Error())
				Assert(t, !strings.Contains(err.Error(), cleanPath), "error should not expose clean candidate path: %q", err.Error())
			} else {
				Ok(t, err)
			}
		})
	}
}

func TestEnsureSubPathResolved(t *testing.T) {
	// t.TempDir may itself live under a symlink (macOS /var -> /private/var),
	// which also exercises resolution of the base directory.
	tmp := t.TempDir()
	base := filepath.Join(tmp, "repo")
	outside := filepath.Join(tmp, "outside")
	Ok(t, os.MkdirAll(filepath.Join(base, "modules", "real"), 0o755))
	Ok(t, os.MkdirAll(filepath.Join(outside, "child"), 0o755))
	Ok(t, os.Symlink(outside, filepath.Join(base, "abs-escape")))
	Ok(t, os.Symlink(filepath.Join("..", "outside"), filepath.Join(base, "rel-escape")))
	Ok(t, os.Symlink(filepath.Join("modules", "real"), filepath.Join(base, "in-repo")))
	Ok(t, os.Symlink(filepath.Join(outside, "nowhere"), filepath.Join(base, "dangling")))

	cases := []struct {
		desc   string
		path   string
		expErr error
	}{
		{
			desc: "path equals base",
			path: base,
		},
		{
			desc: "regular directory within base",
			path: filepath.Join(base, "modules", "real"),
		},
		{
			desc: "symlink resolving within base",
			path: filepath.Join(base, "in-repo"),
		},
		{
			desc:   "symlink with absolute target outside base",
			path:   filepath.Join(base, "abs-escape"),
			expErr: utils.ErrPathEscapesBase,
		},
		{
			desc:   "symlink with relative .. target outside base",
			path:   filepath.Join(base, "rel-escape"),
			expErr: utils.ErrPathEscapesBase,
		},
		{
			desc:   "path below an escaping symlink",
			path:   filepath.Join(base, "abs-escape", "child"),
			expErr: utils.ErrPathEscapesBase,
		},
		{
			desc:   "lexical escape is rejected without touching the filesystem",
			path:   filepath.Join(base, "..", "outside"),
			expErr: utils.ErrPathEscapesBase,
		},
		{
			desc:   "missing path",
			path:   filepath.Join(base, "missing"),
			expErr: fs.ErrNotExist,
		},
		{
			desc:   "dangling symlink",
			path:   filepath.Join(base, "dangling"),
			expErr: fs.ErrNotExist,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := utils.EnsureSubPathResolved(base, tc.path)
			if tc.expErr == nil {
				Ok(t, err)
				return
			}
			Assert(t, err != nil, "expected error but got nil")
			Assert(t, errors.Is(err, tc.expErr), "expected %v but got %v", tc.expErr, err)
			if errors.Is(err, utils.ErrPathEscapesBase) {
				Assert(t, !strings.Contains(err.Error(), tmp), "error should not expose paths: %q", err.Error())
			}
		})
	}
}

func TestEnsureSubPathResolved_MissingBase(t *testing.T) {
	base := filepath.Join(t.TempDir(), "missing-base")
	err := utils.EnsureSubPathResolved(base, filepath.Join(base, "project"))
	Assert(t, errors.Is(err, fs.ErrNotExist), "expected fs.ErrNotExist but got %v", err)
}
