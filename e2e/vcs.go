// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import (
	"context"
	"time"
)

type CommitStatus struct {
	State     string
	ID        int64
	UpdatedAt time.Time
}

type VCSClient interface {
	Clone(cloneDir string) error
	CreateAtlantisWebhook(ctx context.Context, hookURL string) (int64, error)
	DeleteAtlantisHook(ctx context.Context, hookID int64) error
	CreatePullRequest(ctx context.Context, title, branchName string) (string, int, error)
	// PostPRComment posts a PR/MR comment that Atlantis treats as a command.
	PostPRComment(ctx context.Context, pullNumber int, body string) error
	// GetAtlantisStatus polls for plan completion. Returns aggregate terminal
	// state across all matching statuses. Used for the polling loop only.
	GetAtlantisStatus(ctx context.Context, branchName string) (string, error)
	// GetAtlantisCommandStatus polls an aggregate Atlantis command status such
	// as "atlantis/plan" or "atlantis/apply".
	GetAtlantisCommandStatus(ctx context.Context, branchName string, command string) (string, error)
	// GetCommitStatus returns a specific commit status context with metadata
	// used to distinguish repeated command runs on the same SHA.
	GetCommitStatus(ctx context.Context, branchName, statusContext string) (CommitStatus, error)
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
