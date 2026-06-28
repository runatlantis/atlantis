// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import "context"

type VCSClient interface {
	Clone(cloneDir string) error
	CreateAtlantisWebhook(ctx context.Context, hookURL string) (int64, error)
	DeleteAtlantisHook(ctx context.Context, hookID int64) error
	CreatePullRequest(ctx context.Context, title, branchName string) (string, int, error)
	// GetAtlantisStatus polls for plan completion. Returns aggregate terminal
	// state across all matching statuses. Used for the polling loop only.
	GetAtlantisStatus(ctx context.Context, branchName string) (string, error)
	// GetProjectStatuses returns all per-project Atlantis status contexts and
	// their states. Contexts have the form "atlantis/plan: <project>".
	// Used for post-completion assertions. Returns nil on GitLab (unsupported).
	GetProjectStatuses(ctx context.Context, branchName string) (map[string]string, error)
	// GetPRComments returns all comment bodies on the PR/MR.
	GetPRComments(ctx context.Context, pullNumber int) ([]string, error)
	ClosePullRequest(ctx context.Context, pullRequestNumber int) error
	DeleteBranch(ctx context.Context, branchName string) error
	IsAtlantisInProgress(state string) bool
	DidAtlantisSucceed(state string) bool
	DidAtlantisFail(state string) bool
}
