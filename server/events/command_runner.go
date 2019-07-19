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
	"sync"

	"github.com/google/go-github/v31/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/pkg/errors"
	"github.com/remeh/sizedwaitgroup"
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/recovery"
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
	RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand)
	RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User)
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

// CommentCommandRunner runs individual command workflows.
type CommentCommandRunner interface {
	Run(*CommandContext, *CommentCommand)
}

func buildCommentCommandRunner(
	cmdRunner *DefaultCommandRunner,
	cmdName models.CommandName,
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
	Logger                   logging.SimpleLogging
	// AllowForkPRs controls whether we operate on pull requests from forks.
	AllowForkPRs bool
	// ParallelPoolSize controls the size of the wait group used to run
	// parallel plans and applies (if enabled).
	ParallelPoolSize int
	// AllowForkPRsFlag is the name of the flag that controls fork PR's. We use
	// this in our error message back to the user on a forked PR so they know
	// how to enable this functionality.
	AllowForkPRsFlag string
	// ParallelPlansPoolSize controls the size of the wait group used to run
	// parallel plans (if enabled).
	ParallelPlansPoolSize int
	ProjectCommandBuilder ProjectCommandBuilder
	ProjectCommandRunner  ProjectCommandRunner
	// GlobalAutomerge is true if we should automatically merge pull requests if all
	// plans have been successfully applied. This is set via a CLI flag.
	GlobalAutomerge   bool
	PendingPlanFinder PendingPlanFinder
	WorkingDir        WorkingDir
	DB                *db.BoltDB
}

// RunAutoplanCommand runs plan and policy_checks when a pull request is opened or updated.
func (c *DefaultCommandRunner) RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pull.Num, ShutdownComment, models.PlanCommand.String()); commentErr != nil {
			c.Logger.Log(logging.Error, "unable to comment that Atlantis is shutting down: %s", commentErr)
		}
		return
	}
	defer c.Drainer.OpDone()

	log := c.buildLogger(baseRepo.FullName, pull.Num)
	defer c.logPanics(baseRepo, pull.Num, log)
	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		Trigger:  Auto,
	}
	if !c.validateCtxAndComment(ctx) {
		return
	}
	if c.DisableAutoplan {
		return
	}

	// Run our plan commands in parallel if enabled
	var result CommandResult
	if c.parallelPlansEnabled(ctx, projectCmds) {
		ctx.Log.Info("Running plans in parallel")
		result = c.runProjectCmdsParallel(projectCmds, models.PlanCommand)
	} else {
		result = c.runProjectCmds(projectCmds, models.PlanCommand)
	}

	if c.automergeEnabled(ctx, projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		c.deletePlans(ctx)
		result.PlansDeleted = true
	}
	c.updatePull(ctx, AutoplanCommand{}, result)
	pullStatus, err := c.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		c.Logger.Err("writing results: %s", err)
	}

	autoPlanRunner.Run(ctx, nil)
}

// RunCommentCommand executes the command.
// We take in a pointer for maybeHeadRepo because for some events there isn't
// enough data to construct the Repo model and callers might want to wait until
// the event is further validated before making an additional (potentially
// wasteful) call to get the necessary data.
func (c *DefaultCommandRunner) RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pullNum, ShutdownComment, ""); commentErr != nil {
			c.Logger.Log(logging.Error, "unable to comment that Atlantis is shutting down: %s", commentErr)
		}
		return
	}
	defer c.Drainer.OpDone()

	log := c.buildLogger(baseRepo.FullName, pullNum)
	defer c.logPanics(baseRepo, pullNum, log)

	headRepo, pull, err := c.ensureValidRepoMetadata(baseRepo, maybeHeadRepo, maybePull, user, pullNum, log)
	if err != nil {
		return
	}

	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		Trigger:  Comment,
	}

	if !c.validateCtxAndComment(ctx) {
		return
	}

	// Run our plan commands in parallel if enabled
	var result CommandResult
	if cmd.Name == models.PlanCommand && c.parallelPlansEnabled(ctx, projectCmds) {
		ctx.Log.Info("Running plans in parallel")
		result = c.runProjectCmdsParallel(projectCmds, cmd.Name)
	} else {
		result = c.runProjectCmds(projectCmds, cmd.Name)
	}

	if cmd.Name == models.PlanCommand && c.automergeEnabled(ctx, projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		c.deletePlans(ctx)
		result.PlansDeleted = true
	}
	c.updatePull(
		ctx,
		cmd,
		result)

	pullStatus, err := c.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		c.Logger.Err("writing results: %s", err)
		return
	}

	c.updateCommitStatus(ctx, cmd.Name, pullStatus)

	if cmd.Name == models.ApplyCommand && c.automergeEnabled(ctx, projectCmds) {
		c.automerge(ctx, pullStatus)
	}
}

func (c *DefaultCommandRunner) updateCommitStatus(ctx *CommandContext, cmd models.CommandName, pullStatus models.PullStatus) {
	var numSuccess int
	var status models.CommitStatus

	if cmd == models.PlanCommand {
		// We consider anything that isn't a plan error as a plan success.
		// For example, if there is an apply error, that means that at least a
		// plan was generated successfully.
		numSuccess = len(pullStatus.Projects) - pullStatus.StatusCount(models.ErroredPlanStatus)
		status = models.SuccessCommitStatus
		if numSuccess != len(pullStatus.Projects) {
			status = models.FailedCommitStatus
		}
	} else {
		numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus)

		numErrored := pullStatus.StatusCount(models.ErroredApplyStatus)
		status = models.SuccessCommitStatus
		if numErrored > 0 {
			status = models.FailedCommitStatus
		} else if numSuccess < len(pullStatus.Projects) {
			// If there are plans that haven't been applied yet, we'll use a pending
			// status.
			status = models.PendingCommitStatus
		}
	}

	if err := c.CommitStatusUpdater.UpdateCombinedCount(ctx.BaseRepo, ctx.Pull, status, cmd, numSuccess, len(pullStatus.Projects)); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}

func (c *DefaultCommandRunner) automerge(ctx *CommandContext, pullStatus models.PullStatus) {
	// We only automerge if all projects have been successfully applied.
	for _, p := range pullStatus.Projects {
		if p.Status != models.AppliedPlanStatus {
			ctx.Log.Info("not automerging because project at dir %q, workspace %q has status %q", p.RepoRelDir, p.Workspace, p.Status.String())
			return
		}
	}

	// Comment that we're automerging the pull request.
	if err := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, automergeComment); err != nil {
		ctx.Log.Err("failed to comment about automerge: %s", err)
		// Commenting isn't required so continue.
	}

	// Make the API call to perform the merge.
	ctx.Log.Info("automerging pull request")
	err := c.VCSClient.MergePull(ctx.Pull)

	if err != nil {
		ctx.Log.Err("automerging failed: %s", err)

		failureComment := fmt.Sprintf("Automerging failed:\n```\n%s\n```", err)
		if commentErr := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, failureComment); commentErr != nil {
			ctx.Log.Err("failed to comment about automerge failing: %s", err)
		}
	}
}

func (c *DefaultCommandRunner) runProjectCmdsParallel(cmds []models.ProjectCommandContext, cmdName models.CommandName) CommandResult {
	var results []models.ProjectResult
	mux := &sync.Mutex{}

	wg := sizedwaitgroup.New(c.ParallelPlansPoolSize)
	for _, pCmd := range cmds {
		pCmd := pCmd
		var execute func()
		wg.Add()

		switch cmdName {
		case models.PlanCommand:
			execute = func() {
				defer wg.Done()
				res := c.ProjectCommandRunner.Plan(pCmd)
				mux.Lock()
				results = append(results, res)
				mux.Unlock()
			}
		case models.ApplyCommand:
			execute = func() {
				defer wg.Done()
				res := c.ProjectCommandRunner.Apply(pCmd)
				mux.Lock()
				results = append(results, res)
				mux.Unlock()
			}
		}
		go execute()
	}

	wg.Wait()
	return CommandResult{ProjectResults: results}
}

func (c *DefaultCommandRunner) runProjectCmds(cmds []models.ProjectCommandContext, cmdName models.CommandName) CommandResult {
	var results []models.ProjectResult
	for _, pCmd := range cmds {
		var res models.ProjectResult
		switch cmdName {
		case models.PlanCommand:
			res = c.ProjectCommandRunner.Plan(pCmd)
		case models.ApplyCommand:
			res = c.ProjectCommandRunner.Apply(pCmd)
		}
		results = append(results, res)
	}
	return CommandResult{ProjectResults: results}
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

func (c *DefaultCommandRunner) buildLogger(repoFullName string, pullNum int) *logging.SimpleLogger {
	src := fmt.Sprintf("%s#%d", repoFullName, pullNum)
	return c.Logger.NewLogger(src, true, c.Logger.GetLevel())
}

func (c *DefaultCommandRunner) ensureValidRepoMetadata(
	baseRepo models.Repo,
	maybeHeadRepo *models.Repo,
	maybePull *models.PullRequest,
	user models.User,
	pullNum int,
	log *logging.SimpleLogger,
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
		log.Err(err.Error())
		if commentErr := c.VCSClient.CreateComment(baseRepo, pullNum, fmt.Sprintf("`Error: %s`", err), ""); commentErr != nil {
			log.Err("unable to comment: %s", commentErr)
		}
	}

	return
}

func (c *DefaultCommandRunner) validateCtxAndComment(ctx *CommandContext) bool {
	if !c.AllowForkPRs && ctx.HeadRepo.Owner != ctx.Pull.BaseRepo.Owner {
		if c.SilenceForkPRErrors {
			return false
		}
		ctx.Log.Info("command was run on a fork pull request which is disallowed")
		if err := c.VCSClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s  or, to disable this message, set --%s", c.AllowForkPRsFlag, c.SilenceForkPRErrorsFlag), ""); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
		return false
	}

	if ctx.Pull.State != models.OpenPullState {
		ctx.Log.Info("command was run on closed pull request")
		if err := c.VCSClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, "Atlantis commands can't be run on closed pull requests", ""); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
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
			baseRepo,
			pullNum,
			fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack),
			"",
		); commentErr != nil {
			logger.Err("unable to comment: %s", commentErr)
		}
	}
}

// deletePlans deletes all plans generated in this ctx.
func (c *DefaultCommandRunner) deletePlans(ctx *CommandContext) {
	pullDir, err := c.WorkingDir.GetPullDir(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		ctx.Log.Err("getting pull dir: %s", err)
	}
	if err := c.PendingPlanFinder.DeletePlans(pullDir); err != nil {
		ctx.Log.Err("deleting pending plans: %s", err)
	}
}

func (c *DefaultCommandRunner) updateDB(ctx *CommandContext, pull models.PullRequest, results []models.ProjectResult) (models.PullStatus, error) {
	// Filter out results that errored due to the directory not existing. We
	// don't store these in the database because they would never be "applyable"
	// and so the pull request would always have errors.
	var filtered []models.ProjectResult
	for _, r := range results {
		if _, ok := r.Error.(DirNotExistErr); ok {
			ctx.Log.Debug("ignoring error result from project at dir %q workspace %q because it is dir not exist error", r.RepoRelDir, r.Workspace)
			continue
		}
		filtered = append(filtered, r)
	}
	ctx.Log.Debug("updating DB with pull results")
	return c.DB.UpdatePullWithResults(pull, filtered)
}

// automergeEnabled returns true if automerging is enabled in this context.
func (c *DefaultCommandRunner) automergeEnabled(ctx *CommandContext, projectCmds []models.ProjectCommandContext) bool {
	// If the global automerge is set, we always automerge.
	return c.GlobalAutomerge ||
		// Otherwise we check if this repo is configured for automerging.
		(len(projectCmds) > 0 && projectCmds[0].AutomergeEnabled)
}

// parallelPlansEnabled returns true if parallel plans is enabled in this context.
func (c *DefaultCommandRunner) parallelPlansEnabled(ctx *CommandContext, projectCmds []models.ProjectCommandContext) bool {
	return len(projectCmds) > 0 && projectCmds[0].ParallelPlansEnabled
}

// automergeComment is the comment that gets posted when Atlantis automatically
// merges the PR.
var automergeComment = `Automatically merging because all plans have been successfully applied.`
