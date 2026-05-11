// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package github

import (
	"testing"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// If the hostname is github.com, should use normal BaseURL.
func TestNew_GithubCom(t *testing.T) {
	client, err := New("github.com", &UserCredentials{"user", "pass", ""}, Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	Equals(t, "https://api.github.com/", client.client.BaseURL.String())
}

// If the hostname is a non-github hostname should use the right BaseURL.
func TestNew_NonGithub(t *testing.T) {
	client, err := New("example.com", &UserCredentials{"user", "pass", ""}, Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	Equals(t, "https://example.com/api/v3/", client.client.BaseURL.String())
	// If possible in the future, test the GraphQL client's URL as well. But at the
	// moment the shurcooL library doesn't expose it.
}
