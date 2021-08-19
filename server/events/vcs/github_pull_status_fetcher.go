package vcs

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

type PullReqStatusFetcher interface {
	FetchPullStatus(repo models.Repo, pull models.PullRequest) (models.PullReqStatus, error)
}

type SQBasedPullStatusFetcher struct {
	GithubClient IGithubClient
}

func (s SQBasedPullStatusFetcher) FetchPullStatus(repo models.Repo, pull models.PullRequest) (pullStatus models.PullReqStatus, err error) {
	statuses, err := s.GithubClient.GetRepoStatuses(repo, pull)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching repo statuses for repo: %s, and pull number: %d", repo.FullName, pull.Num)
	}

	approved, err := s.GithubClient.PullIsApproved(repo, pull)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching pull approval status for repo: %s, and pull number: %d", repo.FullName, pull.Num)
	}

	sqLocked, err := s.GithubClient.PullIsSQLocked(repo, pull, statuses)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching pull locked status for repo: %s, and pull number: %d", repo.FullName, pull.Num)
	}

	mergeable, err := s.GithubClient.PullIsSQMergeable(repo, pull, statuses)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "fetching mergeability status for repo: %s, and pull number: %d", repo.FullName, pull.Num)
	}

	return models.PullReqStatus{
		Approved:  approved,
		Mergeable: mergeable,
		SQLocked:  sqLocked,
	}, err
}
