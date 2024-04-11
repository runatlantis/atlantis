package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type PullUpdater struct {
	HidePrevPlanComments bool
	VCSClient            vcs.Client
	MarkdownRenderer     *MarkdownRenderer
}

func (c *PullUpdater) updatePull(ctx *command.Context, cmd PullCommand, res command.Result) {
	// Log if we got any errors or failures.
	if res.Error != nil {
		ctx.Log.Err(res.Error.Error())
	} else if res.Failure != "" {
		ctx.Log.Warn(res.Failure)
	}

	// HidePrevCommandComments will hide old comments left from previous runs to reduce
	// clutter in a pull/merge request. This will not delete the comment, since the
	// comment trail may be useful in auditing or backtracing problems.
	if c.HidePrevPlanComments {
		ctx.Log.Debug("Hiding previous plan comments for command: '%v', directory: '%v'", cmd.CommandName().TitleString(), cmd.Dir())
		if err := c.VCSClient.HidePrevCommandComments(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, cmd.CommandName().TitleString(), cmd.Dir()); err != nil {
			ctx.Log.Err("unable to hide old comments: %s", err)
		}
	}

	comment := c.MarkdownRenderer.Render(ctx, res, cmd)
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, comment, cmd.CommandName().String()); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}
