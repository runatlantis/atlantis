// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestLocksPageController_Get(t *testing.T) {
	t.Run("renders locks page with locks", func(t *testing.T) {
		locks := map[string]models.ProjectLock{
			"test-lock": {
				Project: models.Project{
					RepoFullName: "owner/repo",
					Path:         "path/to/project",
				},
				Pull: models.PullRequest{
					Num: 123,
				},
				User: models.User{
					Username: "testuser",
				},
				Workspace: "default",
				Time:      time.Now(),
			},
		}

		tmpl := web_templates.LocksPageTemplate
		controller := controllers.NewLocksPageController(
			tmpl,
			func() (map[string]models.ProjectLock, error) { return locks, nil },
			func() bool { return false },
			"v0.1.0",
			"",
			nil, // logger
		)

		req := httptest.NewRequest(http.MethodGet, "/locks", nil)
		w := httptest.NewRecorder()

		controller.Get(w, req)

		Equals(t, http.StatusOK, w.Code)
		Assert(t, strings.Contains(w.Body.String(), "owner/repo"), "should contain repo name")
		Assert(t, strings.Contains(w.Body.String(), "testuser"), "should contain username")
	})

	t.Run("renders empty state when no locks", func(t *testing.T) {
		tmpl := web_templates.LocksPageTemplate
		controller := controllers.NewLocksPageController(
			tmpl,
			func() (map[string]models.ProjectLock, error) { return nil, nil },
			func() bool { return false },
			"v0.1.0",
			"",
			nil, // logger
		)

		req := httptest.NewRequest(http.MethodGet, "/locks", nil)
		w := httptest.NewRecorder()

		controller.Get(w, req)

		Equals(t, http.StatusOK, w.Code)
		Assert(t, strings.Contains(w.Body.String(), "No Active Locks"), "should show empty state")
	})
}
