package server

import (
	"net/http"

	"strings"

	"github.com/hootsuite/atlantis/logging"
	"github.com/urfave/negroni"
)

func NewRequestLogger(logger *logging.SimpleLogger) *RequestLogger {
	return &RequestLogger{logger}
}

// RequestLogger logs requests and their response codes
type RequestLogger struct {
	logger *logging.SimpleLogger
}

func (l *RequestLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	if !strings.HasPrefix(r.URL.RequestURI(), "/static") {
		l.logger.Info("%d | %s %s", res.Status(), r.Method, r.URL.RequestURI())
	}
}
