package gateway_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/gateway"
	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	req, _ := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	gateway.Healthz(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	body, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, "application/json", w.Result().Header["Content-Type"][0])
	assert.Equal(t,
		`{
  "status": "ok"
}`, string(body))
}
