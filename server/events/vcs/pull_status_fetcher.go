package vcs

import (
	"time"

	"github.com/pkg/errors"
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

	statusRetryTimes   int
	statusRetryBackoff time.Duration
}

type ReqStatusFetcherFuncOpt func(*pullReqStatusFetcher)

func WithRetry(times int) ReqStatusFetcherFuncOpt {
	return func(prsf *pullReqStatusFetcher) {
		prsf.statusRetryTimes = times
	}
}

func WithBackoff(backoff time.Duration) ReqStatusFetcherFuncOpt {
	return func(prsf *pullReqStatusFetcher) {
		prsf.statusRetryBackoff = backoff
	}
}

func NewPullReqStatusFetcher(client Client, vcsStatusName string, ignoreVCSStatusNames []string, opts ...ReqStatusFetcherFuncOpt) PullReqStatusFetcher {
	fetcher := &pullReqStatusFetcher{
		client:               client,
		vcsStatusName:        vcsStatusName,
		ignoreVCSStatusNames: ignoreVCSStatusNames,
	}
	for _, opt := range opts {
		opt(fetcher)
	}

	return fetcher
}

func (f *pullReqStatusFetcher) FetchPullStatus(logger logging.SimpleLogging, pull models.PullRequest) (pullStatus models.PullReqStatus, err error) {
	approvalStatus, err := f.client.PullIsApproved(logger, pull.BaseRepo, pull)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching pull approval status for repo: %s, and pull number: %d", pull.BaseRepo.FullName, pull.Num)
	}

	mergeable, err := f.client.PullIsMergeable(logger, pull.BaseRepo, pull, f.vcsStatusName, f.ignoreVCSStatusNames)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching mergeability status for repo: %s, and pull number: %d", pull.BaseRepo.FullName, pull.Num)
	}
	if mergeable {
		return models.PullReqStatus{
			ApprovalStatus: approvalStatus,
			Mergeable:      mergeable,
		}, nil
	}

	retries := f.statusRetryTimes
	ticker := time.NewTicker(f.statusRetryBackoff)
	defer ticker.Stop()
	for retries != 0 {
		<-ticker.C
		mergeable, err = f.client.PullIsMergeable(logger, pull.BaseRepo, pull, f.vcsStatusName, f.ignoreVCSStatusNames)
		if err != nil {
			return pullStatus, errors.Wrapf(err, "fetching mergeability status for repo: %s, and pull number: %d", pull.BaseRepo.FullName, pull.Num)
		}
		if mergeable {
			break
		}
		retries--
	}

	return models.PullReqStatus{
		ApprovalStatus: approvalStatus,
		Mergeable:      mergeable,
	}, err
}
