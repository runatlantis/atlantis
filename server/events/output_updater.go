package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
)

type OutputUpdater interface {
	UpdateOutput(ctx *command.Context, cmd PullCommand, res command.Result)
}

// [WENGINES-4643] TODO: Remove PullOutputUpdater and default to checks once github checks is stable
// defaults to pull comments if checks is turned off
type FeatureAwareChecksOutputUpdater struct {
	PullOutputUpdater
	ChecksOutputUpdater

	FeatureAllocator feature.Allocator
	Logger           logging.Logger
}

func (c *FeatureAwareChecksOutputUpdater) UpdateOutput(ctx *command.Context, cmd PullCommand, res command.Result) {
	shouldAllocate, err := c.FeatureAllocator.ShouldAllocate(feature.GithubChecks, ctx.HeadRepo.FullName)
	if err != nil {
		c.Logger.Error(fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.GithubChecks, err))
	}

	// Github Checks turned on and github provider
	if ctx.HeadRepo.VCSHost.Type == models.Github && shouldAllocate {
		c.ChecksOutputUpdater.UpdateOutput(ctx, cmd, res)
		return
	}
	c.PullOutputUpdater.UpdateOutput(ctx, cmd, res)
}

// Used to support checks type output (Github checks for example)
type ChecksOutputUpdater struct {
	VCSClient        vcs.Client
	MarkdownRenderer *MarkdownRenderer
	TitleBuilder     vcs.StatusTitleBuilder
	GlobalCfg        valid.GlobalCfg
}

func (c *ChecksOutputUpdater) UpdateOutput(ctx *command.Context, cmd PullCommand, res command.Result) {
	var templateOverrides map[string]string

	// retrieve template override if configured
	repoCfg := c.GlobalCfg.MatchingRepo(ctx.Pull.BaseRepo.ID())
	if repoCfg != nil {
		templateOverrides = repoCfg.TemplateOverrides
	}

	// iterate through all project results and the update the github check
	for _, projectResult := range res.ProjectResults {
		statusName := c.TitleBuilder.Build(cmd.CommandName().String(), vcs.StatusTitleOptions{
			ProjectName: projectResult.ProjectName,
		})

		output := c.MarkdownRenderer.Render(res, cmd.CommandName(), ctx.Pull.BaseRepo.VCSHost.Type, templateOverrides)
		updateStatusReq := types.UpdateStatusRequest{
			Repo:        ctx.HeadRepo,
			Ref:         ctx.Pull.HeadCommit,
			StatusName:  statusName,
			PullNum:     ctx.Pull.Num,
			Description: output,
		}

		if err := c.VCSClient.UpdateStatus(ctx.RequestCtx, updateStatusReq); err != nil {
			ctx.Log.Error("updable to update check run", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

}

// Default prj output updater which writes to the pull req comment
type PullOutputUpdater struct {
	HidePrevPlanComments bool
	VCSClient            vcs.Client
	MarkdownRenderer     *MarkdownRenderer
	GlobalCfg            valid.GlobalCfg
}

func (c *PullOutputUpdater) UpdateOutput(ctx *command.Context, cmd PullCommand, res command.Result) {
	// Log if we got any errors or failures.
	if res.Error != nil {
		ctx.Log.Error("", map[string]interface{}{
			"error": res.Error.Error(),
		})
	} else if res.Failure != "" {
		ctx.Log.Warn("", map[string]interface{}{
			"failiure": res.Failure,
		})
	}

	// HidePrevCommandComments will hide old comments left from previous runs to reduce
	// clutter in a pull/merge request. This will not delete the comment, since the
	// comment trail may be useful in auditing or backtracing problems.
	if c.HidePrevPlanComments {
		if err := c.VCSClient.HidePrevCommandComments(ctx.Pull.BaseRepo, ctx.Pull.Num, cmd.CommandName().TitleString()); err != nil {
			ctx.Log.Error("unable to hide old comments", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	var templateOverrides map[string]string
	repoCfg := c.GlobalCfg.MatchingRepo(ctx.Pull.BaseRepo.ID())
	if repoCfg != nil {
		templateOverrides = repoCfg.TemplateOverrides
	}

	comment := c.MarkdownRenderer.Render(res, cmd.CommandName(), ctx.Pull.BaseRepo.VCSHost.Type, templateOverrides)
	if err := c.VCSClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, comment, cmd.CommandName().String()); err != nil {
		ctx.Log.Error("unable to comment", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
