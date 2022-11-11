package testing

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func ResponseContains(t *testing.T, r *httptest.ResponseRecorder, status int, bodySubstr string) {
	t.Helper()
	body, err := io.ReadAll(r.Result().Body)
	Ok(t, err)
	Assert(t, status == r.Result().StatusCode, "exp %d got %d, body: %s", status, r.Result().StatusCode, string(body))
	Assert(t, strings.Contains(string(body), bodySubstr), "exp %q to be contained in %q", bodySubstr, string(body))
}
