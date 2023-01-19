package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally"
)

type IntrumentedCommandRunner interface {
	Plan(ctx command.ProjectContext) command.ProjectResult
	PolicyCheck(ctx command.ProjectContext) command.ProjectResult
	Apply(ctx command.ProjectContext) command.ProjectResult
	ApprovePolicies(ctx command.ProjectContext) command.ProjectResult
	Import(ctx command.ProjectContext) command.ProjectResult
	StateRm(ctx command.ProjectContext) command.ProjectResult
}

type InstrumentedProjectCommandRunner struct {
	projectCommandRunner ProjectCommandRunner
	scope                tally.Scope
}

func NewInstrumentedProjectCommandRunner(scope tally.Scope, projectCommandRunner ProjectCommandRunner) *InstrumentedProjectCommandRunner {
	projectTags := command.ProjectScopeTags{}
	scope = scope.Tagged(projectTags.Loadtags())

	for _, m := range []string{metrics.ExecutionSuccessMetric, metrics.ExecutionErrorMetric, metrics.ExecutionFailureMetric} {
		metrics.InitCounter(scope, m)
	}

	return &InstrumentedProjectCommandRunner{
		projectCommandRunner: projectCommandRunner,
		scope:                scope,
	}
}

func (p *InstrumentedProjectCommandRunner) Plan(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats("plan", ctx, p.projectCommandRunner.Plan, p.scope)
}

func (p *InstrumentedProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats("policy check", ctx, p.projectCommandRunner.PolicyCheck, p.scope)
}

func (p *InstrumentedProjectCommandRunner) Apply(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats("apply", ctx, p.projectCommandRunner.Apply, p.scope)
}

func (p *InstrumentedProjectCommandRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats("approve policies", ctx, p.projectCommandRunner.ApprovePolicies, p.scope)
}

func (p *InstrumentedProjectCommandRunner) Import(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats("import", ctx, p.projectCommandRunner.Import, p.scope)
}

func (p *InstrumentedProjectCommandRunner) StateRm(ctx command.ProjectContext) command.ProjectResult {
	return RunAndEmitStats("state rm", ctx, p.projectCommandRunner.StateRm, p.scope)
}

func RunAndEmitStats(commandName string, ctx command.ProjectContext, execute func(ctx command.ProjectContext) command.ProjectResult, scope tally.Scope) command.ProjectResult {
	// ensures we are differentiating between project level command and overall command
	scope = ctx.SetProjectScopeTags(scope)
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
