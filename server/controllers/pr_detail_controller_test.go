// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// mockPRDetailDB implements the PRDetailDatabase interface for tests
type mockPRDetailDB struct {
	outputs []models.ProjectOutput
	err     error
}

func (m *mockPRDetailDB) GetProjectOutputsByPull(repoFullName string, pullNum int) ([]models.ProjectOutput, error) {
	return m.outputs, m.err
}

func TestPRDetailController_PRDetail_Success(t *testing.T) {
	mockDB := &mockPRDetailDB{
		outputs: []models.ProjectOutput{
			{
				Path:          "terraform/staging",
				Workspace:     "default",
				CommandName:   "policy_check",
				Status:        models.SuccessOutputStatus,
				PolicyPassed:  true,
				ResourceStats: models.ResourceStats{Add: 2, Change: 1},
				CompletedAt:   time.Now().Add(-5 * time.Minute),
			},
			{
				Path:         "terraform/prod",
				Workspace:    "default",
				CommandName:  "plan",
				Status:       models.FailedOutputStatus,
				PolicyPassed: false,
				Error:        "plan failed",
				CompletedAt:  time.Now().Add(-10 * time.Minute),
			},
		},
	}

	controller := controllers.NewPRDetailController(
		mockDB,
		web_templates.PRDetailTemplate,
		web_templates.PRDetailProjectsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
	)

	req := httptest.NewRequest("GET", "/pr/owner/repo/pulls/123", nil)
	req = mux.SetURLVars(req, map[string]string{
		"owner":    "owner",
		"repo":     "repo",
		"pull_num": "123",
	})
	w := httptest.NewRecorder()

	controller.PRDetail(w, req)

	Equals(t, http.StatusOK, w.Code)
	body := w.Body.String()
	t.Logf("Body length: %d, first 500 chars: %s", len(body), body[:min(500, len(body))])
	Assert(t, strings.Contains(body, "owner/repo"), "should contain repo name")
	Assert(t, strings.Contains(body, "#123"), "should contain PR number")
	Assert(t, strings.Contains(body, "terraform/staging"), "should contain project path")
	Assert(t, strings.Contains(body, "terraform/prod"), "should contain failed project")
}

func TestPRDetailController_PRDetail_InvalidPullNum(t *testing.T) {
	controller := controllers.NewPRDetailController(
		&mockPRDetailDB{},
		web_templates.PRDetailTemplate,
		web_templates.PRDetailProjectsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
	)

	req := httptest.NewRequest("GET", "/pr/owner/repo/pulls/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{
		"owner":    "owner",
		"repo":     "repo",
		"pull_num": "invalid",
	})
	w := httptest.NewRecorder()

	controller.PRDetail(w, req)

	Equals(t, http.StatusBadRequest, w.Code)
}

func TestPRDetailController_PRDetailProjects_WithFilter(t *testing.T) {
	mockDB := &mockPRDetailDB{
		outputs: []models.ProjectOutput{
			{Path: "staging", CommandName: "policy_check", Status: models.SuccessOutputStatus, PolicyPassed: true},
			{Path: "prod", CommandName: "plan", Status: models.FailedOutputStatus, PolicyPassed: false},
		},
	}

	controller := controllers.NewPRDetailController(
		mockDB,
		web_templates.PRDetailTemplate,
		web_templates.PRDetailProjectsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
	)

	tests := []struct {
		name          string
		filter        string
		expectedPaths []string
	}{
		{"all", "", []string{"staging", "prod"}},
		{"failed", "failed", []string{"prod"}},
		{"passed", "passed", []string{"staging"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/pr/owner/repo/pulls/123/projects"
			if tt.filter != "" {
				url += "?filter=" + tt.filter
			}
			req := httptest.NewRequest("GET", url, nil)
			req = mux.SetURLVars(req, map[string]string{
				"owner":    "owner",
				"repo":     "repo",
				"pull_num": "123",
			})
			w := httptest.NewRecorder()

			controller.PRDetailProjects(w, req)

			Equals(t, http.StatusOK, w.Code)
			body := w.Body.String()
			for _, path := range tt.expectedPaths {
				Assert(t, strings.Contains(body, path), "should contain %s", path)
			}
		})
	}
}

func TestPRDetailController_PRDetail_ShowsErrorStateOnDBFailure(t *testing.T) {
	mockDB := &mockPRDetailDB{
		err: errors.New("database connection failed"),
	}

	controller := controllers.NewPRDetailController(
		mockDB,
		web_templates.PRDetailTemplate,
		web_templates.PRDetailProjectsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
	)

	req := httptest.NewRequest("GET", "/pr/owner/repo/pulls/123", nil)
	req = mux.SetURLVars(req, map[string]string{
		"owner":    "owner",
		"repo":     "repo",
		"pull_num": "123",
	})
	w := httptest.NewRecorder()

	controller.PRDetail(w, req)

	// Should return 200 OK, not 500
	Equals(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Should contain the repo name and PR number (header is always shown)
	Assert(t, strings.Contains(body, "owner/repo"), "should contain repo name")
	Assert(t, strings.Contains(body, "#123"), "should contain PR number")

	// Should contain the error message warning
	Assert(t, strings.Contains(body, "Unable to Load Project Data"), "should contain error header")
	Assert(t, strings.Contains(body, "temporary database issue"), "should contain error message")

	// Should NOT contain the toolbar (filter buttons) since we're showing error state
	Assert(t, !strings.Contains(body, "filter-toggle"), "should not contain filter buttons")
}

func TestBuildDetailProject(t *testing.T) {
	output := models.ProjectOutput{
		Path:          "terraform/staging",
		Workspace:     "default",
		ProjectName:   "staging-vpc",
		CommandName:   "policy_check",
		Status:        models.SuccessOutputStatus,
		PolicyPassed:  true,
		ResourceStats: models.ResourceStats{Add: 5, Change: 2, Destroy: 1},
		CompletedAt:   time.Now().Add(-5 * time.Minute),
	}

	project := controllers.BuildDetailProject(output)

	Equals(t, "terraform/staging", project.Path)
	Equals(t, "default", project.Workspace)
	Equals(t, "staging-vpc", project.ProjectName)
	Equals(t, "success", project.Status)
	Equals(t, true, project.PolicyPassed)
	Equals(t, true, project.HasPolicyCheck)
	Equals(t, 5, project.AddCount)
	Equals(t, 2, project.ChangeCount)
	Equals(t, 1, project.DestroyCount)
}

func TestBuildDetailProject_NonPolicyCommand(t *testing.T) {
	output := models.ProjectOutput{
		Path:          "terraform/staging",
		Workspace:     "default",
		CommandName:   "plan",
		Status:        models.SuccessOutputStatus,
		PolicyPassed:  false, // Default false for non-policy commands
		ResourceStats: models.ResourceStats{Add: 1},
		CompletedAt:   time.Now().Add(-5 * time.Minute),
	}

	project := controllers.BuildDetailProject(output)

	Equals(t, false, project.HasPolicyCheck)
	Equals(t, false, project.PolicyPassed)
}

func TestBuildDetailProject_PolicyOutputSetsHasPolicyCheck(t *testing.T) {
	output := models.ProjectOutput{
		Path:          "terraform/staging",
		Workspace:     "default",
		CommandName:   "plan",
		Status:        models.SuccessOutputStatus,
		PolicyPassed:  true,
		PolicyOutput:  "Policies passed",
		ResourceStats: models.ResourceStats{Add: 1},
		CompletedAt:   time.Now().Add(-5 * time.Minute),
	}

	project := controllers.BuildDetailProject(output)

	Equals(t, true, project.HasPolicyCheck)
	Equals(t, true, project.PolicyPassed)
}
