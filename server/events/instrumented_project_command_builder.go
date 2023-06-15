package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

type InstrumentedProjectCommandBuilder struct {
	ProjectCommandBuilder
	Logger logging.SimpleLogging
	scope  tally.Scope
}

func (b *InstrumentedProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return b.buildAndEmitStats(
		"apply",
		func() ([]command.ProjectContext, error) {
			return b.ProjectCommandBuilder.BuildApplyCommands(ctx, comment)
		},
	)
}

func (b *InstrumentedProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	return b.buildAndEmitStats(
		"auto plan",
		func() ([]command.ProjectContext, error) {
			return b.ProjectCommandBuilder.BuildAutoplanCommands(ctx)
		},
	)
}

func (b *InstrumentedProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return b.buildAndEmitStats(
		"plan",
		func() ([]command.ProjectContext, error) {
			return b.ProjectCommandBuilder.BuildPlanCommands(ctx, comment)
		},
	)
}

func (b *InstrumentedProjectCommandBuilder) BuildImportCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return b.buildAndEmitStats(
		"import",
		func() ([]command.ProjectContext, error) {
			return b.ProjectCommandBuilder.BuildImportCommands(ctx, comment)
		},
	)
}

func (b *InstrumentedProjectCommandBuilder) BuildStateRmCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return b.buildAndEmitStats(
		"state rm",
		func() ([]command.ProjectContext, error) {
			return b.ProjectCommandBuilder.BuildStateRmCommands(ctx, comment)
		},
	)
}

func (b *InstrumentedProjectCommandBuilder) buildAndEmitStats(
	command string,
	execute func() ([]command.ProjectContext, error),
) ([]command.ProjectContext, error) {
	timer := b.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := b.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := b.scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := execute()

	if err != nil {
		executionError.Inc(1)
		b.Logger.Err("Error building %s commands: %s", command, err)
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err
}
