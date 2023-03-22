package middleware

import (
	"context"
	"net/http"

	key "github.com/runatlantis/atlantis/server/neptune/context"
)

const ghRequestIDHeader = "X-Github-Delivery"

// RequestID is responsible for extract various types of request IDs from request headers
// and plumbing those through the context
type RequestID struct{}

func (m *RequestID) Middleware(next http.Handler) http.Handler {
	return &requestIDExtractor{
		next: next,
	}
}

type requestIDExtractor struct {
	next http.Handler
}

func (e *requestIDExtractor) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if id, ok := r.Header[ghRequestIDHeader]; ok {
		ctx := context.WithValue(
			r.Context(),
			key.RequestIDKey,
			id,
		)

		e.next.ServeHTTP(rw, r.WithContext(ctx))
		return
	}

	e.next.ServeHTTP(rw, r)
}
