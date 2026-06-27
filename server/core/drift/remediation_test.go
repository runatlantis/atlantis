// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift_test

import (
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
	planCalls  []remediationExecutorCall
	applyCalls []remediationExecutorCall
	planDrift  *models.DriftSummary
}

func (r *recordingRemediationExecutor) ExecutePlan(_, _, _, projectName, path, workspace string) (string, *models.DriftSummary, error) {
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

func TestInMemoryRemediationService_AutoApplyRefreshesCachedDrift(t *testing.T) {
	storage := drift.NewInMemoryStorage()
	oldDrift := models.ProjectDrift{
		ProjectName: "app",
		Path:        "modules/app",
		Workspace:   "default",
		Ref:         "main",
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
