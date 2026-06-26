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
	GetAtlantisStatus(ctx context.Context, branchName string) (string, error)
	ClosePullRequest(ctx context.Context, pullRequestNumber int) error
	DeleteBranch(ctx context.Context, branchName string) error
	IsAtlantisInProgress(state string) bool
	DidAtlantisSucceed(state string) bool
}
