// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package web_templates

import (
	"io"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectJobsErrorTemplate(t *testing.T) {
	err := ProjectJobsErrorTemplate.Execute(io.Discard, ProjectJobsError{
		LayoutData: LayoutData{
			AtlantisVersion: "v0.0.0",
			CleanedBasePath: "/path",
		},
	})
	Ok(t, err)
}

func TestGithubAppSetupTemplate(t *testing.T) {
	err := GithubAppSetupTemplate.Execute(io.Discard, GithubSetupData{
		Target:          "target",
		Manifest:        "manifest",
		ID:              1,
		Key:             "key",
		WebhookSecret:   "webhook secret",
		URL:             "https://example.com",
		CleanedBasePath: "/path",
	})
	Ok(t, err)
}
