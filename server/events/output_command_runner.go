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
	runCommentWorkflow(ctx, cmd, r.prjCmdBuilder.BuildOutputCommands, r.prjCmdRunner.Output, r.pullUpdater, r.silenceNoProjects)
}

func (r *OutputCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, r.prjCmdBuilder)
}
