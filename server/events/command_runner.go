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
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/logging/fields"
	"github.com/runatlantis/atlantis/server/recovery"
	"github.com/uber-go/tally/v4"
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
	RunCommentCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, pullNum int, cmd *command.Comment, timestamp time.Time, installationToken int64)
	RunAutoplanCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, timestamp time.Time, installationToken int64)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_stale_command_checker.go StaleCommandChecker

// StaleCommandChecker handles checks to validate if current command is stale and can be dropped.
type StaleCommandChecker interface {
	// CommandIsStale returns true if currentEventTimestamp is earlier than timestamp set in DB's latest pull model.
	CommandIsStale(ctx *command.Context) bool
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
	VCSClient       vcs.Client
	DisableAutoplan bool
	GlobalCfg       valid.GlobalCfg
	StatsScope      tally.Scope
	// ParallelPoolSize controls the size of the wait group used to run
	// parallel plans and applies (if enabled).
	ParallelPoolSize              int
	CommentCommandRunnerByCmd     map[command.Name]command.Runner
	Drainer                       *Drainer
	PreWorkflowHooksCommandRunner PreWorkflowHooksCommandRunner
	VCSStatusUpdater              VCSStatusUpdater
	PullStatusFetcher             PullStatusFetcher
	StaleCommandChecker           StaleCommandChecker
	Logger                        logging.Logger
}

// RunAutoplanCommand runs plan and policy_checks when a pull request is opened or updated.
func (c *DefaultCommandRunner) RunAutoplanCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, timestamp time.Time, installationToken int64) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pull.Num, ShutdownComment, command.Plan.String()); commentErr != nil {
			c.Logger.ErrorContext(ctx, commentErr.Error())
		}
		return
	}
	defer c.Drainer.OpDone()

	defer c.logPanics(ctx)
	status, err := c.PullStatusFetcher.GetPullStatus(pull)

	if err != nil {
		c.Logger.ErrorContext(ctx, err.Error())
	}

	scope := c.StatsScope.SubScope("autoplan")
	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	cmdCtx := &command.Context{
		User:              user,
		Log:               c.Logger,
		Scope:             scope,
		Pull:              pull,
		HeadRepo:          headRepo,
		PullStatus:        status,
		Trigger:           command.AutoTrigger,
		TriggerTimestamp:  timestamp,
		RequestCtx:        ctx,
		InstallationToken: installationToken,
	}
	if !c.validateCtxAndComment(cmdCtx) {
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
		_, err := c.VCSStatusUpdater.UpdateCombined(ctx, cmdCtx.HeadRepo, cmdCtx.Pull, models.FailedVCSStatus, command.Plan, "", err.Error())
		if err != nil {
			c.Logger.ErrorContext(ctx, err.Error())
		}
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
func (c *DefaultCommandRunner) RunCommentCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, pullNum int, cmd *command.Comment, timestamp time.Time, installationToken int64) {
	if opStarted := c.Drainer.StartOp(); !opStarted {
		if commentErr := c.VCSClient.CreateComment(baseRepo, pullNum, ShutdownComment, ""); commentErr != nil {
			c.Logger.ErrorContext(ctx, commentErr.Error())
		}
		return
	}
	defer c.Drainer.OpDone()

	defer c.logPanics(ctx)

	scope := c.StatsScope.SubScope("comment")

	if cmd != nil {
		scope = scope.SubScope(cmd.Name.String())
	}
	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	status, err := c.PullStatusFetcher.GetPullStatus(pull)

	if err != nil {
		c.Logger.ErrorContext(ctx, err.Error())
	}

	cmdCtx := &command.Context{
		User:              user,
		Log:               c.Logger,
		Pull:              pull,
		PullStatus:        status,
		HeadRepo:          headRepo,
		Trigger:           command.CommentTrigger,
		Scope:             scope,
		TriggerTimestamp:  timestamp,
		RequestCtx:        ctx,
		InstallationToken: installationToken,
	}

	if !c.validateCtxAndComment(cmdCtx) {
		return
	}

	// Drop request if a more recent VCS event updated Atlantis state
	if c.StaleCommandChecker.CommandIsStale(cmdCtx) {
		return
	}

	if err := c.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmdCtx); err != nil {
		// Replace approve policies command with policy check if preworkflow hook fails since we don't use
		// approve policies statuses
		cmdName := cmd.Name
		if cmdName == command.ApprovePolicies {
			cmdName = command.PolicyCheck
		}

		c.Logger.ErrorContext(ctx, "Error running pre-workflow hooks", fields.PullRequestWithErr(pull, err))
		_, err := c.VCSStatusUpdater.UpdateCombined(ctx, cmdCtx.HeadRepo, cmdCtx.Pull, models.FailedVCSStatus, cmdName, "", err.Error())
		if err != nil {
			c.Logger.ErrorContext(ctx, err.Error())
		}
		return
	}

	cmdRunner := buildCommentCommandRunner(c, cmd.CommandName())

	cmdRunner.Run(cmdCtx, cmd)
}

func (c *DefaultCommandRunner) validateCtxAndComment(cmdCtx *command.Context) bool {
	if cmdCtx.HeadRepo.Owner != cmdCtx.Pull.BaseRepo.Owner {
		c.Logger.InfoContext(cmdCtx.RequestCtx, "command was run on a fork pull request which is disallowed")
		if err := c.VCSClient.CreateComment(cmdCtx.Pull.BaseRepo, cmdCtx.Pull.Num, "Atlantis commands can't be run on fork pull requests.", ""); err != nil {
			c.Logger.ErrorContext(cmdCtx.RequestCtx, err.Error())
		}
		return false
	}

	if cmdCtx.Pull.State != models.OpenPullState {
		c.Logger.InfoContext(cmdCtx.RequestCtx, "command was run on closed pull request")
		if err := c.VCSClient.CreateComment(cmdCtx.Pull.BaseRepo, cmdCtx.Pull.Num, "Atlantis commands can't be run on closed pull requests", ""); err != nil {
			c.Logger.ErrorContext(cmdCtx.RequestCtx, err.Error())
		}
		return false
	}

	repo := c.GlobalCfg.MatchingRepo(cmdCtx.Pull.BaseRepo.ID())
	if !repo.BranchMatches(cmdCtx.Pull.BaseBranch) {
		c.Logger.InfoContext(cmdCtx.RequestCtx, "command was run on a pull request which doesn't match base branches")
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

func (f *ForceApplyCommandRunner) RunCommentCommand(ctx context.Context, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User, pullNum int, cmd *command.Comment, timestamp time.Time, installationToken int64) {
	if cmd.ForceApply {
		warningMessage := "⚠️ WARNING ⚠️\n\n You have bypassed all apply requirements for this PR 🚀 . This can have unpredictable consequences 🙏🏽 and should only be used in an emergency 🆘 .\n\n 𝐓𝐡𝐢𝐬 𝐚𝐜𝐭𝐢𝐨𝐧 𝐰𝐢𝐥𝐥 𝐛𝐞 𝐚𝐮𝐝𝐢𝐭𝐞𝐝.\n"
		if commentErr := f.VCSClient.CreateComment(baseRepo, pullNum, warningMessage, ""); commentErr != nil {
			f.Logger.ErrorContext(ctx, commentErr.Error())
		}
	}
	f.CommandRunner.RunCommentCommand(ctx, baseRepo, headRepo, pull, user, pullNum, cmd, timestamp, installationToken)
}
