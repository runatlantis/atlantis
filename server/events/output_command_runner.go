// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import "github.com/runatlantis/atlantis/server/events/command"

type OutputCommandRunner struct {
	pullUpdater       *PullUpdater
	prjCmdBuilder     ProjectOutputCommandBuilder
	prjCmdRunner      ProjectOutputCommandRunner
	silenceNoProjects bool
}

func NewOutputCommandRunner(pullUpdater *PullUpdater, prjCmdBuilder ProjectOutputCommandBuilder, prjCmdRunner ProjectOutputCommandRunner, silenceNoProjects bool) *OutputCommandRunner {
	return &OutputCommandRunner{
		pullUpdater:       pullUpdater,
		prjCmdBuilder:     prjCmdBuilder,
		prjCmdRunner:      prjCmdRunner,
		silenceNoProjects: silenceNoProjects,
	}
}

func (r *OutputCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	cmds, err := r.prjCmdBuilder.BuildOutputCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		return
	}
	result := command.Result{Error: err}
	if err == nil {
		if len(cmds) == 0 && r.silenceNoProjects {
			return
		}
		result = runProjectCmds(cmds, r.prjCmdRunner.Output)
	}
	r.pullUpdater.updatePull(ctx, cmd, result)
}

func (r *OutputCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, r.prjCmdBuilder)
}
