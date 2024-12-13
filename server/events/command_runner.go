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
	"fmt"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/recovery"
	"github.com/runatlantis/atlantis/server/utils"
	tally "github.com/uber-go/tally/v4"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	ShutdownComment = "Atlantis server is shutting down, please try again later."
)

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_command_runner.go CommandRunner

// CommandRunner is the first step after a command request has been parsed.
type CommandRunner interface {
	// RunCommentCommand is the first step after a command request has been parsed.
	// It handles gathering additional information needed to execute the command
	// and then calling the appropriate services to finish executing the command.
	RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand)
	RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User)
}

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_github_pull_getter.go GithubPullGetter

// GithubPullGetter makes API calls to get pull requests.
type GithubPullGetter interface {
	// GetPullRequest gets the pull request with id pullNum for the repo.
	GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*github.PullRequest, error)
}

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_azuredevops_pull_getter.go AzureDevopsPullGetter

// AzureDevopsPullGetter makes API calls to get pull requests.
type AzureDevopsPullGetter interface {
	// GetPullRequest gets the pull request with id pullNum for the repo.
	GetPullRequest(logger logging.SimpleLogging, repo models.Repo, pullNum int) (*azuredevops.GitPullRequest, error)
}

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_gitlab_merge_request_getter.go GitlabMergeRequestGetter

// GitlabMergeRequestGetter makes API calls to get merge requests.
type GitlabMergeRequestGetter interface {
	// GetMergeRequest gets the pull request with the id pullNum for the repo.
	GetMergeRequest(logger logging.SimpleLogging, repoFullName string, pullNum int) (*gitlab.MergeRequest, error)
}

// CommentCommandRunner runs individual command workflows.
type CommentCommandRunner interface {
	Run(*command.Context, *CommentCommand)
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
	GiteaPullGetter          *gitea.GiteaClient
	// User config option: Disables autoplan when a pull request is opened or updated.
	DisableAutoplan      bool
	DisableAutoplanLabel string
	EventParser          EventParsing
	// User config option: Fail and do not run the Atlantis command request if any of the pre workflow hooks error
	FailOnPreWorkflowHookError bool
	Logger                     logging.SimpleLogging
	GlobalCfg                  valid.GlobalCfg
	StatsScope                 tally.Scope
	// User config option: controls whether to operate on pull requests from forks.
	AllowForkPRs bool
	// ParallelPoolSize controls the size of the wait group used to run
	// parallel plans and applies (if enabled).
	ParallelPoolSize int
	// AllowForkPRsFlag is the name of the flag that controls fork PR's. We use
	// this in our error message back to the user on a forked PR so they know
	// how to enable this functionality.
	AllowForkPRsFlag string
	// User config option: controls whether to comment on Fork PRs when AllowForkPRs = False
	SilenceForkPRErrors bool
	// SilenceForkPRErrorsFlag is the name of the flag that controls fork PR's. We use
	// this in our error message back to the user on a forked PR so they know
	// how to disable error comment
	SilenceForkPRErrorsFlag        string
	CommentCommandRunnerByCmd      map[command.Name]CommentCommandRunner
	Drainer                        *Drainer
	PreWorkflowHooksCommandRunner  PreWorkflowHooksCommandRunner
	PostWorkflowHooksCommandRunner PostWorkflowHooksCommandRunner
	PullStatusFetcher              PullStatusFetcher
	TeamAllowlistChecker           command.TeamAllowlistChecker
	VarFileAllowlistChecker        *VarFileAllowlistChecker
	CommitStatusUpdater            CommitStatusUpdater
}

// RunAutoplanCommand runs plan and policy_checks when a pull request is opened or updated.
func (c *DefaultCommandRunner) RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(c.Logger, baseRepo, pull.Num, ShutdownComment, command.Plan.String()); commentErr != nil {
			c.Logger.Log(logging.Error, "unable to comment that Atlantis is shutting down: %s", commentErr)
		}
		return
	}
	defer c.Drainer.OpDone()

	log := c.buildLogger(baseRepo.FullName, pull.Num)
	defer c.logPanics(baseRepo, pull.Num, log)
	status, err := c.PullStatusFetcher.GetPullStatus(pull)

	if err != nil {
		log.Err("Unable to fetch pull status, this is likely a bug.", err)
	}

	scope := c.StatsScope.SubScope("autoplan")
	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	// Check if the user who triggered the autoplan has permissions to run 'plan'.
	if c.TeamAllowlistChecker != nil && c.TeamAllowlistChecker.HasRules() {
		err := c.fetchUserTeams(baseRepo, &user)
		if err != nil {
			log.Err("Unable to fetch user teams: %s", err)
			return
		}

		ok, err := c.checkUserPermissions(baseRepo, user, "plan")
		if err != nil {
			log.Err("Unable to check user permissions: %s", err)
			return
		}
		if !ok {
			return
		}
	}

	ctx := &command.Context{
		User:       user,
		Log:        log,
		Scope:      scope,
		Pull:       pull,
		HeadRepo:   headRepo,
		PullStatus: status,
		Trigger:    command.AutoTrigger,
	}
	if !c.validateCtxAndComment(ctx, command.Autoplan) {
		return
	}
	if c.DisableAutoplan {
		return
	}
	if len(c.DisableAutoplanLabel) > 0 {
		labels, err := c.VCSClient.GetPullLabels(ctx.Log, baseRepo, pull)
		if err != nil {
			ctx.Log.Err("Unable to get VCS pull/merge request labels: %s. Proceeding with autoplan.", err)
		} else if utils.SlicesContains(labels, c.DisableAutoplanLabel) {
			ctx.Log.Info("Pull/merge request has disable auto plan label '%s' so not running autoplan.", c.DisableAutoplanLabel)
			return
		}
	}

	ctx.Log.Info("Running autoplan...")
	cmd := &CommentCommand{
		Name: command.Autoplan,
	}
	err = c.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmd)

	if err != nil {
		if c.FailOnPreWorkflowHookError {
			ctx.Log.Err("'fail-on-pre-workflow-hook-error' set, so not running %s command.", command.Plan)

			// Update the plan or apply commit status to failed
			switch cmd.Name {
			case command.Plan:
				if err := c.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); err != nil {
					ctx.Log.Warn("Unable to update plan commit status: %s", err)
				}
			case command.Apply:
				if err := c.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Apply); err != nil {
					ctx.Log.Warn("Unable to update apply commit status: %s", err)
				}
			}

			return
		}

		ctx.Log.Err("'fail-on-pre-workflow-hook-error' not set so running %s command.", command.Plan)
	}

	autoPlanRunner := buildCommentCommandRunner(c, command.Plan)

	autoPlanRunner.Run(ctx, nil)

	c.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, cmd) // nolint: errcheck
}

// commentUserDoesNotHavePermissions comments on the pull request that the user
// is not allowed to execute the command.
func (c *DefaultCommandRunner) commentUserDoesNotHavePermissions(baseRepo models.Repo, pullNum int, user models.User, cmd *CommentCommand) {
	errMsg := fmt.Sprintf("```\nError: User @%s does not have permissions to execute '%s' command.\n```", user.Username, cmd.Name.String())
	if err := c.VCSClient.CreateComment(c.Logger, baseRepo, pullNum, errMsg, ""); err != nil {
		c.Logger.Err("unable to comment on pull request: %s", err)
	}
}

// checkUserPermissions checks if the user has permissions to execute the command
func (c *DefaultCommandRunner) checkUserPermissions(repo models.Repo, user models.User, cmdName string) (bool, error) {
	if c.TeamAllowlistChecker == nil || !c.TeamAllowlistChecker.HasRules() {
		// allowlist restriction is not enabled
		return true, nil
	}
	ctx := models.TeamAllowlistCheckerContext{
		BaseRepo:    repo,
		CommandName: cmdName,
		Log:         c.Logger,
		Pull:        models.PullRequest{},
		User:        user,
		Verbose:     false,
		API:         false,
	}
	ok := c.TeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, user.Teams, cmdName)
	if !ok {
		return false, nil
	}
	return true, nil
}

// checkVarFilesInPlanCommandAllowlisted checks if paths in a 'plan' command are allowlisted.
func (c *DefaultCommandRunner) checkVarFilesInPlanCommandAllowlisted(cmd *CommentCommand) error {
	if cmd == nil || cmd.CommandName() != command.Plan {
		return nil
	}

	return c.VarFileAllowlistChecker.Check(cmd.Flags)
}

// RunCommentCommand executes the command.
// We take in a pointer for maybeHeadRepo because for some events there isn't
// enough data to construct the Repo model and callers might want to wait until
// the event is further validated before making an additional (potentially
// wasteful) call to get the necessary data.
func (c *DefaultCommandRunner) RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(c.Logger, baseRepo, pullNum, ShutdownComment, ""); commentErr != nil {
			c.Logger.Log(logging.Error, "unable to comment that Atlantis is shutting down: %s", commentErr)
		}
		return
	}
	defer c.Drainer.OpDone()

	log := c.buildLogger(baseRepo.FullName, pullNum)
	defer c.logPanics(baseRepo, pullNum, log)

	scope := c.StatsScope.SubScope("comment")

	if cmd != nil {
		scope = scope.SubScope(cmd.Name.String())
	}
	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	// Check if the user who commented has the permissions to execute the 'plan' or 'apply' commands
	if c.TeamAllowlistChecker != nil && c.TeamAllowlistChecker.HasRules() {
		err := c.fetchUserTeams(baseRepo, &user)
		if err != nil {
			c.Logger.Err("Unable to fetch user teams: %s", err)
			return
		}

		ok, err := c.checkUserPermissions(baseRepo, user, cmd.Name.String())
		if err != nil {
			c.Logger.Err("Unable to check user permissions: %s", err)
			return
		}
		if !ok {
			c.commentUserDoesNotHavePermissions(baseRepo, pullNum, user, cmd)
			return
		}
	}

	// Check if the provided var files in a 'plan' command are allowlisted
	if err := c.checkVarFilesInPlanCommandAllowlisted(cmd); err != nil {
		errMsg := fmt.Sprintf("```\n%s\n```", err.Error())
		if commentErr := c.VCSClient.CreateComment(c.Logger, baseRepo, pullNum, errMsg, ""); commentErr != nil {
			c.Logger.Err("unable to comment on pull request: %s", commentErr)
		}
		return
	}

	headRepo, pull, err := c.ensureValidRepoMetadata(baseRepo, maybeHeadRepo, maybePull, user, pullNum, log)
	if err != nil {
		return
	}

	status, err := c.PullStatusFetcher.GetPullStatus(pull)

	if err != nil {
		log.Err("Unable to fetch pull status, this is likely a bug.", err)
	}

	ctx := &command.Context{
		User:                 user,
		Log:                  log,
		Pull:                 pull,
		PullStatus:           status,
		HeadRepo:             headRepo,
		Scope:                scope,
		Trigger:              command.CommentTrigger,
		PolicySet:            cmd.PolicySet,
		ClearPolicyApproval:  cmd.ClearPolicyApproval,
		TeamAllowlistChecker: c.TeamAllowlistChecker,
	}

	if !c.validateCtxAndComment(ctx, cmd.Name) {
		return
	}

	err = c.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmd)

	if err != nil {
		if c.FailOnPreWorkflowHookError {
			ctx.Log.Err("'fail-on-pre-workflow-hook-error' set, so not running %s command.", cmd.Name.String())

			// Update the plan or apply commit status to failed
			switch cmd.Name {
			case command.Plan:
				if err := c.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); err != nil {
					ctx.Log.Warn("unable to update plan commit status: %s", err)
				}
			case command.Apply:
				if err := c.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Apply); err != nil {
					ctx.Log.Warn("unable to update apply commit status: %s", err)
				}
			}

			return
		}

		ctx.Log.Err("'fail-on-pre-workflow-hook-error' not set so running %s command.", cmd.Name.String())
	}

	cmdRunner := buildCommentCommandRunner(c, cmd.CommandName())

	cmdRunner.Run(ctx, cmd)

	c.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, cmd) // nolint: errcheck
}

func (c *DefaultCommandRunner) getGithubData(logger logging.SimpleLogging, baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if c.GithubPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("Atlantis not configured to support GitHub")
	}
	ghPull, err := c.GithubPullGetter.GetPullRequest(logger, baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to GitHub")
	}
	pull, _, headRepo, err := c.EventParser.ParseGithubPull(logger, ghPull)
	if err != nil {
		return pull, headRepo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, headRepo, nil
}

func (c *DefaultCommandRunner) getGiteaData(logger logging.SimpleLogging, baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if c.GiteaPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("Atlantis not configured to support Gitea")
	}
	giteaPull, err := c.GiteaPullGetter.GetPullRequest(logger, baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to Gitea")
	}
	pull, _, headRepo, err := c.EventParser.ParseGiteaPull(giteaPull)
	if err != nil {
		return pull, headRepo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, headRepo, nil
}

func (c *DefaultCommandRunner) getGitlabData(logger logging.SimpleLogging, baseRepo models.Repo, pullNum int) (models.PullRequest, error) {
	if c.GitlabMergeRequestGetter == nil {
		return models.PullRequest{}, errors.New("Atlantis not configured to support GitLab")
	}
	mr, err := c.GitlabMergeRequestGetter.GetMergeRequest(logger, baseRepo.FullName, pullNum)
	if err != nil {
		return models.PullRequest{}, errors.Wrap(err, "making merge request API call to GitLab")
	}
	pull := c.EventParser.ParseGitlabMergeRequest(mr, baseRepo)
	return pull, nil
}

func (c *DefaultCommandRunner) getAzureDevopsData(logger logging.SimpleLogging, baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if c.AzureDevopsPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("atlantis not configured to support Azure DevOps")
	}
	adPull, err := c.AzureDevopsPullGetter.GetPullRequest(logger, baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to Azure DevOps")
	}
	pull, _, headRepo, err := c.EventParser.ParseAzureDevopsPull(adPull)
	if err != nil {
		return pull, headRepo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, headRepo, nil
}

func (c *DefaultCommandRunner) buildLogger(repoFullName string, pullNum int) logging.SimpleLogging {

	return c.Logger.WithHistory(
		"repo", repoFullName,
		"pull", strconv.Itoa(pullNum),
	)
}

func (c *DefaultCommandRunner) ensureValidRepoMetadata(
	baseRepo models.Repo,
	maybeHeadRepo *models.Repo,
	maybePull *models.PullRequest,
	_ models.User,
	pullNum int,
	log logging.SimpleLogging,
) (headRepo models.Repo, pull models.PullRequest, err error) {
	if maybeHeadRepo != nil {
		headRepo = *maybeHeadRepo
	}

	switch baseRepo.VCSHost.Type {
	case models.Github:
		pull, headRepo, err = c.getGithubData(log, baseRepo, pullNum)
	case models.Gitlab:
		pull, err = c.getGitlabData(log, baseRepo, pullNum)
	case models.BitbucketCloud, models.BitbucketServer:
		if maybePull == nil {
			err = errors.New("pull request should not be nil–this is a bug")
			break
		}
		pull = *maybePull
	case models.AzureDevops:
		pull, headRepo, err = c.getAzureDevopsData(log, baseRepo, pullNum)
	case models.Gitea:
		pull, headRepo, err = c.getGiteaData(log, baseRepo, pullNum)
	default:
		err = errors.New("Unknown VCS type–this is a bug")
	}

	if err != nil {
		log.Err(err.Error())
		if commentErr := c.VCSClient.CreateComment(c.Logger, baseRepo, pullNum, fmt.Sprintf("`Error: %s`", err), ""); commentErr != nil {
			log.Err("unable to comment: %s", commentErr)
		}
	}

	return
}

func (c *DefaultCommandRunner) fetchUserTeams(repo models.Repo, user *models.User) error {
	teams, err := c.VCSClient.GetTeamNamesForUser(repo, *user)
	if err != nil {
		return err
	}

	user.Teams = teams
	return nil
}

func (c *DefaultCommandRunner) validateCtxAndComment(ctx *command.Context, commandName command.Name) bool {
	if !c.AllowForkPRs && ctx.HeadRepo.Owner != ctx.Pull.BaseRepo.Owner {
		if c.SilenceForkPRErrors {
			return false
		}
		ctx.Log.Info("command was run on a fork pull request which is disallowed")
		if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s  or, to disable this message, set --%s", c.AllowForkPRsFlag, c.SilenceForkPRErrorsFlag), ""); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
		return false
	}

	if ctx.Pull.State != models.OpenPullState && commandName != command.Unlock {
		ctx.Log.Info("command was run on closed pull request")
		if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, "Atlantis commands can't be run on closed pull requests", ""); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
		return false
	}

	repo := c.GlobalCfg.MatchingRepo(ctx.Pull.BaseRepo.ID())
	if !repo.BranchMatches(ctx.Pull.BaseBranch) {
		ctx.Log.Info("command was run on a pull request which doesn't match base branches")
		// just ignore it to allow us to use any git workflows without malicious intentions.
		return false
	}
	return true
}

// logPanics logs and creates a comment on the pull request for panics.
func (c *DefaultCommandRunner) logPanics(baseRepo models.Repo, pullNum int, logger logging.SimpleLogging) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		logger.Err("PANIC: %s\n%s", err, stack)
		if commentErr := c.VCSClient.CreateComment(
			logger,
			baseRepo,
			pullNum,
			fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack),
			"",
		); commentErr != nil {
			logger.Err("unable to comment: %s", commentErr)
		}
	}
}

var automergeComment = `Automatically merging because all plans have been successfully applied.`
