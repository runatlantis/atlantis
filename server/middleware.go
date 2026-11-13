// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/urfave/negroni/v3"
)

// NewRequestLogger creates a RequestLogger.
func NewRequestLogger(s *Server) *RequestLogger {
	return &RequestLogger{
		s.Logger,
		s.WebAuthentication,
		s.WebUsername,
		s.WebPassword,
	}
}

// RequestLogger logs requests and their response codes.
// as well as handle the basicauth on the requests
type RequestLogger struct {
	logger            logging.SimpleLogging
	WebAuthentication bool
	WebUsername       string
	WebPassword       string
}

// ServeHTTP implements the middleware function. It logs all requests at DEBUG level.
func (l *RequestLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.logger.Debug("%s %s – from %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	allowed := false
	if !l.WebAuthentication ||
		r.URL.Path == "/events" ||
		r.URL.Path == "/healthz" ||
		r.URL.Path == "/readyz" ||
		r.URL.Path == "/status" ||
		strings.HasPrefix(r.URL.Path, "/api/") {
		allowed = true
	} else {
		user, pass, ok := r.BasicAuth()
		if ok {
			r.SetBasicAuth(user, pass)
			if l.credentialsMatch(user, pass) {
				l.logger.Debug("[VALID] log in: >> url: %s", r.URL.RequestURI())
				allowed = true
			} else {
				allowed = false
				l.logger.Info("[INVALID] log in attempt: >> url: %s", r.URL.RequestURI())
			}
		}
	}
	if !allowed {
		rw.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
	} else {
		next(rw, r)
	}
	l.logger.Debug("%s %s – respond HTTP %d", r.Method, r.URL.RequestURI(), rw.(negroni.ResponseWriter).Status())
}

// credentialsMatch reports whether the supplied basic-auth credentials equal
// the configured ones. Each value is reduced to a SHA-256 digest before the
// constant-time comparison: ConstantTimeCompare returns early when the lengths
// differ, so comparing the raw values would reveal whether the supplied
// length was right. Both comparisons always run, even when the username is
// wrong, so the time taken to answer reveals neither how much of a value was
// correct nor which of the two was wrong.
func (l *RequestLogger) credentialsMatch(user, pass string) bool {
	userMatch := digestsEqual(user, l.WebUsername)
	passMatch := digestsEqual(pass, l.WebPassword)
	return userMatch && passMatch
}

// digestsEqual compares the SHA-256 digests of got and want in constant time.
func digestsEqual(got, want string) bool {
	gotSum := sha256.Sum256([]byte(got))
	wantSum := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gotSum[:], wantSum[:]) == 1
}
