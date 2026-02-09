// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// mockProjectOutputDB implements the db.Database interface for tests
type mockProjectOutputDB struct {
	history []models.ProjectOutput
	err     error
}

func (m *mockProjectOutputDB) GetProjectOutputHistory(repoFullName string, pullNum int, path string, workspace string, projectName string) ([]models.ProjectOutput, error) {
	return m.history, m.err
}

func (m *mockProjectOutputDB) GetProjectOutputRun(repoFullName string, pullNum int, path string, workspace string, projectName string, command string, runTimestamp int64) (*models.ProjectOutput, error) {
	for i := range m.history {
		if m.history[i].RunTimestamp == runTimestamp {
			return &m.history[i], nil
		}
	}
	return nil, nil
}

// Stub implementations to satisfy db.Database interface
func (m *mockProjectOutputDB) TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	return false, models.ProjectLock{}, nil
}
func (m *mockProjectOutputDB) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) List() ([]models.ProjectLock, error) { return nil, nil }
func (m *mockProjectOutputDB) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error {
	return nil
}
func (m *mockProjectOutputDB) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) DeletePullStatus(pull models.PullRequest) error { return nil }
func (m *mockProjectOutputDB) UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error) {
	return models.PullStatus{}, nil
}
func (m *mockProjectOutputDB) LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) UnlockCommand(cmdName command.Name) error { return nil }
func (m *mockProjectOutputDB) CheckCommandLock(cmdName command.Name) (*command.Lock, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) SaveProjectOutput(output models.ProjectOutput) error { return nil }
func (m *mockProjectOutputDB) GetProjectOutputsByPull(repoFullName string, pullNum int) ([]models.ProjectOutput, error) {
	return nil, nil
}
func (m *mockProjectOutputDB) DeleteProjectOutputsByPull(repoFullName string, pullNum int) error {
	return nil
}
func (m *mockProjectOutputDB) GetActivePullRequests() ([]models.PullRequest, error) { return nil, nil }
func (m *mockProjectOutputDB) Close() error                                         { return nil }
func (m *mockProjectOutputDB) GetProjectOutputByJobID(jobID string) (*models.ProjectOutput, error) {
	return nil, nil
}

func TestProjectOutputController_ProjectOutput_Success(t *testing.T) {
	now := time.Now()
	mockDB := &mockProjectOutputDB{
		history: []models.ProjectOutput{
			{
				RepoFullName:  "owner/repo",
				PullNum:       123,
				Path:          "terraform/staging",
				Workspace:     "default",
				CommandName:   "plan",
				RunTimestamp:  now.UnixMilli(),
				Output:        "Terraform will perform the following actions:\n\n  # aws_instance.example will be created\n  + resource \"aws_instance\" \"example\" {\n      + ami           = \"ami-12345\"\n      + instance_type = \"t2.micro\"\n    }\n\nPlan: 1 to add, 0 to change, 0 to destroy.",
				Status:        models.SuccessOutputStatus,
				ResourceStats: models.ResourceStats{Add: 1, Change: 0, Destroy: 0},
				PolicyPassed:  true,
				TriggeredBy:   "jwalton",
				StartedAt:     now.Add(-2 * time.Minute),
				CompletedAt:   now,
			},
		},
	}

	controller := controllers.NewProjectOutputController(
		mockDB,
		web_templates.ProjectOutputTemplate,
		web_templates.ProjectOutputPartialTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // outputHandler
	)

	req := httptest.NewRequest("GET", "/pr/owner/repo/pulls/123/project/terraform/staging?workspace=default", nil)
	req = mux.SetURLVars(req, map[string]string{
		"owner":    "owner",
		"repo":     "repo",
		"pull_num": "123",
		"path":     "terraform/staging",
	})
	w := httptest.NewRecorder()

	controller.ProjectOutput(w, req)

	Equals(t, http.StatusOK, w.Code)
	body := w.Body.String()
	Assert(t, strings.Contains(body, "terraform/staging"), "should contain project path")
	Assert(t, strings.Contains(body, "owner/repo"), "should contain repo name")
	Assert(t, strings.Contains(body, "#123"), "should contain PR number")
	Assert(t, strings.Contains(body, "aws_instance"), "should contain terraform resource")
	Assert(t, strings.Contains(body, "Planned"), "should contain status label")
}

func TestProjectOutputController_ProjectOutput_NotFound(t *testing.T) {
	mockDB := &mockProjectOutputDB{
		history: nil,
		err:     nil,
	}

	controller := controllers.NewProjectOutputController(
		mockDB,
		web_templates.ProjectOutputTemplate,
		web_templates.ProjectOutputPartialTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // outputHandler
	)

	req := httptest.NewRequest("GET", "/pr/owner/repo/pulls/123/project/terraform/staging?workspace=default", nil)
	req = mux.SetURLVars(req, map[string]string{
		"owner":    "owner",
		"repo":     "repo",
		"pull_num": "123",
		"path":     "terraform/staging",
	})
	w := httptest.NewRecorder()

	controller.ProjectOutput(w, req)

	Equals(t, http.StatusNotFound, w.Code)
}

func TestProjectOutputController_ProjectOutput_Failed(t *testing.T) {
	now := time.Now()
	mockDB := &mockProjectOutputDB{
		history: []models.ProjectOutput{
			{
				RepoFullName:  "owner/repo",
				PullNum:       123,
				Path:          "terraform/staging",
				Workspace:     "default",
				CommandName:   "plan",
				RunTimestamp:  now.UnixMilli(),
				Output:        "Error: Invalid provider configuration",
				Status:        models.FailedOutputStatus,
				Error:         "exit status 1",
				ResourceStats: models.ResourceStats{},
				PolicyPassed:  false,
				StartedAt:     now.Add(-1 * time.Minute),
				CompletedAt:   now,
			},
		},
	}

	controller := controllers.NewProjectOutputController(
		mockDB,
		web_templates.ProjectOutputTemplate,
		web_templates.ProjectOutputPartialTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // outputHandler
	)

	req := httptest.NewRequest("GET", "/pr/owner/repo/pulls/123/project/terraform/staging?workspace=default", nil)
	req = mux.SetURLVars(req, map[string]string{
		"owner":    "owner",
		"repo":     "repo",
		"pull_num": "123",
		"path":     "terraform/staging",
	})
	w := httptest.NewRecorder()

	controller.ProjectOutput(w, req)

	Equals(t, http.StatusOK, w.Code)
	body := w.Body.String()
	Assert(t, strings.Contains(body, "Plan Failed"), "should contain failure message")
	Assert(t, strings.Contains(body, "exit status 1"), "should contain error message")
}

func TestProjectOutputController_FormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"one_minute", 60 * time.Second, "1m 0s"},
		{"minutes_and_seconds", 2*time.Minute + 34*time.Second, "2m 34s"},
		{"many_minutes", 15*time.Minute + 7*time.Second, "15m 7s"},
		{"hours", 1*time.Hour + 5*time.Minute + 30*time.Second, "1h 5m 30s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controllers.FormatDuration(tt.duration)
			Equals(t, tt.expected, result)
		})
	}
}

func TestProjectOutputController_HighlightTerraformOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"add_line",
			"  + resource \"aws_instance\" \"example\"",
			"tf-add",
		},
		{
			"destroy_line",
			"  - resource \"aws_instance\" \"example\"",
			"tf-destroy",
		},
		{
			"change_line",
			"  ~ ami = \"old\" -> \"new\"",
			"tf-change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controllers.HighlightTerraformOutput(tt.input)
			Assert(t, strings.Contains(string(result), tt.contains), "should contain %s", tt.contains)
		})
	}
}
