package events

import (
	"context"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
)

type OutputUpdater interface {
	Update(ctx *command.Context, cmd PullCommand, res command.Result)
}

func NewOutputUpdater(client vcs.Client, markdownRenderer *MarkdownRenderer, globalCfg valid.GlobalCfg, useCheckApi bool, hidePrevComments bool) OutputUpdater {
	if useCheckApi {
		return &ChecksOutputUpdater{
			VCSClient:        client,
			MarkdownRenderer: markdownRenderer,
			GlobalCfg:        globalCfg,
		}
	}

	return &PullOutputUpdater{
		VCSClient:            client,
		MarkdownRenderer:     markdownRenderer,
		GlobalCfg:            globalCfg,
		HidePrevPlanComments: hidePrevComments,
	}
}

type ChecksOutputUpdater struct {
	VCSClient        vcs.Client
	MarkdownRenderer *MarkdownRenderer
	GlobalCfg        valid.GlobalCfg
}

func (c *ChecksOutputUpdater) Update(ctx *command.Context, cmd PullCommand, res command.Result) {
	logErrorsAndFailures(ctx, res)
	templateOverrides := getTemplateOverridesForRepo(ctx.Pull.BaseRepo, c.GlobalCfg)

	// Update all projects
	for _, projectResult := range res.ProjectResults {
		comment := c.MarkdownRenderer.Render(res, cmd.CommandName(), ctx.Pull.BaseRepo.VCSHost.Type, templateOverrides)
		request := types.UpdateStatusRequest{
			Description: comment,
			StatusId:    projectResult.StatusId,
		}

		// Can be certain check Id is present
		if _, err := c.VCSClient.UpdateStatus(context.TODO(), request); err != nil {
			ctx.Log.Errorf("unable to comment: %s", err)
		}
	}
}

type PullOutputUpdater struct {
	HidePrevPlanComments bool
	VCSClient            vcs.Client
	MarkdownRenderer     *MarkdownRenderer
	GlobalCfg            valid.GlobalCfg
}

func (c *PullOutputUpdater) Update(ctx *command.Context, cmd PullCommand, res command.Result) {
	logErrorsAndFailures(ctx, res)
	templateOverrides := getTemplateOverridesForRepo(ctx.Pull.BaseRepo, c.GlobalCfg)

	// HidePrevCommandComments will hide old comments left from previous runs to reduce
	// clutter in a pull/merge request. This will not delete the comment, since the
	// comment trail may be useful in auditing or backtracing problems.
	if c.HidePrevPlanComments {
		if err := c.VCSClient.HidePrevCommandComments(ctx.Pull.BaseRepo, ctx.Pull.Num, cmd.CommandName().TitleString()); err != nil {
			ctx.Log.Errorf("unable to hide old comments: %s", err)
		}
	}

	comment := c.MarkdownRenderer.Render(res, cmd.CommandName(), ctx.Pull.BaseRepo.VCSHost.Type, templateOverrides)
	if err := c.VCSClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, comment, cmd.CommandName().String()); err != nil {
		ctx.Log.Errorf("unable to comment: %s", err)
	}
}

func logErrorsAndFailures(ctx *command.Context, res command.Result) {
	if res.Error != nil {
		ctx.Log.Errorf(res.Error.Error())
	} else if res.Failure != "" {
		ctx.Log.Warnf(res.Failure)
	}
}

func getTemplateOverridesForRepo(repo models.Repo, cfg valid.GlobalCfg) map[string]string {
	repoCfg := cfg.MatchingRepo(repo.ID())
	if repoCfg != nil {
		return repoCfg.TemplateOverrides
	}
	return map[string]string{}
}
