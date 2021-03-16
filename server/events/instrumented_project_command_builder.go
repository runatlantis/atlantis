package events

import (
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type InstrumentedProjectCommandBuilder struct {
	ProjectCommandBuilder
	Logger logging.SimpleLogging
}

func (b *InstrumentedProjectCommandBuilder) BuildApplyCommands(ctx *CommandContext, comment *CommentCommand) ([]models.ProjectCommandContext, error) {
	scope := ctx.Scope.Scope("builder")

	timer := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer timer.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildApplyCommands(ctx, comment)

	if err != nil {
		executionError.Inc()
		b.Logger.Err("Error building apply commands: %s", err)
	} else {
		executionSuccess.Inc()
	}

	return projectCmds, err

}
func (b *InstrumentedProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	scope := ctx.Scope.Scope("builder")

	timer := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer timer.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildAutoplanCommands(ctx)

	if err != nil {
		executionError.Inc()
		b.Logger.Err("Error building auto plan commands: %s", err)
	} else {
		executionSuccess.Inc()
	}

	return projectCmds, err

}
func (b *InstrumentedProjectCommandBuilder) BuildPlanCommands(ctx *CommandContext, comment *CommentCommand) ([]models.ProjectCommandContext, error) {
	scope := ctx.Scope.Scope("builder")

	timer := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer timer.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildPlanCommands(ctx, comment)

	if err != nil {
		executionError.Inc()
		b.Logger.Err("Error building plan commands: %s", err)
	} else {
		executionSuccess.Inc()
	}

	return projectCmds, err

}
