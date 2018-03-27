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
	// ExecuteCommand is the first step after a command request has been parsed.
	// It handles gathering additional information needed to execute the command
	// and then calling the appropriate services to finish executing the command.
	ExecuteCommand(baseRepo models.Repo, headRepo models.Repo, user models.User, pullNum int, cmd *Command, vcsHost models.Host)
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

// CommandHandler is the first step when processing a comment command.
type CommandHandler struct {
	PlanExecutor             Executor
	ApplyExecutor            Executor
	LockURLGenerator         LockURLGenerator
	VCSClient                vcs.ClientProxy
	GithubPullGetter         GithubPullGetter
	GitlabMergeRequestGetter GitlabMergeRequestGetter
	CommitStatusUpdater      CommitStatusUpdater
	EventParser              EventParsing
	AtlantisWorkspaceLocker  AtlantisWorkspaceLocker
	MarkdownRenderer         *MarkdownRenderer
	Logger                   logging.SimpleLogging
	// AllowForkPRs controls whether we operate on pull requests from forks.
	AllowForkPRs bool
	// AllowForkPRsFlag is the name of the flag that controls fork PR's. We use
	// this in our error message back to the user on a forked PR so they know
	// how to enable this functionality.
	AllowForkPRsFlag string
}

// ExecuteCommand executes the command.
// If vcsHost is GitHub, we don't use headRepo and instead make an API call
// to get the headRepo. This is because the caller is unable to pass in a
// headRepo since there's not enough data available on the initial webhook
// payload.
func (c *CommandHandler) ExecuteCommand(baseRepo models.Repo, headRepo models.Repo, user models.User, pullNum int, cmd *Command, vcsHost models.Host) {
	var err error
	var pull models.PullRequest
	if vcsHost == models.Github {
		pull, headRepo, err = c.getGithubData(baseRepo, pullNum)
	} else if vcsHost == models.Gitlab {
		pull, err = c.getGitlabData(baseRepo.FullName, pullNum)
	}

	log := c.buildLogger(baseRepo.FullName, pullNum)
	if err != nil {
		log.Err(err.Error())
		return
	}
	ctx := &CommandContext{
		User:     user,
		Log:      log,
		Pull:     pull,
		HeadRepo: headRepo,
		Command:  cmd,
		VCSHost:  vcsHost,
		BaseRepo: baseRepo,
	}
	c.run(ctx)
}

func (c *CommandHandler) getGithubData(baseRepo models.Repo, pullNum int) (models.PullRequest, models.Repo, error) {
	if c.GithubPullGetter == nil {
		return models.PullRequest{}, models.Repo{}, errors.New("Atlantis not configured to support GitHub")
	}
	ghPull, err := c.GithubPullGetter.GetPullRequest(baseRepo, pullNum)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, errors.Wrap(err, "making pull request API call to GitHub")
	}
	pull, repo, err := c.EventParser.ParseGithubPull(ghPull)
	if err != nil {
		return pull, repo, errors.Wrap(err, "extracting required fields from comment data")
	}
	return pull, repo, nil
}

func (c *CommandHandler) getGitlabData(repoFullName string, pullNum int) (models.PullRequest, error) {
	if c.GitlabMergeRequestGetter == nil {
		return models.PullRequest{}, errors.New("Atlantis not configured to support GitLab")
	}
	mr, err := c.GitlabMergeRequestGetter.GetMergeRequest(repoFullName, pullNum)
	if err != nil {
		return models.PullRequest{}, errors.Wrap(err, "making merge request API call to GitLab")
	}
	pull := c.EventParser.ParseGitlabMergeRequest(mr)
	return pull, nil
}

func (c *CommandHandler) buildLogger(repoFullName string, pullNum int) *logging.SimpleLogger {
	src := fmt.Sprintf("%s#%d", repoFullName, pullNum)
	return logging.NewSimpleLogger(src, c.Logger.Underlying(), true, c.Logger.GetLevel())
}

// SetLockURL sets a function that's used to return the URL for a lock.
func (c *CommandHandler) SetLockURL(f func(id string) (url string)) {
	c.LockURLGenerator.SetLockURL(f)
}

func (c *CommandHandler) run(ctx *CommandContext) {
	log := c.buildLogger(ctx.BaseRepo.FullName, ctx.Pull.Num)
	ctx.Log = log
	defer c.logPanics(ctx)

	if !c.AllowForkPRs && ctx.HeadRepo.Owner != ctx.BaseRepo.Owner {
		ctx.Log.Info("command was run on a fork pull request which is disallowed")
		c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, fmt.Sprintf("Atlantis commands can't be run on fork pull requests. To enable, set --%s", c.AllowForkPRsFlag), ctx.VCSHost) // nolint: errcheck
		return
	}

	if ctx.Pull.State != models.Open {
		ctx.Log.Info("command was run on closed pull request")
		c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, "Atlantis commands can't be run on closed pull requests", ctx.VCSHost) // nolint: errcheck
		return
	}

	if err := c.CommitStatusUpdater.Update(ctx.BaseRepo, ctx.Pull, vcs.Pending, ctx.Command, ctx.VCSHost); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
	if !c.AtlantisWorkspaceLocker.TryLock(ctx.BaseRepo.FullName, ctx.Command.Workspace, ctx.Pull.Num) {
		errMsg := fmt.Sprintf(
			"The %s workspace is currently locked by another"+
				" command that is running for this pull request."+
				" Wait until the previous command is complete and try again.",
			ctx.Command.Workspace)
		ctx.Log.Warn(errMsg)
		c.updatePull(ctx, CommandResponse{Failure: errMsg})
		return
	}
	defer c.AtlantisWorkspaceLocker.Unlock(ctx.BaseRepo.FullName, ctx.Command.Workspace, ctx.Pull.Num)

	var cr CommandResponse
	switch ctx.Command.Name {
	case Plan:
		cr = c.PlanExecutor.Execute(ctx)
	case Apply:
		cr = c.ApplyExecutor.Execute(ctx)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
	}
	c.updatePull(ctx, cr)
}

func (c *CommandHandler) updatePull(ctx *CommandContext, res CommandResponse) {
	// Log if we got any errors or failures.
	if res.Error != nil {
		ctx.Log.Err(res.Error.Error())
	} else if res.Failure != "" {
		ctx.Log.Warn(res.Failure)
	}

	// Update the pull request's status icon and comment back.
	if err := c.CommitStatusUpdater.UpdateProjectResult(ctx, res); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
	comment := c.MarkdownRenderer.Render(res, ctx.Command.Name, ctx.Log.History.String(), ctx.Command.Verbose)
	c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, comment, ctx.VCSHost) // nolint: errcheck
}

// logPanics logs and creates a comment on the pull request for panics.
func (c *CommandHandler) logPanics(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		c.VCSClient.CreateComment(ctx.BaseRepo, ctx.Pull.Num, // nolint: errcheck
			fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack), ctx.VCSHost)
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
