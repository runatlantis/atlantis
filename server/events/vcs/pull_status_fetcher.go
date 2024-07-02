package vcs

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type PullReqStatusFetcher interface {
	FetchPullStatus(logger logging.SimpleLogging, pull models.PullRequest) (models.PullReqStatus, error)
}

type pullReqStatusFetcher struct {
	client        Client
	vcsStatusName string
}

func NewPullReqStatusFetcher(client Client, vcsStatusName string) PullReqStatusFetcher {
	return &pullReqStatusFetcher{
		client:        client,
		vcsStatusName: vcsStatusName,
	}
}

func (f *pullReqStatusFetcher) FetchPullStatus(logger logging.SimpleLogging, pull models.PullRequest) (pullStatus models.PullReqStatus, err error) {
	approvalStatus, err := f.client.PullIsApproved(logger, pull.BaseRepo, pull)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching pull approval status for repo: %s, and pull number: %d", pull.BaseRepo.FullName, pull.Num)
	}

	mergeable, err := f.client.PullIsMergeable(logger, pull.BaseRepo, pull, f.vcsStatusName)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching mergeability status for repo: %s, and pull number: %d", pull.BaseRepo.FullName, pull.Num)
	}

	return models.PullReqStatus{
		ApprovalStatus: approvalStatus,
		Mergeable:      mergeable,
	}, err
}
