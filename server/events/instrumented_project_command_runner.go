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
	ctx.Log = ctx.Log.WithHistory("project", ctx.ProjectName)
	logger := ctx.Log

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)
	executionFailure := scope.Counter(metrics.ExecutionFailureMetric)

	result := execute(ctx)

	if result.Error != nil {
		executionError.Inc(1)
		logger.Err("Error running %s operation: %s", commandName, result.Error.Error())
		return result
	}

	if result.Failure != "" {
		executionFailure.Inc(1)
		logger.Err("Failure running %s operation: %s", commandName, result.Failure)
		return result
	}

	logger.Info("%s success. output available at: %s", commandName, ctx.Pull.URL)

	executionSuccess.Inc(1)
	return result

}
