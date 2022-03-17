package vcs

import (
	"fmt"
	"strconv"

	"github.com/google/go-github/v31/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally"
)

// NewInstrumentedGithubClient creates a client proxy responsible for gathering stats and logging
func NewInstrumentedGithubClient(client *GithubClient, statsScope tally.Scope, logger logging.SimpleLogging) IGithubClient {
	scope := statsScope.SubScope("github")

	instrumentedGHClient := &InstrumentedClient{
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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_github_pull_request_getter.go GithubPullRequestGetter

type GithubPullRequestGetter interface {
	GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error)
}

// IGithubClient exists to bridge the gap between GithubPullRequestGetter and Client interface to allow
// for a single instrumented client
type IGithubClient interface {
	Client
	GithubPullRequestGetter
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement soley any github specific interfaces.
type InstrumentedGithubClient struct {
	*InstrumentedClient
	PullRequestGetter GithubPullRequestGetter
	StatsScope        tally.Scope
	Logger            logging.SimpleLogging
}

func (c *InstrumentedGithubClient) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.SubScope("get_pull_request")
	logger := c.Logger.WithHistory([]interface{}{
		"repository", fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
		"pull-num", strconv.Itoa(pullNum),
	}...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	pull, err := c.PullRequestGetter.GetPullRequest(repo, pullNum)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get pull number for repo, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return pull, err

}

type InstrumentedClient struct {
	Client
	StatsScope tally.Scope
	Logger     logging.SimpleLogging
}

func (c *InstrumentedClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := c.StatsScope.SubScope("get_modified_files")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	files, err := c.Client.GetModifiedFiles(repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get modified files, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return files, err

}
func (c *InstrumentedClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	scope := c.StatsScope.SubScope("create_comment")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pullNum)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.CreateComment(repo, pullNum, comment, command); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to create comment for command %s, error: %s", command, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil
}
func (c *InstrumentedClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	scope := c.StatsScope.SubScope("hide_prev_plan_comments")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pullNum)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.HidePrevCommandComments(repo, pullNum, command); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to hide previous %s comments, error: %s", command, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil

}
func (c *InstrumentedClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	scope := c.StatsScope.SubScope("pull_is_approved")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	approved, err := c.Client.PullIsApproved(repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to check pull approval status, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return approved, err

}
func (c *InstrumentedClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	scope := c.StatsScope.SubScope("pull_is_mergeable")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	mergeable, err := c.Client.PullIsMergeable(repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to check pull mergeable status, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return mergeable, err
}

func (c *InstrumentedClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	scope := c.StatsScope.SubScope("update_status")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.UpdateStatus(repo, pull, state, src, description, url); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to update status at url: %s, error: %s", url, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil

}
func (c *InstrumentedClient) MergePull(pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	scope := c.StatsScope.SubScope("merge_pull")
	logger := c.Logger.WithHistory("pull-num", pull.Num)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.MergePull(pull, pullOptions); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to merge pull, error: %s", err.Error())
	}

	executionSuccess.Inc(1)
	return nil

}

// taken from other parts of the code, would be great to have this in a shared spot
func fmtLogSrc(repo models.Repo, pullNum int) []interface{} {
	return []interface{}{
		"repository", repo.FullName,
		"pull-num", strconv.Itoa(pullNum),
	}
}
