package events

import (
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type CommitOutputUpdater interface {
	Update(ctx *command.Context, cmd PullCommand, res command.Result)
}

func NewCommitOutputUpdater(client vcs.Client, markdownRenderer *MarkdownRenderer, globalCfg valid.GlobalCfg, useCheckApi bool, hidePrevComments bool) CommitOutputUpdater {
	if useCheckApi {
		return &ChecksCommitOutputUpdater{
			VCSClient:        client,
			MarkdownRenderer: markdownRenderer,
			GlobalCfg:        globalCfg,
		}
	}

	return &PullCommitOutputUpdater{
		VCSClient:            client,
		MarkdownRenderer:     markdownRenderer,
		GlobalCfg:            globalCfg,
		HidePrevPlanComments: hidePrevComments,
	}
}

type ChecksCommitOutputUpdater struct {
	VCSClient        vcs.Client
	MarkdownRenderer *MarkdownRenderer
	GlobalCfg        valid.GlobalCfg
}

func (c *ChecksCommitOutputUpdater) Update(ctx *command.Context, cmd PullCommand, res command.Result) {
	logErrorsAndFailures(ctx, res)
	getTemplateOverridesForRepo(ctx.Pull.BaseRepo, c.GlobalCfg)

	// TODO: Update GithubChecks output
}

type PullCommitOutputUpdater struct {
	HidePrevPlanComments bool
	VCSClient            vcs.Client
	MarkdownRenderer     *MarkdownRenderer
	GlobalCfg            valid.GlobalCfg
}

func (c *PullCommitOutputUpdater) Update(ctx *command.Context, cmd PullCommand, res command.Result) {
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
