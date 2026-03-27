// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
)

// newInstrumentedClient is a test helper that wires up an InstrumentedClient
// with a mock backing client and a real (but silent) metrics scope.
func newInstrumentedClient(t *testing.T, inner *vcsmocks.MockClient) *common.InstrumentedClient {
	t.Helper()
	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	return &common.InstrumentedClient{
		Client:     inner,
		StatsScope: scope,
		Logger:     logger,
	}
}

// makeTestRepo returns a minimal Repo with the given full name.
func makeTestRepo(fullName string) models.Repo {
	return models.Repo{FullName: fullName}
}

// makeTestPull returns a minimal PullRequest with the given Num.
func makeTestPull(num int) models.PullRequest {
	return models.PullRequest{Num: num, BaseRepo: makeTestRepo("owner/repo")}
}

// ---------------------------------------------------------------------------
// GetModifiedFiles
// ---------------------------------------------------------------------------

func TestInstrumentedClient_GetModifiedFiles_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantFiles := []string{"main.tf", "variables.tf"}

	When(inner.GetModifiedFiles(logger, repo, pull)).ThenReturn(wantFiles, nil)

	files, err := client.GetModifiedFiles(logger, repo, pull)
	Ok(t, err)
	Equals(t, wantFiles, files)
	inner.VerifyWasCalledOnce().GetModifiedFiles(logger, repo, pull)
}

func TestInstrumentedClient_GetModifiedFiles_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantErr := errors.New("api failure")

	When(inner.GetModifiedFiles(logger, repo, pull)).ThenReturn(nil, wantErr)

	_, err := client.GetModifiedFiles(logger, repo, pull)
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// CreateComment
// ---------------------------------------------------------------------------

func TestInstrumentedClient_CreateComment_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	When(inner.CreateComment(logger, repo, 1, "msg", "plan")).ThenReturn(nil)

	err := client.CreateComment(logger, repo, 1, "msg", "plan")
	Ok(t, err)
	inner.VerifyWasCalledOnce().CreateComment(logger, repo, 1, "msg", "plan")
}

func TestInstrumentedClient_CreateComment_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	wantErr := errors.New("create comment failed")
	When(inner.CreateComment(logger, repo, 1, "msg", "plan")).ThenReturn(wantErr)

	err := client.CreateComment(logger, repo, 1, "msg", "plan")
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// ReactToComment
// ---------------------------------------------------------------------------

func TestInstrumentedClient_ReactToComment_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	When(inner.ReactToComment(logger, repo, 1, int64(42), "eyes")).ThenReturn(nil)

	err := client.ReactToComment(logger, repo, 1, int64(42), "eyes")
	Ok(t, err)
	inner.VerifyWasCalledOnce().ReactToComment(logger, repo, 1, int64(42), "eyes")
}

func TestInstrumentedClient_ReactToComment_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	wantErr := errors.New("react failed")
	When(inner.ReactToComment(logger, repo, 1, int64(42), "eyes")).ThenReturn(wantErr)

	err := client.ReactToComment(logger, repo, 1, int64(42), "eyes")
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// HidePrevCommandComments
// ---------------------------------------------------------------------------

func TestInstrumentedClient_HidePrevCommandComments_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	When(inner.HidePrevCommandComments(logger, repo, 1, "plan", "")).ThenReturn(nil)

	err := client.HidePrevCommandComments(logger, repo, 1, "plan", "")
	Ok(t, err)
	inner.VerifyWasCalledOnce().HidePrevCommandComments(logger, repo, 1, "plan", "")
}

func TestInstrumentedClient_HidePrevCommandComments_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	wantErr := errors.New("hide failed")
	When(inner.HidePrevCommandComments(logger, repo, 1, "plan", "")).ThenReturn(wantErr)

	err := client.HidePrevCommandComments(logger, repo, 1, "plan", "")
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// PullIsApproved
// ---------------------------------------------------------------------------

func TestInstrumentedClient_PullIsApproved_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantStatus := models.ApprovalStatus{IsApproved: true, ApprovedBy: "alice"}

	When(inner.PullIsApproved(logger, repo, pull)).ThenReturn(wantStatus, nil)

	got, err := client.PullIsApproved(logger, repo, pull)
	Ok(t, err)
	Equals(t, wantStatus, got)
}

func TestInstrumentedClient_PullIsApproved_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantErr := errors.New("approval check failed")

	When(inner.PullIsApproved(logger, repo, pull)).ThenReturn(models.ApprovalStatus{}, wantErr)

	_, err := client.PullIsApproved(logger, repo, pull)
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// PullIsMergeable
// ---------------------------------------------------------------------------

func TestInstrumentedClient_PullIsMergeable_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantStatus := models.MergeableStatus{IsMergeable: true}

	When(inner.PullIsMergeable(logger, repo, pull, "atlantis", []string{})).ThenReturn(wantStatus, nil)

	got, err := client.PullIsMergeable(logger, repo, pull, "atlantis", []string{})
	Ok(t, err)
	Equals(t, wantStatus, got)
}

func TestInstrumentedClient_PullIsMergeable_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantErr := errors.New("mergeable check failed")

	When(inner.PullIsMergeable(logger, repo, pull, "atlantis", []string{})).ThenReturn(models.MergeableStatus{}, wantErr)

	_, err := client.PullIsMergeable(logger, repo, pull, "atlantis", []string{})
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// UpdateStatus
// ---------------------------------------------------------------------------

func TestInstrumentedClient_UpdateStatus_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	When(inner.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")).ThenReturn(nil)

	err := client.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
	Ok(t, err)
	inner.VerifyWasCalledOnce().UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
}

func TestInstrumentedClient_UpdateStatus_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(1)
	wantErr := errors.New("status update failed")
	When(inner.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")).ThenReturn(wantErr)

	err := client.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
	Equals(t, wantErr, err)
}

// TestInstrumentedClient_UpdateStatus_NoPR verifies that UpdateStatus is a
// no-op when the pull request number is 0 (i.e. plan triggered outside a PR).
func TestInstrumentedClient_UpdateStatus_NoPR(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	repo := makeTestRepo("owner/repo")
	pull := makeTestPull(0) // no PR number

	// The underlying client must NOT be called.
	err := client.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
	Ok(t, err)
	inner.VerifyWasCalled(Never()).UpdateStatus(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[models.CommitStatus](), Any[string](), Any[string](), Any[string](),
	)
}

// ---------------------------------------------------------------------------
// MergePull
// ---------------------------------------------------------------------------

func TestInstrumentedClient_MergePull_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	pull := makeTestPull(1)
	pull.BaseRepo = makeTestRepo("owner/repo")
	opts := models.PullRequestOptions{DeleteSourceBranchOnMerge: true}
	When(inner.MergePull(logger, pull, opts)).ThenReturn(nil)

	err := client.MergePull(logger, pull, opts)
	Ok(t, err)
	inner.VerifyWasCalledOnce().MergePull(logger, pull, opts)
}

func TestInstrumentedClient_MergePull_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	inner := vcsmocks.NewMockClient()
	client := newInstrumentedClient(t, inner)

	pull := makeTestPull(1)
	pull.BaseRepo = makeTestRepo("owner/repo")
	opts := models.PullRequestOptions{}
	wantErr := errors.New("merge failed")
	When(inner.MergePull(logger, pull, opts)).ThenReturn(wantErr)

	err := client.MergePull(logger, pull, opts)
	Equals(t, wantErr, err)
}

// ---------------------------------------------------------------------------
// SetGitScopeTags
// ---------------------------------------------------------------------------

func TestSetGitScopeTags_ReturnsTaggedScope(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	// Calling SetGitScopeTags must not panic and must return a non-nil scope.
	tagged := common.SetGitScopeTags(scope, "owner/repo", 42)
	Assert(t, tagged != nil, "expected non-nil tagged scope")
}
