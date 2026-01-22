// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewDriftSummaryFromPlanStats(t *testing.T) {
	cases := []struct {
		name     string
		stats    models.PlanSuccessStats
		summary  string
		expected models.DriftSummary
	}{
		{
			name: "no drift",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 0,
				Import:  0,
			},
			summary: "No changes",
			expected: models.DriftSummary{
				HasDrift:  false,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  0,
				Summary:   "No changes",
			},
		},
		{
			name: "with additions",
			stats: models.PlanSuccessStats{
				Add:     3,
				Change:  0,
				Destroy: 0,
				Import:  0,
			},
			summary: "Plan: 3 to add",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     3,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  0,
				Summary:   "Plan: 3 to add",
			},
		},
		{
			name: "with changes",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  5,
				Destroy: 0,
				Import:  0,
			},
			summary: "Plan: 5 to change",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  5,
				ToDestroy: 0,
				ToImport:  0,
				Summary:   "Plan: 5 to change",
			},
		},
		{
			name: "with destroys",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 2,
				Import:  0,
			},
			summary: "Plan: 2 to destroy",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 2,
				ToImport:  0,
				Summary:   "Plan: 2 to destroy",
			},
		},
		{
			name: "with imports",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 0,
				Import:  1,
			},
			summary: "Plan: 1 to import",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  1,
				Summary:   "Plan: 1 to import",
			},
		},
		{
			name: "with changes outside terraform",
			stats: models.PlanSuccessStats{
				Add:            0,
				Change:         1,
				Destroy:        0,
				Import:         0,
				ChangesOutside: true,
			},
			summary: "Plan: 1 to change (changes outside Terraform)",
			expected: models.DriftSummary{
				HasDrift:       true,
				ToAdd:          0,
				ToChange:       1,
				ToDestroy:      0,
				ToImport:       0,
				Summary:        "Plan: 1 to change (changes outside Terraform)",
				ChangesOutside: true,
			},
		},
		{
			name: "mixed drift",
			stats: models.PlanSuccessStats{
				Add:     2,
				Change:  3,
				Destroy: 1,
				Import:  1,
			},
			summary: "Plan: 2 to add, 3 to change, 1 to destroy, 1 to import",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     2,
				ToChange:  3,
				ToDestroy: 1,
				ToImport:  1,
				Summary:   "Plan: 2 to add, 3 to change, 1 to destroy, 1 to import",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := models.NewDriftSummaryFromPlanStats(c.stats, c.summary)
			Equals(t, c.expected, result)
		})
	}
}

func TestNewDriftSummaryFromPlanSuccess_Nil(t *testing.T) {
	result := models.NewDriftSummaryFromPlanSuccess(nil)
	Equals(t, models.DriftSummary{}, result)
}

func TestNewDriftStatusResponse(t *testing.T) {
	projects := []models.ProjectDrift{
		{
			ProjectName: "project1",
			Path:        "path1",
			Workspace:   "default",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: true,
				ToAdd:    2,
			},
			LastChecked: time.Now(),
		},
		{
			ProjectName: "project2",
			Path:        "path2",
			Workspace:   "default",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: false,
			},
			LastChecked: time.Now(),
		},
		{
			ProjectName: "project3",
			Path:        "path3",
			Workspace:   "staging",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: true,
				ToChange: 1,
			},
			LastChecked: time.Now(),
		},
	}

	result := models.NewDriftStatusResponse("owner/repo", projects)

	Equals(t, "owner/repo", result.Repository)
	Equals(t, 3, result.TotalProjects)
	Equals(t, 2, result.ProjectsWithDrift)
	Equals(t, projects, result.Projects)
	Assert(t, !result.CheckedAt.IsZero(), "CheckedAt should be set")
}

func TestNewDriftStatusResponse_Empty(t *testing.T) {
	result := models.NewDriftStatusResponse("owner/repo", []models.ProjectDrift{})

	Equals(t, "owner/repo", result.Repository)
	Equals(t, 0, result.TotalProjects)
	Equals(t, 0, result.ProjectsWithDrift)
}
