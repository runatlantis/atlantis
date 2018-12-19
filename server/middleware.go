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
	"github.com/urfave/negroni"
)

// NewRequestLogger creates a RequestLogger.
func NewRequestLogger(logger *logging.SimpleLogger) *RequestLogger {
	return &RequestLogger{logger}
}

// RequestLogger logs requests and their response codes.
type RequestLogger struct {
	logger *logging.SimpleLogger
}

// ServeHTTP implements the middleware function. It logs a request at INFO
// level unless it's a request to /static/*.
func (l *RequestLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if l.shouldLog(r) {
		l.logger.Info("%s %s – from %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	}
	next(rw, r)
	if l.shouldLog(r) {
		l.logger.Info("%s %s – respond HTTP %d", r.Method, r.URL.RequestURI(), rw.(negroni.ResponseWriter).Status())
	}
}

func (l *RequestLogger) shouldLog(r *http.Request) bool {
	return !strings.HasPrefix(r.URL.RequestURI(), "/static")
}
