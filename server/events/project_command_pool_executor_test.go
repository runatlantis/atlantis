// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRunProjectCmdsWithCancellationTracker_UpdatesPullStatusBetweenGroups(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	ctx := &command.Context{
		Log: logger,
		PullStatus: &models.PullStatus{
			Projects: []models.ProjectStatus{
				{
					Workspace:   "default",
					RepoRelDir:  "project1",
					ProjectName: "project1",
					Status:      models.PlannedPlanStatus,
				},
			},
		},
	}

	projectCmds := []command.ProjectContext{
		{
			CommandName:         command.Apply,
			Workspace:           "default",
			RepoRelDir:          "project1",
			ProjectName:         "project1",
			ExecutionOrderGroup: 0,
		},
		{
			CommandName:         command.Apply,
			Workspace:           "default",
			RepoRelDir:          "project2",
			ProjectName:         "project2",
			ExecutionOrderGroup: 1,
			DependsOn:           []string{"project1"},
			PullStatus:          ctx.PullStatus,
		},
	}

	runnerFunc := func(ctx command.ProjectContext) command.ProjectCommandOutput {
		return command.ProjectCommandOutput{
			ApplySuccess: "success",
		}
	}

	result := runProjectCmdsWithCancellationTracker(
		ctx,
		projectCmds,
		nil,
		1,
		false,
		runnerFunc,
	)

	// Verify both projects ran
	if len(result.ProjectResults) != 2 {
		t.Fatalf("expected 2 project results, got %d", len(result.ProjectResults))
	}

	// Verify ctx.PullStatus was updated after first group
	// The status should now be AppliedPlanStatus for project1
	foundProject1 := false
	for _, proj := range ctx.PullStatus.Projects {
		if proj.ProjectName == "project1" {
			foundProject1 = true
			Equals(t, models.AppliedPlanStatus, proj.Status)
		}
	}

	Assert(t, foundProject1, "project1 should be in PullStatus")
}
