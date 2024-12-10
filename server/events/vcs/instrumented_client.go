package vcs

import (
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
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

//go:generate pegomock generate --package mocks -o mocks/mock_github_pull_request_getter.go GithubPullRequestGetter

type GithubPullRequestGetter interface {
	GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*github.PullRequest, error)
}

// IGithubClient exists to bridge the gap between GithubPullRequestGetter and Client interface to allow
// for a single instrumented client
type IGithubClient interface {
	Client
	GithubPullRequestGetter
}

// InstrumentedGithubClient should delegate to the underlying InstrumentedClient for vcs provider-agnostic
// methods and implement solely any github specific interfaces.
type InstrumentedGithubClient struct {
	*InstrumentedClient
	PullRequestGetter GithubPullRequestGetter
	StatsScope        tally.Scope
	Logger            logging.SimpleLogging
}

func (c *InstrumentedGithubClient) GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.SubScope("get_pull_request")
	scope = SetGitScopeTags(scope, repo.FullName, pullNum)

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

type InstrumentedClient struct {
	Client
	StatsScope tally.Scope
	Logger     logging.SimpleLogging
}

func (c *InstrumentedClient) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := c.StatsScope.SubScope("get_modified_files")
	scope = SetGitScopeTags(scope, repo.FullName, pull.Num)

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
	scope := c.StatsScope.SubScope("create_comment")
	scope = SetGitScopeTags(scope, repo.FullName, pullNum)

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
	scope := c.StatsScope.SubScope("react_to_comment")

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
	scope := c.StatsScope.SubScope("hide_prev_plan_comments")
	scope = SetGitScopeTags(scope, repo.FullName, pullNum)

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
	scope := c.StatsScope.SubScope("pull_is_approved")
	scope = SetGitScopeTags(scope, repo.FullName, pull.Num)

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

func (c *InstrumentedClient) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (bool, error) {
	scope := c.StatsScope.SubScope("pull_is_mergeable")
	scope = SetGitScopeTags(scope, repo.FullName, pull.Num)

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
	scope := c.StatsScope.SubScope("update_status")
	scope = SetGitScopeTags(scope, repo.FullName, pull.Num)

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
	scope := c.StatsScope.SubScope("merge_pull")
	scope = SetGitScopeTags(scope, pull.BaseRepo.FullName, pull.Num)

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

func SetGitScopeTags(scope tally.Scope, repoFullName string, pullNum int) tally.Scope {
	return scope.Tagged(map[string]string{
		"base_repo": repoFullName,
		"pr_number": strconv.Itoa(pullNum),
	})
}
