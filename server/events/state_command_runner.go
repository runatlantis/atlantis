// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewStateCommandRunner(
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
	prjCmdBuilder ProjectStateCommandBuilder,
	prjCmdRunner ProjectStateCommandRunner,
) *StateCommandRunner {
	return &StateCommandRunner{
		pullUpdater:          pullUpdater,
		dbUpdater:            dbUpdater,
		pullReqStatusFetcher: pullReqStatusFetcher,
		prjCmdBuilder:        prjCmdBuilder,
		prjCmdRunner:         prjCmdRunner,
	}
}

type StateCommandRunner struct {
	pullUpdater          *PullUpdater
	dbUpdater            *DBUpdater
	pullReqStatusFetcher vcs.PullReqStatusFetcher
	prjCmdBuilder        ProjectStateCommandBuilder
	prjCmdRunner         ProjectStateCommandRunner
}

func (v *StateCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var result command.Result
	switch cmd.SubName {
	case "rm":
		result = v.runRm(ctx, cmd)
	default:
		result = command.Result{
			Failure: fmt.Sprintf("unknown state subcommand %s", cmd.SubName),
		}
	}
	if ctx.CommandSkipped {
		return
	}
	if err := v.dbUpdater.updateDBForDiscardedPlans(ctx, ctx.Pull, result.ProjectResults); err != nil {
		result.Error = fmt.Errorf("writing discarded plan status: %w", err)
		ctx.CommandHasErrors = true
	}
	v.pullUpdater.updatePull(ctx, cmd, result)
}

func (v *StateCommandRunner) runRm(ctx *command.Context, cmd *CommentCommand) command.Result {
	// Fetch the approved and mergeable status before building the project
	// contexts, as import does, so the state rm requirement check sees the
	// pull request's real state rather than the zero value.
	var err error
	ctx.PullRequestStatus, err = v.pullReqStatusFetcher.FetchPullStatus(ctx.Log, ctx.Pull)
	if err != nil {
		// Continue with approved and mergeable assumed false: a requirement
		// that needs them then fails closed.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}

	projectCmds, err := v.prjCmdBuilder.BuildStateRmCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		return command.Result{}
	}
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}
	return runProjectCmds(projectCmds, v.prjCmdRunner.StateRm)
}

func (v *StateCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, v.prjCmdBuilder)
}
