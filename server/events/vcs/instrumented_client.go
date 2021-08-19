package vcs

import (
	"fmt"
	"strconv"

	"github.com/google/go-github/v31/github"
	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// NewInstrumentedGithubClient creates a client proxy responsible for gathering stats and logging
func NewInstrumentedGithubClient(client *GithubClient, statsScope stats.Scope, logger logging.SimpleLogging) IGithubClient {
	scope := statsScope.Scope("github")

	instrumentedGHClient := &InstrumentedClient{
		Client:     client,
		StatsScope: scope,
		Logger:     logger,
	}

	return &InstrumentedGithubClient{
		InstrumentedClient: instrumentedGHClient,
		GhClient:           client,
		StatsScope:         scope,
		Logger:             logger,
	}
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_github_pull_status_fetcher.go PullApprovalChecker
type SQPullStatusChecker interface {
	GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error)
	PullIsSQMergeable(repo models.Repo, pull models.PullRequest, statuses []*github.RepoStatus) (bool, error)
	PullIsSQLocked(baseRepo models.Repo, pull models.PullRequest, statuses []*github.RepoStatus) (bool, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_github_pull_request_getter.go GithubPullRequestGetter

type GithubPullRequestGetter interface {
	GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error)
	GetPullRequestFromName(repoName string, repoOwner string, pullNum int) (*github.PullRequest, error)
}

// IGithubClient exists to bridge the gap between GithubPullRequestGetter and Client interface to allow
// for a single instrumented client

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_IGithub_client.go IGithubClient
type IGithubClient interface {
	Client
	GithubPullRequestGetter
	SQPullStatusChecker

	GetContents(owner, repo, branch, path string) ([]byte, error)
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement soley any github specific interfaces.
type InstrumentedGithubClient struct {
	*InstrumentedClient
	GhClient   *GithubClient
	StatsScope stats.Scope
	Logger     logging.SimpleLogging
}

func (c *InstrumentedGithubClient) GetContents(owner, repo, branch, path string) ([]byte, error) {
	scope := c.StatsScope.Scope("get_contents")
	logger := c.Logger.WithHistory([]interface{}{
		"repository", fmt.Sprintf("%s/%s", owner, repo),
	}...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	contents, err := c.GhClient.GetContents(owner, repo, branch, path)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to get contents, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return contents, err
}

func (c *InstrumentedGithubClient) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	return c.GetPullRequestFromName(repo.Name, repo.Owner, pullNum)

}

func (c *InstrumentedGithubClient) GetPullRequestFromName(repoName string, repoOwner string, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.Scope("get_pull_request")
	logger := c.Logger.WithHistory([]interface{}{
		"repository", fmt.Sprintf("%s/%s", repoOwner, repoName),
		"pull-num", strconv.Itoa(pullNum),
	}...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	pull, err := c.GhClient.GetPullRequestFromName(repoName, repoOwner, pullNum)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to get pull number for repo, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return pull, err
}

func (c *InstrumentedGithubClient) PullIsSQMergeable(repo models.Repo, pull models.PullRequest, statuses []*github.RepoStatus) (bool, error) {
	scope := c.StatsScope.Scope("pull_is_sq_mergeable")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	sqMergeable, err := c.GhClient.PullIsSQMergeable(repo, pull, statuses)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to check pull sq mergeable status, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return sqMergeable, err
}

func (c *InstrumentedGithubClient) PullIsSQLocked(repo models.Repo, pull models.PullRequest, statuses []*github.RepoStatus) (bool, error) {
	scope := c.StatsScope.Scope("pull_is_locked")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	locked, err := c.GhClient.PullIsLocked(repo, pull, statuses)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to check pull lock status, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return locked, err
}

func (c *InstrumentedGithubClient) GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error) {
	scope := c.StatsScope.Scope("get_repo_status")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	statuses, err := c.GhClient.GetRepoStatuses(repo, pull)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to get repo status: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return statuses, err
}

type InstrumentedClient struct {
	Client
	StatsScope stats.Scope
	Logger     logging.SimpleLogging
}

func (c *InstrumentedClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := c.StatsScope.Scope("get_modified_files")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	files, err := c.Client.GetModifiedFiles(repo, pull)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to get modified files, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return files, err

}
func (c *InstrumentedClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	scope := c.StatsScope.Scope("create_comment")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pullNum)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	if err := c.Client.CreateComment(repo, pullNum, comment, command); err != nil {
		executionError.Inc()
		logger.Err("Unable to create comment for command %s, error: %s", command, err.Error())
		return err
	}

	executionSuccess.Inc()
	return nil
}
func (c *InstrumentedClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	scope := c.StatsScope.Scope("hide_prev_plan_comments")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pullNum)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	if err := c.Client.HidePrevCommandComments(repo, pullNum, command); err != nil {
		executionError.Inc()
		logger.Err("Unable to hide previous %s comments, error: %s", command, err.Error())
		return err
	}

	executionSuccess.Inc()
	return nil

}
func (c *InstrumentedClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	scope := c.StatsScope.Scope("pull_is_approved")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	approved, err := c.Client.PullIsApproved(repo, pull)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to check pull approval status, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return approved, err

}
func (c *InstrumentedClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	scope := c.StatsScope.Scope("pull_is_mergeable")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	mergeable, err := c.Client.PullIsMergeable(repo, pull)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to check pull mergeable status, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return mergeable, err
}

func (c *InstrumentedClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	scope := c.StatsScope.Scope("update_status")
	logger := c.Logger.WithHistory(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	if err := c.Client.UpdateStatus(repo, pull, state, src, description, url); err != nil {
		executionError.Inc()
		logger.Err("Unable to update status at url: %s, error: %s", url, err.Error())
		return err
	}

	executionSuccess.Inc()
	return nil

}
func (c *InstrumentedClient) MergePull(pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	scope := c.StatsScope.Scope("merge_pull")
	logger := c.Logger.WithHistory("pull-num", pull.Num)

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	if err := c.Client.MergePull(pull, pullOptions); err != nil {
		executionError.Inc()
		logger.Err("Unable to merge pull, error: %s", err.Error())
	}

	executionSuccess.Inc()
	return nil

}

// taken from other parts of the code, would be great to have this in a shared spot
func fmtLogSrc(repo models.Repo, pullNum int) []interface{} {
	return []interface{}{
		"repository", repo.FullName,
		"pull-num", strconv.Itoa(pullNum),
	}
}
