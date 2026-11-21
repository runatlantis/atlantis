// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import "github.com/runatlantis/atlantis/server/events/command"

type StateShowCommandRunner struct {
	pullUpdater       *PullUpdater
	prjCmdBuilder     ProjectStateShowCommandBuilder
	prjCmdRunner      ProjectStateShowCommandRunner
	silenceNoProjects bool
}

func NewStateShowCommandRunner(pullUpdater *PullUpdater, prjCmdBuilder ProjectStateShowCommandBuilder, prjCmdRunner ProjectStateShowCommandRunner, silenceNoProjects bool) *StateShowCommandRunner {
	return &StateShowCommandRunner{
		pullUpdater:       pullUpdater,
		prjCmdBuilder:     prjCmdBuilder,
		prjCmdRunner:      prjCmdRunner,
		silenceNoProjects: silenceNoProjects,
	}
}

func (r *StateShowCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	runCommentWorkflow(ctx, cmd, r.prjCmdBuilder.BuildStateShowCommands, r.prjCmdRunner.StateShow, r.pullUpdater, r.silenceNoProjects)
}

func (r *StateShowCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, r.prjCmdBuilder)
}
