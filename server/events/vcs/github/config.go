// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import "time"

const (
	defaultAsyncMergePollInterval = 3 * time.Second
	defaultAsyncMergeTimeout      = 5 * time.Minute
)

// GithubConfig allows for custom github-specific functionality and behavior
type Config struct {
	AllowMergeableBypassApply bool
	// AsyncMergePollInterval is how often the result of an asynchronous merge
	// (used for stacked pull requests) is polled. Defaults to 3 seconds.
	AsyncMergePollInterval time.Duration
	// AsyncMergeTimeout is how long to wait for an asynchronous merge to
	// complete. Defaults to 5 minutes.
	AsyncMergeTimeout time.Duration
}

func (c Config) asyncMergePollInterval() time.Duration {
	if c.AsyncMergePollInterval > 0 {
		return c.AsyncMergePollInterval
	}
	return defaultAsyncMergePollInterval
}

func (c Config) asyncMergeTimeout() time.Duration {
	if c.AsyncMergeTimeout > 0 {
		return c.AsyncMergeTimeout
	}
	return defaultAsyncMergeTimeout
}
