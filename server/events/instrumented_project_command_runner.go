// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
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
	prScopeManager       *metrics.PRScopeManager
	scope                tally.Scope // fallback scope if PRScopeManager is nil
}

func NewInstrumentedProjectCommandRunner(scope tally.Scope, projectCommandRunner ProjectCommandRunner, prScopeManager *metrics.PRScopeManager) *InstrumentedProjectCommandRunner {
	projectTags := command.ProjectScopeTags{}
	scope = scope.SubScope("project").Tagged(projectTags.Loadtags())

	for _, m := range []string{metrics.ExecutionSuccessMetric, metrics.ExecutionErrorMetric, metrics.ExecutionFailureMetric} {
		metrics.InitCounter(scope, m)
	}

	return &InstrumentedProjectCommandRunner{
		projectCommandRunner: projectCommandRunner,
		prScopeManager:       prScopeManager,
		scope:                scope,
	}
}

func (p *InstrumentedProjectCommandRunner) Plan(ctx command.ProjectContext) command.ProjectCommandOutput {
	return RunAndEmitStats(ctx, p.projectCommandRunner.Plan, p.prScopeManager, p.scope)
}

func (p *InstrumentedProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectCommandOutput {
	return RunAndEmitStats(ctx, p.projectCommandRunner.PolicyCheck, p.prScopeManager, p.scope)
}

func (p *InstrumentedProjectCommandRunner) Apply(ctx command.ProjectContext) command.ProjectCommandOutput {
	return RunAndEmitStats(ctx, p.projectCommandRunner.Apply, p.prScopeManager, p.scope)
}

func (p *InstrumentedProjectCommandRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectCommandOutput {
	return RunAndEmitStats(ctx, p.projectCommandRunner.ApprovePolicies, p.prScopeManager, p.scope)
}

func (p *InstrumentedProjectCommandRunner) Import(ctx command.ProjectContext) command.ProjectCommandOutput {
	return RunAndEmitStats(ctx, p.projectCommandRunner.Import, p.prScopeManager, p.scope)
}

func (p *InstrumentedProjectCommandRunner) StateRm(ctx command.ProjectContext) command.ProjectCommandOutput {
	return RunAndEmitStats(ctx, p.projectCommandRunner.StateRm, p.prScopeManager, p.scope)
}

func RunAndEmitStats(ctx command.ProjectContext, execute func(ctx command.ProjectContext) command.ProjectCommandOutput, prScopeManager *metrics.PRScopeManager, fallbackScope tally.Scope) command.ProjectCommandOutput {
	commandName := ctx.CommandName.String()
	// ensures we are differentiating between project level command and overall command

	scope := ctx.SetProjectScopeTags(prScopeManager).SubScope(commandName)
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
