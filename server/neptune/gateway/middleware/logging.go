package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/urfave/negroni"
)

// Logger logs request properties in addition to the response
type Logger struct {
	Logger logging.Logger
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return &loggerHandler{
		logger: l.Logger,
		next:   next,
	}

}

type loggerHandler struct {
	logger logging.Logger
	next   http.Handler
}

func (h *loggerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	wrappedRW := negroni.NewResponseWriter(rw)

	defer func() {
		status := wrappedRW.Status()
		duration := time.Since(start)

		h.logger.InfoContext(r.Context(), fmt.Sprintf("%s %s %s request complete", r.Method, r.Host, r.URL.Path), map[string]interface{}{
			"start-time": start,
			"duration":   duration,
			"status":     status,
		})
	}()

	h.next.ServeHTTP(wrappedRW, r)
}
