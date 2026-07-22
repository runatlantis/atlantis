// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestContext_ErroredPlanProjects(t *testing.T) {
	ctx := command.Context{
		PullStatus: &models.PullStatus{
			Projects: []models.ProjectStatus{
				{
					Workspace:  "default",
					RepoRelDir: ".",
					Status:     models.PlannedPlanStatus,
				},
				{
					Workspace:   "default",
					RepoRelDir:  "staging",
					ProjectName: "staging",
					Status:      models.ErroredPlanStatus,
				},
			},
		},
	}

	failedProjects := ctx.ErroredPlanProjects()

	Equals(t, 1, len(failedProjects))
	Equals(t, "staging", failedProjects[0].RepoRelDir)
}

func TestContext_ErroredPlanProjects_NilPullStatus(t *testing.T) {
	ctx := command.Context{}

	Equals(t, 0, len(ctx.ErroredPlanProjects()))
}
