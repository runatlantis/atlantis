// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestFetchPullStatus_Success(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	mockClient := vcsmocks.NewMockClient()
	pull := models.PullRequest{
		Num:      1,
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	wantApproval := models.ApprovalStatus{IsApproved: true, ApprovedBy: "alice"}
	wantMergeable := models.MergeableStatus{IsMergeable: true}

	When(mockClient.PullIsApproved(logger, pull.BaseRepo, pull)).ThenReturn(wantApproval, nil)
	When(mockClient.PullIsMergeable(logger, pull.BaseRepo, pull, "atlantis", []string{"skip-me"})).ThenReturn(wantMergeable, nil)

	fetcher := vcs.NewPullReqStatusFetcher(mockClient, "atlantis", []string{"skip-me"})
	got, err := fetcher.FetchPullStatus(logger, pull)

	Ok(t, err)
	Equals(t, wantApproval, got.ApprovalStatus)
	Equals(t, wantMergeable, got.MergeableStatus)
}

func TestFetchPullStatus_ApprovalError(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	mockClient := vcsmocks.NewMockClient()
	pull := models.PullRequest{
		Num:      2,
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	wantErr := errors.New("approval API failed")

	When(mockClient.PullIsApproved(logger, pull.BaseRepo, pull)).ThenReturn(models.ApprovalStatus{}, wantErr)

	fetcher := vcs.NewPullReqStatusFetcher(mockClient, "atlantis", nil)
	_, err := fetcher.FetchPullStatus(logger, pull)

	Assert(t, err != nil, "expected error")
	// The error message should mention the repo name and pull number.
	ErrContains(t, "owner/repo", err)
	ErrContains(t, "2", err)
}

func TestFetchPullStatus_MergeableError(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	mockClient := vcsmocks.NewMockClient()
	pull := models.PullRequest{
		Num:      3,
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	wantApproval := models.ApprovalStatus{IsApproved: false}
	wantErr := errors.New("mergeable API failed")

	When(mockClient.PullIsApproved(logger, pull.BaseRepo, pull)).ThenReturn(wantApproval, nil)
	When(mockClient.PullIsMergeable(logger, pull.BaseRepo, pull, "atlantis", []string(nil))).ThenReturn(models.MergeableStatus{}, wantErr)

	fetcher := vcs.NewPullReqStatusFetcher(mockClient, "atlantis", nil)
	_, err := fetcher.FetchPullStatus(logger, pull)

	Assert(t, err != nil, "expected error")
	// The error message should mention the repo name and pull number.
	ErrContains(t, "owner/repo", err)
	ErrContains(t, "3", err)
}

func TestFetchPullStatus_PassesVCSStatusNameAndIgnoreList(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	mockClient := vcsmocks.NewMockClient()
	pull := models.PullRequest{
		Num:      4,
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	ignoreList := []string{"ci/skip", "bot-check"}

	When(mockClient.PullIsApproved(logger, pull.BaseRepo, pull)).ThenReturn(models.ApprovalStatus{}, nil)
	When(mockClient.PullIsMergeable(logger, pull.BaseRepo, pull, "my-status", ignoreList)).ThenReturn(models.MergeableStatus{IsMergeable: true}, nil)

	fetcher := vcs.NewPullReqStatusFetcher(mockClient, "my-status", ignoreList)
	got, err := fetcher.FetchPullStatus(logger, pull)

	Ok(t, err)
	Equals(t, true, got.MergeableStatus.IsMergeable)

	// Verify the correct args were passed to PullIsMergeable.
	mockClient.VerifyWasCalledOnce().PullIsMergeable(logger, pull.BaseRepo, pull, "my-status", ignoreList)
}
