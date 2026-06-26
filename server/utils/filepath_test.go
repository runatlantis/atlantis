// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"errors"
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
