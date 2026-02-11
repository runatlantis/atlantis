// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strconv"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

type InstrumentedClient struct {
	vcs.Client
	StatsScope     tally.Scope
	PRScopeManager *metrics.PRScopeManager
	Logger         logging.SimpleLogging
}

func (c *InstrumentedClient) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pull.Num).SubScope("get_modified_files")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	files, err := c.Client.GetModifiedFiles(logger, repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get modified files, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return files, err
}

func (c *InstrumentedClient) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pullNum).SubScope("create_comment")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.CreateComment(logger, repo, pullNum, comment, command); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to create comment for command %s, error: %s", command, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil
}

func (c *InstrumentedClient) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pullNum).SubScope("react_to_comment")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.ReactToComment(logger, repo, pullNum, commentID, reaction); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to react to comment, error: %s", err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil
}

func (c *InstrumentedClient) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pullNum).SubScope("hide_prev_plan_comments")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.HidePrevCommandComments(logger, repo, pullNum, command, dir); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to hide previous %s comments, error: %s", command, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil

}

func (c *InstrumentedClient) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pull.Num).SubScope("pull_is_approved")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	approved, err := c.Client.PullIsApproved(logger, repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to check pull approval status, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return approved, err
}

func (c *InstrumentedClient) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (models.MergeableStatus, error) {
	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pull.Num).SubScope("pull_is_mergeable")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	mergeable, err := c.Client.PullIsMergeable(logger, repo, pull, vcsstatusname, ignoreVCSStatusNames)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to check pull mergeable status, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return mergeable, err
}

func (c *InstrumentedClient) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	// If the plan isn't coming from a pull request,
	// don't attempt to update the status.
	if pull.Num == 0 {
		return nil
	}

	scope := SetGitScopeTags(c.PRScopeManager, repo.FullName, pull.Num).SubScope("update_status")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.UpdateStatus(logger, repo, pull, state, src, description, url); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to update status at url: %s, error: %s", url, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil
}

func (c *InstrumentedClient) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	scope := SetGitScopeTags(c.PRScopeManager, pull.BaseRepo.FullName, pull.Num).SubScope("merge_pull")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.MergePull(logger, pull, pullOptions); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to merge pull, error: %s", err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil
}

// SetGitScopeTags sets git-level tags (repo and PR) on a scope using the PR scope manager.
// Creates a closeable PR-specific root scope with git-level tags.
func SetGitScopeTags(prScopeManager *metrics.PRScopeManager, repoFullName string, pullNum int) tally.Scope {
	tags := map[string]string{
		"base_repo": repoFullName,
		"pr_number": strconv.Itoa(pullNum),
	}

	return prScopeManager.GetOrCreatePRScope(repoFullName, pullNum, tags)
}
