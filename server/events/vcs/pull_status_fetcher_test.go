// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestFetchPullStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	repo := models.Repo{FullName: "owner/repo"}
	pull := models.PullRequest{Num: 1, BaseRepo: repo}
	approved := models.ApprovalStatus{
		IsApproved: true,
		ApprovedBy: "reviewer",
		Date:       time.Date(2026, time.November, 24, 12, 0, 0, 0, time.UTC),
	}
	mergeable := models.MergeableStatus{IsMergeable: true}
	approvalErr := errors.New("approval lookup failed")
	mergeableErr := errors.New("mergeability lookup failed")

	cases := []struct {
		description      string
		approval         models.ApprovalStatus
		approvalErr      error
		mergeable        models.MergeableStatus
		mergeableErr     error
		expStatus        models.PullReqStatus
		expErr           error
		expErrContains   string
		expMergeableCall bool
	}{
		{
			description:      "approved and mergeable",
			approval:         approved,
			mergeable:        mergeable,
			expStatus:        models.PullReqStatus{ApprovalStatus: approved, MergeableStatus: mergeable},
			expMergeableCall: true,
		},
		{
			description:    "approval lookup fails",
			approvalErr:    approvalErr,
			expErr:         approvalErr,
			expErrContains: "fetching pull approval status for repo: owner/repo, and pull number: 1",
		},
		{
			// The pull request is approved, so a failing mergeability check
			// must not make it look unapproved.
			description:      "mergeability lookup fails on an approved pull request",
			approval:         approved,
			mergeableErr:     mergeableErr,
			expStatus:        models.PullReqStatus{ApprovalStatus: approved},
			expErr:           mergeableErr,
			expErrContains:   "fetching mergeability status for repo: owner/repo, and pull number: 1",
			expMergeableCall: true,
		},
		{
			description:      "mergeability lookup fails on an unapproved pull request",
			mergeableErr:     mergeableErr,
			expErr:           mergeableErr,
			expErrContains:   "fetching mergeability status for repo: owner/repo, and pull number: 1",
			expMergeableCall: true,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := vcsmocks.NewMockClient()
			When(client.PullIsApproved(Any[logging.SimpleLogging](), Eq(repo), Eq(pull))).
				ThenReturn(c.approval, c.approvalErr)
			When(client.PullIsMergeable(Any[logging.SimpleLogging](), Eq(repo), Eq(pull), Eq("atlantis"), Eq([]string{"ignored"}))).
				ThenReturn(c.mergeable, c.mergeableErr)

			fetcher := vcs.NewPullReqStatusFetcher(client, "atlantis", []string{"ignored"})
			status, err := fetcher.FetchPullStatus(logger, pull)

			Equals(t, c.expStatus, status)
			if c.expErr == nil {
				Ok(t, err)
			} else {
				Assert(t, errors.Is(err, c.expErr), "expected error to wrap %q, got %v", c.expErr, err)
				ErrContains(t, c.expErrContains, err)
			}

			if c.expMergeableCall {
				client.VerifyWasCalledOnce().PullIsMergeable(Any[logging.SimpleLogging](), Eq(repo), Eq(pull), Eq("atlantis"), Eq([]string{"ignored"}))
			} else {
				client.VerifyWasCalled(Never()).PullIsMergeable(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[[]string]())
			}
		})
	}
}
