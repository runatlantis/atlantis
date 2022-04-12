// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/logging/fields"
	"github.com/runatlantis/atlantis/server/recovery"
	"github.com/uber-go/tally"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	ShutdownComment = "Atlantis server is shutting down, please try again later."
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_command_runner.go CommandRunner

// CommandRunner is the first step after a command request has been parsed.
type CommandRunner interface {
	// RunCommentCommand is the first step after a command request has been parsed.
	// It handles gathering additional information needed to execute the command
	// and then calling the appropriate services to finish executing the command.
	RunCommentCommand(ctx context.Context, baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *command.Comment, timestamp time.Time)
	RunAutoplanCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, timestamp time.Time)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_stale_command_checker.go StaleCommandChecker

// StaleCommandChecker handles checks to validate if current command is stale and can be dropped.
type StaleCommandChecker interface {
	// CommandIsStale returns true if currentEventTimestamp is earlier than timestamp set in DB's latest pull model.
	CommandIsStale(ctx *command.Context) bool
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_github_pull_getter.go GithubPullGetter

// GithubPullGetter makes API calls to get pull requests.
type GithubPullGetter interface {
	// GetPullRequest gets the pull request with id pullNum for the repo.
	GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_azuredevops_pull_getter.go AzureDevopsPullGetter

// AzureDevopsPullGetter makes API calls to get pull requests.
type AzureDevopsPullGetter interface {
	// GetPullRequest gets the pull request with id pullNum for the repo.
	GetPullRequest(repo models.Repo, pullNum int) (*azuredevops.GitPullRequest, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_gitlab_merge_request_getter.go GitlabMergeRequestGetter

// GitlabMergeRequestGetter makes API calls to get merge requests.
type GitlabMergeRequestGetter interface {
	// GetMergeRequest gets the pull request with the id pullNum for the repo.
	GetMergeRequest(repoFullName string, pullNum int) (*gitlab.MergeRequest, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_comment_command_runner.go CommentCommandRunner

// CommentCommandRunner runs individual command workflows.
type CommentCommandRunner interface {
	Run(*command.Context, *command.Comment)
}

func buildCommentCommandRunner(
	cmdRunner *DefaultCommandRunner,
	cmdName command.Name,
) CommentCommandRunner {
	// panic here, we want to fail fast and hard since
	// this would be an internal service configuration error.
	runner, ok := cmdRunner.CommentCommandRunnerByCmd[cmdName]

	if !ok {
		panic(fmt.Sprintf("command runner not configured for command %s", cmdName.String()))
	}

	return runner
}

// DefaultCommandRunner is the first step when processing a comment command.
type DefaultCommandRunner struct {
	VCSClient                vcs.Client
	GithubPullGetter         GithubPullGetter
	AzureDevopsPullGetter    AzureDevopsPullGetter
	GitlabMergeRequestGetter GitlabMergeRequestGetter
	DisableAutoplan          bool
	EventParser              EventParsing
	GlobalCfg                valid.GlobalCfg
	StatsScope               tally.Scope
	// ParallelPoolSize controls the size of the wait group used to run
	// parallel plans and applies (if enabled).
	ParallelPoolSize              int
	CommentCommandRunnerByCmd     map[command.Name]command.Runner
	Drainer                       *Drainer
	PreWorkflowHooksCommandRunner PreWorkflowHooksCommandRunner
	CommitStatusUpdater           CommitStatusUpdater
	PullStatusFetcher             PullStatusFetcher
	StaleCommandChecker           StaleCommandChecker
	Logger                        logging.Logger
	LegacyLogger                  logging.SimpleLogging
}

// RunAutoplanCommand runs plan and policy_checks when a pull request is opened or updated.
func (c *DefaultCommandRunner) RunAutoplanCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, timestamp time.Time) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pull.Num, ShutdownComment, command.Plan.String()); commentErr != nil {
			c.Logger.ErrorContext(ctx, commentErr.Error())
		}
		return
	}
	defer c.Drainer.OpDone()

	ctx = newCtx(ctx, baseRepo.FullName, pull.Num)
	defer c.logPanics(ctx)
	status, err := c.PullStatusFetcher.GetPullStatus(pull)

	if err != nil {
		c.Logger.ErrorContext(ctx, err.Error())
	}

	scope := c.StatsScope.SubScope("autoplan")
	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	cmdCtx := &command.Context{
		User:             user,
		Log:              c.buildLegacyLogger(ctx, c.LegacyLogger, baseRepo.FullName, pull.Num),
		Scope:            scope,
		Pull:             pull,
		HeadRepo:         headRepo,
		PullStatus:       status,
		Trigger:          command.AutoTrigger,
		TriggerTimestamp: timestamp,
	}
	if !c.validateCtxAndComment(ctx, cmdCtx) {
		return
	}
	if c.DisableAutoplan {
		return
	}
	// Drop request if a more recent VCS event updated Atlantis state
	if c.StaleCommandChecker.CommandIsStale(cmdCtx) {
		return
	}

	if err := c.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmdCtx); err != nil {
		c.Logger.ErrorContext(ctx, "Error running pre-workflow hooks", fields.PullRequestWithErr(pull, err))
		c.CommitStatusUpdater.UpdateCombined(ctx, cmdCtx.HeadRepo, cmdCtx.Pull, models.FailedCommitStatus, command.Plan)
		return
	}

	autoPlanRunner := buildCommentCommandRunner(c, command.Plan)

	autoPlanRunner.Run(cmdCtx, nil)
}

// RunCommentCommand executes the command.
// We take in a pointer for maybeHeadRepo because for some events there isn't
// enough data to construct the Repo model and callers might want to wait until
// the event is further validated before making an additional (potentially
// wasteful) call to get the necessary data.
func (c *DefaultCommandRunner) RunCommentCommand(ctx context.Context, baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *command.Comment, timestamp time.Time) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pullNum, ShutdownComment, ""); commentErr != nil {
			c.Logger.ErrorContext(ctx, commentErr.Error())
		}
		return
	}
	defer c.Drainer.OpDone()

	ctx = newCtx(ctx, baseRepo.FullName, pullNum)
	defer c.logPanics(ctx)

	scope := c.StatsScope.SubScope("comment")

	if cmd != nil {
		scope = scope.SubScope(cmd.Name.String())
	}
	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	headRepo, pull, err := c.ensureValidRepoMetadata(ctx, baseRepo, maybeHeadRepo, maybePull, user, pullNum)
	if err != nil {
		return
	}

	status, err := c.PullStatusFetcher.GetPullStatus(pull)

	if err != nil {
		c.Logger.ErrorContext(ctx, err.Error())
	}

	cmdCtx := &command.Context{
		User:             user,
		Log:              c.buildLegacyLogger(ctx, c.LegacyLogger, baseRepo.FullName, pull.Num),
		Pull:             pull,
		PullStatus:       status,
		HeadRepo:         headRepo,
		Trigger:          command.CommentTrigger,
		Scope:            scope,
		TriggerTimestamp: timestamp,
	}

	if !c.validateCtxAndComment(ctx, cmdCtx) {
		return
	}

	// Drop request if a more recent VCS event updated Atlantis state
	if c.StaleCommandChecker.CommandIsStale(cmdCtx) {
		return
	}

	if err := c.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmdCtx); err != nil {
		c.Logger.ErrorContext(ctx, "Error running pre-workflow hooks", fields.PullRequestWithErr(pull, err))
		c.CommitStatusUpdater.UpdateCombined(ctx, cmdCtx.HeadRepo, cmdCtx.Pull, models.FailedCommitStatus, cmd.Name)
		return
	}

	cmdRunner := buildCommentCommandRunner(c, cmd.CommandName())

	cmdRunner.Run(cmdCtx, cmd)
}

func (c *DefaultCommandRunner) getGithubData(baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if c.GithubPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("Atlantis not configured to support GitHub")
	}
	ghPull, err := c.GithubPullGetter.GetPullRequest(baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to GitHub")
	}
	pull, _, headRepo, err := c.EventParser.ParseGithubPull(ghPull)
	if err != nil {
		return pull, headRepo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, headRepo, nil
}

func (c *DefaultCommandRunner) getGitlabData(baseRepo models.Repo, pullNum int) (models.PullRequest, error) {
	if c.GitlabMergeRequestGetter == nil {
		return models.PullRequest{}, errors.New("Atlantis not configured to support GitLab")
	}
	mr, err := c.GitlabMergeRequestGetter.GetMergeRequest(baseRepo.FullName, pullNum)
	if err != nil {
		return models.PullRequest{}, errors.Wrap(err, "making merge request API call to GitLab")
	}
	pull := c.EventParser.ParseGitlabMergeRequest(mr, baseRepo)
	return pull, nil
}

func (c *DefaultCommandRunner) getAzureDevopsData(baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if c.AzureDevopsPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("atlantis not configured to support Azure DevOps")
	}
	adPull, err := c.AzureDevopsPullGetter.GetPullRequest(baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to Azure DevOps")
	}
	pull, _, headRepo, err := c.EventParser.ParseAzureDevopsPull(adPull)
	if err != nil {
		return pull, headRepo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, headRepo, nil
}

func (c *DefaultCommandRunner) buildLegacyLogger(ctx context.Context, log logging.SimpleLogging, repoFullName string, pullNum int) logging.SimpleLogging {

	args := []interface{}{
		"repository", repoFullName,
		"pull-num", strconv.Itoa(pullNum),
	}
	// remove this once we drop support for this legacy logger
	if requestId, ok := ctx.Value(logging.RequestIDKey).(string); ok {
		args = append(args, "gh-request-id", requestId)
	}

	return log.With(args...)
}

func newCtx(ctx context.Context, repoFullName string, pullNum int) context.Context {
	ctx = context.WithValue(ctx, logging.RepositoryKey, repoFullName)
	return context.WithValue(ctx, logging.PullNumKey, strconv.Itoa(pullNum))
}

func (c *DefaultCommandRunner) ensureValidRepoMetadata(
	ctx context.Context,
	baseRepo models.Repo,
	maybeHeadRepo *models.Repo,
	maybePull *models.PullRequest,
	user models.User,
	pullNum int,
) (headRepo models.Repo, pull models.PullRequest, err error) {
	if maybeHeadRepo != nil {
		headRepo = *maybeHeadRepo
	}

	switch baseRepo.VCSHost.Type {
	case models.Github:
		pull, headRepo, err = c.getGithubData(baseRepo, pullNum)
	case models.Gitlab:
		pull, err = c.getGitlabData(baseRepo, pullNum)
	case models.BitbucketCloud, models.BitbucketServer:
		if maybePull == nil {
			err = errors.New("pull request should not be nil–this is a bug")
			break
		}
		pull = *maybePull
	case models.AzureDevops:
		pull, headRepo, err = c.getAzureDevopsData(baseRepo, pullNum)
	default:
		err = errors.New("Unknown VCS type–this is a bug")
	}

	if err != nil {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pullNum, fmt.Sprintf("`Error: %s`", err), ""); commentErr != nil {
			c.Logger.ErrorContext(ctx, commentErr.Error())
		}
	}

	return
}

func (c *DefaultCommandRunner) validateCtxAndComment(ctx context.Context, cmdCtx *command.Context) bool {
	if cmdCtx.HeadRepo.Owner != cmdCtx.Pull.BaseRepo.Owner {
		c.Logger.InfoContext(ctx, "command was run on a fork pull request which is disallowed")
		if err := c.VCSClient.CreateComment(cmdCtx.Pull.BaseRepo, cmdCtx.Pull.Num, "Atlantis commands can't be run on fork pull requests.", ""); err != nil {
			c.Logger.ErrorContext(ctx, err.Error())
		}
		return false
	}

	if cmdCtx.Pull.State != models.OpenPullState {
		c.Logger.InfoContext(ctx, "command was run on closed pull request")
		if err := c.VCSClient.CreateComment(cmdCtx.Pull.BaseRepo, cmdCtx.Pull.Num, "Atlantis commands can't be run on closed pull requests", ""); err != nil {
			c.Logger.ErrorContext(ctx, err.Error())
		}
		return false
	}

	repo := c.GlobalCfg.MatchingRepo(cmdCtx.Pull.BaseRepo.ID())
	if !repo.BranchMatches(cmdCtx.Pull.BaseBranch) {
		c.Logger.InfoContext(ctx, "command was run on a pull request which doesn't match base branches")
		// just ignore it to allow us to use any git workflows without malicious intentions.
		return false
	}
	return true
}

// logPanics logs and creates a comment on the pull request for panics.
func (c *DefaultCommandRunner) logPanics(ctx context.Context) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		c.Logger.ErrorContext(ctx, fmt.Sprintf("PANIC: %s\n%s", err, stack))
	}
}

type ForceApplyCommandRunner struct {
	CommandRunner
	VCSClient vcs.Client
	Logger    logging.Logger
}

func (f *ForceApplyCommandRunner) RunCommentCommand(ctx context.Context, baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *command.Comment, timestamp time.Time) {
	if cmd.ForceApply {
		warningMessage := "⚠️ WARNING ⚠️\n\n You have bypassed all apply requirements for this PR 🚀 . This can have unpredictable consequences 🙏🏽 and should only be used in an emergency 🆘 .\n\n 𝐓𝐡𝐢𝐬 𝐚𝐜𝐭𝐢𝐨𝐧 𝐰𝐢𝐥𝐥 𝐛𝐞 𝐚𝐮𝐝𝐢𝐭𝐞𝐝.\n"
		if commentErr := f.VCSClient.CreateComment(baseRepo, pullNum, warningMessage, ""); commentErr != nil {
			f.Logger.ErrorContext(ctx, commentErr.Error())
		}
	}
	f.CommandRunner.RunCommentCommand(ctx, baseRepo, maybeHeadRepo, maybePull, user, pullNum, cmd, timestamp)
}
