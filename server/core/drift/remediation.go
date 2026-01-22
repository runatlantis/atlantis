// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift

//go:generate pegomock generate --package mocks -o mocks/mock_remediation_service.go RemediationService

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RemediationService handles drift remediation operations.
type RemediationService interface {
	// Remediate executes drift remediation for the given projects.
	// It returns a result ID that can be used to track the remediation status.
	Remediate(req models.RemediationRequest, executor RemediationExecutor) (*models.RemediationResult, error)

	// GetResult retrieves a remediation result by ID.
	GetResult(id string) (*models.RemediationResult, error)

	// ListResults returns all remediation results for a repository.
	ListResults(repository string, limit int) ([]*models.RemediationResult, error)
}

// RemediationExecutor executes the actual plan/apply operations.
// This interface allows the service to be decoupled from the API controller.
type RemediationExecutor interface {
	// ExecutePlan runs a plan for the given project.
	ExecutePlan(repository, ref, vcsType, projectName, path, workspace string) (output string, driftSummary *models.DriftSummary, err error)

	// ExecuteApply runs an apply for the given project.
	ExecuteApply(repository, ref, vcsType, projectName, path, workspace string) (output string, err error)
}

// InMemoryRemediationService implements RemediationService with in-memory storage.
type InMemoryRemediationService struct {
	mu           sync.RWMutex
	results      map[string]*models.RemediationResult
	repoResults  map[string][]string // repository -> result IDs
	driftStorage Storage
}

// NewInMemoryRemediationService creates a new in-memory remediation service.
func NewInMemoryRemediationService(driftStorage Storage) *InMemoryRemediationService {
	return &InMemoryRemediationService{
		results:      make(map[string]*models.RemediationResult),
		repoResults:  make(map[string][]string),
		driftStorage: driftStorage,
	}
}

// Remediate executes drift remediation for the given projects.
func (s *InMemoryRemediationService) Remediate(req models.RemediationRequest, executor RemediationExecutor) (*models.RemediationResult, error) {
	// Validate action, default to plan-only
	if req.Action == "" {
		req.Action = models.RemediationPlanOnly
	}
	if req.Action != models.RemediationPlanOnly && req.Action != models.RemediationAutoApply {
		return nil, fmt.Errorf("invalid action: %s, must be 'plan' or 'apply'", req.Action)
	}

	// Generate unique ID
	id := uuid.New().String()

	// Create result
	result := models.NewRemediationResult(id, req.Repository, req.Ref, req.Action)
	result.Status = models.RemediationStatusRunning

	// Store initial result
	s.storeResult(result)

	// Get projects to remediate
	projects, err := s.getProjectsToRemediate(req)
	if err != nil {
		result.Error = err.Error()
		result.Status = models.RemediationStatusFailed
		result.Complete()
		s.storeResult(result)
		return result, nil
	}

	if len(projects) == 0 {
		result.Complete()
		s.storeResult(result)
		return result, nil
	}

	// Execute remediation for each project
	for _, proj := range projects {
		projectResult := s.remediateProject(req, proj, executor)
		result.AddProjectResult(projectResult)
	}

	// Mark as complete
	result.Complete()
	s.storeResult(result)

	return result, nil
}

// getProjectsToRemediate determines which projects to remediate based on the request.
func (s *InMemoryRemediationService) getProjectsToRemediate(req models.RemediationRequest) ([]models.ProjectDrift, error) {
	var projects []models.ProjectDrift

	// If drift storage is available and DriftOnly is true, get projects with drift
	if s.driftStorage != nil && req.DriftOnly {
		opts := GetOptions{}
		drifts, err := s.driftStorage.Get(req.Repository, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get drift data: %w", err)
		}

		for _, d := range drifts {
			// Apply filters
			if !s.matchesFilters(d, req) {
				continue
			}
			// Only include projects with drift
			if d.Drift.HasDrift {
				projects = append(projects, d)
			}
		}
	} else if len(req.Projects) > 0 {
		// Specific projects requested
		for _, projName := range req.Projects {
			projects = append(projects, models.ProjectDrift{
				ProjectName: projName,
				Ref:         req.Ref,
			})
		}
	} else if s.driftStorage != nil {
		// Get all projects from drift storage
		drifts, err := s.driftStorage.Get(req.Repository, GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get drift data: %w", err)
		}

		for _, d := range drifts {
			if !s.matchesFilters(d, req) {
				continue
			}
			projects = append(projects, d)
		}
	}

	return projects, nil
}

// matchesFilters checks if a project matches the request filters.
func (s *InMemoryRemediationService) matchesFilters(proj models.ProjectDrift, req models.RemediationRequest) bool {
	// Check project name filter
	if len(req.Projects) > 0 {
		found := false
		for _, p := range req.Projects {
			if p == proj.ProjectName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check workspace filter
	if len(req.Workspaces) > 0 {
		found := false
		for _, w := range req.Workspaces {
			if w == proj.Workspace {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// remediateProject executes remediation for a single project.
func (s *InMemoryRemediationService) remediateProject(req models.RemediationRequest, proj models.ProjectDrift, executor RemediationExecutor) models.ProjectRemediationResult {
	result := models.ProjectRemediationResult{
		ProjectName: proj.ProjectName,
		Path:        proj.Path,
		Workspace:   proj.Workspace,
		Status:      models.RemediationStatusRunning,
	}

	// Store drift before remediation
	if proj.Drift.HasDrift {
		result.DriftBefore = &proj.Drift
	}

	// Execute plan
	planOutput, driftSummary, err := executor.ExecutePlan(
		req.Repository,
		req.Ref,
		req.Type,
		proj.ProjectName,
		proj.Path,
		proj.Workspace,
	)
	result.PlanOutput = planOutput

	if err != nil {
		result.Error = err.Error()
		result.Status = models.RemediationStatusFailed
		return result
	}

	// If plan-only, we're done
	if req.Action == models.RemediationPlanOnly {
		result.Status = models.RemediationStatusSuccess
		if driftSummary != nil {
			result.DriftAfter = driftSummary
		}
		return result
	}

	// Execute apply for auto-apply action
	applyOutput, err := executor.ExecuteApply(
		req.Repository,
		req.Ref,
		req.Type,
		proj.ProjectName,
		proj.Path,
		proj.Workspace,
	)
	result.ApplyOutput = applyOutput

	if err != nil {
		result.Error = err.Error()
		result.Status = models.RemediationStatusFailed
		return result
	}

	result.Status = models.RemediationStatusSuccess

	// After apply, drift should be resolved
	result.DriftAfter = &models.DriftSummary{
		HasDrift: false,
		Summary:  "Apply completed successfully",
	}

	return result
}

// storeResult stores a remediation result.
func (s *InMemoryRemediationService) storeResult(result *models.RemediationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results[result.ID] = result

	// Track by repository
	if _, ok := s.repoResults[result.Repository]; !ok {
		s.repoResults[result.Repository] = []string{}
	}

	// Check if ID already exists in repo results
	found := false
	for _, id := range s.repoResults[result.Repository] {
		if id == result.ID {
			found = true
			break
		}
	}
	if !found {
		s.repoResults[result.Repository] = append(s.repoResults[result.Repository], result.ID)
	}
}

// GetResult retrieves a remediation result by ID.
func (s *InMemoryRemediationService) GetResult(id string) (*models.RemediationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.results[id]
	if !ok {
		return nil, fmt.Errorf("remediation result not found: %s", id)
	}
	return result, nil
}

// ListResults returns all remediation results for a repository.
func (s *InMemoryRemediationService) ListResults(repository string, limit int) ([]*models.RemediationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids, ok := s.repoResults[repository]
	if !ok {
		return []*models.RemediationResult{}, nil
	}

	results := make([]*models.RemediationResult, 0, len(ids))
	for i := len(ids) - 1; i >= 0 && (limit <= 0 || len(results) < limit); i-- {
		if result, ok := s.results[ids[i]]; ok {
			results = append(results, result)
		}
	}

	return results, nil
}

// RemediationHistory represents the history of remediations for tracking.
type RemediationHistory struct {
	// ID is the unique identifier for this history entry.
	ID string `json:"id"`
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// RemediationID links to the remediation result.
	RemediationID string `json:"remediation_id"`
	// Action is the remediation action performed.
	Action models.RemediationAction `json:"action"`
	// Status is the final status.
	Status models.RemediationStatus `json:"status"`
	// Timestamp is when the remediation occurred.
	Timestamp time.Time `json:"timestamp"`
	// ProjectCount is the number of projects remediated.
	ProjectCount int `json:"project_count"`
	// SuccessCount is the number of successful remediations.
	SuccessCount int `json:"success_count"`
}
