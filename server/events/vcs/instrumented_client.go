package vcs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v31/github"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
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
		GhClient:           client,
		StatsScope:         scope,
		Logger:             logger,
	}
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

	GetContents(owner, repo, branch, path string) ([]byte, error)
	GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error)
	GetRepoChecks(repo models.Repo, pull models.PullRequest) ([]*github.CheckRun, error)
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement soley any github specific interfaces.
type InstrumentedGithubClient struct {
	*InstrumentedClient
	GhClient   *GithubClient
	StatsScope tally.Scope
	Logger     logging.SimpleLogging
}

func (c *InstrumentedGithubClient) GetContents(owner, repo, branch, path string) ([]byte, error) {
	scope := c.StatsScope.SubScope("get_contents")
	logger := c.Logger.With([]interface{}{
		"repository", fmt.Sprintf("%s/%s", owner, repo),
	}...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	contents, err := c.GhClient.GetContents(owner, repo, branch, path)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get contents, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return contents, err
}

func (c *InstrumentedGithubClient) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	return c.GetPullRequestFromName(repo.Name, repo.Owner, pullNum)

}

func (c *InstrumentedGithubClient) GetPullRequestFromName(repoName string, repoOwner string, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.SubScope("get_pull_request")
	logger := c.Logger.With([]interface{}{
		"repository", fmt.Sprintf("%s/%s", repoOwner, repoName),
		"pull-num", strconv.Itoa(pullNum),
	}...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	pull, err := c.GhClient.GetPullRequestFromName(repoName, repoOwner, pullNum)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get pull number for repo, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return pull, err
}

func (c *InstrumentedGithubClient) GetRepoChecks(repo models.Repo, pull models.PullRequest) ([]*github.CheckRun, error) {
	scope := c.StatsScope.SubScope("get_repo_checks")
	logger := c.Logger.With(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	statuses, err := c.GhClient.GetRepoChecks(repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get repo status: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return statuses, err
}

func (c *InstrumentedGithubClient) GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error) {
	scope := c.StatsScope.SubScope("get_repo_status")
	logger := c.Logger.With(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	statuses, err := c.GhClient.GetRepoStatuses(repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to get repo status: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return statuses, err
}

type InstrumentedClient struct {
	Client
	StatsScope tally.Scope
	Logger     logging.SimpleLogging
}

func (c *InstrumentedClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := c.StatsScope.SubScope("get_modified_files")
	logger := c.Logger.With(fmtLogSrc(repo, pull.Num)...)

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
	logger := c.Logger.With(fmtLogSrc(repo, pullNum)...)

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
	logger := c.Logger.With(fmtLogSrc(repo, pullNum)...)

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
	logger := c.Logger.With(fmtLogSrc(repo, pull.Num)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	approvalStatus, err := c.Client.PullIsApproved(repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("Unable to check pull approval status, error: %s", err.Error())
	} else {
		executionSuccess.Inc(1)
	}

	return approvalStatus, err

}
func (c *InstrumentedClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	scope := c.StatsScope.SubScope("pull_is_mergeable")
	logger := c.Logger.With(fmtLogSrc(repo, pull.Num)...)

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

func (c *InstrumentedClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) error {
	scope := c.StatsScope.SubScope("update_status")

	repo := request.Repo
	logger := c.Logger.With(fmtLogSrc(repo, request.PullNum)...)

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.UpdateStatus(ctx, request); err != nil {
		executionError.Inc(1)
		logger.Err("Unable to update status at url: %s, error: %s", request.DetailsURL, err.Error())
		return err
	}

	executionSuccess.Inc(1)
	return nil

}
func (c *InstrumentedClient) MergePull(pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	scope := c.StatsScope.SubScope("merge_pull")
	logger := c.Logger.With("pull-num", pull.Num)

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
