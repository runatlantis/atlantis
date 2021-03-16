package events

import "github.com/runatlantis/atlantis/server/events/vcs"

type PullUpdater struct {
	HidePrevPlanComments bool
	VCSClient            vcs.Client
	MarkdownRenderer     *MarkdownRenderer
}

func (c *PullUpdater) updatePull(ctx *CommandContext, command PullCommand, res CommandResult) {
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
		if err := c.VCSClient.HidePrevCommandComments(ctx.Pull.BaseRepo, ctx.Pull.Num, command.CommandName().TitleString()); err != nil {
			ctx.Log.Err("unable to hide old comments: %s", err)
		}
	}

	comment := c.MarkdownRenderer.Render(res, command.CommandName(), ctx.Log.GetHistory(), command.IsVerbose(), ctx.Pull.BaseRepo.VCSHost.Type)
	if err := c.VCSClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, comment, command.CommandName().String()); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}
