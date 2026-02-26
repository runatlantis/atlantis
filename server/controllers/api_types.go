// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// API DTOs (Data Transfer Objects)
// These types define the public API contract, separate from internal models.
// This provides:
// - Security: Internal fields never leak to API consumers
// - Stability: Internal refactors don't break API contracts
// - Clarity: API types are self-documenting for consumers

// ProjectResultAPI is the API representation of a project command result.
// It exposes only the fields appropriate for API consumers.
type ProjectResultAPI struct {
	// ProjectName is the name of the project (from atlantis.yaml).
	ProjectName string `json:"project_name"`
	// Directory is the relative path to the project root.
	Directory string `json:"directory"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// Command is the command that was executed.
	Command string `json:"command"`
	// Status indicates whether the command succeeded or failed.
	Status string `json:"status"`
	// Output contains the command output (plan/apply output).
	Output string `json:"output,omitempty"`
	// Error contains any error message.
	Error string `json:"error,omitempty"`
	// Failure contains any failure message (different from error).
	Failure string `json:"failure,omitempty"`
	// Plan contains plan-specific details if this was a plan command.
	Plan *PlanDetailsAPI `json:"plan,omitempty"`
}

// PlanDetailsAPI is the API representation of plan output details.
type PlanDetailsAPI struct {
	// HasChanges indicates whether the plan contains changes.
	HasChanges bool `json:"has_changes"`
	// ToAdd is the number of resources to add.
	ToAdd int `json:"to_add"`
	// ToChange is the number of resources to change.
	ToChange int `json:"to_change"`
	// ToDestroy is the number of resources to destroy.
	ToDestroy int `json:"to_destroy"`
	// ToImport is the number of resources to import.
	ToImport int `json:"to_import"`
	// Summary is a human-readable summary.
	Summary string `json:"summary,omitempty"`
}

// NewProjectResultAPI converts an internal ProjectResult to its API representation.
func NewProjectResultAPI(pr command.ProjectResult) ProjectResultAPI {
	result := ProjectResultAPI{
		ProjectName: pr.ProjectName,
		Directory:   pr.RepoRelDir,
		Workspace:   pr.Workspace,
		Command:     pr.Command.String(),
		Status:      "success",
	}

	// Handle error
	if pr.Error != nil {
		result.Status = "error"
		result.Error = pr.Error.Error()
	}

	// Handle failure
	if pr.Failure != "" {
		result.Status = "failed"
		result.Failure = pr.Failure
	}

	// Handle plan success
	if pr.PlanSuccess != nil {
		stats := pr.PlanSuccess.Stats()
		result.Output = pr.PlanSuccess.TerraformOutput
		result.Plan = &PlanDetailsAPI{
			HasChanges: stats.Changes,
			ToAdd:      stats.Add,
			ToChange:   stats.Change,
			ToDestroy:  stats.Destroy,
			ToImport:   stats.Import,
			Summary:    pr.PlanSuccess.DiffSummary(),
		}
	}

	// Handle apply success
	if pr.ApplySuccess != "" {
		result.Output = pr.ApplySuccess
	}

	return result
}

// CommandResultAPI is the API representation of a command result with multiple projects.
type CommandResultAPI struct {
	// Command is the command that was executed.
	Command string `json:"command"`
	// Projects contains results for each project.
	Projects []ProjectResultAPI `json:"projects"`
	// TotalProjects is the total number of projects.
	TotalProjects int `json:"total_projects"`
	// SuccessCount is the number of successful projects.
	SuccessCount int `json:"success_count"`
	// FailureCount is the number of failed projects.
	FailureCount int `json:"failure_count"`
	// ExecutedAt is when the command was executed.
	ExecutedAt time.Time `json:"executed_at"`
}

// NewCommandResultAPI converts an internal CommandResult to its API representation.
func NewCommandResultAPI(cr *command.Result, cmdName string) CommandResultAPI {
	result := CommandResultAPI{
		Command:    cmdName,
		Projects:   make([]ProjectResultAPI, 0, len(cr.ProjectResults)),
		ExecutedAt: time.Now(),
	}

	for _, pr := range cr.ProjectResults {
		apiResult := NewProjectResultAPI(pr)
		result.Projects = append(result.Projects, apiResult)

		if apiResult.Status == "success" {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}

	result.TotalProjects = len(result.Projects)
	return result
}

// DriftProjectAPI is the API representation of drift for a single project.
type DriftProjectAPI struct {
	// ProjectName is the name of the project.
	ProjectName string `json:"project_name"`
	// Directory is the relative path to the project.
	Directory string `json:"directory"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// Ref is the git reference that was checked.
	Ref string `json:"ref"`
	// DetectionID links this result to the detection run that produced it.
	DetectionID string `json:"detection_id,omitempty"`
	// HasDrift indicates whether drift was detected.
	HasDrift bool `json:"has_drift"`
	// Drift contains drift details if drift was detected.
	Drift *DriftDetailsAPI `json:"drift,omitempty"`
	// LastChecked is when drift was last checked.
	LastChecked time.Time `json:"last_checked"`
	// Error contains any error message if detection failed.
	Error string `json:"error,omitempty"`
}

// DriftDetailsAPI is the API representation of drift details.
type DriftDetailsAPI struct {
	// ToAdd is the number of resources to add.
	ToAdd int `json:"to_add"`
	// ToChange is the number of resources to change.
	ToChange int `json:"to_change"`
	// ToDestroy is the number of resources to destroy.
	ToDestroy int `json:"to_destroy"`
	// ToImport is the number of resources to import.
	ToImport int `json:"to_import"`
	// TotalChanges is the total number of changes.
	TotalChanges int `json:"total_changes"`
	// Summary is a human-readable summary.
	Summary string `json:"summary,omitempty"`
	// ChangesOutside indicates if changes were made outside Terraform.
	ChangesOutside bool `json:"changes_outside"`
}

// NewDriftProjectAPI converts an internal ProjectDrift to its API representation.
func NewDriftProjectAPI(pd models.ProjectDrift) DriftProjectAPI {
	result := DriftProjectAPI{
		ProjectName: pd.ProjectName,
		Directory:   pd.Path,
		Workspace:   pd.Workspace,
		Ref:         pd.Ref,
		DetectionID: pd.DetectionID,
		HasDrift:    pd.Drift.HasDrift,
		LastChecked: pd.LastChecked,
		Error:       pd.Error,
	}

	if pd.Drift.HasDrift {
		result.Drift = &DriftDetailsAPI{
			ToAdd:          pd.Drift.ToAdd,
			ToChange:       pd.Drift.ToChange,
			ToDestroy:      pd.Drift.ToDestroy,
			ToImport:       pd.Drift.ToImport,
			TotalChanges:   pd.Drift.TotalChanges(),
			Summary:        pd.Drift.Summary,
			ChangesOutside: pd.Drift.ChangesOutside,
		}
	}

	return result
}

// DriftStatusAPI is the API response for drift status.
type DriftStatusAPI struct {
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// Projects contains drift status for each project.
	Projects []DriftProjectAPI `json:"projects"`
	// CheckedAt is when the status was checked.
	CheckedAt time.Time `json:"checked_at"`
	// Summary provides aggregate drift statistics.
	Summary DriftSummaryAPI `json:"summary"`
}

// DriftSummaryAPI is the aggregate drift summary.
type DriftSummaryAPI struct {
	// TotalProjects is the total number of projects.
	TotalProjects int `json:"total_projects"`
	// ProjectsWithDrift is the number of projects with drift.
	ProjectsWithDrift int `json:"projects_with_drift"`
	// ProjectsWithoutDrift is the number of projects without drift.
	ProjectsWithoutDrift int `json:"projects_without_drift"`
	// ProjectsWithErrors is the number of projects with errors.
	ProjectsWithErrors int `json:"projects_with_errors"`
}

// NewDriftStatusAPI converts an internal DriftStatusResponse to its API representation.
func NewDriftStatusAPI(ds models.DriftStatusResponse) DriftStatusAPI {
	result := DriftStatusAPI{
		Repository: ds.Repository,
		Projects:   make([]DriftProjectAPI, 0, len(ds.Projects)),
		CheckedAt:  ds.CheckedAt,
	}

	var withErrors int
	for _, p := range ds.Projects {
		result.Projects = append(result.Projects, NewDriftProjectAPI(p))
		if p.Error != "" {
			withErrors++
		}
	}

	result.Summary = DriftSummaryAPI{
		TotalProjects:        ds.TotalProjects,
		ProjectsWithDrift:    ds.ProjectsWithDrift,
		ProjectsWithoutDrift: ds.TotalProjects - ds.ProjectsWithDrift - withErrors,
		ProjectsWithErrors:   withErrors,
	}

	return result
}

// DriftDetectionResultAPI is the API response for drift detection.
type DriftDetectionResultAPI struct {
	// ID is a unique identifier for this detection run.
	ID string `json:"id"`
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// Projects contains drift status for each project checked.
	Projects []DriftProjectAPI `json:"projects"`
	// DetectedAt is when drift detection was performed.
	DetectedAt time.Time `json:"detected_at"`
	// Summary provides aggregate drift statistics.
	Summary DriftSummaryAPI `json:"summary"`
}

// NewDriftDetectionResultAPI converts an internal DriftDetectionResult to its API representation.
func NewDriftDetectionResultAPI(dr *models.DriftDetectionResult) DriftDetectionResultAPI {
	result := DriftDetectionResultAPI{
		ID:         dr.ID,
		Repository: dr.Repository,
		Projects:   make([]DriftProjectAPI, 0, len(dr.Projects)),
		DetectedAt: dr.DetectedAt,
	}

	var withErrors int
	for _, p := range dr.Projects {
		result.Projects = append(result.Projects, NewDriftProjectAPI(p))
		if p.Error != "" {
			withErrors++
		}
	}

	result.Summary = DriftSummaryAPI{
		TotalProjects:        dr.TotalProjects,
		ProjectsWithDrift:    dr.ProjectsWithDrift,
		ProjectsWithoutDrift: dr.TotalProjects - dr.ProjectsWithDrift - withErrors,
		ProjectsWithErrors:   withErrors,
	}

	return result
}

// RemediationResultAPI is the API response for remediation.
type RemediationResultAPI struct {
	// ID is the unique identifier for this remediation.
	ID string `json:"id"`
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// Ref is the git reference that was remediated.
	Ref string `json:"ref"`
	// Action is the remediation action performed.
	Action string `json:"action"`
	// Status is the overall remediation status.
	Status string `json:"status"`
	// StartedAt is when the remediation started.
	StartedAt time.Time `json:"started_at"`
	// CompletedAt is when the remediation completed.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Projects contains results for each project.
	Projects []RemediationProjectAPI `json:"projects"`
	// Summary provides aggregate statistics.
	Summary RemediationSummaryAPI `json:"summary"`
	// Error contains any top-level error.
	Error string `json:"error,omitempty"`
}

// RemediationProjectAPI is the API representation of a project remediation result.
type RemediationProjectAPI struct {
	// ProjectName is the name of the project.
	ProjectName string `json:"project_name"`
	// Directory is the relative path to the project.
	Directory string `json:"directory"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// Status is the remediation status for this project.
	Status string `json:"status"`
	// PlanOutput contains the plan output.
	PlanOutput string `json:"plan_output,omitempty"`
	// ApplyOutput contains the apply output.
	ApplyOutput string `json:"apply_output,omitempty"`
	// Error contains any error message.
	Error string `json:"error,omitempty"`
	// DriftBefore contains drift state before remediation.
	DriftBefore *DriftDetailsAPI `json:"drift_before,omitempty"`
	// DriftAfter contains drift state after remediation.
	DriftAfter *DriftDetailsAPI `json:"drift_after,omitempty"`
}

// RemediationSummaryAPI provides aggregate remediation statistics.
type RemediationSummaryAPI struct {
	// TotalProjects is the total number of projects targeted.
	TotalProjects int `json:"total_projects"`
	// SuccessCount is the number of successful remediations.
	SuccessCount int `json:"success_count"`
	// FailureCount is the number of failed remediations.
	FailureCount int `json:"failure_count"`
}

// RemediationListAPI is the API response for listing remediation results.
type RemediationListAPI struct {
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// Count is the number of results returned.
	Count int `json:"count"`
	// Results contains the remediation results.
	Results []RemediationResultAPI `json:"results"`
}

// NewRemediationResultAPI converts an internal RemediationResult to its API representation.
func NewRemediationResultAPI(rr *models.RemediationResult) RemediationResultAPI {
	result := RemediationResultAPI{
		ID:         rr.ID,
		Repository: rr.Repository,
		Ref:        rr.Ref,
		Action:     string(rr.Action),
		Status:     string(rr.Status),
		StartedAt:  rr.StartedAt,
		Projects:   make([]RemediationProjectAPI, 0, len(rr.Projects)),
		Error:      rr.Error,
	}

	// Only include completed time if it's set
	if !rr.CompletedAt.IsZero() {
		result.CompletedAt = &rr.CompletedAt
	}

	for _, p := range rr.Projects {
		apiProject := RemediationProjectAPI{
			ProjectName: p.ProjectName,
			Directory:   p.Path,
			Workspace:   p.Workspace,
			Status:      string(p.Status),
			PlanOutput:  p.PlanOutput,
			ApplyOutput: p.ApplyOutput,
			Error:       p.Error,
		}

		if p.DriftBefore != nil && p.DriftBefore.HasDrift {
			apiProject.DriftBefore = &DriftDetailsAPI{
				ToAdd:          p.DriftBefore.ToAdd,
				ToChange:       p.DriftBefore.ToChange,
				ToDestroy:      p.DriftBefore.ToDestroy,
				ToImport:       p.DriftBefore.ToImport,
				TotalChanges:   p.DriftBefore.TotalChanges(),
				Summary:        p.DriftBefore.Summary,
				ChangesOutside: p.DriftBefore.ChangesOutside,
			}
		}

		if p.DriftAfter != nil {
			apiProject.DriftAfter = &DriftDetailsAPI{
				ToAdd:          p.DriftAfter.ToAdd,
				ToChange:       p.DriftAfter.ToChange,
				ToDestroy:      p.DriftAfter.ToDestroy,
				ToImport:       p.DriftAfter.ToImport,
				TotalChanges:   p.DriftAfter.TotalChanges(),
				Summary:        p.DriftAfter.Summary,
				ChangesOutside: p.DriftAfter.ChangesOutside,
			}
		}

		result.Projects = append(result.Projects, apiProject)
	}

	result.Summary = RemediationSummaryAPI{
		TotalProjects: rr.TotalProjects,
		SuccessCount:  rr.SuccessCount,
		FailureCount:  rr.FailureCount,
	}

	return result
}

// LockDetailAPI is the API representation of a project lock.
type LockDetailAPI struct {
	// ID is the unique lock identifier.
	ID string `json:"id"`
	// ProjectName is the name of the locked project.
	ProjectName string `json:"project_name"`
	// Repository is the full repository name (owner/repo).
	Repository string `json:"repository"`
	// Path is the relative path to the project within the repository.
	Path string `json:"path"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// PullRequestID is the PR number that holds the lock.
	PullRequestID int `json:"pull_request_id"`
	// PullRequestURL is the URL to the pull request.
	PullRequestURL string `json:"pull_request_url"`
	// LockedBy is the username who created the lock.
	LockedBy string `json:"locked_by"`
	// LockedAt is when the lock was acquired.
	LockedAt time.Time `json:"locked_at"`
}

// ListLocksResultAPI is the API response for listing locks.
type ListLocksResultAPI struct {
	// Locks contains the list of active locks.
	Locks []LockDetailAPI `json:"locks"`
	// TotalCount is the total number of locks.
	TotalCount int `json:"total_count"`
}

// NewListLocksResultAPI creates a ListLocksResultAPI from the internal lock map.
// This converts from the Locker's internal representation to the API representation.
func NewListLocksResultAPI(locks map[string]models.ProjectLock) ListLocksResultAPI {
	result := ListLocksResultAPI{
		Locks: make([]LockDetailAPI, 0, len(locks)),
	}

	for name, lock := range locks {
		result.Locks = append(result.Locks, LockDetailAPI{
			ID:             name,
			ProjectName:    lock.Project.ProjectName,
			Repository:     lock.Project.RepoFullName,
			Path:           lock.Project.Path,
			Workspace:      lock.Workspace,
			PullRequestID:  lock.Pull.Num,
			PullRequestURL: lock.Pull.URL,
			LockedBy:       lock.User.Username,
			LockedAt:       lock.Time,
		})
	}

	result.TotalCount = len(result.Locks)
	return result
}
