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
			name: "with forgets",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 0,
				Import:  0,
				Forget:  2,
			},
			summary: "Plan: 0 to add, 0 to change, 0 to destroy, 2 to forget",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  0,
				ToForget:  2,
				Summary:   "Plan: 0 to add, 0 to change, 0 to destroy, 2 to forget",
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
			name: "with only changes outside terraform",
			stats: models.PlanSuccessStats{
				ChangesOutside: true,
			},
			summary: "Objects have changed outside of Terraform",
			expected: models.DriftSummary{
				HasDrift:       true,
				Summary:        "Objects have changed outside of Terraform",
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
				Forget:  1,
			},
			summary: "Plan: 1 to import, 2 to add, 3 to change, 1 to destroy, 1 to forget",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     2,
				ToChange:  3,
				ToDestroy: 1,
				ToImport:  1,
				ToForget:  1,
				Summary:   "Plan: 1 to import, 2 to add, 3 to change, 1 to destroy, 1 to forget",
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

func TestNewDriftSummaryFromPlanSuccess_OutsideOnlyDriftIncludesNote(t *testing.T) {
	result := models.NewDriftSummaryFromPlanSuccess(&models.PlanSuccess{
		TerraformOutput: "Note: Objects have changed outside of Terraform\ndummy\nNo changes. Infrastructure is up-to-date.",
	})

	Equals(t, true, result.HasDrift)
	Equals(t, true, result.ChangesOutside)
	Equals(t, 0, result.TotalChanges())
	Equals(t, "\n**Note: Objects have changed outside of Terraform**\nNo changes. Infrastructure is up-to-date.", result.Summary)
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

func TestDriftDetectionRequestValidateRejectsEmptySelectors(t *testing.T) {
	for _, c := range []struct {
		name    string
		request models.DriftDetectionRequest
		field   string
	}{
		{
			name: "empty project",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Projects:   []string{""},
			},
			field: "projects",
		},
		{
			name: "space-only project",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Projects:   []string{"   "},
			},
			field: "projects",
		},
		{
			name: "empty path",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: ""}},
			},
			field: "paths",
		},
		{
			name: "space-only path",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "   "}},
			},
			field: "paths",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs := c.request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, c.field, errs[0].Field)
		})
	}
}
