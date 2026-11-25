// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/drift"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

type remediationExecutorCall struct {
	projectName string
	path        string
	workspace   string
}

type recordingRemediationExecutor struct {
	planCalls           []remediationExecutorCall
	applyCalls          []remediationExecutorCall
	applyProjectCalls   [][]remediationExecutorCall
	planRefs            []string
	applyProjectRefs    []string
	applyProjectResults []models.ProjectRemediationResult
	applyProjectErr     error
	planDrift           *models.DriftSummary
}

type failingRemediationStorage struct{}

func (f failingRemediationStorage) Store(string, models.ProjectDrift) error {
	return nil
}

func (f failingRemediationStorage) Get(string, drift.GetOptions) ([]models.ProjectDrift, error) {
	return nil, errors.New("storage unavailable")
}

func (f failingRemediationStorage) Delete(string, string) error {
	return nil
}

func (f failingRemediationStorage) DeleteMatching(string, drift.GetOptions) error {
	return nil
}

func (f failingRemediationStorage) GetAll() (map[string][]models.ProjectDrift, error) {
	return nil, nil
}

type staticRemediationStorage struct {
	drifts []models.ProjectDrift
}

func (s staticRemediationStorage) Store(string, models.ProjectDrift) error {
	return nil
}

func (s staticRemediationStorage) Get(string, drift.GetOptions) ([]models.ProjectDrift, error) {
	return s.drifts, nil
}

func (s staticRemediationStorage) Delete(string, string) error {
	return nil
}

func (s staticRemediationStorage) DeleteMatching(string, drift.GetOptions) error {
	return nil
}

func (s staticRemediationStorage) GetAll() (map[string][]models.ProjectDrift, error) {
	return nil, nil
}

func (r *recordingRemediationExecutor) ExecutePlan(_, ref, _ string, projectName, path, workspace string) (string, *models.DriftSummary, error) {
	r.planRefs = append(r.planRefs, ref)
	r.planCalls = append(r.planCalls, remediationExecutorCall{
		projectName: projectName,
		path:        path,
		workspace:   workspace,
	})
	if r.planDrift != nil {
		return "plan", r.planDrift, nil
	}
	return "plan", &models.DriftSummary{}, nil
}

func (r *recordingRemediationExecutor) ExecuteApply(_, _, _, projectName, path, workspace string) (string, error) {
	r.applyCalls = append(r.applyCalls, remediationExecutorCall{
		projectName: projectName,
		path:        path,
		workspace:   workspace,
	})
	return "apply", nil
}

func (r *recordingRemediationExecutor) ExecuteApplyProjects(_, ref, _ string, projects []models.ProjectDrift) ([]models.ProjectRemediationResult, error) {
	r.applyProjectRefs = append(r.applyProjectRefs, ref)
	calls := make([]remediationExecutorCall, 0, len(projects))
	results := make([]models.ProjectRemediationResult, 0, len(projects))
	for _, project := range projects {
		call := remediationExecutorCall{
			projectName: project.ProjectName,
			path:        project.Path,
			workspace:   project.Workspace,
		}
		calls = append(calls, call)
		r.planCalls = append(r.planCalls, call)
		r.applyCalls = append(r.applyCalls, call)
		if len(r.applyProjectResults) > 0 {
			continue
		}
		results = append(results, models.ProjectRemediationResult{
			ProjectName: project.ProjectName,
			Path:        project.Path,
			Workspace:   project.Workspace,
			Status:      models.RemediationStatusSuccess,
			PlanOutput:  "plan",
			ApplyOutput: "apply",
		})
	}
	r.applyProjectCalls = append(r.applyProjectCalls, calls)
	if r.applyProjectErr != nil && len(r.applyProjectResults) == 0 {
		return nil, r.applyProjectErr
	}
	if len(r.applyProjectResults) > 0 {
		results = make([]models.ProjectRemediationResult, len(r.applyProjectResults))
		copy(results, r.applyProjectResults)
	}
	return results, r.applyProjectErr
}

func TestInMemoryRemediationService_ExplicitProjectsHonorWorkspaceFilters(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app"},
		Workspaces: []string{"staging", "production"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 2, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", workspace: "staging"}, executor.planCalls[0])
	Equals(t, remediationExecutorCall{projectName: "app", workspace: "production"}, executor.planCalls[1])
}

func TestInMemoryRemediationService_GetProjectsErrorPreservesFailureStatus(t *testing.T) {
	service := drift.NewInMemoryRemediationService(failingRemediationStorage{})
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		DriftOnly:  true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusFailed, result.Status)
	Equals(t, "failed to get drift data: storage unavailable", result.Error)
	Equals(t, 0, result.TotalProjects)
	Equals(t, 0, len(result.Projects))
	Assert(t, result.CompletedAt != nil && !result.CompletedAt.IsZero(), "expected completed timestamp")
}

func TestInMemoryRemediationService_AutoApplyRefreshesCachedDrift(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	oldDrift := models.ProjectDrift{
		ProjectName:    "app",
		Path:           "modules/app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift: models.DriftSummary{
			HasDrift: true,
			ToChange: 1,
		},
		LastChecked: time.Now().Add(-time.Hour),
	}
	Ok(t, storage.Store("owner/repo", oldDrift))

	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{
		planDrift: &models.DriftSummary{
			HasDrift: true,
			ToChange: 1,
		},
	}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		DriftOnly:  true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, 1, len(executor.applyCalls))
	Equals(t, 1, len(executor.applyProjectCalls))
	Assert(t, result.Projects[0].DriftBefore != nil, "expected drift before remediation")
	Assert(t, result.Projects[0].DriftAfter != nil, "expected drift after remediation")
	Equals(t, false, result.Projects[0].DriftAfter.HasDrift)

	cached, err := storage.Get("owner/repo", drift.GetOptions{
		ProjectName: "app",
		Path:        "modules/app",
		Workspace:   "default",
		Ref:         "main",
	})
	Ok(t, err)
	Equals(t, 1, len(cached))
	Equals(t, false, cached[0].Drift.HasDrift)
	Equals(t, "Apply completed successfully", cached[0].Drift.Summary)
}

func TestInMemoryRemediationService_AutoApplyUsesSingleProjectBatch(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "network",
		Path:           "network",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))

	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		DriftOnly:  true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.applyProjectCalls))
	Equals(t, 2, len(executor.applyProjectCalls[0]))
	Equals(t, 2, len(result.Projects))
}

func TestInMemoryRemediationService_AutoApplyExecutorErrorWithoutResultsFailsTargets(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToAdd: 1},
		LastChecked:    time.Now(),
	}))
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "db",
		Path:           "db",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{applyProjectErr: errors.New("apply locked")}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
	}, executor)

	Ok(t, err)
	Equals(t, models.RemediationStatusFailed, result.Status)
	Equals(t, 2, len(result.Projects))
	for _, project := range result.Projects {
		Equals(t, models.RemediationStatusFailed, project.Status)
		Assert(t, strings.Contains(project.Error, "apply locked"), "expected executor error, got %q", project.Error)
	}
}

func TestInMemoryRemediationService_AutoApplyRequiresCachedDriftBeforeExecutor(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app"},
		Action:     models.RemediationAutoApply,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusFailed, result.Status)
	Equals(t, 0, len(executor.planCalls))
	Equals(t, 0, len(executor.applyCalls))
	Equals(t, 0, len(executor.applyProjectCalls))
	Equals(t, 1, len(result.Projects))
	Equals(t, models.RemediationStatusFailed, result.Projects[0].Status)
	Assert(t, strings.Contains(result.Projects[0].Error, "cached drift with a resolved commit is required"), "expected cached drift error, got %q", result.Projects[0].Error)

	cached, err := storage.Get("owner/repo", drift.GetOptions{ProjectName: "app", Ref: "main"})
	Ok(t, err)
	Equals(t, 0, len(cached))
}

func TestInMemoryRemediationService_AutoApplyWithoutCachedTargetsFails(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusFailed, result.Status)
	Assert(t, strings.Contains(result.Error, "cached drift with has_drift=true is required"), "expected cached drift error, got %q", result.Error)
	Equals(t, 0, len(executor.planCalls))
	Equals(t, 0, len(executor.applyCalls))
	Equals(t, 0, len(executor.applyProjectCalls))
}

func TestInMemoryRemediationService_AutoApplyRequiresCachedPositiveDrift(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: false, Summary: "No changes."},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app"},
		Action:     models.RemediationAutoApply,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusFailed, result.Status)
	Equals(t, 0, len(executor.planCalls))
	Equals(t, 0, len(executor.applyCalls))
	Equals(t, 0, len(executor.applyProjectCalls))
	Equals(t, 1, len(result.Projects))
	Equals(t, models.RemediationStatusFailed, result.Projects[0].Status)
	Assert(t, strings.Contains(result.Projects[0].Error, "cached drift with has_drift=true is required"), "expected positive drift error, got %q", result.Projects[0].Error)
}

func TestInMemoryRemediationService_AutoApplyDeduplicatesProjectSelectors(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		DetectionID:    "detection-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		Projects:   []string{"app", "app"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.applyProjectCalls))
	Equals(t, 1, len(executor.applyProjectCalls[0]))
	Equals(t, remediationExecutorCall{projectName: "app", path: "app", workspace: "default"}, executor.applyProjectCalls[0][0])
}

func TestInMemoryRemediationService_AutoApplyDeduplicatesPathSelectors(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "env",
		Workspace:      "prod",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		DetectionID:    "detection-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		Paths: []models.DriftDetectionPath{
			{Directory: "env", Workspace: "prod"},
			{Directory: "env", Workspace: "prod"},
		},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.applyProjectCalls))
	Equals(t, 1, len(executor.applyProjectCalls[0]))
	Equals(t, remediationExecutorCall{projectName: "app", path: "env", workspace: "prod"}, executor.applyProjectCalls[0][0])
}

func TestInMemoryRemediationService_AutoApplyDeduplicatesMixedSelectors(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "env",
		Workspace:      "prod",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		DetectionID:    "detection-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		Projects:   []string{"app"},
		Paths: []models.DriftDetectionPath{
			{Directory: "env", Workspace: "prod"},
			{Directory: "env", Workspace: "prod"},
		},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.applyProjectCalls))
	Equals(t, 1, len(executor.applyProjectCalls[0]))
	Equals(t, remediationExecutorCall{projectName: "app", path: "env", workspace: "prod"}, executor.applyProjectCalls[0][0])
}

func TestInMemoryRemediationService_AutoApplyRejectsConflictingDuplicateTargets(t *testing.T) {
	checkedAt := time.Now()
	service := drift.NewInMemoryRemediationService(staticRemediationStorage{drifts: []models.ProjectDrift{
		{
			ProjectName:    "app",
			Path:           "app",
			Workspace:      "default",
			Ref:            "main",
			BaseBranch:     "main",
			ResolvedCommit: "commit-a",
			DetectionID:    "detection-a",
			Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
			LastChecked:    checkedAt,
		},
		{
			ProjectName:    "app",
			Path:           "app",
			Workspace:      "default",
			Ref:            "main",
			BaseBranch:     "main",
			ResolvedCommit: "commit-b",
			DetectionID:    "detection-a",
			Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
			LastChecked:    checkedAt,
		},
	}})
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		DriftOnly:  true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusFailed, result.Status)
	Assert(t, strings.Contains(result.Error, "conflicting cached drift records"), "expected conflict error, got %q", result.Error)
	Equals(t, 0, len(executor.applyProjectCalls))
}

func TestInMemoryRemediationService_PlanOnlyExplicitWithoutCachedDriftStillRuns(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app"},
		Action:     models.RemediationPlanOnly,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, 0, len(executor.applyCalls))
	cached, err := storage.Get("owner/repo", drift.GetOptions{ProjectName: "app", Ref: "main"})
	Ok(t, err)
	Equals(t, 0, len(cached))
}

func TestInMemoryRemediationService_ExplicitProjectsPreserveCachedTargetMetadata(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	oldDrift := models.ProjectDrift{
		ProjectName:    "app",
		Path:           "modules/app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift: models.DriftSummary{
			HasDrift: true,
			ToChange: 1,
		},
		LastChecked: time.Now().Add(-time.Hour),
	}
	Ok(t, storage.Store("owner/repo", oldDrift))

	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		Projects:   []string{"app"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "modules/app", workspace: "default"}, executor.planCalls[0])
	Equals(t, 1, len(executor.applyCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "modules/app", workspace: "default"}, executor.applyCalls[0])

	cached, err := storage.Get("owner/repo", drift.GetOptions{ProjectName: "app", Ref: "main"})
	Ok(t, err)
	Equals(t, 1, len(cached))
	Equals(t, "modules/app", cached[0].Path)
	Equals(t, "default", cached[0].Workspace)
	Equals(t, false, cached[0].Drift.HasDrift)
}

func TestInMemoryRemediationService_ExplicitProjectNamesWithDotsMatchExactly(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app.prod",
		Path:           "modules/app-prod",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "appXprod",
		Path:           "modules/app-x-prod",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))

	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		Projects:   []string{"app.prod"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app.prod", path: "modules/app-prod", workspace: "default"}, executor.planCalls[0])
}

func TestInMemoryRemediationService_PathSelectorsTargetCachedDrift(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName: "app",
		Path:        "apps/staging",
		Workspace:   "default",
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName: "app",
		Path:        "apps/prod",
		Workspace:   "default",
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "apps/prod",
			Workspace: "default",
		}},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "apps/prod", workspace: "default"}, executor.planCalls[0])
}

func TestInMemoryRemediationService_PathSelectorsNormalizeBeforeCacheLookup(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "modules/vpc",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		DriftOnly:  true,
		Paths: []models.DriftDetectionPath{{
			Directory: "modules/vpc/",
		}},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "modules/vpc", workspace: "default"}, executor.planCalls[0])
}

func TestInMemoryRemediationService_PathSelectorsNormalizeFallbackTargets(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "modules/vpc/",
		}},
		Workspaces: []string{"prod"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{path: "modules/vpc", workspace: "prod"}, executor.planCalls[0])
}

func TestInMemoryRemediationService_PathSelectorsHonorTopLevelWorkspaceFallbacks(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
		}},
		Workspaces: []string{"prod", "stage"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 2, len(executor.planCalls))
	Equals(t, remediationExecutorCall{path: "env", workspace: "prod"}, executor.planCalls[0])
	Equals(t, remediationExecutorCall{path: "env", workspace: "stage"}, executor.planCalls[1])
}

func TestInMemoryRemediationService_PathSelectorsPreserveProjectFallbacks(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app"},
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
		}},
		Workspaces: []string{"prod"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "env", workspace: "prod"}, executor.planCalls[0])
}

func TestInMemoryRemediationService_PathSelectorsHonorCachedWorkspaceFilters(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	for _, workspace := range []string{"prod", "stage", "dev"} {
		Ok(t, storage.Store("owner/repo", models.ProjectDrift{
			ProjectName: "app",
			Path:        "env",
			Workspace:   workspace,
			Ref:         "main",
			BaseBranch:  "main",
			Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
			LastChecked: time.Now(),
		}))
	}
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
		}},
		Workspaces: []string{"prod", "stage"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 2, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "env", workspace: "prod"}, executor.planCalls[0])
	Equals(t, remediationExecutorCall{projectName: "app", path: "env", workspace: "stage"}, executor.planCalls[1])
}

func TestInMemoryRemediationService_PathSelectorWithoutWorkspaceTargetsDefaultOnly(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName: "app",
		Path:        "env",
		Workspace:   "prod",
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
		}},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{path: "env", workspace: "default"}, executor.planCalls[0])
	Equals(t, 1, len(result.Projects))
	Assert(t, result.Projects[0].DriftBefore == nil, "default workspace fallback must not consume cached prod drift")
}

func TestInMemoryRemediationService_PathSelectorRejectsConflictingWorkspaceFilter(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
			Workspace: "dev",
		}},
		Workspaces: []string{"prod"},
	}, executor)

	Ok(t, err)
	Equals(t, models.RemediationStatusFailed, result.Status)
	Assert(t, strings.Contains(result.Error, "path workspace"), "expected path workspace error, got %q", result.Error)
	Equals(t, 0, len(executor.planCalls))
}

func TestInMemoryRemediationService_PathOnlyApplyMatchesResolvedProjectResult(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		Path:           "env",
		Workspace:      "prod",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{
		applyProjectResults: []models.ProjectRemediationResult{{
			ProjectName: "app",
			Path:        "env",
			Workspace:   "prod",
			Status:      models.RemediationStatusSuccess,
			PlanOutput:  "plan",
			ApplyOutput: "apply",
		}},
	}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Action:     models.RemediationAutoApply,
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
		}},
		Workspaces: []string{"prod"},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(result.Projects))
	Equals(t, "app", result.Projects[0].ProjectName)
	Equals(t, "env", result.Projects[0].Path)
	Equals(t, "prod", result.Projects[0].Workspace)
	Equals(t, models.RemediationStatusSuccess, result.Projects[0].Status)
}

func TestInMemoryRemediationService_PathSelectorsUseMatchingBaseBranch(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName: "app",
		Path:        "apps/app",
		Workspace:   "default",
		Ref:         "v1.0.0",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName: "app",
		Path:        "apps/app",
		Workspace:   "default",
		Ref:         "v1.0.0",
		BaseBranch:  "release",
		Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
		LastChecked: time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "v1.0.0",
		BaseBranch: "release",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "apps/app",
			Workspace: "default",
		}},
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "apps/app", workspace: "default"}, executor.planCalls[0])
	Assert(t, result.Projects[0].DriftBefore != nil, "expected cached drift before remediation")
}

func TestInMemoryRemediationService_NormalizesBranchRefsForCacheLookup(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName: "app",
		Path:        "apps/app",
		Workspace:   "default",
		Ref:         "main",
		BaseBranch:  "main",
		Drift:       models.DriftSummary{HasDrift: true, ToAdd: 1},
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "refs/heads/main",
		Type:       "Github",
		Action:     models.RemediationPlanOnly,
		DriftOnly:  true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "apps/app", workspace: "default"}, executor.planCalls[0])
}

func TestInMemoryRemediationService_PreservesExecutionRefForAutoApply(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("owner/repo", models.ProjectDrift{
		ProjectName:    "app",
		Path:           "apps/app",
		Workspace:      "default",
		Ref:            "main",
		BaseBranch:     "main",
		ResolvedCommit: "commit-a",
		Drift:          models.DriftSummary{HasDrift: true, ToAdd: 1},
		LastChecked:    time.Now(),
	}))
	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository:   "owner/repo",
		Ref:          "main",
		ExecutionRef: "refs/heads/main",
		Type:         "Github",
		Action:       models.RemediationAutoApply,
		DriftOnly:    true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.applyProjectRefs))
	Equals(t, "refs/heads/main", executor.applyProjectRefs[0])
	Equals(t, 1, len(executor.applyProjectCalls))
	Equals(t, remediationExecutorCall{projectName: "app", path: "apps/app", workspace: "default"}, executor.applyProjectCalls[0][0])
}

func TestInMemoryRemediationService_UsesHostQualifiedStorageRepository(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	Ok(t, storage.Store("github.com/acme/infra", models.ProjectDrift{
		ProjectName: "github-app",
		Path:        "github/app",
		Workspace:   "default",
		Ref:         "main",
		BaseBranch:  "main",
		Drift: models.DriftSummary{
			HasDrift: true,
			ToChange: 1,
		},
		LastChecked: time.Now(),
	}))
	Ok(t, storage.Store("gitlab.com/acme/infra", models.ProjectDrift{
		ProjectName: "gitlab-app",
		Path:        "gitlab/app",
		Workspace:   "default",
		Ref:         "main",
		BaseBranch:  "main",
		Drift: models.DriftSummary{
			HasDrift: true,
			ToChange: 1,
		},
		LastChecked: time.Now(),
	}))

	service := drift.NewInMemoryRemediationService(storage)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository:        "acme/infra",
		StorageRepository: "gitlab.com/acme/infra",
		Ref:               "main",
		Type:              "Gitlab",
		DriftOnly:         true,
	}, executor)
	Ok(t, err)

	Equals(t, models.RemediationStatusSuccess, result.Status)
	Equals(t, 1, len(executor.planCalls))
	Equals(t, remediationExecutorCall{projectName: "gitlab-app", path: "gitlab/app", workspace: "default"}, executor.planCalls[0])
	Equals(t, "acme/infra", result.Repository)
}

func TestInMemoryRemediationService_ListResultsUsesHostQualifiedStorageRepository(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	githubResult, err := service.Remediate(models.RemediationRequest{
		Repository:        "acme/infra",
		StorageRepository: "github.com/acme/infra",
		Ref:               "main",
		Type:              "Github",
		Projects:          []string{"app"},
	}, executor)
	Ok(t, err)

	gitlabResult, err := service.Remediate(models.RemediationRequest{
		Repository:        "acme/infra",
		StorageRepository: "gitlab.com/acme/infra",
		Ref:               "main",
		Type:              "Gitlab",
		Projects:          []string{"app"},
	}, executor)
	Ok(t, err)

	githubResults, err := service.ListResults("github.com/acme/infra", 10)
	Ok(t, err)
	Equals(t, 1, len(githubResults))
	Equals(t, githubResult.ID, githubResults[0].ID)

	gitlabResults, err := service.ListResults("gitlab.com/acme/infra", 10)
	Ok(t, err)
	Equals(t, 1, len(gitlabResults))
	Equals(t, gitlabResult.ID, gitlabResults[0].ID)

	rawResults, err := service.ListResults("acme/infra", 10)
	Ok(t, err)
	Equals(t, 0, len(rawResults))
}

func TestInMemoryRemediationService_ReturnsResultSnapshots(t *testing.T) {
	service := drift.NewInMemoryRemediationService(nil)
	executor := &recordingRemediationExecutor{}

	result, err := service.Remediate(models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app"},
	}, executor)
	Ok(t, err)

	got, err := service.GetResult(result.ID)
	Ok(t, err)
	got.Status = models.RemediationStatusFailed
	got.Projects[0].Status = models.RemediationStatusFailed
	got.Projects[0].DriftAfter = &models.DriftSummary{HasDrift: true, ToAdd: 99}

	gotAgain, err := service.GetResult(result.ID)
	Ok(t, err)
	Equals(t, models.RemediationStatusSuccess, gotAgain.Status)
	Equals(t, models.RemediationStatusSuccess, gotAgain.Projects[0].Status)
	Equals(t, false, gotAgain.Projects[0].DriftAfter.HasDrift)

	listed, err := service.ListResults("owner/repo", 10)
	Ok(t, err)
	Equals(t, 1, len(listed))
	listed[0].Status = models.RemediationStatusFailed

	listedAgain, err := service.ListResults("owner/repo", 10)
	Ok(t, err)
	Equals(t, models.RemediationStatusSuccess, listedAgain[0].Status)
}
