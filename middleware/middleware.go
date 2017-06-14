package middleware

import (
	"net/http"
	"github.com/urfave/negroni"
	"github.com/hootsuite/atlantis/logging"
)

func NewNon200Logger(logger *logging.SimpleLogger) *FailedRequestLogger {
	return &FailedRequestLogger{logger}
}

// FailedRequestLogger logs the request when a response code >= 400 is sent
type FailedRequestLogger struct{
	logger *logging.SimpleLogger
}

func (l *FailedRequestLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	if res.Status() >= 400 {
		l.logger.Info("%s %s - Response code %d", r.Method, r.URL.RequestURI(), res.Status())
	}
}
