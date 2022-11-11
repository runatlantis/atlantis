package http_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	httputil "github.com/runatlantis/atlantis/server/http"
	"github.com/stretchr/testify/assert"
)

func TestGetBody(t *testing.T) {
	requestBody := "body"
	rawRequest, err := http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBuffer([]byte(requestBody))))
	assert.NoError(t, err)

	subject, err := httputil.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	// read first time
	body, err := subject.GetBody()
	assert.NoError(t, err)

	payload1, err := io.ReadAll(body)
	assert.NoError(t, err)

	// read second time
	body, err = subject.GetBody()
	assert.NoError(t, err)

	payload2, err := io.ReadAll(body)
	assert.NoError(t, err)

	assert.Equal(t, payload1, payload2)
}

func TestGetRequest(t *testing.T) {
	requestBody := "body"
	rawRequest, err := http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBuffer([]byte(requestBody))))
	assert.NoError(t, err)

	subject1, err := httputil.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)

	// read from raw request first
	body := subject1.GetRequest().Body
	payload1, err := io.ReadAll(body)
	assert.NoError(t, err)

	// read from wrapper next
	body, err = subject1.GetBody()
	assert.NoError(t, err)

	payload2, err := io.ReadAll(body)
	assert.NoError(t, err)

	assert.Equal(t, payload1, payload2)
}
