package vcs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/logging/fields"
	keys "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/uber-go/tally/v4"
)

// NewInstrumentedGithubClient creates a client proxy responsible for gathering stats and logging
func NewInstrumentedGithubClient(client *GithubClient, statsScope tally.Scope, logger logging.Logger) IGithubClient {
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
	GetRepoChecks(repo models.Repo, commitSHA string) ([]*github.CheckRun, error)
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement soley any github specific interfaces.
type InstrumentedGithubClient struct {
	*InstrumentedClient
	GhClient   *GithubClient
	StatsScope tally.Scope
	Logger     logging.Logger
}

func (c *InstrumentedGithubClient) GetContents(owner, repo, branch, path string) ([]byte, error) {
	scope := c.StatsScope.SubScope("get_contents")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	contents, err := c.GhClient.GetContents(owner, repo, branch, path)

	if err != nil {
		executionError.Inc(1)
		return contents, err
	}
	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched contents", map[string]interface{}{
		keys.RepositoryKey.String(): repo,
	})

	return contents, err
}

func (c *InstrumentedGithubClient) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	return c.GetPullRequestFromName(repo.Name, repo.Owner, pullNum)
}

func (c *InstrumentedGithubClient) GetPullRequestFromName(repoName string, repoOwner string, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.SubScope("get_pull_request")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	pull, err := c.GhClient.GetPullRequestFromName(repoName, repoOwner, pullNum)

	if err != nil {
		executionError.Inc(1)
		return pull, err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched pull request", map[string]interface{}{
		keys.RepositoryKey.String(): fmt.Sprintf("%s/%s", repoOwner, repoName),
		keys.PullNumKey.String():    strconv.Itoa(pullNum),
	})

	return pull, err
}

func (c *InstrumentedGithubClient) GetRepoChecks(repo models.Repo, commitSHA string) ([]*github.CheckRun, error) {
	scope := c.StatsScope.SubScope("get_repo_checks")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	statuses, err := c.GhClient.GetRepoChecks(repo, commitSHA)

	if err != nil {
		executionError.Inc(1)
		return statuses, err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched vcs repo checks", map[string]interface{}{"commitSHA": commitSHA})

	return statuses, err
}

func (c *InstrumentedGithubClient) GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error) {
	scope := c.StatsScope.SubScope("get_repo_status")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	statuses, err := c.GhClient.GetRepoStatuses(repo, pull)

	if err != nil {
		executionError.Inc(1)
		return statuses, err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched vcs repo statuses", fields.PullRequest(pull))

	return statuses, err
}

type InstrumentedClient struct {
	Client
	StatsScope tally.Scope
	Logger     logging.Logger
}

func (c *InstrumentedClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := c.StatsScope.SubScope("get_modified_files")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	files, err := c.Client.GetModifiedFiles(repo, pull)

	if err != nil {
		executionError.Inc(1)
		return files, err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched pull request modified files", fields.PullRequest(pull))

	return files, err

}
func (c *InstrumentedClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	scope := c.StatsScope.SubScope("create_comment")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.CreateComment(repo, pullNum, comment, command); err != nil {
		executionError.Inc(1)
		return err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("created pull request comment", map[string]interface{}{
		keys.RepositoryKey.String(): repo.FullName,
		keys.PullNumKey.String():    strconv.Itoa(pullNum),
	})
	return nil
}
func (c *InstrumentedClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	scope := c.StatsScope.SubScope("hide_prev_plan_comments")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	if err := c.Client.HidePrevCommandComments(repo, pullNum, command); err != nil {
		executionError.Inc(1)
		return err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("hid previous comments", map[string]interface{}{
		keys.RepositoryKey.String(): repo.FullName,
		keys.PullNumKey.String():    strconv.Itoa(pullNum),
	})
	return nil

}
func (c *InstrumentedClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	scope := c.StatsScope.SubScope("pull_is_approved")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	approvalStatus, err := c.Client.PullIsApproved(repo, pull)

	if err != nil {
		executionError.Inc(1)
		return approvalStatus, err
	}

	executionSuccess.Inc(1)

	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched pull request approval status", fields.PullRequest(pull))

	return approvalStatus, err

}
func (c *InstrumentedClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	scope := c.StatsScope.SubScope("pull_is_mergeable")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	mergeable, err := c.Client.PullIsMergeable(repo, pull)

	if err != nil {
		executionError.Inc(1)
		return mergeable, err
	}

	executionSuccess.Inc(1)
	//TODO: thread context and use related logging methods.
	c.Logger.Info("fetched pull request mergeability", fields.PullRequest(pull))

	return mergeable, err
}

func (c *InstrumentedClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	scope := c.StatsScope.SubScope("update_status")

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	statusID, err := c.Client.UpdateStatus(ctx, request)
	if err != nil {
		executionError.Inc(1)
		return "", err
	}

	//TODO: thread context and use related logging methods.
	// for now keeping this at info to debug weirdness we've been
	// seeing with status api calls.
	c.Logger.Info("updated vcs status", map[string]interface{}{
		keys.RepositoryKey.String(): request.Repo.FullName,
		keys.PullNumKey.String():    strconv.Itoa(request.PullNum),
		keys.SHAKey.String():        request.Ref,
		"status-name":               request.StatusName,
		"status-id":                 request.StatusID,
		"state":                     request.State.String(),
	})

	executionSuccess.Inc(1)
	return statusID, nil

}
