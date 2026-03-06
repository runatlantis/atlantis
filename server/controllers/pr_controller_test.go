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

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	tMocks "github.com/runatlantis/atlantis/server/controllers/web_templates/mocks"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestPRController_PRList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	now := time.Now()

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo1"}},
		{Num: 456, BaseRepo: models.Repo{FullName: "owner/repo2"}},
	}, nil)

	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo1", 123).Return([]models.ProjectOutput{
		{
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 1, Change: 2},
			CompletedAt:   now.Add(-5 * time.Minute),
		},
	}, nil)

	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo2", 456).Return([]models.ProjectOutput{
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

	mockTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil)

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"/basepath",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusOK, w.Code)
}

func TestPRController_PRList_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	mockDB.EXPECT().GetActivePullRequests().Return(nil, errors.New("database error"))

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "Internal server error"), "should contain generic error message")
}

func TestPRController_PRList_TemplateError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{}, nil)
	mockTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(errors.New("template error"))

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
	)

	req := httptest.NewRequest("GET", "/prs", nil)
	w := httptest.NewRecorder()

	controller.PRList(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "Internal server error"), "should contain generic error message")
}

func TestPRController_PRListPartial_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockRowsTemplate := tMocks.NewMockTemplateWriter(ctrl)

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo"}},
	}, nil)

	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo", 123).Return([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus},
	}, nil)

	mockRowsTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil)

	controller := controllers.NewPRController(
		mockDB,
		tMocks.NewMockTemplateWriter(ctrl),
		mockRowsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
	)

	req := httptest.NewRequest("GET", "/prs/partial", nil)
	w := httptest.NewRecorder()

	controller.PRListPartial(w, req)

	Equals(t, http.StatusOK, w.Code)
}

func TestPRController_PRListPartial_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockRowsTemplate := tMocks.NewMockTemplateWriter(ctrl)

	mockDB.EXPECT().GetActivePullRequests().Return(nil, errors.New("database error"))

	controller := controllers.NewPRController(
		mockDB,
		tMocks.NewMockTemplateWriter(ctrl),
		mockRowsTemplate,
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
	)

	req := httptest.NewRequest("GET", "/prs/partial", nil)
	w := httptest.NewRecorder()

	controller.PRListPartial(w, req)

	Equals(t, http.StatusInternalServerError, w.Code)
	Assert(t, strings.Contains(w.Body.String(), "Internal server error"), "should contain generic error message")
}

func TestPRController_PRList_ShowsErrorStateForPRsWithOutputErrors(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo1"}},
		{Num: 456, BaseRepo: models.Repo{FullName: "owner/repo2"}},
	}, nil)

	// First PR fails to get outputs
	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo1", 123).Return(nil, errors.New("output error"))

	// Second PR succeeds
	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo2", 456).Return([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus},
	}, nil)

	var capturedData web_templates.PRListData
	mockTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(
		func(wr io.Writer, data any) error {
			if d, ok := data.(web_templates.PRListData); ok {
				capturedData = d
			}
			return nil
		})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
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
	Assert(t, errorPR.ErrorMessage != "", "error PR should have error message")
	Equals(t, 123, errorPR.PullNum)

	// Verify the successful PR is normal
	Equals(t, "passed", successPR.Status)
	Equals(t, "", successPR.ErrorMessage)
	Equals(t, 456, successPR.PullNum)
}

func TestPRController_PRList_SortsByLastActivity(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	now := time.Now()

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{
		{Num: 1, BaseRepo: models.Repo{FullName: "owner/old"}},
		{Num: 2, BaseRepo: models.Repo{FullName: "owner/new"}},
	}, nil)

	// Old PR - older activity
	mockDB.EXPECT().GetProjectOutputsByPull("owner/old", 1).Return([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus, CompletedAt: now.Add(-2 * time.Hour)},
	}, nil)

	// New PR - recent activity
	mockDB.EXPECT().GetProjectOutputsByPull("owner/new", 2).Return([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus, CompletedAt: now.Add(-5 * time.Minute)},
	}, nil)

	var capturedData web_templates.PRListData
	mockTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(
		func(wr io.Writer, data any) error {
			if d, ok := data.(web_templates.PRListData); ok {
				capturedData = d
			}
			return nil
		})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
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
		name     string
		success  int
		failed   int
		pending  int
		expected string
	}{
		{"all_success", 3, 0, 0, "passed"},
		{"all_failed", 0, 2, 0, "failed"},
		{"all_pending", 0, 0, 2, "pending"},
		{"mixed_with_failure", 2, 1, 0, "failed"},
		{"mixed_with_pending", 2, 0, 1, "pending"},
		{"failure_takes_precedence", 1, 1, 1, "failed"},
		{"no_projects", 0, 0, 0, "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := controllers.DetermineStatus(tt.success, tt.failed, tt.pending)
			Equals(t, tt.expected, status)
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
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	now := time.Now()

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo"}},
	}, nil)

	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo", 123).Return([]models.ProjectOutput{
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
	mockTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(
		func(wr io.Writer, data any) error {
			if d, ok := data.(web_templates.PRListData); ok {
				capturedData = d
			}
			return nil
		})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
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
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockDatabase(ctrl)
	mockTemplate := tMocks.NewMockTemplateWriter(ctrl)

	now := time.Now()

	mockDB.EXPECT().GetActivePullRequests().Return([]models.PullRequest{
		{Num: 123, BaseRepo: models.Repo{FullName: "owner/repo"}},
	}, nil)

	mockDB.EXPECT().GetProjectOutputsByPull("owner/repo", 123).Return([]models.ProjectOutput{
		{Status: models.SuccessOutputStatus, CompletedAt: now},
		{Status: models.FailedOutputStatus, CompletedAt: now},
		{Status: models.RunningOutputStatus, CompletedAt: now},
	}, nil)

	var capturedData web_templates.PRListData
	mockTemplate.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(
		func(wr io.Writer, data any) error {
			if d, ok := data.(web_templates.PRListData); ok {
				capturedData = d
			}
			return nil
		})

	controller := controllers.NewPRController(
		mockDB,
		mockTemplate,
		tMocks.NewMockTemplateWriter(ctrl),
		"1.0.0",
		"",
		func() bool { return false },
		nil, // getJobsForPull
		nil, // logger
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
