package vcs

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

type PullReqStatusFetcher interface {
	FetchPullStatus(repo models.Repo, pull models.PullRequest) (models.PullReqStatus, error)
}

type pullReqStatusFetcher struct {
	client Client
}

func NewPullReqStatusFetcher(client Client) PullReqStatusFetcher {
	return &pullReqStatusFetcher{
		client: client,
	}
}

func (f *pullReqStatusFetcher) FetchPullStatus(repo models.Repo, pull models.PullRequest) (pullStatus models.PullReqStatus, err error) {
	approvalStatus, err := f.client.PullIsApproved(repo, pull)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching pull approval status for repo: %s, and pull number: %d", repo.FullName, pull.Num)
	}

	mergeable, err := f.client.PullIsMergeable(repo, pull)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching mergeability status for repo: %s, and pull number: %d", repo.FullName, pull.Num)
	}

	return models.PullReqStatus{
		ApprovalStatus: approvalStatus,
		Mergeable:      mergeable,
	}, err
}
