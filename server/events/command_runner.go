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
	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/recovery"
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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_gitlab_merge_request_getter.go GitlabMergeRequestGetter

// GitlabMergeRequestGetter makes API calls to get merge requests.
type GitlabMergeRequestGetter interface {
	// GetMergeRequest gets the pull request with the id pullNum for the repo.
	GetMergeRequest(repoFullName string, pullNum int) (*gitlab.MergeRequest, error)
}

// DefaultCommandRunner is the first step when processing a comment command.
type DefaultCommandRunner struct {
	VCSClient                vcs.ClientProxy
	GithubPullGetter         GithubPullGetter
	GitlabMergeRequestGetter GitlabMergeRequestGetter
	CommitStatusUpdater      CommitStatusUpdater
	EventParser              EventParsing
	MarkdownRenderer         *MarkdownRenderer
	Logger                   logging.SimpleLogging
	// AllowForkPRs controls whether we operate on pull requests from forks.
	AllowForkPRs bool
	// AllowForkPRsFlag is the name of the flag that controls fork PR's. We use
	// this in our error message back to the user on a forked PR so they know
	// how to enable this functionality.
	AllowForkPRsFlag      string
	ProjectCommandBuilder ProjectCommandBuilder
	ProjectCommandRunner  ProjectCommandRunner
	// RequireAllPlansSucceed is true if we require all plans succeed in each
	// run. If all plans don't succeed, we delete the ones that did.
	RequireAllPlansSucceed bool
	// Automerge is true if we should automatically merge pull requests if all
	// plans have been successfully applied.
	Automerge         bool
	PendingPlanFinder PendingPlanFinder
	WorkingDir        WorkingDir
	DB                *db.BoltDB
}

// RunAutoplanCommand runs plan when a pull request is opened or updated.
func (c *DefaultCommandRunner) RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) {
	log := c.buildLogger(baseRepo.FullName, pull.Num)
	defer c.logPanics(baseRepo, pull.Num, log)
	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		BaseRepo: baseRepo,
	}
	if !c.validateCtxAndComment(ctx) {
		return
	}
	if err := c.CommitStatusUpdater.Update(ctx.BaseRepo, ctx.Pull, models.PendingCommitStatus, PlanCommand); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	projectCmds, err := c.ProjectCommandBuilder.BuildAutoplanCommands(ctx)
	if err != nil {
		c.updatePull(ctx, AutoplanCommand{}, CommandResult{Error: err})
		return
	}
	if len(projectCmds) == 0 {
		log.Info("determined there was no project to run plan in")
		if err := c.CommitStatusUpdater.Update(baseRepo, pull, models.SuccessCommitStatus, PlanCommand); err != nil {
			ctx.Log.Warn("unable to update commit status: %s", err)
		}
		return
	}

	result := c.runProjectCmds(projectCmds, PlanCommand)
	if c.RequireAllPlansSucceed && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		c.deletePlans(ctx)
	}
	c.updatePull(ctx, AutoplanCommand{}, result)
	_, err = c.updateDB(ctx.Pull, result.ProjectResults)
	if err != nil {
		c.Logger.Err("writing results: %s", err)
	}
}

// RunCommentCommand executes the command.
// We take in a pointer for maybeHeadRepo because for some events there isn't
// enough data to construct the Repo model and callers might want to wait until
// the event is further validated before making an additional (potentially
// wasteful) call to get the necessary data.
func (c *DefaultCommandRunner) RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand) {
	log := c.buildLogger(baseRepo.FullName, pullNum)
	defer c.logPanics(baseRepo, pullNum, log)

	var headRepo models.Repo
	if maybeHeadRepo != nil {
		headRepo = *maybeHeadRepo
	}

	var err error
	var pull models.PullRequest
	switch baseRepo.VCSHost.Type {
	case models.Github:
		pull, headRepo, err = c.getGithubData(baseRepo, pullNum)
	case models.Gitlab:
		pull, err = c.getGitlabData(baseRepo, pullNum)
	case models.BitbucketCloud, models.BitbucketServer:
		if maybePull == nil {
			err = errors.New("pull request should not be nil–this is a bug")
		}
		pull = *maybePull
	default:
		err = errors.New("Unknown VCS type–this is a bug")
	}
	if err != nil {
		log.Err(err.Error())
		if commentErr := c.VCSClient.CreateComment(baseRepo, pullNum, fmt.Sprintf("`Error: %s`", err)); commentErr != nil {
			log.Err("unable to comment: %s", commentErr)
		}
		return
	}
	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		BaseRepo: baseRepo,
	}
	if !c.validateCtxAndComment(ctx) {
		return
	}
	if err = c.CommitStatusUpdater.Update(baseRepo, pull, models.PendingCommitStatus, cmd.CommandName()); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	var projectCmds []models.ProjectCommandContext
	switch cmd.Name {
	case PlanCommand:
		projectCmds, err = c.ProjectCommandBuilder.BuildPlanCommands(ctx, cmd)
	case ApplyCommand:
		projectCmds, err = c.ProjectCommandBuilder.BuildApplyCommands(ctx, cmd)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
		return
	}
	if err != nil {
		c.updatePull(ctx, cmd, CommandResult{Error: err})
		return
	}

	result := c.runProjectCmds(projectCmds, cmd.Name)
	if cmd.Name == PlanCommand && c.RequireAllPlansSucceed && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		c.deletePlans(ctx)
	}
	c.updatePull(
		ctx,
		cmd,
		result)

	pullStatus, err := c.updateDB(pull, result.ProjectResults)
	if err != nil {
		c.Logger.Err("writing results: %s", err)
		return
	}

	if cmd.Name == ApplyCommand && c.Automerge {
		c.automerge(ctx, pullStatus)
	}
}

func (c *DefaultCommandRunner) automerge(ctx *CommandContext, pullStatus *db.PullStatus) {
	// We only automerge if all projects have been successfully applied.
	for _, p := range pullStatus.Projects {
		if p.Status != db.AppliedPlanStatus {
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

		failureComment := fmt.Sprintf("Automerging failed: %s", err)
		if commentErr := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, failureComment); commentErr != nil {
			ctx.Log.Err("failed to comment about automerge failing: %s", err)
		}
	}
}

func (c *DefaultCommandRunner) runProjectCmds(cmds []models.ProjectCommandContext, cmdName CommandName) CommandResult {
	var results []models.ProjectResult
	for _, pCmd := range cmds {
		var res models.ProjectResult
		switch cmdName {
		case PlanCommand:
			res = c.ProjectCommandRunner.Plan(pCmd)
		case ApplyCommand:
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

func (c *DefaultCommandRunner) buildLogger(repoFullName string, pullNum int) *logging.SimpleLogger {
	src := fmt.Sprintf("%s#%d", repoFullName, pullNum)
	return c.Logger.NewLogger(src, true, c.Logger.GetLevel())
}

func (c *DefaultCommandRunner) validateCtxAndComment(ctx *CommandContext) bool {
	if !c.AllowForkPRs && ctx.HeadRepo.Owner != ctx.BaseRepo.Owner {
		ctx.Log.Info("command was run on a fork pull request which is disallowed")
		if err := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s", c.AllowForkPRsFlag)); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
		return false
	}

	if ctx.Pull.State != models.OpenPullState {
		ctx.Log.Info("command was run on closed pull request")
		if err := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, "Atlantis commands can't be run on closed pull requests"); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
		return false
	}
	return true
}

func (c *DefaultCommandRunner) updatePull(ctx *CommandContext, command PullCommand, res CommandResult) {
	// Log if we got any errors or failures.
	if res.Error != nil {
		ctx.Log.Err(res.Error.Error())
	} else if res.Failure != "" {
		ctx.Log.Warn(res.Failure)
	}

	// Update the pull request's status icon and comment back.
	if err := c.CommitStatusUpdater.UpdateProjectResult(ctx, command.CommandName(), res); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
	comment := c.MarkdownRenderer.Render(res, command.CommandName(), ctx.Log.History.String(), command.IsVerbose(), ctx.BaseRepo.VCSHost.Type)
	if err := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, comment); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
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

func (c *DefaultCommandRunner) updateDB(pull models.PullRequest, results []models.ProjectResult) (*db.PullStatus, error) {
	// Filter out results that errored due to the directory not existing. We
	// don't store these in the database because they would never be "applyable"
	// and so the pull request would always have errors.
	var filtered []models.ProjectResult
	for _, r := range results {
		if _, ok := r.Error.(DirNotExistErr); ok {
			continue
		}
		filtered = append(filtered, r)
	}

	return c.DB.UpdatePullWithResults(pull, filtered)
}

var automergeComment = `Automatically merging because all plans have been successfully applied.`
