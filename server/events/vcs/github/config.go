// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import "github.com/runatlantis/atlantis/server/events/vcs/common"

// GithubConfig allows for custom github-specific functionality and behavior
type Config struct {
	AllowMergeableBypassApply bool
	CommentNamespace          common.CommentNamespace
}
