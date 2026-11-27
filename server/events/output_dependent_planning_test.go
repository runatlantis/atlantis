// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type recordingDependentBuilder struct {
	responses [][]command.ProjectContext
	err       error

	producerCalls  [][]OutputDependentTrigger
	alreadyPlanned [][]ProjectKey
}

func (f *recordingDependentBuilder) ShouldIgnoreTargetedDir(*command.Context, *CommentCommand) bool {
	return false
}

func (f *recordingDependentBuilder) BuildAutoplanCommands(*command.Context) ([]command.ProjectContext, error) {
	return nil, nil
}

func (f *recordingDependentBuilder) BuildPlanCommands(*command.Context, *CommentCommand) ([]command.ProjectContext, error) {
	return nil, nil
}

func (f *recordingDependentBuilder) BuildOutputDependentPlanCommands(_ *command.Context, producers []OutputDependentTrigger, alreadyPlanned []command.ProjectContext) ([]command.ProjectContext, error) {
	f.producerCalls = append(f.producerCalls, producers)
	keys := make([]ProjectKey, 0, len(alreadyPlanned))
	for _, cmd := range alreadyPlanned {
		keys = append(keys, NewProjectKey(cmd.RepoRelDir, cmd.Workspace, cmd.ProjectName))
	}
	f.alreadyPlanned = append(f.alreadyPlanned, keys)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.responses) == 0 {
		return nil, nil
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

type recordingPlanOutputRunner struct {
	results map[string]*models.PlanSuccess
}

func (f *recordingPlanOutputRunner) Plan(ctx command.ProjectContext) command.ProjectCommandOutput {
	return command.ProjectCommandOutput{PlanSuccess: f.results[ctx.RepoRelDir]}
}

func planCmd(dir string) command.ProjectContext {
	return command.ProjectContext{
		CommandName: command.Plan,
		RepoRelDir:  dir,
		Workspace:   DefaultWorkspace,
		ProjectName: dir,
	}
}

func planResult(dir string, changedOutputs ...string) command.ProjectResult {
	changes := make(map[string]models.OutputChange, len(changedOutputs))
	for _, output := range changedOutputs {
		changes[output] = models.OutputChange{Actions: []string{"update"}}
	}
	return command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  dir,
		Workspace:   DefaultWorkspace,
		ProjectName: dir,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{OutputChanges: changes},
		},
	}
}

func TestPlanOutputDependents_FollowsCascade(t *testing.T) {
	builder := &recordingDependentBuilder{
		responses: [][]command.ProjectContext{
			{planCmd("app")},
			{planCmd("web")},
		},
	}
	planRunner := &recordingPlanOutputRunner{results: map[string]*models.PlanSuccess{
		"app": {OutputChanges: map[string]models.OutputChange{"db_endpoint": {Actions: []string{"update"}}}},
		// web is planned and changes nothing, so the cascade stops.
	}}
	runner := &PlanCommandRunner{
		prjCmdBuilder:    builder,
		prjCmdRunner:     planRunner,
		parallelPoolSize: 1,
	}
	ctx := &command.Context{
		Log:  logging.NewNoopLogger(t),
		Pull: models.PullRequest{Num: 1},
	}

	planCmds, policyCheckCmds, results := runner.planOutputDependents(
		ctx,
		[]command.ProjectContext{planCmd("network")},
		[]command.ProjectResult{planResult("network", "vpc_id")},
	)

	if len(policyCheckCmds) != 0 {
		t.Fatalf("expected no policy check commands, got %v", policyCheckCmds)
	}
	if len(planCmds) != 2 || planCmds[0].RepoRelDir != "app" || planCmds[1].RepoRelDir != "web" {
		t.Fatalf("expected plans for app then web, got %v", planCmds)
	}
	if len(results) != 2 || results[0].RepoRelDir != "app" || results[1].RepoRelDir != "web" {
		t.Fatalf("expected results for app then web, got %v", results)
	}
	if len(builder.producerCalls) != 2 {
		t.Fatalf("expected two rounds of dependency discovery, got %d", len(builder.producerCalls))
	}
	first := builder.producerCalls[0]
	if len(first) != 1 || first[0].RepoRelDir != "network" || !first[0].ChangedOutputs["vpc_id"] {
		t.Fatalf("unexpected first round triggers: %v", first)
	}
	second := builder.producerCalls[1]
	if len(second) != 1 || second[0].RepoRelDir != "app" || !second[0].ChangedOutputs["db_endpoint"] {
		t.Fatalf("unexpected second round triggers: %v", second)
	}
	if got := builder.alreadyPlanned[1]; len(got) != 2 {
		t.Fatalf("expected the dependent's dependencies to see both planned projects, got %v", got)
	}
}

func TestPlanOutputDependents_NoOutputChangesDoesNothing(t *testing.T) {
	builder := &recordingDependentBuilder{}
	runner := &PlanCommandRunner{prjCmdBuilder: builder, prjCmdRunner: &recordingPlanOutputRunner{}, parallelPoolSize: 1}
	ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: models.PullRequest{Num: 1}}

	planCmds, policyCheckCmds, results := runner.planOutputDependents(
		ctx,
		[]command.ProjectContext{planCmd("network")},
		[]command.ProjectResult{planResult("network")},
	)

	if len(planCmds) != 0 || len(policyCheckCmds) != 0 || len(results) != 0 {
		t.Fatalf("expected no dependent work, got plans=%v policy=%v results=%v", planCmds, policyCheckCmds, results)
	}
	if len(builder.producerCalls) != 0 {
		t.Fatalf("expected no dependency discovery, got %v", builder.producerCalls)
	}
}

func TestPlanOutputDependents_DiscoveryErrorLeavesPlansAlone(t *testing.T) {
	builder := &recordingDependentBuilder{err: errors.New("boom")}
	runner := &PlanCommandRunner{prjCmdBuilder: builder, prjCmdRunner: &recordingPlanOutputRunner{}, parallelPoolSize: 1}
	ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: models.PullRequest{Num: 1}}

	planCmds, _, results := runner.planOutputDependents(
		ctx,
		[]command.ProjectContext{planCmd("network")},
		[]command.ProjectResult{planResult("network", "vpc_id")},
	)

	if len(planCmds) != 0 || len(results) != 0 {
		t.Fatalf("expected dependency errors to be swallowed, got plans=%v results=%v", planCmds, results)
	}
}
