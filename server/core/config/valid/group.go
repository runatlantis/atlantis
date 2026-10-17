// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid

import (
	"errors"
	"net/url"
	"strings"
)

// ValidateGroupName reports whether group is usable both as an atlantis.yaml
// `group` value and as a `-g` / API group selector.
//
// Config and selectors must agree exactly: a group that parses in the repo
// config but is rejected by every selector is unreachable config, with no error
// raised anywhere. The rule is therefore applied to the raw value, unlike
// project names, which are checked with '/' replaced -- a project name is never
// matched against a selector verbatim, a group always is.
//
// The returned error carries only the reason, so that callers can add the
// context that suits them: the repo-config validator renders it under the
// "group" yaml key, and the comment parser prefixes it with the offending
// value.
func ValidateGroupName(group string) error {
	if group == "" {
		return errors.New("cannot be empty")
	}
	// Redundant against the URL-safe check below, which already rejects these,
	// but it gives the likeliest mistake a message that names the cause.
	if strings.ContainsAny(group, "*?[") {
		return errors.New("cannot contain glob pattern characters ('*', '?', '[')")
	}
	if group != url.QueryEscape(group) {
		return errors.New("must contain only URL safe characters")
	}
	return nil
}
