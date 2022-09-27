package events

import (
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewPRProjectCommandContextBuilder(
	commentBuilder CommentBuilder,
) ProjectCommandContextBuilder {
	return &prProjectContextBuilder{
		CommentBuilder: commentBuilder,
	}
}

type prProjectContextBuilder struct {
	ProjectCommandContextBuilder
	CommentBuilder CommentBuilder
}

func (p *prProjectContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	prjCfg valid.MergedProjectCfg,
	commentArgs []string,
	repoDir string,
	contextFlags *command.ContextFlags,
) []command.ProjectContext {
	return buildContext(
		ctx,
		cmdName,
		getSteps(cmdName, prjCfg.PullRequestWorkflow, contextFlags.LogLevel),
		p.CommentBuilder,
		prjCfg,
		commentArgs,
		repoDir,
		contextFlags,
	)
}
