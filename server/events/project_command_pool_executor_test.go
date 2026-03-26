// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tally "github.com/uber-go/tally/v4"
)

func makeProjectContext(name string) command.ProjectContext {
	return command.ProjectContext{
		CommandName: command.Plan,
		ProjectName: name,
		RepoRelDir:  name,
		Workspace:   "default",
	}
}

func successRunner(_ command.ProjectContext) command.ProjectCommandOutput {
	return command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}
}

func TestRunProjectCmdsParallel_AllComplete(t *testing.T) {
	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
		makeProjectContext("p3"),
	}

	result := runProjectCmdsParallel(cmds, successRunner, 3, nil, models.PullRequest{})

	require.Len(t, result.ProjectResults, 3)
	assert.False(t, result.HasErrors())
}

func TestRunProjectCmdsParallel_CancelledBeforeExecution(t *testing.T) {
	tracker := NewCancellationTracker()
	pull := models.PullRequest{Num: 1}
	tracker.Cancel(pull)

	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
		makeProjectContext("p3"),
	}

	result := runProjectCmdsParallel(cmds, successRunner, 3, tracker, pull)

	require.Len(t, result.ProjectResults, 3)
	for _, r := range result.ProjectResults {
		// it was directly cancelled before execution, so all results should be cancelled
		require.Error(t, r.Error)
		assert.Contains(t, r.Error.Error(), "cancelled")
	}
}

// TestRunProjectCmdsParallel_CancelledWhileExceedingPoolSize verifies the bug fix for
// https://github.com/runatlantis/atlantis/pull/5813#issuecomment-3893031969:
// When the number of commands exceeds the pool size, commands queued behind a full pool
// must be cancelled rather than left waiting indefinitely.
//
// The slow runner simulates real workload by sleeping, keeping the pool saturated long
// enough for the cancellation to be registered before any worker finishes. Queued
// commands (p3–p5) pick up the cancellation in the post-wg.Add() check.
func TestRunProjectCmdsParallel_CancelledWhileExceedingPoolSize(t *testing.T) {
	const poolSize = 2
	const totalCmds = 5
	const expectedCancelled = totalCmds - poolSize

	tracker := NewCancellationTracker()
	pull := models.PullRequest{Num: 1}

	slowRunner := func(_ command.ProjectContext) command.ProjectCommandOutput {
		time.Sleep(500 * time.Millisecond)
		return command.ProjectCommandOutput{}
	}

	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
		makeProjectContext("p3"),
		makeProjectContext("p4"),
		makeProjectContext("p5"),
	}

	go func() {
		// Cancel after a short delay, while p1 and p2 are still running
		// and p3–p5 are queued waiting for a free pool slot.
		time.Sleep(50 * time.Millisecond)
		tracker.Cancel(pull)
	}()

	result := runProjectCmdsParallel(cmds, slowRunner, poolSize, tracker, pull)

	require.Len(t, result.ProjectResults, totalCmds)

	cancelledCount := 0
	for _, r := range result.ProjectResults {
		if r.Error != nil {
			assert.Contains(t, r.Error.Error(), "cancelled")
			cancelledCount++
		}
	}
	assert.Equal(t, expectedCancelled, cancelledCount)
}

func TestRunProjectCmds_Sequential(t *testing.T) {
	order := make([]string, 0)
	var mu sync.Mutex

	orderedRunner := func(cmd command.ProjectContext) command.ProjectCommandOutput {
		mu.Lock()
		order = append(order, cmd.ProjectName)
		mu.Unlock()
		return command.ProjectCommandOutput{}
	}

	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
		makeProjectContext("p3"),
	}

	result := runProjectCmds(cmds, orderedRunner)

	require.Len(t, result.ProjectResults, 3)
	assert.Equal(t, []string{"p1", "p2", "p3"}, order)
}

func TestSplitByExecutionOrderGroup(t *testing.T) {
	cmds := []command.ProjectContext{
		{ProjectName: "a", ExecutionOrderGroup: 1},
		{ProjectName: "b", ExecutionOrderGroup: 0},
		{ProjectName: "c", ExecutionOrderGroup: 1},
		{ProjectName: "d", ExecutionOrderGroup: 2},
	}

	groups := splitByExecutionOrderGroup(cmds)

	require.Len(t, groups, 3)
	assert.Len(t, groups[0], 1)
	assert.Equal(t, "b", groups[0][0].ProjectName)
	assert.Len(t, groups[1], 2)
	assert.Len(t, groups[2], 1)
	assert.Equal(t, "d", groups[2][0].ProjectName)
}

func TestPrepareExecutionGroups_SingleGroupNonParallel(t *testing.T) {
	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
	}

	groups := prepareExecutionGroups(cmds, false)

	assert.Len(t, groups, 2, "non-parallel single group should be split into individual groups")
	for _, g := range groups {
		assert.Len(t, g, 1)
	}
}

func TestPrepareExecutionGroups_SingleGroupParallel(t *testing.T) {
	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
	}

	groups := prepareExecutionGroups(cmds, true)

	assert.Len(t, groups, 1, "parallel single group should stay as one group")
	assert.Len(t, groups[0], 2)
}

func TestPrepareExecutionGroups_MultipleGroups(t *testing.T) {
	cmds := []command.ProjectContext{
		{ProjectName: "p1", ExecutionOrderGroup: 0},
		{ProjectName: "p2", ExecutionOrderGroup: 1},
	}

	groups := prepareExecutionGroups(cmds, false)

	assert.Len(t, groups, 2, "multiple execution order groups should always be preserved")
}

func makeCtx(t *testing.T, pull models.PullRequest) *command.Context {
	t.Helper()
	return &command.Context{
		Log:   logging.NewNoopLogger(t),
		Pull:  pull,
		Scope: tally.NoopScope,
	}
}

func TestRunProjectCmdsWithCancellationTracker_NoCancellation(t *testing.T) {
	tracker := NewCancellationTracker()
	pull := models.PullRequest{Num: 1}
	ctx := makeCtx(t, pull)

	cmds := []command.ProjectContext{
		makeProjectContext("p1"),
		makeProjectContext("p2"),
	}

	result := runProjectCmdsWithCancellationTracker(ctx, cmds, tracker, 2, false, successRunner)

	require.Len(t, result.ProjectResults, 2)
	assert.False(t, result.HasErrors())
}

func TestRunProjectCmdsWithCancellationTracker_CancelBetweenGroups(t *testing.T) {
	tracker := NewCancellationTracker()
	pull := models.PullRequest{Num: 2}
	ctx := makeCtx(t, pull)

	cmds := []command.ProjectContext{
		{CommandName: command.Plan, ProjectName: "p1", RepoRelDir: "p1", Workspace: "default", ExecutionOrderGroup: 0},
		{CommandName: command.Plan, ProjectName: "p2", RepoRelDir: "p2", Workspace: "default", ExecutionOrderGroup: 1},
		{CommandName: command.Plan, ProjectName: "p3", RepoRelDir: "p3", Workspace: "default", ExecutionOrderGroup: 1},
	}

	cancellingRunner := func(cmd command.ProjectContext) command.ProjectCommandOutput {
		if cmd.ProjectName == "p1" {
			tracker.Cancel(pull)
		}
		return command.ProjectCommandOutput{}
	}

	result := runProjectCmdsWithCancellationTracker(ctx, cmds, tracker, 2, true, cancellingRunner)

	require.Len(t, result.ProjectResults, 3)

	p1Result := findResult(result, "p1")
	require.NotNil(t, p1Result)
	require.NoError(t, p1Result.Error)

	p2Result := findResult(result, "p2")
	require.NotNil(t, p2Result)
	require.Error(t, p2Result.Error)
	assert.Contains(t, p2Result.Error.Error(), "cancelled")

	p3Result := findResult(result, "p3")
	require.NotNil(t, p3Result)
	require.Error(t, p3Result.Error)
	assert.Contains(t, p3Result.Error.Error(), "cancelled")
}

func TestRunProjectCmdsWithCancellationTracker_ClearsTrackerOnDone(t *testing.T) {
	tracker := NewCancellationTracker()
	pull := models.PullRequest{Num: 3}
	ctx := makeCtx(t, pull)

	tracker.Cancel(pull)
	assert.True(t, tracker.IsCancelled(pull))

	cmds := []command.ProjectContext{makeProjectContext("p1")}
	runProjectCmdsWithCancellationTracker(ctx, cmds, tracker, 2, false, successRunner)

	assert.False(t, tracker.IsCancelled(pull), "tracker should be cleared after execution")
}

func TestCreateCancelledResults(t *testing.T) {
	groups := [][]command.ProjectContext{
		{
			{CommandName: command.Plan, ProjectName: "p1", RepoRelDir: "dir1", Workspace: "ws1"},
			{CommandName: command.Plan, ProjectName: "p2", RepoRelDir: "dir2", Workspace: "ws2"},
		},
		{
			{CommandName: command.Plan, ProjectName: "p3", RepoRelDir: "dir3", Workspace: "ws3"},
		},
	}

	results := createCancelledResults(groups)

	require.Len(t, results, 3)
	for _, r := range results {
		require.Error(t, r.Error)
		assert.Contains(t, r.Error.Error(), "cancelled")
	}

	assert.Equal(t, "p1", results[0].ProjectName)
	assert.Equal(t, "dir1", results[0].RepoRelDir)
	assert.Equal(t, "ws1", results[0].Workspace)
	assert.Equal(t, "p3", results[2].ProjectName)
	assert.Equal(t, "dir3", results[2].RepoRelDir)
	assert.Equal(t, "ws3", results[2].Workspace)
}

func findResult(result command.Result, projectName string) *command.ProjectResult {
	for i := range result.ProjectResults {
		if result.ProjectResults[i].ProjectName == projectName {
			return &result.ProjectResults[i]
		}
	}
	return nil
}
