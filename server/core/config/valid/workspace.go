// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid

import (
	"errors"
	"strings"
	"unicode"
)

// ValidateWorkspaceName checks that a workspace name is safe for the two things
// Atlantis does with it: using it as a Terraform command argument, and using it
// as a component of the paths Atlantis writes (plan files, env/<workspace>.tfvars,
// the plan directory).
//
// Atlantis builds its Terraform command lines as argument vectors rather than
// shell source, so this check deliberately does not exclude shell
// metacharacters. Terraform accepts names such as `foo&bar` and `foo@bar`, and
// Atlantis accepted them before, so rejecting them here would break existing
// workspaces without adding safety.
//
// What it does exclude is what would be unsafe or unusable regardless of any
// shell: path separators and traversal, a leading '-' that Terraform would read
// as a flag, a leading '~' that gets expanded into a home directory, and
// whitespace or control characters, which cannot survive the comment parser and
// make unusable filenames.
//
// An empty name is accepted because callers substitute the default workspace
// later.
//
// The returned error carries only the reason, so that callers can add the
// context that suits them: the repo-config validator renders it under the
// "workspace" yaml key, and the comment parser prefixes it with the offending
// value.
func ValidateWorkspaceName(workspace string) error {
	if workspace == "" {
		return nil
	}
	if strings.ContainsAny(workspace, `/\`) {
		return errors.New(`cannot contain '/' or '\'`)
	}
	if strings.Contains(workspace, "..") {
		return errors.New("cannot contain '..'")
	}
	if workspace == "." {
		return errors.New("cannot be '.'")
	}
	if strings.HasPrefix(workspace, "-") {
		return errors.New("cannot start with '-'")
	}
	if strings.HasPrefix(workspace, "~") {
		return errors.New("cannot start with '~'")
	}
	// Atlantis expands environment variables in Terraform arguments, so a '$'
	// in the name would be replaced rather than passed through. The previous
	// shell-based command line expanded it too, meaning such a name never
	// worked; rejecting it is clearer than silently mangling it.
	if strings.Contains(workspace, "$") {
		return errors.New("cannot contain '$'")
	}
	for _, r := range workspace {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return errors.New("cannot contain whitespace or control characters")
		}
	}
	return nil
}
