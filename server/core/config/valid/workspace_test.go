// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestValidateWorkspaceName(t *testing.T) {
	cases := []struct {
		description string
		workspace   string
		expErr      string
	}{
		// An unset workspace is filled in with the default later on, so it is
		// not this function's job to reject it.
		{"empty is allowed", "", ""},
		{"simple", "default", ""},
		{"dashes", "my-workspace", ""},
		{"underscores", "my_workspace", ""},
		{"dots", "my.workspace", ""},
		{"digits", "ws123", ""},
		{"leading digit", "1ws", ""},
		{"mixed", "My.work_space-1", ""},

		// Names Terraform itself accepts, and that Atlantis accepted before
		// workspace validation was tightened. Verified against Terraform
		// 1.15.9: `terraform workspace new` creates all of these.
		{"leading underscore", "_foo", ""},
		{"leading dot", ".hidden", ""},
		{"at sign", "foo@bar", ""},
		{"plus", "foo+bar", ""},
		{"colon", "foo:bar", ""},
		{"equals", "foo=bar", ""},
		{"ampersand", "foo&bar", ""},

		// '$' is excluded because Atlantis expands environment variables in
		// Terraform arguments. Under the previous shell-based command line the
		// shell did the same, so a name like foo$bar was silently truncated to
		// foo rather than working; rejecting it is clearer than mangling it.
		{"dollar", "foo$bar", workspaceDollarErr()},
		{"embedded tilde", "foo~bar", ""},

		// Path separators: the workspace is used to build file paths.
		{"forward slash", "sub/dir", workspacePathErr()},
		{"backslash", "sub\\dir", workspacePathErr()},
		{"absolute path", "/etc", workspacePathErr()},

		// Path traversal.
		{"parent dir", "../evil", workspacePathErr()},
		{"embedded dotdot", "abc..abc", workspaceDotDotErr()},
		{"trailing dotdot", "abc..", workspaceDotDotErr()},
		{"just dotdot", "..", workspaceDotDotErr()},
		{"just dot", ".", workspaceDotErr()},

		// A leading '-' is parsed by Terraform as a flag rather than a
		// workspace name; Terraform rejects it too.
		{"leading dash", "-chdir", workspaceDashErr()},

		// A leading '~' is expanded by shells and by some tools that build
		// paths from the name.
		{"leading tilde", "~root", workspaceTildeErr()},

		// Whitespace and control characters would break the comment parser,
		// which splits on whitespace, and make unusable filenames. Terraform
		// rejects these as well.
		{"space", "a b", workspaceCharErr()},
		{"tab", "a\tb", workspaceCharErr()},
		{"newline", "a\nb", workspaceCharErr()},
		{"null", "a\x00b", workspaceCharErr()},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := valid.ValidateWorkspaceName(c.workspace)
			if c.expErr == "" {
				Ok(t, err)
				return
			}
			ErrEquals(t, c.expErr, err)
		})
	}
}

// The error carries only the reason. Callers add their own context: the
// repo-config validator renders it under the "workspace" yaml key, and the
// comment parser prefixes it with the offending value.
func workspacePathErr() string {
	return "cannot contain '/' or '\\'"
}

func workspaceDotDotErr() string {
	return "cannot contain '..'"
}

func workspaceDotErr() string {
	return "cannot be '.'"
}

func workspaceDashErr() string {
	return "cannot start with '-'"
}

func workspaceTildeErr() string {
	return "cannot start with '~'"
}

func workspaceDollarErr() string {
	return "cannot contain '$'"
}

func workspaceCharErr() string {
	return "cannot contain whitespace or control characters"
}
