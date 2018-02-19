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
	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	if !strings.HasPrefix(r.URL.RequestURI(), "/static") {
		l.logger.Info("%d | %s %s", res.Status(), r.Method, r.URL.RequestURI())
	}
}
