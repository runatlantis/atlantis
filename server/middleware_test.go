// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/oidc"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/urfave/negroni/v3"
)

func newMiddleware(t *testing.T, basicAuth bool, username, password string, oidcAuth oidc.OIDCAuthProvider, sm *oidc.SessionManager) *server.RequestLogger {
	t.Helper()
	return &server.RequestLogger{
		Logger:             logging.NewNoopLogger(t),
		WebAuthentication:  basicAuth,
		WebUsername:        username,
		WebPassword:        password,
		WebOIDCAuth:        oidcAuth,
		OIDCSessionManager: sm,
	}
}

func serveMiddleware(mw *server.RequestLogger, r *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	w := negroni.NewResponseWriter(rec)
	mw.ServeHTTP(w, r, func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK")) //nolint:errcheck
	})
	return rec
}

func TestMiddleware_NoAuthEnabled(t *testing.T) {
	mw := newMiddleware(t, false, "", "", "", nil)
	r := httptest.NewRequest("GET", "/", nil)
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusOK, w.Code)
}

func TestMiddleware_ExemptPaths(t *testing.T) {
	mw := newMiddleware(t, true, "user", "pass", oidc.OIDCAuthEntraID, nil)

	for _, path := range []string{"/events", "/healthz", "/status", "/api/plan", "/auth/login", "/auth/callback"} {
		r := httptest.NewRequest("GET", path, nil)
		w := serveMiddleware(mw, r)
		Equals(t, http.StatusOK, w.Code)
	}
}

func TestMiddleware_BasicAuth_ValidCredentials(t *testing.T) {
	mw := newMiddleware(t, true, "admin", "secret", "", nil)

	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("admin", "secret")
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusOK, w.Code)
}

func TestMiddleware_BasicAuth_InvalidCredentials(t *testing.T) {
	mw := newMiddleware(t, true, "admin", "secret", "", nil)

	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("admin", "wrong")
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusUnauthorized, w.Code)
	Assert(t, w.Header().Get("WWW-Authenticate") != "", "expected WWW-Authenticate header")
}

func TestMiddleware_BasicAuth_NoCreds(t *testing.T) {
	mw := newMiddleware(t, true, "admin", "secret", "", nil)

	r := httptest.NewRequest("GET", "/", nil)
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_OIDC_ValidSession(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	cookieW := httptest.NewRecorder()
	Ok(t, sm.SetSession(cookieW, "user@example.com"))

	mw := newMiddleware(t, false, "", "", oidc.OIDCAuthEntraID, sm)

	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookieW.Result().Cookies() {
		r.AddCookie(c)
	}
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusOK, w.Code)
}

func TestMiddleware_OIDC_NoSession_RedirectsToLogin(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	mw := newMiddleware(t, false, "", "", oidc.OIDCAuthEntraID, sm)

	r := httptest.NewRequest("GET", "/", nil)
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusFound, w.Code)
	Equals(t, "/auth/login", w.Header().Get("Location"))
}

func TestMiddleware_BothAuth_BasicAuthWins(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	mw := newMiddleware(t, true, "admin", "secret", oidc.OIDCAuthEntraID, sm)

	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("admin", "secret")
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusOK, w.Code)
}

func TestMiddleware_BothAuth_OIDCFallback(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	cookieW := httptest.NewRecorder()
	Ok(t, sm.SetSession(cookieW, "user@example.com"))

	mw := newMiddleware(t, true, "admin", "secret", oidc.OIDCAuthEntraID, sm)

	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookieW.Result().Cookies() {
		r.AddCookie(c)
	}
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusOK, w.Code)
}

func TestMiddleware_BothAuth_InvalidBasicAuth_Returns401(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	sm := oidc.NewSessionManager([]byte("test-secret-32-characters-long!!"), false, "", logger)

	mw := newMiddleware(t, true, "admin", "secret", oidc.OIDCAuthEntraID, sm)

	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("admin", "wrong")
	w := serveMiddleware(mw, r)

	Equals(t, http.StatusUnauthorized, w.Code)
}
