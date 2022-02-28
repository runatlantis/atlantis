package events

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
)

type SizeLimitedProjectCommandBuilder struct {
	Limit int
	ProjectCommandBuilder
}

func (b *SizeLimitedProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	projects, err := b.ProjectCommandBuilder.BuildAutoplanCommands(ctx)

	if err != nil {
		return projects, err
	}

	return projects, b.CheckAgainstLimit(projects)
}

func (b *SizeLimitedProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	projects, err := b.ProjectCommandBuilder.BuildPlanCommands(ctx, comment)

	if err != nil {
		return projects, err
	}

	return projects, b.CheckAgainstLimit(projects)
}

func (b *SizeLimitedProjectCommandBuilder) CheckAgainstLimit(projects []command.ProjectContext) error {

	var planCommands []command.ProjectContext

	for _, project := range projects {

		if project.CommandName == command.Plan {
			planCommands = append(planCommands, project)
		}
	}

	if b.Limit != InfiniteProjectsPerPR && len(planCommands) > b.Limit {
		return errors.New(
			fmt.Sprintf(
				"Number of projects cannot exceed %d.  This can either be caused by:\n"+
					"1) GH failure in recognizing the diff\n"+
					"2) Pull Request batch is too large for the given Atlantis instance\n\n"+
					"Please break this pull request into smaller batches and try again.",
				b.Limit,
			),
		)
	}
	return nil
}
