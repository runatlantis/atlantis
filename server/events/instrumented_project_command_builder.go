package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally"
)

type InstrumentedProjectCommandBuilder struct {
	ProjectCommandBuilder
	Logger logging.SimpleLogging
	scope  tally.Scope
}

func (b *InstrumentedProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	timer := b.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := b.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := b.scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildApplyCommands(ctx, comment)

	if err != nil {
		executionError.Inc(1)
		b.Logger.Err("Error building apply commands: %s", err)
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err

}
func (b *InstrumentedProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	timer := b.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := b.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := b.scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildAutoplanCommands(ctx)

	if err != nil {
		executionError.Inc(1)
		b.Logger.Err("Error building auto plan commands: %s", err)
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err

}
func (b *InstrumentedProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	timer := b.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()

	executionSuccess := b.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := b.scope.Counter(metrics.ExecutionErrorMetric)

	projectCmds, err := b.ProjectCommandBuilder.BuildPlanCommands(ctx, comment)

	if err != nil {
		executionError.Inc(1)
		b.Logger.Err("Error building plan commands: %s", err)
	} else {
		executionSuccess.Inc(1)
	}

	return projectCmds, err

}
