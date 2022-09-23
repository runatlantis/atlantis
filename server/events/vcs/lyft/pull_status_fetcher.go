package lyft

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

const LockValue = "lock"

type PullClient interface {
	GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error)
	GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error)
	GetRepoChecks(repo models.Repo, commitSHA string) ([]*github.CheckRun, error)
	PullIsApproved(repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error)
}

type SQBasedPullStatusFetcher struct {
	client  PullClient
	checker MergeabilityChecker
}

func NewSQBasedPullStatusFetcher(client PullClient, checker MergeabilityChecker) *SQBasedPullStatusFetcher {
	return &SQBasedPullStatusFetcher{
		client:  client,
		checker: checker,
	}
}

func (s *SQBasedPullStatusFetcher) FetchPullStatus(repo models.Repo, pull models.PullRequest) (models.PullReqStatus, error) {
	pullStatus := models.PullReqStatus{}

	approvalStatus, err := s.client.PullIsApproved(repo, pull)
	if err != nil {
		return models.PullReqStatus{}, errors.Wrap(err, "fetching pull approval status")
	}

	githubPR, err := s.client.GetPullRequest(repo, pull.Num)
	if err != nil {
		return pullStatus, errors.Wrap(err, "fetching pull request")
	}

	statuses, err := s.client.GetRepoStatuses(repo, pull)
	if err != nil {
		return pullStatus, errors.Wrap(err, "fetching repo statuses")
	}

	checks, err := s.client.GetRepoChecks(repo, pull.HeadCommit)
	if err != nil {
		return pullStatus, errors.Wrap(err, "fetching repo checks")
	}

	mergeable := s.checker.Check(githubPR, statuses, checks)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "checking mergeability")
	}

	sqLocked, err := s.isPRLocked(statuses, checks)
	if err != nil {
		return pullStatus, errors.Wrapf(err, "checking sq lock status")
	}

	return models.PullReqStatus{
		ApprovalStatus: approvalStatus,
		Mergeable:      mergeable,
		SQLocked:       sqLocked,
	}, nil
}

func (s SQBasedPullStatusFetcher) isPRLocked(statuses []*github.RepoStatus, checks []*github.CheckRun) (bool, error) {
	rawMetadata := ""

	// First check statuses
	for _, status := range statuses {
		if status.GetContext() == SubmitQueueReadinessContext {
			rawMetadata = status.GetDescription()
			break
		}
	}
	// Next try check runs if no statuses
	if len(rawMetadata) == 0 {
		for _, check := range checks {
			if check.GetName() == SubmitQueueReadinessContext {
				output := check.GetOutput()
				if output != nil {
					rawMetadata = output.GetTitle()
				}
			}
			break
		}
	}

	// No metadata found, assume not locked
	if len(rawMetadata) == 0 {
		return false, nil
	}

	// Not using struct tags because there's no predefined schema for description.
	description := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawMetadata), &description)
	if err != nil {
		return false, errors.Wrapf(err, "parsing status description")
	}

	waitingList, ok := description["waiting"]
	if !ok {
		// No waiting key means no lock.
		return false, nil
	}

	typedWaitingList, ok := waitingList.([]interface{})
	if !ok {
		return false, fmt.Errorf("cast failed for %v", waitingList)
	}
	for _, item := range typedWaitingList {
		if item == LockValue {
			return true, nil
		}
	}

	// No Lock found.
	return false, nil
}
