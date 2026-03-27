// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/oidc"
	. "github.com/runatlantis/atlantis/testing"
)

func TestOIDCController_Login_Redirect(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	w := httptest.NewRecorder()
	_, err := sm.CreateState(w)
	Ok(t, err)

	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "atlantis_oidc_state" {
			stateCookie = c
		}
	}
	Assert(t, stateCookie != nil, "expected state cookie to be set")
}

func TestOIDCController_Callback_MissingCode(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	ctrl := &controllers.OIDCController{
		Provider:       nil,
		SessionManager: sm,
		Logger:         logger,
	}

	r := httptest.NewRequest("GET", "/auth/callback?state=test", nil)
	r.AddCookie(&http.Cookie{
		Name:  "atlantis_oidc_state",
		Value: "test",
	})
	w := httptest.NewRecorder()

	ctrl.Callback(w, r)

	Equals(t, http.StatusBadRequest, w.Code)
}

func TestOIDCController_Callback_MissingState(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	ctrl := &controllers.OIDCController{
		Provider:       nil,
		SessionManager: sm,
		Logger:         logger,
	}

	r := httptest.NewRequest("GET", "/auth/callback", nil)
	w := httptest.NewRecorder()

	ctrl.Callback(w, r)

	Equals(t, http.StatusBadRequest, w.Code)
}

func TestOIDCController_Callback_IDPError(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	ctrl := &controllers.OIDCController{
		Provider:       nil,
		SessionManager: sm,
		Logger:         logger,
	}

	r := httptest.NewRequest("GET", "/auth/callback?error=access_denied&error_description=User+cancelled", nil)
	w := httptest.NewRecorder()

	ctrl.Callback(w, r)

	Equals(t, http.StatusBadRequest, w.Code)
}

func TestOIDCController_Logout(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	ctrl := &controllers.OIDCController{
		Provider:       nil,
		SessionManager: sm,
		Logger:         logger,
	}

	r := httptest.NewRequest("GET", "/auth/logout", nil)
	w := httptest.NewRecorder()

	ctrl.Logout(w, r)

	Equals(t, http.StatusFound, w.Code)

	var cleared bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "atlantis_oidc" && c.MaxAge == -1 {
			cleared = true
		}
	}
	Assert(t, cleared, "expected session cookie to be cleared")
}
