package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/metrics"
)

const (
	Plan        = "plan"
	PolicyCheck = "policy check"
	Apply       = "apply"
)

type InstrumentedProjectCommandRunner struct {
	ProjectCommandRunner
}

func (p *InstrumentedProjectCommandRunner) Plan(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats(Plan, ctx, p.ProjectCommandRunner.Plan)
}

func (p *InstrumentedProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats(PolicyCheck, ctx, p.ProjectCommandRunner.PolicyCheck)
}

func (p *InstrumentedProjectCommandRunner) Apply(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats(Apply, ctx, p.ProjectCommandRunner.Apply)
}

func RunAndEmitStats(commandName string, ctx command.ProjectContext, execute func(ctx command.ProjectContext) command.ProjectResult) command.ProjectResult {

	// ensures we are differentiating between project level command and overall command
	ctx.SetScope("project")

	scope := ctx.Scope
	logger := ctx.Log

	executionTime := scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)
	executionFailure := scope.Counter(metrics.ExecutionFailureMetric)

	result := execute(ctx)

	if result.Error != nil {
		executionError.Inc(1)
		logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("Error running %s operation: %s", commandName, result.Error.Error()), map[string]interface{}{"project": ctx.ProjectName})
		return result
	}

	if result.Failure != "" {
		executionFailure.Inc(1)
		logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("Failure running %s operation: %s", commandName, result.Failure), map[string]interface{}{"project": ctx.ProjectName})
		return result
	}

	// Log successful policy check results
	if commandName == PolicyCheck && result.PolicyCheckSuccess != nil {
		logger.InfoContext(ctx.RequestCtx, fmt.Sprintf("Success running %s operation: %s", commandName, result.PolicyCheckSuccess.PolicyCheckOutput), map[string]interface{}{"project": ctx.ProjectName})
	}
	logger.InfoContext(ctx.RequestCtx, fmt.Sprintf("%s success. output available at: %s", commandName, ctx.Pull.URL), map[string]interface{}{"project": ctx.ProjectName})

	executionSuccess.Inc(1)
	return result

}
