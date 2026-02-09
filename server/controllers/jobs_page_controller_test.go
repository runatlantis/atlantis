// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/jobs"
	. "github.com/runatlantis/atlantis/testing"
)

type mockJobsTemplate struct {
	data    any
	execErr error
}

func (m *mockJobsTemplate) Execute(w io.Writer, data any) error {
	m.data = data
	if m.execErr != nil {
		return m.execErr
	}
	_, _ = w.Write([]byte("ok"))
	return nil
}

func TestJobsPageController_Get(t *testing.T) {
	t.Run("renders jobs page with jobs", func(t *testing.T) {
		tmpl := &mockJobsTemplate{}
		getJobs := func() []jobs.PullInfoWithJobIDs {
			return []jobs.PullInfoWithJobIDs{
				{
					Pull: jobs.PullInfo{
						PullNum:      123,
						Repo:         "repo",
						RepoFullName: "owner/repo",
					},
					JobIDInfos: []jobs.JobIDInfo{
						{JobID: "job1"},
						{JobID: "job2"},
					},
				},
			}
		}

		ctrl := controllers.NewJobsPageController(
			tmpl,
			tmpl, // partialTemplate - can reuse for simple tests
			getJobs,
			func() bool { return false },
			"v1.0.0",
			"/base",
		)

		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		w := httptest.NewRecorder()

		ctrl.Get(w, req)

		Equals(t, http.StatusOK, w.Code)
		data := tmpl.data.(web_templates.JobsPageData)
		Equals(t, 2, data.TotalCount)
		Equals(t, 1, len(data.Jobs))
	})

	t.Run("renders empty state when no jobs", func(t *testing.T) {
		tmpl := &mockJobsTemplate{}
		getJobs := func() []jobs.PullInfoWithJobIDs {
			return []jobs.PullInfoWithJobIDs{}
		}

		ctrl := controllers.NewJobsPageController(
			tmpl,
			tmpl, // partialTemplate - can reuse for simple tests
			getJobs,
			func() bool { return false },
			"v1.0.0",
			"/base",
		)

		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		w := httptest.NewRecorder()

		ctrl.Get(w, req)

		Equals(t, http.StatusOK, w.Code)
		data := tmpl.data.(web_templates.JobsPageData)
		Equals(t, 0, data.TotalCount)
		Equals(t, 0, len(data.Jobs))
	})

	t.Run("returns 500 on template error", func(t *testing.T) {
		tmpl := &mockJobsTemplate{execErr: bytes.ErrTooLarge}
		getJobs := func() []jobs.PullInfoWithJobIDs {
			return []jobs.PullInfoWithJobIDs{}
		}

		ctrl := controllers.NewJobsPageController(
			tmpl,
			tmpl, // partialTemplate - can reuse for simple tests
			getJobs,
			func() bool { return false },
			"v1.0.0",
			"/base",
		)

		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		w := httptest.NewRecorder()

		ctrl.Get(w, req)

		Equals(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("extracts unique repositories sorted alphabetically", func(t *testing.T) {
		tmpl := &mockJobsTemplate{}
		getJobs := func() []jobs.PullInfoWithJobIDs {
			return []jobs.PullInfoWithJobIDs{
				{
					Pull:       jobs.PullInfo{RepoFullName: "org/repo-b"},
					JobIDInfos: []jobs.JobIDInfo{{JobID: "job1"}},
				},
				{
					Pull:       jobs.PullInfo{RepoFullName: "org/repo-a"},
					JobIDInfos: []jobs.JobIDInfo{{JobID: "job2"}},
				},
				{
					Pull:       jobs.PullInfo{RepoFullName: "org/repo-b"},
					JobIDInfos: []jobs.JobIDInfo{{JobID: "job3"}},
				},
			}
		}

		ctrl := controllers.NewJobsPageController(
			tmpl,
			tmpl, // partialTemplate - can reuse for simple tests
			getJobs,
			func() bool { return false },
			"v1.0.0",
			"/base",
		)

		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		w := httptest.NewRecorder()

		ctrl.Get(w, req)

		data := tmpl.data.(web_templates.JobsPageData)
		Equals(t, 2, len(data.Repositories))
		Equals(t, "org/repo-a", data.Repositories[0])
		Equals(t, "org/repo-b", data.Repositories[1])
	})
}

func TestJobsPageController_GetPartial(t *testing.T) {
	t.Run("renders partial with jobs data", func(t *testing.T) {
		tmpl := &mockJobsTemplate{}
		partialTmpl := &mockJobsTemplate{}
		getJobs := func() []jobs.PullInfoWithJobIDs {
			return []jobs.PullInfoWithJobIDs{
				{
					Pull: jobs.PullInfo{
						PullNum:      123,
						RepoFullName: "owner/repo",
					},
					JobIDInfos: []jobs.JobIDInfo{
						{JobID: "job1", JobStep: "plan"},
					},
				},
			}
		}

		ctrl := controllers.NewJobsPageController(
			tmpl,
			partialTmpl,
			getJobs,
			func() bool { return false },
			"v1.0.0",
			"/base",
		)

		req := httptest.NewRequest(http.MethodGet, "/jobs/partial", nil)
		w := httptest.NewRecorder()

		ctrl.GetPartial(w, req)

		Equals(t, http.StatusOK, w.Code)
		data := partialTmpl.data.(web_templates.JobsPageData)
		Equals(t, 1, data.TotalCount)
	})
}
