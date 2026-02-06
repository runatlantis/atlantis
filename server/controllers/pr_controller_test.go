// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	tMocks "github.com/runatlantis/atlantis/server/controllers/web_templates/mocks"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPRController_PRList_Success(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	now := time.Now()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo1"}},
		{Num: 456, BaseRepo: models.Repo{FullName: "owner/repo2"}},
	}, nil)

	When(mockDB.GetProjectOutputsByPull("owner/repo1", 123)).ThenReturn([]models.ProjectOutput{
		{
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 1, Change: 2},
			CompletedAt:   now.Add(-5 * time.Minute),
		},
	}, nil)

	When(mockDB.GetProjectOutputsByPull("owner/repo2", 456)).ThenReturn([]models.ProjectOutput{
		{
			Status:        models.FailedOutputStatus,
			ResourceStats: models.ResourceStats{},
			CompletedAt:   now.Add(-1 * time.Hour),
		},
		{
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 5},
			CompletedAt:   now.Add(-30 * time.Minute),
		},
	}, nil)

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"/basepath",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusOK, w.Code)

	// Verify template was called
	mockTemplate.VerifyWasCalledOnce().Execute(Any[io.Writer](), Any[interface{}]())
}

func TestPRController_PRList_DBError(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	When(mockDB.GetActivePullRequests()).ThenReturn(nil, errors.New("database error"))

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "error loading PR data"), "should contain error message")
}

func TestPRController_PRList_TemplateError(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{}, nil)
	When(mockTemplate.Execute(Any[io.Writer](), Any[interface{}]())).ThenReturn(errors.New("template error"))

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "error rendering template"), "should contain template error message")
}

func TestPRController_PRListPartial_Success(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockRowsTemplate := tMocks.NewMockTemplateWriter()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo"}},
	}, nil)

	When(mockDB.GetProjectOutputsByPull("owner/repo", 123)).ThenReturn([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus},
	}, nil)

	controller := controllers.NewPRController(
		mockDB,
		tMocks.NewMockTemplateWriter(),
		mockRowsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs/partial", nil)
	w := httptest.NewRecorder()

	controller.PRListPartial(w, req)

	Equals(t, http.StatusOK, w.Code)
	mockRowsTemplate.VerifyWasCalledOnce().Execute(Any[io.Writer](), Any[interface{}]())
}

func TestPRController_PRListPartial_DBError(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockRowsTemplate := tMocks.NewMockTemplateWriter()

	When(mockDB.GetActivePullRequests()).ThenReturn(nil, errors.New("database error"))

	controller := controllers.NewPRController(
		mockDB,
		tMocks.NewMockTemplateWriter(),
		mockRowsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs/partial", nil)
	w := httptest.NewRecorder()

	controller.PRListPartial(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "error loading PR data"), "should contain error message")
}

func TestPRController_PRList_ShowsErrorStateForPRsWithOutputErrors(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo1"}},
		{Num: 456, BaseRepo: models.Repo{FullName: "owner/repo2"}},
	}, nil)

	// First PR fails to get outputs
	When(mockDB.GetProjectOutputsByPull("owner/repo1", 123)).ThenReturn(nil, errors.New("output error"))

	// Second PR succeeds
	When(mockDB.GetProjectOutputsByPull("owner/repo2", 456)).ThenReturn([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus},
	}, nil)

	var capturedData web_templates.PRListData
	When(mockTemplate.Execute(Any[io.Writer](), Any[interface{}]())).Then(func(params []Param) ReturnValues {
		if data, ok := params[1].(web_templates.PRListData); ok {
			capturedData = data
		}
		return []ReturnValue{nil}
	})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusOK, w.Code)
	// Both PRs should be in the list - one with error state, one successful
	Equals(t, 2, capturedData.TotalCount)

	// Find the PR with error state
	var errorPR, successPR web_templates.PRListItem
	for _, pr := range capturedData.PullRequests {
		if pr.RepoFullName == "owner/repo1" {
			errorPR = pr
		} else {
			successPR = pr
		}
	}

	// Verify the error PR has the correct error state
	Equals(t, "error", errorPR.Status)
	Equals(t, "alert-circle", errorPR.StatusIcon)
	Assert(t, errorPR.ErrorMessage != "", "error PR should have error message")
	Equals(t, 123, errorPR.PullNum)

	// Verify the successful PR is normal
	Equals(t, "passed", successPR.Status)
	Equals(t, "", successPR.ErrorMessage)
	Equals(t, 456, successPR.PullNum)
}

func TestPRController_PRList_SortsByLastActivity(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	now := time.Now()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{
		{Num: 1, BaseRepo: models.Repo{FullName: "owner/old"}},
		{Num: 2, BaseRepo: models.Repo{FullName: "owner/new"}},
	}, nil)

	// Old PR - older activity
	When(mockDB.GetProjectOutputsByPull("owner/old", 1)).ThenReturn([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus, CompletedAt: now.Add(-2 * time.Hour)},
	}, nil)

	// New PR - recent activity
	When(mockDB.GetProjectOutputsByPull("owner/new", 2)).ThenReturn([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus, CompletedAt: now.Add(-5 * time.Minute)},
	}, nil)

	var capturedData web_templates.PRListData
	When(mockTemplate.Execute(Any[io.Writer](), Any[interface{}]())).Then(func(params []Param) ReturnValues {
		if data, ok := params[1].(web_templates.PRListData); ok {
			capturedData = data
		}
		return []ReturnValue{nil}
	})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusOK, w.Code)
	Equals(t, 2, len(capturedData.PullRequests))
	// Most recent should be first
	Equals(t, "owner/new", capturedData.PullRequests[0].RepoFullName)
	Equals(t, "owner/old", capturedData.PullRequests[1].RepoFullName)
}

func TestDetermineStatus(t *testing.T) {
	tests := []struct {
		name         string
		success      int
		failed       int
		pending      int
		expectedStat string
		expectedIcon string
	}{
		{"all_success", 3, 0, 0, "passed", "passed"},
		{"all_failed", 0, 2, 0, "failed", "failed"},
		{"all_pending", 0, 0, 2, "pending", "pending"},
		{"mixed_with_failure", 2, 1, 0, "failed", "failed"},
		{"mixed_with_pending", 2, 0, 1, "pending", "pending"},
		{"failure_takes_precedence", 1, 1, 1, "failed", "failed"},
		{"no_projects", 0, 0, 0, "pending", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, icon := controllers.DetermineStatus(tt.success, tt.failed, tt.pending)
			Equals(t, tt.expectedStat, status)
			Equals(t, tt.expectedIcon, icon)
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"just_now", -30 * time.Second, "just now"},
		{"1_minute", -1*time.Minute - 30*time.Second, "1 minute ago"},
		{"5_minutes", -5 * time.Minute, "5 minutes ago"},
		{"1_hour", -1*time.Hour - 5*time.Minute, "1 hour ago"},
		{"3_hours", -3 * time.Hour, "3 hours ago"},
		{"1_day", -25 * time.Hour, "1 day ago"},
		{"3_days", -72 * time.Hour, "3 days ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := time.Now().Add(tt.duration)
			result := controllers.FormatRelativeTime(ts)
			Equals(t, tt.expected, result)
		})
	}

	// Test zero time
	t.Run("zero_time", func(t *testing.T) {
		result := controllers.FormatRelativeTime(time.Time{})
		Equals(t, "unknown", result)
	})
}

func TestPRController_PRList_AggregatesResourceStats(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	now := time.Now()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo"}},
	}, nil)

	When(mockDB.GetProjectOutputsByPull("owner/repo", 123)).ThenReturn([]models.ProjectOutput{
		{
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 1, Change: 2, Destroy: 0},
			CompletedAt:   now.Add(-5 * time.Minute),
		},
		{
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 3, Change: 1, Destroy: 2},
			CompletedAt:   now.Add(-10 * time.Minute),
		},
	}, nil)

	var capturedData web_templates.PRListData
	When(mockTemplate.Execute(Any[io.Writer](), Any[interface{}]())).Then(func(params []Param) ReturnValues {
		if data, ok := params[1].(web_templates.PRListData); ok {
			capturedData = data
		}
		return []ReturnValue{nil}
	})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusOK, w.Code)
	Equals(t, 1, len(capturedData.PullRequests))
	prItem := capturedData.PullRequests[0]
	Equals(t, 4, prItem.AddCount)     // 1 + 3
	Equals(t, 3, prItem.ChangeCount)  // 2 + 1
	Equals(t, 2, prItem.DestroyCount) // 0 + 2
	Equals(t, 2, prItem.ProjectCount)
	Equals(t, 2, prItem.SuccessCount)
}

func TestPRController_PRList_MixedStatuses(t *testing.T) {
	RegisterMockTestingT(t)

	mockDB := mocks.NewMockDatabase()
	mockTemplate := tMocks.NewMockTemplateWriter()

	now := time.Now()

	When(mockDB.GetActivePullRequests()).ThenReturn([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo"}},
	}, nil)

	When(mockDB.GetProjectOutputsByPull("owner/repo", 123)).ThenReturn([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus, CompletedAt: now},
		{Status: models.FailedOutputStatus, CompletedAt: now},
		{Status: models.PendingOutputStatus, CompletedAt: now},
	}, nil)

	var capturedData web_templates.PRListData
	When(mockTemplate.Execute(Any[io.Writer](), Any[interface{}]())).Then(func(params []Param) ReturnValues {
		if data, ok := params[1].(web_templates.PRListData); ok {
			capturedData = data
		}
		return []ReturnValue{nil}
	})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusOK, w.Code)
	prItem := capturedData.PullRequests[0]
	Equals(t, 1, prItem.SuccessCount)
	Equals(t, 1, prItem.FailedCount)
	Equals(t, 1, prItem.PendingCount)
	Equals(t, "failed", prItem.Status) // Failed takes precedence
}
