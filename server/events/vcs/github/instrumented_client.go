// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/google/go-github/v92/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally/v4"
)

// NewInstrumentedGithubClient creates a client proxy responsible for gathering stats and logging
func NewInstrumentedGithubClient(client *Client, statsScope tally.Scope, logger logging.SimpleLogging) IGithubClient {
	scope := statsScope.SubScope("github")

	instrumentedGHClient := &common.InstrumentedClient{
		Client:     client,
		StatsScope: scope,
		Logger:     logger,
	}

	return &InstrumentedGithubClient{
		InstrumentedClient: instrumentedGHClient,
		PullRequestGetter:  client,
		StackGetter:        client,
		StatsScope:         scope,
		Logger:             logger,
	}
}

//go:generate go tool pegomock generate --package mocks -o mocks/mock_github_pull_request_getter.go GithubPullRequestGetter

type GithubPullRequestGetter interface {
	GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*github.PullRequest, error)
}

// GithubPullRequestStackGetter fetches the stack a pull request belongs to.
type GithubPullRequestStackGetter interface {
	// GetPullRequestStack returns the pull requests of the stack that pull
	// request pullNum belongs to, ordered from the bottom of the stack to the
	// top. It returns nil if the pull request is not part of a stack.
	GetPullRequestStack(logger logging.SimpleLogging, repo models.Repo, pullNum int) ([]models.StackedPull, error)
}

// IGithubClient exists to bridge the gap between GithubPullRequestGetter and Client interface to allow
// for a single instrumented client
type IGithubClient interface {
	vcs.Client
	GithubPullRequestGetter
	GithubPullRequestStackGetter
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement solely any github specific interfaces.
type InstrumentedGithubClient struct {
	*common.InstrumentedClient
	PullRequestGetter GithubPullRequestGetter
	StackGetter       GithubPullRequestStackGetter
	StatsScope        tally.Scope
	Logger            logging.SimpleLogging
}

func (c *InstrumentedGithubClient) GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.SubScope("get_pull_request")
	scope = common.SetGitScopeTags(scope, repo.FullName, pullNum)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	pull, err := c.PullRequestGetter.GetPullRequest(logger, repo, pullNum)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get pull number for repo, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return pull, err

}

func (c *InstrumentedGithubClient) GetPullRequestStack(logger logging.SimpleLogging, repo models.Repo, pullNum int) ([]models.StackedPull, error) {
	scope := c.StatsScope.SubScope("get_pull_request_stack")
	scope = common.SetGitScopeTags(scope, repo.FullName, pullNum)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	pulls, err := c.StackGetter.GetPullRequestStack(logger, repo, pullNum)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get pull request stack, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return pulls, err
}
