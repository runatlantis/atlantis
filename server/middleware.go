// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package server

import (
	"net/http"
	"strings"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/oidc"
	"github.com/urfave/negroni/v3"
)

// NewRequestLogger creates a RequestLogger.
func NewRequestLogger(s *Server) *RequestLogger {
	rl := &RequestLogger{
		Logger:            s.Logger,
		WebAuthentication: s.WebAuthentication,
		WebUsername:        s.WebUsername,
		WebPassword:        s.WebPassword,
	}
	if s.OIDCController != nil {
		rl.WebOIDCAuth = s.WebOIDCAuth
		rl.OIDCSessionManager = s.OIDCController.SessionManager
	}
	return rl
}

// RequestLogger logs requests and their response codes,
// as well as handling authentication on the requests.
type RequestLogger struct {
	Logger             logging.SimpleLogging
	WebAuthentication  bool
	WebUsername        string
	WebPassword        string
	WebOIDCAuth       oidc.OIDCAuthProvider
	OIDCSessionManager *oidc.SessionManager
}

// isExemptPath returns true if the request path should bypass authentication.
func isExemptPath(path string) bool {
	return path == "/events" ||
		path == "/healthz" ||
		path == "/status" ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/auth/")
}

// authEnabled returns true if any authentication method is configured.
func (l *RequestLogger) authEnabled() bool {
	return l.WebAuthentication || l.WebOIDCAuth != ""
}

// ServeHTTP implements the middleware function. It logs all requests at DEBUG
// level and enforces authentication when configured.
func (l *RequestLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.Logger.Debug("%s %s – from %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)

	allowed := false

	hadBasicAuth := false

	if !l.authEnabled() || isExemptPath(r.URL.Path) {
		allowed = true
	} else {
		// Try basic auth first if enabled.
		if l.WebAuthentication {
			user, pass, ok := r.BasicAuth()
			if ok {
				hadBasicAuth = true
				if user == l.WebUsername && pass == l.WebPassword {
					l.Logger.Debug("[VALID] basic auth log in: >> url: %s", r.URL.RequestURI())
					allowed = true
				} else {
					l.Logger.Info("[INVALID] basic auth log in attempt: >> url: %s", r.URL.RequestURI())
				}
			}
		}

		// Try OIDC session cookie if basic auth didn't succeed.
		if !allowed && l.WebOIDCAuth != "" && l.OIDCSessionManager != nil {
			if _, err := l.OIDCSessionManager.GetSession(r); err == nil {
				l.Logger.Debug("[VALID] OIDC session >> url %q", r.URL.RequestURI())
				allowed = true
			}
		}
	}

	if !allowed {
		l.handleUnauthorized(rw, r, hadBasicAuth)
	} else {
		next(rw, r)
	}
	l.Logger.Debug("%s %s – respond HTTP %d", r.Method, r.URL.RequestURI(), rw.(negroni.ResponseWriter).Status())
}

// handleUnauthorized sends the appropriate response for unauthenticated
// requests depending on which auth methods are configured.
func (l *RequestLogger) handleUnauthorized(rw http.ResponseWriter, r *http.Request, hadBasicAuth bool) {
	// If OIDC is enabled and the request doesn't have basic auth credentials,
	// redirect to the login page instead of returning a 401.
	if l.WebOIDCAuth != "" && !hadBasicAuth {
		http.Redirect(rw, r, "auth/login", http.StatusFound)
		return
	}

	// Fall back to basic auth challenge.
	rw.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(rw, "Unauthorized", http.StatusUnauthorized)
}
