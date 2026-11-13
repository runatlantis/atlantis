// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package models

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseGitlabHostname splits a user-supplied GitLab hostname into its host and
// base path components. Accepted inputs include bare hosts ("gitlab.com"),
// hosts with a subpath ("acme.com/gitlab"), and fully-qualified URLs
// ("https://acme.com:8080/gitlab/").
//
// The returned host includes any port. The returned basePath is either empty
// or starts with "/" and has no trailing slash.
func ParseGitlabHostname(raw string) (host, basePath string, err error) {
	if raw == "" {
		return "", "", fmt.Errorf("gitlab hostname is empty")
	}
	absoluteURL := raw
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		absoluteURL = "https://" + absoluteURL
	}
	u, parseErr := url.Parse(absoluteURL)
	if parseErr != nil {
		return "", "", fmt.Errorf("parsing gitlab hostname %q: %w", raw, parseErr)
	}
	if u.Host == "" {
		return "", "", fmt.Errorf("gitlab hostname %q has no host component", raw)
	}
	return u.Host, strings.TrimRight(u.Path, "/"), nil
}
