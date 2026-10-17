// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/controllers"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/core/locking"
	lockMocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/urfave/negroni/v3"
	"go.uber.org/mock/gomock"
)

type betaJobFixture struct{ jobs.NoopProjectOutputHandler }

func (*betaJobFixture) GetPullToJobMapping() []jobs.PullInfoWithJobIDs {
	return []jobs.PullInfoWithJobIDs{{Pull: jobs.PullInfo{RepoFullName: "org/repo", PullNum: 12, Workspace: "production", Path: "infra"}, JobIDInfos: []jobs.JobIDInfo{{JobID: "plan-job", JobStep: "plan"}}}}
}

func newBetaTestServer(t *testing.T) *server.Server {
	u, err := url.Parse("https://example.com/atlantis")
	Ok(t, err)
	s := &server.Server{
		Router: mux.NewRouter(), AtlantisURL: u, Logger: logging.NewNoopLogger(t),
		IndexTemplate: web_templates.IndexTemplate, ProjectCmdOutputHandler: &betaJobFixture{},
		StatusController: &controllers.StatusController{}, APIController: &controllers.APIController{},
		LocksController: &controllers.LocksController{}, GithubAppController: &controllers.GithubAppController{},
		JobsController: &controllers.JobsController{}, VCSEventsController: &events_controllers.VCSEventsController{},
	}
	s.SetupRoutes()
	return s
}

func TestBetaDashboardRoutesAndClassic(t *testing.T) {
	s := newBetaTestServer(t)
	ctrl := gomock.NewController(t)
	locker := lockMocks.NewMockLocker(ctrl)
	apply := lockMocks.NewMockApplyLocker(ctrl)
	locker.EXPECT().List().Return(map[string]models.ProjectLock{"lock-id": {Pull: models.PullRequest{Num: 12, Author: "owner", URL: "https://github.example.com/org/repo/pull/12"}, Project: models.Project{RepoFullName: "org/repo", Path: "infra"}, Workspace: "production"}}, nil).Times(2)
	apply.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{GlobalApplyLockEnabled: false}, nil).Times(2)
	s.Locker, s.ApplyLocker = locker, apply
	for _, path := range []string{"/beta", "/"} {
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		Equals(t, http.StatusOK, w.Code)
		Assert(t, strings.Contains(w.Body.String(), "/atlantis/jobs/plan-job"), "must preserve job URL")
		Assert(t, strings.Contains(w.Body.String(), "/atlantis/lock?id=lock-id"), "must preserve lock URL")
		Equals(t, path == "/beta", strings.Contains(w.Body.String(), "beta-dashboard.css"))
		if path == "/beta" {
			Assert(t, strings.Contains(w.Body.String(), `href="https://github.example.com/org/repo/pull/12"`), "beta must expose stored request URL")
		}
	}
	for _, asset := range []string{"/static/css/beta-dashboard.css", "/static/js/beta-dashboard-filters.js"} {
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, asset, nil))
		Equals(t, http.StatusOK, w.Code)
		Assert(t, w.Body.Len() > 0, "embedded asset must not be empty")
	}
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/beta", nil))
	Equals(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBetaDashboardBackendErrors(t *testing.T) {
	for _, backend := range []string{"locks", "apply"} {
		t.Run(backend, func(t *testing.T) {
			s := newBetaTestServer(t)
			ctrl := gomock.NewController(t)
			locker := lockMocks.NewMockLocker(ctrl)
			apply := lockMocks.NewMockApplyLocker(ctrl)
			s.Locker, s.ApplyLocker = locker, apply
			message := "Could not retrieve locks: unavailable"
			if backend == "locks" {
				locker.EXPECT().List().Return(nil, errors.New("unavailable"))
			} else {
				locker.EXPECT().List().Return(nil, nil)
				apply.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, errors.New("unavailable"))
				message = "Could not retrieve global apply lock: unavailable"
			}
			w := httptest.NewRecorder()
			s.Router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/beta", nil))
			ResponseContains(t, w, http.StatusServiceUnavailable, message)
		})
	}
}

func TestBetaDashboardUsesWebAuthentication(t *testing.T) {
	s := newBetaTestServer(t)
	s.WebAuthentication, s.WebUsername, s.WebPassword = true, "user", "password"
	for _, authorized := range []bool{false, true} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/beta", nil)
		if authorized {
			r.SetBasicAuth("user", "password")
		}
		called := false
		server.NewRequestLogger(s).ServeHTTP(negroni.NewResponseWriter(w), r, func(w http.ResponseWriter, _ *http.Request) { called = true; w.WriteHeader(http.StatusNoContent) })
		Equals(t, authorized, called)
		if authorized {
			Equals(t, http.StatusNoContent, w.Code)
		} else {
			Equals(t, http.StatusUnauthorized, w.Code)
		}
	}
}
