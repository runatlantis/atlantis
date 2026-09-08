// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"slices"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type PullUpdater struct {
	HidePrevPlanComments bool
	VCSClient            vcs.Client
	MarkdownRenderer     *MarkdownRenderer
}

func (c *PullUpdater) updatePull(ctx *command.Context, cmd PullCommand, res command.Result) error {
	// Log if we got any errors or failures.
	if res.Error != nil {
		ctx.Log.Err("%s", res.Error.Error())
	} else if res.Failure != "" {
		ctx.Log.Warn("%s", res.Failure)
	}

	// HidePrevCommandComments will hide old comments left from previous runs to reduce
	// clutter in a pull/merge request. This will not delete the comment, since the
	// comment trail may be useful in auditing or backtracing problems.
	if c.HidePrevPlanComments && (!ctx.PublicationRequired || ctx.TerminalPublisher != nil) {
		ctx.Log.Debug("hiding previous plan comments for command: '%v', directory: '%v'", cmd.CommandName().TitleString(), cmd.Dir())
		if err := publishTerminal(ctx, func() error {
			return c.VCSClient.HidePrevCommandComments(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, cmd.CommandName().TitleString(), cmd.Dir())
		}); err != nil {
			ctx.Log.Err("unable to hide old comments: %s", err)
			if ctx.PublicationRequired {
				return err
			}
		}
	}

	if len(res.ProjectResults) > 0 {
		var commentOnProjects []command.ProjectResult
		for _, result := range res.ProjectResults {
			if slices.Contains(result.SilencePRComments, cmd.CommandName().String()) {
				ctx.Log.Debug("silenced command '%s' comment for project '%s'", cmd.CommandName().String(), result.ProjectName)
				continue
			}
			commentOnProjects = append(commentOnProjects, result)
		}

		if len(commentOnProjects) == 0 {
			return nil
		}

		res.ProjectResults = commentOnProjects
	}

	comment := c.MarkdownRenderer.Render(ctx, res, cmd)
	// An acquisition/publication error needs a visible historical error comment;
	// it cannot hide previous comments or publish a successful command result.
	if ctx.PublicationRequired && ctx.TerminalPublisher == nil && res.HasErrors() {
		return c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, comment, cmd.CommandName().String())
	}
	return publishTerminal(ctx, func() error {
		if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, comment, cmd.CommandName().String()); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
			return err
		}
		return nil
	})
}
