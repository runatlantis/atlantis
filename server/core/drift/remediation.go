// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift

//go:generate go tool pegomock generate --package mocks -o mocks/mock_remediation_service.go RemediationService

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/events/models"
)

const defaultWorkspace = "default"

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

	// ExecuteApplyProjects plans and applies the given projects as one operation.
	ExecuteApplyProjects(repository, ref, vcsType string, projects []models.ProjectDrift) ([]models.ProjectRemediationResult, error)
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
	result.StorageRepository = remediationStorageRepository(req)
	result.Status = models.RemediationStatusRunning

	// Store initial result
	s.storeResult(result)

	// Get projects to remediate
	projects, err := s.getProjectsToRemediate(req)
	if err != nil {
		result.Error = err.Error()
		result.Status = models.RemediationStatusFailed
		completedAt := time.Now()
		result.CompletedAt = &completedAt
		s.storeResult(result)
		return result, nil
	}
	projects, err = deduplicateRemediationTargets(req, projects)
	if err != nil {
		result.Error = err.Error()
		result.Status = models.RemediationStatusFailed
		completedAt := time.Now()
		result.CompletedAt = &completedAt
		s.storeResult(result)
		return result, nil
	}

	if len(projects) == 0 {
		if req.Action == models.RemediationAutoApply {
			result.Status = models.RemediationStatusFailed
			result.Error = "cached drift with has_drift=true is required before remediation apply; run drift detection for the same ref, base_branch, project, path, and workspace first"
			completedAt := time.Now()
			result.CompletedAt = &completedAt
		} else {
			result.Complete()
		}
		s.storeResult(result)
		return result, nil
	}

	if req.Action == models.RemediationAutoApply {
		if err := requireCachedApplyTargets(projects); err != nil {
			for _, projectResult := range failedProjectRemediationResults(projects, err.Error()) {
				result.AddProjectResult(projectResult)
			}
			result.Complete()
			s.storeResult(result)
			return result, nil
		}
		for _, projectResult := range s.remediateProjectsWithApply(req, projects, executor) {
			result.AddProjectResult(projectResult)
		}
	} else {
		// Execute remediation for each project
		for _, proj := range projects {
			projectResult := s.remediateProject(req, proj, executor)
			result.AddProjectResult(projectResult)
		}
	}

	// Mark as complete
	result.Complete()
	s.storeResult(result)

	return result, nil
}

func (s *InMemoryRemediationService) remediateProjectsWithApply(req models.RemediationRequest, projects []models.ProjectDrift, executor RemediationExecutor) []models.ProjectRemediationResult {
	results, err := executor.ExecuteApplyProjects(req.Repository, remediationExecutionRef(req), req.Type, projects)
	if err != nil && len(results) == 0 {
		return failedProjectRemediationResults(projects, err.Error())
	}
	results = s.completeApplyProjectResults(req, projects, results, err)
	return results
}

func failedProjectRemediationResults(projects []models.ProjectDrift, errorMessage string) []models.ProjectRemediationResult {
	results := make([]models.ProjectRemediationResult, 0, len(projects))
	for _, proj := range projects {
		results = append(results, models.ProjectRemediationResult{
			ProjectName: proj.ProjectName,
			Path:        proj.Path,
			Workspace:   proj.Workspace,
			Status:      models.RemediationStatusFailed,
			Error:       errorMessage,
			DriftBefore: cloneDriftSummaryPtr(&proj.Drift),
		})
	}
	return results
}

func (s *InMemoryRemediationService) completeApplyProjectResults(req models.RemediationRequest, projects []models.ProjectDrift, results []models.ProjectRemediationResult, applyErr error) []models.ProjectRemediationResult {
	projectsByKey := make(map[string]models.ProjectDrift, len(projects))
	for _, proj := range projects {
		projectsByKey[remediationProjectKey(proj.ProjectName, proj.Path, proj.Workspace)] = proj
	}

	matchedProjectKeys := make(map[string]struct{}, len(results))
	for i := range results {
		result := &results[i]
		proj, projectKey, ok := findRemediationProjectForResult(projects, projectsByKey, *result)
		if ok {
			matchedProjectKeys[projectKey] = struct{}{}
		}
		if ok && result.DriftBefore == nil && proj.Drift.HasDrift {
			result.DriftBefore = cloneDriftSummaryPtr(&proj.Drift)
		}
		if result.Status == models.RemediationStatusRunning {
			result.Status = models.RemediationStatusFailed
			if result.Error == "" && applyErr != nil {
				result.Error = applyErr.Error()
			}
		}
		if result.Status == models.RemediationStatusSuccess && result.DriftAfter == nil {
			result.DriftAfter = &models.DriftSummary{
				HasDrift: false,
				Summary:  "Apply completed successfully",
			}
		}
		if result.Status == models.RemediationStatusSuccess && s.driftStorage != nil && ok && proj.ResolvedCommit != "" {
			updatedDrift := proj
			updatedDrift.ProjectName = result.ProjectName
			updatedDrift.Path = result.Path
			updatedDrift.Workspace = result.Workspace
			updatedDrift.Ref = remediationRef(req)
			updatedDrift.BaseBranch = remediationBaseBranch(req)
			updatedDrift.Drift = *result.DriftAfter
			updatedDrift.Error = ""
			updatedDrift.LastChecked = time.Now()
			if err := s.driftStorage.Store(remediationStorageRepository(req), updatedDrift); err != nil {
				result.Error = fmt.Sprintf("updating drift status after apply: %v", err)
				result.Status = models.RemediationStatusFailed
			}
		}
	}

	for _, proj := range projects {
		key := remediationProjectKey(proj.ProjectName, proj.Path, proj.Workspace)
		if _, ok := matchedProjectKeys[key]; ok {
			continue
		}
		results = append(results, models.ProjectRemediationResult{
			ProjectName: proj.ProjectName,
			Path:        proj.Path,
			Workspace:   proj.Workspace,
			Status:      models.RemediationStatusFailed,
			Error:       "apply did not return a result for project",
			DriftBefore: cloneDriftSummaryPtr(&proj.Drift),
		})
	}

	return results
}

// getProjectsToRemediate determines which projects to remediate based on the request.
func (s *InMemoryRemediationService) getProjectsToRemediate(req models.RemediationRequest) ([]models.ProjectDrift, error) {
	var err error
	req, err = normalizeRemediationRequest(req)
	if err != nil {
		return nil, err
	}

	var projects []models.ProjectDrift
	storageRepository := remediationStorageRepository(req)

	// Drift records are keyed by ref, so when reading from storage we restrict
	// to the requested ref. This prevents remediating a project that drifted on
	// a different branch/commit than the one the caller asked about, which is
	// especially risky for auto-apply.
	opts := GetOptions{Ref: remediationRef(req), BaseBranch: remediationBaseBranch(req)}

	// If drift storage is available and DriftOnly is true, get projects with drift
	if s.driftStorage != nil && req.DriftOnly {
		drifts, err := s.driftStorage.Get(storageRepository, opts)
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
	} else if len(req.Paths) > 0 {
		for _, path := range req.Paths {
			workspaces, err := remediationPathWorkspaces(path, req.Workspaces)
			if err != nil {
				return nil, err
			}
			for _, workspace := range workspaces {
				pathOpts := GetOptions{Path: path.Directory, Workspace: workspace, Ref: remediationRef(req), BaseBranch: remediationBaseBranch(req)}
				if s.driftStorage != nil {
					drifts, err := s.driftStorage.Get(storageRepository, pathOpts)
					if err != nil {
						return nil, fmt.Errorf("failed to get drift data for path %q: %w", path.Directory, err)
					}
					for _, d := range drifts {
						if s.matchesFilters(d, req) {
							projects = append(projects, d)
						}
					}
				}
				for _, projectName := range remediationPathProjectNames(req.Projects) {
					if !hasRemediationPathTarget(projects, path.Directory, workspace, projectName) {
						projects = append(projects, models.ProjectDrift{
							ProjectName: projectName,
							Path:        path.Directory,
							Workspace:   workspace,
							Ref:         remediationRef(req),
							BaseBranch:  remediationBaseBranch(req),
						})
					}
				}
			}
		}
	} else if len(req.Projects) > 0 {
		// Specific projects requested
		for _, projName := range req.Projects {
			if s.driftStorage != nil {
				drifts, err := s.driftStorage.Get(storageRepository, GetOptions{ProjectName: projName, Ref: remediationRef(req), BaseBranch: remediationBaseBranch(req)})
				if err != nil {
					return nil, fmt.Errorf("failed to get drift data for project %q: %w", projName, err)
				}
				for _, d := range drifts {
					if s.matchesFilters(d, req) {
						projects = append(projects, d)
					}
				}
			}
			if len(req.Workspaces) == 0 {
				if !hasRemediationTarget(projects, projName, "") {
					projects = append(projects, models.ProjectDrift{
						ProjectName: projName,
						Ref:         remediationRef(req),
						BaseBranch:  remediationBaseBranch(req),
					})
				}
				continue
			}
			for _, workspace := range req.Workspaces {
				if !hasRemediationTarget(projects, projName, workspace) {
					projects = append(projects, models.ProjectDrift{
						ProjectName: projName,
						Workspace:   workspace,
						Ref:         remediationRef(req),
						BaseBranch:  remediationBaseBranch(req),
					})
				}
			}
		}
	} else if s.driftStorage != nil {
		// Get all projects from drift storage for the requested ref
		drifts, err := s.driftStorage.Get(storageRepository, opts)
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

func normalizeRemediationRequest(req models.RemediationRequest) (models.RemediationRequest, error) {
	for i := range req.Paths {
		directory, ok := models.NormalizeAPIPath(req.Paths[i].Directory)
		if !ok {
			return req, fmt.Errorf("path directory %q is invalid", req.Paths[i].Directory)
		}
		req.Paths[i].Directory = directory
	}
	return req, nil
}

func remediationStorageRepository(req models.RemediationRequest) string {
	if req.StorageRepository != "" {
		return req.StorageRepository
	}
	return req.Repository
}

func remediationRef(req models.RemediationRequest) string {
	return models.NormalizeAPIRef(req.Ref)
}

func remediationExecutionRef(req models.RemediationRequest) string {
	if req.ExecutionRef != "" {
		return req.ExecutionRef
	}
	return req.Ref
}

func remediationBaseBranch(req models.RemediationRequest) string {
	if req.BaseBranch != "" {
		return models.NormalizeAPIRef(req.BaseBranch)
	}
	return models.NormalizeAPIRef(req.Ref)
}

func remediationPathWorkspaces(path models.DriftDetectionPath, workspaces []string) ([]string, error) {
	if path.Workspace != "" {
		if len(workspaces) > 0 && !slices.Contains(workspaces, path.Workspace) {
			return nil, fmt.Errorf("path workspace %q must be included in workspaces", path.Workspace)
		}
		return []string{path.Workspace}, nil
	}
	if len(workspaces) > 0 {
		return workspaces, nil
	}
	return []string{defaultWorkspace}, nil
}

type remediationTargetIdentity struct {
	Repository  string
	Type        string
	Ref         string
	BaseBranch  string
	ProjectName string
	Path        string
	Workspace   string
}

func deduplicateRemediationTargets(req models.RemediationRequest, projects []models.ProjectDrift) ([]models.ProjectDrift, error) {
	if len(projects) < 2 {
		return projects, nil
	}
	deduped := make([]models.ProjectDrift, 0, len(projects))
	seen := make(map[remediationTargetIdentity]models.ProjectDrift, len(projects))
	for _, project := range projects {
		identity := newRemediationTargetIdentity(req, project)
		if existing, ok := seen[identity]; ok {
			if !sameRemediationSafetyMetadata(existing, project) {
				return nil, fmt.Errorf("conflicting cached drift records for %s; rerun drift detection before remediation", remediationTargetDescription(project))
			}
			continue
		}
		seen[identity] = project
		deduped = append(deduped, project)
	}
	return deduped, nil
}

func newRemediationTargetIdentity(req models.RemediationRequest, project models.ProjectDrift) remediationTargetIdentity {
	ref := project.Ref
	if ref == "" {
		ref = remediationRef(req)
	}
	baseBranch := project.BaseBranch
	if baseBranch == "" {
		baseBranch = remediationBaseBranch(req)
	}
	return remediationTargetIdentity{
		Repository:  remediationStorageRepository(req),
		Type:        req.Type,
		Ref:         ref,
		BaseBranch:  baseBranch,
		ProjectName: project.ProjectName,
		Path:        project.Path,
		Workspace:   project.Workspace,
	}
}

func sameRemediationSafetyMetadata(a, b models.ProjectDrift) bool {
	return a.ResolvedCommit == b.ResolvedCommit &&
		a.DetectionID == b.DetectionID &&
		a.LastChecked.Equal(b.LastChecked) &&
		a.Drift.HasDrift == b.Drift.HasDrift
}

func requireCachedApplyTargets(projects []models.ProjectDrift) error {
	for _, project := range projects {
		if project.ResolvedCommit == "" || project.LastChecked.IsZero() {
			return fmt.Errorf("cached drift with a resolved commit is required before remediation apply for %s; run drift detection for the same ref, base_branch, project, path, and workspace first", remediationTargetDescription(project))
		}
		if !project.Drift.HasDrift {
			return fmt.Errorf("cached drift with has_drift=true is required before remediation apply for %s; rerun drift detection or use action: plan for a non-destructive preview", remediationTargetDescription(project))
		}
	}
	return nil
}

func remediationTargetDescription(project models.ProjectDrift) string {
	if project.ProjectName != "" {
		return fmt.Sprintf("project %q", project.ProjectName)
	}
	if project.Path != "" {
		return fmt.Sprintf("path %q workspace %q", project.Path, project.Workspace)
	}
	return "selected project"
}

func remediationPathProjectNames(projects []string) []string {
	if len(projects) == 0 {
		return []string{""}
	}
	return projects
}

func hasRemediationTarget(projects []models.ProjectDrift, projectName string, workspace string) bool {
	for _, p := range projects {
		if p.ProjectName != projectName {
			continue
		}
		if workspace == "" || p.Workspace == workspace {
			return true
		}
	}
	return false
}

func hasRemediationPathTarget(projects []models.ProjectDrift, path string, workspace string, projectName string) bool {
	for _, p := range projects {
		if p.Path != path {
			continue
		}
		if projectName != "" && p.ProjectName != projectName {
			continue
		}
		if p.Workspace == workspace {
			return true
		}
	}
	return false
}

// matchesFilters checks if a project matches the request filters.
func (s *InMemoryRemediationService) matchesFilters(proj models.ProjectDrift, req models.RemediationRequest) bool {
	// Check project name filter
	if len(req.Projects) > 0 {
		if !slices.Contains(req.Projects, proj.ProjectName) {
			return false
		}
	}

	// Check workspace filter
	if len(req.Workspaces) > 0 {
		if !slices.Contains(req.Workspaces, proj.Workspace) {
			return false
		}
	}
	if len(req.Paths) > 0 {
		matched := false
		for _, path := range req.Paths {
			if proj.Path != path.Directory {
				continue
			}
			if path.Workspace != "" && proj.Workspace != path.Workspace {
				continue
			}
			if path.Workspace == "" && len(req.Workspaces) == 0 && proj.Workspace != defaultWorkspace {
				continue
			}
			matched = true
			break
		}
		if !matched {
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
		remediationExecutionRef(req),
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

	result.Status = models.RemediationStatusSuccess
	if driftSummary != nil {
		result.DriftAfter = driftSummary
	}
	return result
}

// storeResult stores a remediation result.
func (s *InMemoryRemediationService) storeResult(result *models.RemediationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results[result.ID] = cloneRemediationResult(result)

	repositoryKey := remediationResultRepositoryKey(result)

	// Track by repository
	if _, ok := s.repoResults[repositoryKey]; !ok {
		s.repoResults[repositoryKey] = []string{}
	}

	// Check if ID already exists in repo results
	if !slices.Contains(s.repoResults[repositoryKey], result.ID) {
		s.repoResults[repositoryKey] = append(s.repoResults[repositoryKey], result.ID)
	}
}

func remediationProjectKey(projectName, path, workspace string) string {
	return projectName + "\n" + path + "\n" + workspace
}

func findRemediationProjectForResult(projects []models.ProjectDrift, projectsByKey map[string]models.ProjectDrift, result models.ProjectRemediationResult) (models.ProjectDrift, string, bool) {
	key := remediationProjectKey(result.ProjectName, result.Path, result.Workspace)
	if proj, ok := projectsByKey[key]; ok {
		return proj, key, true
	}
	for _, proj := range projects {
		if proj.ProjectName != "" || proj.Path != result.Path {
			continue
		}
		if proj.Workspace != "" && proj.Workspace != result.Workspace {
			continue
		}
		return proj, remediationProjectKey(proj.ProjectName, proj.Path, proj.Workspace), true
	}
	for _, proj := range projects {
		if proj.ProjectName == "" || proj.ProjectName != result.ProjectName || proj.Path != "" {
			continue
		}
		if proj.Workspace != "" && proj.Workspace != result.Workspace {
			continue
		}
		return proj, remediationProjectKey(proj.ProjectName, proj.Path, proj.Workspace), true
	}
	return models.ProjectDrift{}, "", false
}

func cloneRemediationResult(result *models.RemediationResult) *models.RemediationResult {
	if result == nil {
		return nil
	}
	clone := *result
	clone.Projects = make([]models.ProjectRemediationResult, len(result.Projects))
	for i := range result.Projects {
		clone.Projects[i] = cloneProjectRemediationResult(result.Projects[i])
	}
	return &clone
}

func cloneProjectRemediationResult(result models.ProjectRemediationResult) models.ProjectRemediationResult {
	clone := result
	clone.DriftBefore = cloneDriftSummaryPtr(result.DriftBefore)
	clone.DriftAfter = cloneDriftSummaryPtr(result.DriftAfter)
	return clone
}

func cloneDriftSummaryPtr(summary *models.DriftSummary) *models.DriftSummary {
	if summary == nil {
		return nil
	}
	clone := *summary
	return &clone
}

func remediationResultRepositoryKey(result *models.RemediationResult) string {
	if result.StorageRepository != "" {
		return result.StorageRepository
	}
	return result.Repository
}

// GetResult retrieves a remediation result by ID.
func (s *InMemoryRemediationService) GetResult(id string) (*models.RemediationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.results[id]
	if !ok {
		return nil, fmt.Errorf("remediation result not found: %s", id)
	}
	return cloneRemediationResult(result), nil
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
			results = append(results, cloneRemediationResult(result))
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
