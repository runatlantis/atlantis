// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
)

// OutputPersistingProjectCommandRunner wraps a ProjectCommandRunner and persists
// command outputs to the database after each command completes.
type OutputPersistingProjectCommandRunner struct {
	ProjectCommandRunner
	outputPersister *OutputPersister
}

// NewOutputPersistingProjectCommandRunner creates a new OutputPersistingProjectCommandRunner.
func NewOutputPersistingProjectCommandRunner(
	runner ProjectCommandRunner,
	outputPersister *OutputPersister,
) *OutputPersistingProjectCommandRunner {
	return &OutputPersistingProjectCommandRunner{
		ProjectCommandRunner: runner,
		outputPersister:      outputPersister,
	}
}

func (p *OutputPersistingProjectCommandRunner) Plan(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.Plan)
	result := p.ProjectCommandRunner.Plan(ctx)
	p.persistOutput(ctx, result, command.Plan)
	return result
}

func (p *OutputPersistingProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.PolicyCheck)
	result := p.ProjectCommandRunner.PolicyCheck(ctx)
	p.persistOutput(ctx, result, command.PolicyCheck)
	return result
}

func (p *OutputPersistingProjectCommandRunner) Apply(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.Apply)
	result := p.ProjectCommandRunner.Apply(ctx)
	p.persistOutput(ctx, result, command.Apply)
	return result
}

func (p *OutputPersistingProjectCommandRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.ApprovePolicies)
	result := p.ProjectCommandRunner.ApprovePolicies(ctx)
	p.persistOutput(ctx, result, command.ApprovePolicies)
	return result
}

func (p *OutputPersistingProjectCommandRunner) Version(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.Version)
	result := p.ProjectCommandRunner.Version(ctx)
	p.persistOutput(ctx, result, command.Version)
	return result
}

func (p *OutputPersistingProjectCommandRunner) Import(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.Import)
	result := p.ProjectCommandRunner.Import(ctx)
	p.persistOutput(ctx, result, command.Import)
	return result
}

func (p *OutputPersistingProjectCommandRunner) StateRm(ctx command.ProjectContext) command.ProjectCommandOutput {
	p.persistStub(ctx, command.State)
	result := p.ProjectCommandRunner.StateRm(ctx)
	p.persistOutput(ctx, result, command.State)
	return result
}

// persistStub writes a Running-status stub record before command execution begins.
func (p *OutputPersistingProjectCommandRunner) persistStub(
	ctx command.ProjectContext,
	cmdName command.Name,
) {
	if p.outputPersister == nil {
		return
	}

	if err := p.outputPersister.PersistStub(ctx, cmdName); err != nil {
		ctx.Log.Warn("failed to persist command stub: %v", err)
	}
}

// persistOutput saves the command output to the database.
func (p *OutputPersistingProjectCommandRunner) persistOutput(
	ctx command.ProjectContext,
	output command.ProjectCommandOutput,
	cmdName command.Name,
) {
	if p.outputPersister == nil {
		return
	}

	// Create ProjectResult from ProjectCommandOutput
	result := command.ProjectResult{
		ProjectCommandOutput: output,
		Command:              cmdName,
		RepoRelDir:           ctx.RepoRelDir,
		Workspace:            ctx.Workspace,
		ProjectName:          ctx.ProjectName,
	}

	if err := p.outputPersister.PersistResult(ctx, result); err != nil {
		ctx.Log.Warn("failed to persist command output: %v", err)
	}
}
