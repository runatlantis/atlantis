// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import "github.com/runatlantis/atlantis/server/events/command"

type StateListCommandRunner struct {
	pullUpdater       *PullUpdater
	prjCmdBuilder     ProjectStateListCommandBuilder
	prjCmdRunner      ProjectStateListCommandRunner
	silenceNoProjects bool
}

func NewStateListCommandRunner(pullUpdater *PullUpdater, prjCmdBuilder ProjectStateListCommandBuilder, prjCmdRunner ProjectStateListCommandRunner, silenceNoProjects bool) *StateListCommandRunner {
	return &StateListCommandRunner{
		pullUpdater:       pullUpdater,
		prjCmdBuilder:     prjCmdBuilder,
		prjCmdRunner:      prjCmdRunner,
		silenceNoProjects: silenceNoProjects,
	}
}

func (r *StateListCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	cmds, err := r.prjCmdBuilder.BuildStateListCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		return
	}
	result := command.Result{Error: err}
	if err == nil {
		if len(cmds) == 0 && r.silenceNoProjects {
			return
		}
		result = runProjectCmds(cmds, r.prjCmdRunner.StateList)
	}
	r.pullUpdater.updatePull(ctx, cmd, result)
}

func (r *StateListCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, r.prjCmdBuilder)
}
