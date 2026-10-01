// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

func TestInstrumentedProjectCommandBuilder_IgnoredTargetedDirIsNotErrorMetric(t *testing.T) {
	scope := tally.NewTestScope("builder", nil)
	builder := &InstrumentedProjectCommandBuilder{
		ProjectCommandBuilder: fakeInstrumentedProjectCommandBuilder{err: ErrIgnoredTargetedDir},
		Logger:                logging.NewNoopLogger(t),
		scope:                 scope,
	}

	_, err := builder.BuildPlanCommands(&command.Context{}, &CommentCommand{})

	if !IsIgnoredTargetedDir(err) {
		t.Fatalf("expected ignored targeted dir error, got %v", err)
	}
	counters := scope.Snapshot().Counters()
	if got := counterValue(counters, "builder."+metrics.ExecutionSuccessMetric+"+"); got != 1 {
		t.Fatalf("expected success metric to be 1, got %d", got)
	}
	if got := counterValue(counters, "builder."+metrics.ExecutionErrorMetric+"+"); got != 0 {
		t.Fatalf("expected error metric to be 0, got %d", got)
	}
}

func counterValue(counters map[string]tally.CounterSnapshot, name string) int64 {
	counter, ok := counters[name]
	if !ok {
		return 0
	}
	return counter.Value()
}

type fakeInstrumentedProjectCommandBuilder struct {
	err error
}

func (f fakeInstrumentedProjectCommandBuilder) ShouldIgnoreTargetedDir(ctx *command.Context, comment *CommentCommand) bool {
	return false
}

func (f fakeInstrumentedProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	return nil, f.err
}

func (f fakeInstrumentedProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return nil, f.err
}

func (f fakeInstrumentedProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return nil, f.err
}

func (f fakeInstrumentedProjectCommandBuilder) BuildApprovePoliciesCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return nil, f.err
}

func (f fakeInstrumentedProjectCommandBuilder) BuildVersionCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return nil, f.err
}

func (f fakeInstrumentedProjectCommandBuilder) BuildImportCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return nil, f.err
}

func (f fakeInstrumentedProjectCommandBuilder) BuildStateRmCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error) {
	return nil, f.err
}
