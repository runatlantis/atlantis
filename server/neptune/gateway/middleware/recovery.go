package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/runatlantis/atlantis/server/logging"
)

// Recovery is middleware that recovers from any panics and writes a 500 if there was one.
type Recovery struct {
	Logger logging.Logger
}

func (m *Recovery) Middleware(next http.Handler) http.Handler {
	return &recoveryHandler{
		logger: m.Logger,
	}
}

type recoveryHandler struct {
	logger logging.Logger
	next   http.Handler
}

func (h *recoveryHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)

			// this value was taken from negroni, unsure why we use this specifically though
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]

			h.logger.ErrorContext(r.Context(), fmt.Sprintf("PANIC: %s\n%s", err, stack))
		}
	}()

	h.next.ServeHTTP(rw, r)
}
