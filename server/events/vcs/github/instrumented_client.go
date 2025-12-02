package github

import (
	"github.com/google/go-github/v71/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally/v4"
)

// NewInstrumentedGithubClient creates a client proxy responsible for gathering stats and logging
func NewInstrumentedGithubClient(client *GithubClient, statsScope tally.Scope, logger logging.SimpleLogging) IGithubClient {
	scope := statsScope.SubScope("github")

	instrumentedGHClient := &common.InstrumentedClient{
		Client:     client,
		StatsScope: scope,
		Logger:     logger,
	}

	return &InstrumentedGithubClient{
		InstrumentedClient: instrumentedGHClient,
		PullRequestGetter:  client,
		StatsScope:         scope,
		Logger:             logger,
	}
}

//go:generate pegomock generate --package mocks -o mocks/mock_github_pull_request_getter.go GithubPullRequestGetter

type GithubPullRequestGetter interface {
	GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*github.PullRequest, error)
}

// IGithubClient exists to bridge the gap between GithubPullRequestGetter and Client interface to allow
// for a single instrumented client
type IGithubClient interface {
	vcs.Client
	GithubPullRequestGetter
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement solely any github specific interfaces.
type InstrumentedGithubClient struct {
	*common.InstrumentedClient
	PullRequestGetter GithubPullRequestGetter
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
