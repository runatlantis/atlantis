// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import "time"

// GithubConfig allows for custom github-specific functionality and behavior
type Config struct {
	AllowMergeableBypassApply bool
	CommentInterval           time.Duration
}
