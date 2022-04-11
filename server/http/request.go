package http

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// ClonableRequest wraps an http request to provide a safe method of cloning before
// a request body is closed.
// It does this by providing a safe way to read a request body.  Since this is a server
// request we do not need to close the Body as per the documentation:
//
// " For server requests, the Request Body is always non-nil
//   but will return EOF immediately when no body is present.
//   The Server will close the request body. The ServeHTTP
//   Handler does not need to. "
//
// Note: This should not be used for client requests at this time.
type CloneableRequest struct {
	request *http.Request
}

func NewCloneableRequest(r *http.Request) (*CloneableRequest, error) {
	wrapped := &CloneableRequest{
		request: r,
	}

	clone, err := wrapped.Clone(r.Context())

	if err != nil {
		return nil, errors.Wrap(err, "creating request clone")
	}
	return clone, nil
}

// GetHeader gets a specific header given a key
func (r *CloneableRequest) GetHeader(key string) string {
	return r.request.Header.Get(key)
}

// GetBody returns a copy of the request body
func (r *CloneableRequest) GetBody() (io.ReadCloser, error) {
	b, err := getBody(r.request)
	if err != nil {
		return nil, errors.Wrap(err, "reading request body")
	}
	return io.NopCloser(b), nil
}

// GetRequest returns the underlying request as an escape hatch.
// Note: Using this is risky since the body can be modified/closed
// directly through this object.
// Use with caution.
func (r *CloneableRequest) GetRequest() *http.Request {
	return r.request
}

// Clone's a request and provides a new CloneableRequest
func (r *CloneableRequest) Clone(ctx context.Context) (*CloneableRequest, error) {
	clone := r.request.Clone(ctx)

	body, err := getBody(r.request)

	if err != nil {
		return nil, errors.Wrap(err, "cloning request body")
	}
	clone.Body = io.NopCloser(body)
	return &CloneableRequest{request: clone}, nil
}

func getBody(request *http.Request) (*bytes.Buffer, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(request.Body)
	if err != nil {
		return nil, err
	}

	return &b, nil
}
