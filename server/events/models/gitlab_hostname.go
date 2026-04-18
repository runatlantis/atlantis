// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
	return u.Host, strings.TrimSuffix(u.Path, "/"), nil
}
