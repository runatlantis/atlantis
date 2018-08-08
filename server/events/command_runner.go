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
//
package events

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/pkg/errors"
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
}

func (c *DefaultCommandRunner) RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) {
	log := c.buildLogger(baseRepo.FullName, pull.Num)
	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		BaseRepo: baseRepo,
	}
	defer c.logPanics(ctx)
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

	results := c.runProjectCmds(projectCmds, PlanCommand)
	c.updatePull(ctx, AutoplanCommand{}, CommandResult{ProjectResults: results})
}

// RunCommentCommand executes the command.
// We take in a pointer for maybeHeadRepo because for some events there isn't
// enough data to construct the Repo model and callers might want to wait until
// the event is further validated before making an additional (potentially
// wasteful) call to get the necessary data.
func (c *DefaultCommandRunner) RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand) {
	log := c.buildLogger(baseRepo.FullName, pullNum)
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
			err = errors.New("pull request should not be nil, this is a bug!")
		}
		pull = *maybePull
	default:
		err = errors.New("Unknown VCS type, this is a bug!")
	}
	if err != nil {
		log.Err(err.Error())
		return
	}
	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		BaseRepo: baseRepo,
	}
	defer c.logPanics(ctx)
	if !c.validateCtxAndComment(ctx) {
		return
	}
	if err := c.CommitStatusUpdater.Update(ctx.BaseRepo, ctx.Pull, models.PendingCommitStatus, cmd.CommandName()); err != nil {
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
	results := c.runProjectCmds(projectCmds, cmd.Name)
	c.updatePull(
		ctx,
		cmd,
		CommandResult{
			ProjectResults: results})
}

func (c *DefaultCommandRunner) runProjectCmds(cmds []models.ProjectCommandContext, cmdName CommandName) []ProjectResult {
	var results []ProjectResult
	for _, pCmd := range cmds {
		var res ProjectResult
		switch cmdName {
		case PlanCommand:
			res = c.ProjectCommandRunner.Plan(pCmd)
		case ApplyCommand:
			res = c.ProjectCommandRunner.Apply(pCmd)
		}
		results = append(results, res)
	}
	return results
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
	return logging.NewSimpleLogger(src, c.Logger.Underlying(), true, c.Logger.GetLevel())
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

func (c *DefaultCommandRunner) updatePull(ctx *CommandContext, command CommandInterface, res CommandResult) {
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
	comment := c.MarkdownRenderer.Render(res, command.CommandName(), ctx.Log.History.String(), command.IsVerbose())
	if err := c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, comment); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}

// logPanics logs and creates a comment on the pull request for panics.
func (c *DefaultCommandRunner) logPanics(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, // nolint: errcheck
			fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
