package events

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

func NewPolicyCheckProjectCommandBuilder(p *DefaultProjectCommandBuilder) *PolicyCheckProjectCommandBuilder {
	return &PolicyCheckProjectCommandBuilder{
		ProjectCommandBuilder: p,
		ParserValidator:       p.ParserValidator,
		ProjectFinder:         p.ProjectFinder,
		VCSClient:             p.VCSClient,
		WorkingDir:            p.WorkingDir,
		WorkingDirLocker:      p.WorkingDirLocker,
		CommentBuilder:        p.CommentBuilder,
		GlobalCfg:             p.GlobalCfg,
	}
}

type PolicyCheckProjectCommandBuilder struct {
	ProjectCommandBuilder *DefaultProjectCommandBuilder
	ParserValidator       *yaml.ParserValidator
	ProjectFinder         ProjectFinder
	VCSClient             vcs.Client
	WorkingDir            WorkingDir
	WorkingDirLocker      WorkingDirLocker
	GlobalCfg             valid.GlobalCfg
	PendingPlanFinder     *DefaultPendingPlanFinder
	CommentBuilder        CommentBuilder
	SkipCloneNoChanges    bool
}

func (p *PolicyCheckProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	projectCmds, err := p.ProjectCommandBuilder.BuildAutoplanCommands(ctx)
	if err != nil {
		return nil, err
	}

	commentCmd := &CommentCommand{
		Verbose: false,
		Flags:   nil,
	}

	policyCheckCmds, err := p.buildProjectCommands(ctx, models.PolicyCheckCommand, commentCmd)

	projectCmds = append(projectCmds, policyCheckCmds...)
	return policyCheckCmds, nil
}

func (p *PolicyCheckProjectCommandBuilder) BuildPlanCommands(ctx *CommandContext, commentCmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	projectCmds, err := p.ProjectCommandBuilder.BuildPlanCommands(ctx, commentCmd)
	if err != nil {
		return nil, err
	}

	policyCheckCmds, err := p.buildProjectCommands(ctx, models.PolicyCheckCommand, commentCmd)

	if err != nil {
		return nil, err
	}

	projectCmds = append(projectCmds, policyCheckCmds...)
	return policyCheckCmds, nil
}

func (p *PolicyCheckProjectCommandBuilder) BuildApplyCommands(ctx *CommandContext, commentCmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	return p.ProjectCommandBuilder.BuildApplyCommands(ctx, commentCmd)
}

func (p *PolicyCheckProjectCommandBuilder) buildProjectCommands(ctx *CommandContext, cmdName models.CommandName, commentCmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	policyCheckCmds, err := buildProjectCommands(
		ctx,
		models.PolicyCheckCommand,
		commentCmd,
		p.CommentBuilder,
		p.GlobalCfg,
		p.WorkingDirLocker,
		p.WorkingDir,
	)

	if err != nil {
		return nil, err
	}

	return policyCheckCmds, nil
}
