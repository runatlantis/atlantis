package http

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// BufferedRequest wraps an http request and contains a buffer of the request body,
// in addition to safe access to the request body and underlying request.
// BufferedRequest does not provide access to the original http request and instead
// vends copies of it.  This is to ensure that the original request body can be read
// multiple times and removes the need to think about this from the consumer end.
//
// Note: the OG request body must not have been closed before construction of this object.
//
// Since this is a server request we do not need to close the original Body as per the documentation:
//
// " For server requests, the Request Body is always non-nil
//
//	but will return EOF immediately when no body is present.
//	The Server will close the request body. The ServeHTTP
//	Handler does not need to. "
//
// Note: This should not be used for client requests at this time.
type BufferedRequest struct {
	request *http.Request
	body    *bytes.Buffer
}

func NewBufferedRequest(r *http.Request) (*BufferedRequest, error) {
	body, err := getBody(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading request body")
	}

	wrapped := &BufferedRequest{
		// clone the request because we've already read and closed the body of the OG.
		request: clone(r.Context(), r, body.Bytes()),
		body:    body,
	}

	return wrapped, nil
}

// GetHeader gets a specific header given a key
func (r *BufferedRequest) GetHeader(key string) string {
	return r.request.Header.Get(key)
}

// GetBody returns a copy of the request body
func (r *BufferedRequest) GetBody() (io.ReadCloser, error) {
	copy := bytes.NewBuffer(r.body.Bytes())
	return io.NopCloser(copy), nil
}

// GetRequest returns a clone of the underlying request in this struct
// Note: reading the request body directly from the returned object, will close it
// it's recommended to always be reading the body from GetBody instead.
func (r *BufferedRequest) GetRequest() *http.Request {
	return r.GetRequestWithContext(r.request.Context())
}

func (r *BufferedRequest) GetRequestWithContext(ctx context.Context) *http.Request {
	return clone(ctx, r.request, r.body.Bytes())
}

// Clone's a request and provides a new BufferedRequest
func clone(ctx context.Context, request *http.Request, body []byte) *http.Request {
	clone := request.Clone(ctx)

	// create one copy for underlying request and one for the new wrapper
	clone.Body = io.NopCloser(bytes.NewBuffer(body))
	return clone
}

func getBody(request *http.Request) (*bytes.Buffer, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(request.Body)
	if err != nil {
		return nil, err
	}

	return &b, nil
}
