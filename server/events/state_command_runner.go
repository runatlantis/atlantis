// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
)

func NewStateCommandRunner(
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	prjCmdBuilder ProjectStateCommandBuilder,
	prjCmdRunner ProjectStateCommandRunner,
) *StateCommandRunner {
	return &StateCommandRunner{
		pullUpdater:   pullUpdater,
		dbUpdater:     dbUpdater,
		prjCmdBuilder: prjCmdBuilder,
		prjCmdRunner:  prjCmdRunner,
	}
}

type StateCommandRunner struct {
	pullUpdater   *PullUpdater
	dbUpdater     *DBUpdater
	prjCmdBuilder ProjectStateCommandBuilder
	prjCmdRunner  ProjectStateCommandRunner
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
