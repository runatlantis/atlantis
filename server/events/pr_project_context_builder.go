package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
)

func NewPlatformModeProjectCommandContextBuilder(
	commentBuilder CommentBuilder,
	delegate ProjectCommandContextBuilder,
	logger logging.Logger,
	allocator feature.Allocator,
) *PlatformModeProjectContextBuilder {
	return &PlatformModeProjectContextBuilder{
		CommentBuilder: commentBuilder,
		delegate:       delegate,
		Logger:         logger,
		allocator:      allocator,
	}
}

type PlatformModeProjectContextBuilder struct {
	delegate       ProjectCommandContextBuilder
	allocator      feature.Allocator
	CommentBuilder CommentBuilder
	Logger         logging.Logger
}

func (p *PlatformModeProjectContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	prjCfg valid.MergedProjectCfg,
	commentArgs []string,
	repoDir string,
	contextFlags *command.ContextFlags,
) []command.ProjectContext {
	shouldAllocate, err := p.allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{RepoName: ctx.HeadRepo.FullName})
	if err != nil {
		p.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.PlatformMode, err))
	}

	if shouldAllocate && prjCfg.WorkflowMode == valid.PlatformWorkflowMode {
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

	return p.delegate.BuildProjectContext(ctx, cmdName, prjCfg, commentArgs, repoDir, contextFlags)
}
