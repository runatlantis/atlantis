// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import "time"

// RemediationAction specifies the type of remediation to perform.
type RemediationAction string

const (
	// RemediationPlanOnly runs a plan to preview remediation without applying.
	RemediationPlanOnly RemediationAction = "plan"
	// RemediationAutoApply runs both plan and apply to fix drift automatically.
	RemediationAutoApply RemediationAction = "apply"
)

// IsValid returns true if the action is a recognized value.
func (a RemediationAction) IsValid() bool {
	switch a {
	case RemediationPlanOnly, RemediationAutoApply, "":
		return true
	default:
		return false
	}
}

// Default returns the default action if empty.
func (a RemediationAction) Default() RemediationAction {
	if a == "" {
		return RemediationPlanOnly
	}
	return a
}

// RemediationStatus represents the current status of a remediation.
type RemediationStatus string

const (
	// RemediationStatusPending indicates remediation is queued.
	RemediationStatusPending RemediationStatus = "pending"
	// RemediationStatusRunning indicates remediation is in progress.
	RemediationStatusRunning RemediationStatus = "running"
	// RemediationStatusSuccess indicates remediation completed successfully.
	RemediationStatusSuccess RemediationStatus = "success"
	// RemediationStatusFailed indicates remediation failed.
	RemediationStatusFailed RemediationStatus = "failed"
	// RemediationStatusPartial indicates some projects succeeded, some failed.
	RemediationStatusPartial RemediationStatus = "partial"
)

// IsTerminal returns true if the status represents a final state.
func (s RemediationStatus) IsTerminal() bool {
	switch s {
	case RemediationStatusSuccess, RemediationStatusFailed, RemediationStatusPartial:
		return true
	default:
		return false
	}
}

// RemediationRequest is the API request for POST /api/drift/remediate.
type RemediationRequest struct {
	// Repository is the full repository name (owner/repo). Required.
	Repository string `json:"repository"`
	// Ref is the git reference (branch/tag/commit) to remediate. Required.
	Ref string `json:"ref"`
	// Type is the VCS provider type (Github/Gitlab). Required.
	Type string `json:"type"`
	// Action specifies plan-only or auto-apply. Defaults to "plan".
	Action RemediationAction `json:"action"`
	// Projects is a list of project names to remediate. If empty, all projects with drift are remediated.
	Projects []string `json:"projects,omitempty"`
	// Workspaces filters remediation to specific workspaces. If empty, all workspaces.
	Workspaces []string `json:"workspaces,omitempty"`
	// DriftOnly if true, only remediates projects that have detected drift.
	DriftOnly bool `json:"drift_only"`
}

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate checks the request and returns any validation errors.
func (r *RemediationRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Repository == "" {
		errors = append(errors, FieldError{Field: "repository", Message: "repository is required"})
	}
	if r.Ref == "" {
		errors = append(errors, FieldError{Field: "ref", Message: "ref is required"})
	}
	if r.Type == "" {
		errors = append(errors, FieldError{Field: "type", Message: "type is required"})
	} else if r.Type != "Github" && r.Type != "Gitlab" && r.Type != "BitbucketCloud" && r.Type != "BitbucketServer" && r.Type != "AzureDevops" && r.Type != "Gitea" {
		errors = append(errors, FieldError{Field: "type", Message: "type must be one of: Github, Gitlab, BitbucketCloud, BitbucketServer, AzureDevops, Gitea"})
	}
	if !r.Action.IsValid() {
		errors = append(errors, FieldError{Field: "action", Message: "action must be 'plan' or 'apply'"})
	}

	return errors
}

// ApplyDefaults applies default values to optional fields.
func (r *RemediationRequest) ApplyDefaults() {
	r.Action = r.Action.Default()
}

// ProjectRemediationResult represents the remediation result for a single project.
type ProjectRemediationResult struct {
	// ProjectName is the name of the project.
	ProjectName string `json:"project_name"`
	// Path is the relative path to the project.
	Path string `json:"path"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// Status is the remediation status for this project.
	Status RemediationStatus `json:"status"`
	// PlanOutput contains the plan output if available.
	PlanOutput string `json:"plan_output,omitempty"`
	// ApplyOutput contains the apply output if action was auto-apply.
	ApplyOutput string `json:"apply_output,omitempty"`
	// Error contains any error message.
	Error string `json:"error,omitempty"`
	// DriftBefore contains the drift state before remediation.
	DriftBefore *DriftSummary `json:"drift_before,omitempty"`
	// DriftAfter contains the drift state after remediation (nil if not re-checked).
	DriftAfter *DriftSummary `json:"drift_after,omitempty"`
}

// RemediationResult is the API response for POST /api/drift/remediate.
type RemediationResult struct {
	// ID is a unique identifier for this remediation run.
	ID string `json:"id"`
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// Ref is the git reference that was remediated.
	Ref string `json:"ref"`
	// Action is the remediation action performed.
	Action RemediationAction `json:"action"`
	// Status is the overall remediation status.
	Status RemediationStatus `json:"status"`
	// StartedAt is when the remediation started.
	StartedAt time.Time `json:"started_at"`
	// CompletedAt is when the remediation completed (zero if still running).
	CompletedAt time.Time `json:"completed_at,omitempty"`
	// TotalProjects is the total number of projects targeted.
	TotalProjects int `json:"total_projects"`
	// SuccessCount is the number of projects successfully remediated.
	SuccessCount int `json:"success_count"`
	// FailureCount is the number of projects that failed remediation.
	FailureCount int `json:"failure_count"`
	// Projects contains the remediation result for each project.
	Projects []ProjectRemediationResult `json:"projects"`
	// Error contains any top-level error message.
	Error string `json:"error,omitempty"`
}

// NewRemediationResult creates a new RemediationResult with initial values.
func NewRemediationResult(id, repository, ref string, action RemediationAction) *RemediationResult {
	return &RemediationResult{
		ID:         id,
		Repository: repository,
		Ref:        ref,
		Action:     action,
		Status:     RemediationStatusPending,
		StartedAt:  time.Now(),
		Projects:   []ProjectRemediationResult{},
	}
}

// Complete marks the remediation as complete and calculates final status.
func (r *RemediationResult) Complete() {
	r.CompletedAt = time.Now()
	r.TotalProjects = len(r.Projects)

	// Reset counts before recomputing to avoid double-counting if Complete() is called multiple times
	r.SuccessCount = 0
	r.FailureCount = 0

	for _, p := range r.Projects {
		switch p.Status {
		case RemediationStatusSuccess:
			r.SuccessCount++
		case RemediationStatusFailed:
			r.FailureCount++
		}
	}

	// Determine overall status
	if r.FailureCount == 0 && r.SuccessCount > 0 {
		r.Status = RemediationStatusSuccess
	} else if r.SuccessCount == 0 && r.FailureCount > 0 {
		r.Status = RemediationStatusFailed
	} else if r.SuccessCount > 0 && r.FailureCount > 0 {
		r.Status = RemediationStatusPartial
	} else {
		// No projects processed
		r.Status = RemediationStatusSuccess
	}
}

// AddProjectResult adds a project remediation result.
func (r *RemediationResult) AddProjectResult(result ProjectRemediationResult) {
	r.Projects = append(r.Projects, result)
}
