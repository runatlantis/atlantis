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

	runnerFunc := func(projCtx command.ProjectContext) command.ProjectCommandOutput {
		// before group 1 runs, project1 from group 0 should be in PullStatus with AppliedPlanStatus
		if projCtx.ExecutionOrderGroup == 1 {
			found := false
			for _, proj := range ctx.PullStatus.Projects {
				if proj.ProjectName == "project1" {
					found = true
					Equals(t, models.AppliedPlanStatus, proj.Status)
				}
			}
			Assert(t, found, "dependant project1 should be in PullStatus before ExecutionOrderGroup 1 runs")
		}
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
	Assert(t, len(result.ProjectResults) == 2, "expected 2 project results, got %d", len(result.ProjectResults))
}
