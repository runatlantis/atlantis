// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import "time"

// DriftSummary represents detected infrastructure drift from a plan output.
// It captures the changes detected between the Terraform state and actual infrastructure.
type DriftSummary struct {
	// HasDrift indicates whether any drift was detected.
	HasDrift bool `json:"has_drift"`
	// ToAdd is the number of resources to add.
	ToAdd int `json:"to_add"`
	// ToChange is the number of resources to change.
	ToChange int `json:"to_change"`
	// ToDestroy is the number of resources to destroy.
	ToDestroy int `json:"to_destroy"`
	// ToImport is the number of resources to import (Terraform 1.5+).
	ToImport int `json:"to_import"`
	// Summary is a human-readable summary of the drift.
	Summary string `json:"summary"`
	// ChangesOutside indicates if Terraform detected changes made outside of Terraform.
	ChangesOutside bool `json:"changes_outside"`
}

// NewDriftSummaryFromPlanStats creates a DriftSummary from PlanSuccessStats.
// This leverages the existing plan output parsing infrastructure.
func NewDriftSummaryFromPlanStats(stats PlanSuccessStats, summary string) DriftSummary {
	hasDrift := stats.Add > 0 || stats.Change > 0 || stats.Destroy > 0 || stats.Import > 0
	return DriftSummary{
		HasDrift:       hasDrift,
		ToAdd:          stats.Add,
		ToChange:       stats.Change,
		ToDestroy:      stats.Destroy,
		ToImport:       stats.Import,
		Summary:        summary,
		ChangesOutside: stats.ChangesOutside,
	}
}

// NewDriftSummaryFromPlanSuccess creates a DriftSummary from a PlanSuccess result.
func NewDriftSummaryFromPlanSuccess(plan *PlanSuccess) DriftSummary {
	if plan == nil {
		return DriftSummary{}
	}
	stats := plan.Stats()
	return NewDriftSummaryFromPlanStats(stats, plan.DiffSummary())
}

// ProjectDrift represents the drift status for a specific project.
type ProjectDrift struct {
	// ProjectName is the name of the project (from atlantis.yaml).
	ProjectName string `json:"project_name"`
	// Path is the relative path to the project root within the repository.
	Path string `json:"path"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// Ref is the git reference (branch/tag/commit) that was checked.
	Ref string `json:"ref"`
	// Drift contains the drift summary for this project.
	Drift DriftSummary `json:"drift"`
	// LastChecked is when the drift was last detected.
	LastChecked time.Time `json:"last_checked"`
	// Error contains any error message if drift detection failed for this project.
	Error string `json:"error,omitempty"`
}

// DriftStatusResponse is the API response for GET /api/drift/status.
type DriftStatusResponse struct {
	// Repository is the full repository name (owner/repo).
	Repository string `json:"repository"`
	// Projects contains the drift status for each project.
	Projects []ProjectDrift `json:"projects"`
	// CheckedAt is when the drift check was performed.
	CheckedAt time.Time `json:"checked_at"`
	// TotalProjects is the total number of projects checked.
	TotalProjects int `json:"total_projects"`
	// ProjectsWithDrift is the number of projects that have drift.
	ProjectsWithDrift int `json:"projects_with_drift"`
}

// NewDriftStatusResponse creates a DriftStatusResponse from a list of ProjectDrift.
func NewDriftStatusResponse(repository string, projects []ProjectDrift) DriftStatusResponse {
	driftCount := 0
	for _, p := range projects {
		if p.Drift.HasDrift {
			driftCount++
		}
	}
	return DriftStatusResponse{
		Repository:        repository,
		Projects:          projects,
		CheckedAt:         time.Now(),
		TotalProjects:     len(projects),
		ProjectsWithDrift: driftCount,
	}
}
