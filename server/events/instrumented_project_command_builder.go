package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/logging"
)

type InstrumentedProjectCommandBuilder struct {
	ProjectCommandBuilder
	Logger logging.Logger
}

func (b *InstrumentedProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	scope := ctx.Scope.SubScope("builder")

	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildApplyCommands(ctx, comment)

	if err != nil {
		executionError.Inc(1)
		b.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("Error building apply commands: %s", err))
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err

}
func (b *InstrumentedProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	scope := ctx.Scope.SubScope("builder")

	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildAutoplanCommands(ctx)

	if err != nil {
		executionError.Inc(1)
		b.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("Error building auto plan commands: %s", err))
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err

}
func (b *InstrumentedProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	scope := ctx.Scope.SubScope("builder")

	timer := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildPlanCommands(ctx, comment)

	if err != nil {
		executionError.Inc(1)
		b.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("Error building plan commands: %s", err))
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err

}
