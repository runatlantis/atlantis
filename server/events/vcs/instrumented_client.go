package vcs

import (
	"fmt"

	"github.com/google/go-github/v31/github"
	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// NewInstrumentedGithubClient creates a client proxy responsible for gathering stats and logging
func NewInstrumentedGithubClient(client *GithubClient, statsScope stats.Scope, logger *logging.SimpleLogger) IGithubClient {
	scope := statsScope.Scope("github")

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
	StatsScope        stats.Scope
	Logger            *logging.SimpleLogger
}

func (c *InstrumentedGithubClient) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	scope := c.StatsScope.Scope("get_pull_request")
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pullNum), true, c.Logger.GetLevel())

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	pull, err := c.PullRequestGetter.GetPullRequest(repo, pullNum)

	if err != nil {
		executionError.Inc()
		logger.Err("Unable to get pull number for repo, error: %s", err.Error())
	} else {
		executionSuccess.Inc()
	}

	return pull, err

}

type InstrumentedClient struct {
	Client
	StatsScope stats.Scope
	Logger     *logging.SimpleLogger
}

func (c *InstrumentedClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	scope := c.StatsScope.Scope("get_modified_files")
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pull.Num), true, c.Logger.GetLevel())

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
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pullNum), true, c.Logger.GetLevel())

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
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pullNum), true, c.Logger.GetLevel())

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
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pull.Num), true, c.Logger.GetLevel())

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
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pull.Num), true, c.Logger.GetLevel())

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
	logger := c.Logger.NewLogger(fmtLogSrc(repo, pull.Num), true, c.Logger.GetLevel())

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
func (c *InstrumentedClient) MergePull(pull models.PullRequest) error {
	scope := c.StatsScope.Scope("merge_pull")
	logger := c.Logger.NewLogger(fmt.Sprintf("#%d", pull.Num), true, c.Logger.GetLevel())

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	if err := c.Client.MergePull(pull); err != nil {
		executionError.Inc()
		logger.Err("Unable to merge pull, error: %s", err.Error())
	}

	executionSuccess.Inc()
	return nil

}

// taken from other parts of the code, would be great to have this in a shared spot
func fmtLogSrc(repo models.Repo, pullNum int) string {
	return fmt.Sprintf("%s#%d", repo.FullName, pullNum)
}
