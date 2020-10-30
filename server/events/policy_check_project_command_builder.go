package events

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

type PolicyCheckProjectCommandBuilder struct {
	projectCommandBuilder *DefaultProjectCommandBuilder
	workingDirLocker      WorkingDirLocker
}

func (p *PolicyCheckProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	projectCmds, err := p.projectCommandBuilder.BuildAutoplanCommands(ctx)
	if err != nil {
		return nil, err
	}

	policyCheckCmds, err := p.projectCommandBuilder.buildCommandsFromPlanFiles(ctx, models.PolicyCheckCommand, &CommentCommand{
		Verbose: false,
		Flags:   nil,
	})

	projectCmds = concat(projectCmds, policyChekcCmds)
	return policyChekcCmds, nil
}

func (p *PolicyCheckProjectCommandBuilder) BuildPlanCommands(ctx *CommandContext, cmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	projectCmds, err := p.projectCommandBuilder.BuildPlanCommands(ctx, cmd)
	if err != nil {
		return nil, err
	}

	policyCheckCmds, err := p.projectCommandBuilder.buildCommandsFromPlanFiles(ctx, models.PolicyCheckCommand, cmd)
	if err != nil {
		return nil, err
	}

	projectCmds = concat(projectCmds, policyChekcCmds)
	return policyChekcCmds, nil
}

func (p *PolicyCheckProjectCommandBuilder) BuildApplyCommands(ctx *CommandContext, cmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	return p.projectCommandBuilder.BuildApplyCommands(ctx, cmd)
}
