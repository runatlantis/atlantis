package events

import (
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
)

type InstrumentedProjectCommandRunner struct {
	ProjectCommandRunner
}

func (p *InstrumentedProjectCommandRunner) Plan(ctx models.ProjectCommandContext) models.ProjectResult {
	return RunAndEmitStats("plan", ctx, p.ProjectCommandRunner.Plan)
}

func (p *InstrumentedProjectCommandRunner) PolicyCheck(ctx models.ProjectCommandContext) models.ProjectResult {
	return RunAndEmitStats("policy check", ctx, p.ProjectCommandRunner.PolicyCheck)
}

func (p *InstrumentedProjectCommandRunner) Apply(ctx models.ProjectCommandContext) models.ProjectResult {
	return RunAndEmitStats("apply", ctx, p.ProjectCommandRunner.Apply)
}

func RunAndEmitStats(commandName string, ctx models.ProjectCommandContext, execute func(ctx models.ProjectCommandContext) models.ProjectResult) models.ProjectResult {

	// ensures we are differentiating between project level command and overall command
	ctx.SetScope("project")

	scope := ctx.Scope
	logger := ctx.Log

	executionTime := scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	executionSuccess := scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := scope.NewCounter(metrics.ExecutionErrorMetric)
	executionFailure := scope.NewCounter(metrics.ExecutionFailureMetric)

	result := execute(ctx)

	if result.Error != nil {
		executionError.Inc()
		logger.Err("Error running %s operation: %s", commandName, result.Error.Error())
		return result
	}

	if result.Failure == "" {
		executionFailure.Inc()
		logger.Err("Failure running %s operation: %s", commandName, result.Failure)
		return result
	}

	executionSuccess.Inc()
	return result

}
