// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events/vcs --package mocks -o mocks/mock_pull_req_status_fetcher.go PullReqStatusFetcher

type PullReqStatusFetcher interface {
	FetchPullStatus(logger logging.SimpleLogging, pull models.PullRequest) (models.PullReqStatus, error)
}

type pullReqStatusFetcher struct {
	client               Client
	vcsStatusName        string
	ignoreVCSStatusNames []string
}

func NewPullReqStatusFetcher(client Client, vcsStatusName string, ignoreVCSStatusNames []string) PullReqStatusFetcher {
	return &pullReqStatusFetcher{
		client:               client,
		vcsStatusName:        vcsStatusName,
		ignoreVCSStatusNames: ignoreVCSStatusNames,
	}
}

func (f *pullReqStatusFetcher) FetchPullStatus(logger logging.SimpleLogging, pull models.PullRequest) (pullStatus models.PullReqStatus, err error) {
	approvalStatus, err := f.client.PullIsApproved(logger, pull.BaseRepo, pull)
	if err != nil {
		return pullStatus, fmt.Errorf("fetching pull approval status for repo: %s, and pull number: %d: %w", pull.BaseRepo.FullName, pull.Num, err)
	}

	mergeable, err := f.client.PullIsMergeable(logger, pull.BaseRepo, pull, f.vcsStatusName, f.ignoreVCSStatusNames)
	if err != nil {
		return pullStatus, fmt.Errorf("fetching mergeability status for repo: %s, and pull number: %d: %w", pull.BaseRepo.FullName, pull.Num, err)
	}

	return models.PullReqStatus{
		ApprovalStatus:  approvalStatus,
		MergeableStatus: mergeable,
	}, err
}
