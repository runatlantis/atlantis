// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

// recorder is a concurrency-safe record of the basic-auth tokens seen by the
// test server.
type recorder struct {
	mu     sync.Mutex
	hit    bool
	tokens []string
}

func (r *recorder) record(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hit = true
	r.tokens = append(r.tokens, token)
}

func (r *recorder) wasHit() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hit
}

func (r *recorder) seen() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.tokens))
	copy(out, r.tokens)
	return out
}

// newRecordingServer returns an httptest server that records the basic-auth
// password (the Azure DevOps token) of every request it receives.
func newRecordingServer(t *testing.T, rec *recorder) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, password, _ := r.BasicAuth()
		rec.record(password)
		w.WriteHeader(http.StatusOK)
	}))
}

func doGet(t *testing.T, c *http.Client, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	Ok(t, err)
	return c.Do(req)
}

func TestTokenTransport_RoundTrip_SendsStaticToken(t *testing.T) {
	rec := &recorder{}
	ts := newRecordingServer(t, rec)
	defer ts.Close()

	tr := &tokenTransport{
		credentials: &PATCredentials{User: "user", Token: "static-token"},
	}
	client := &http.Client{Transport: tr}

	resp, err := doGet(t, client, ts.URL)
	Ok(t, err)
	defer resp.Body.Close()

	Equals(t, http.StatusOK, resp.StatusCode)
	Equals(t, []string{"static-token"}, rec.seen())
}

func TestTokenTransport_RoundTrip_RereadsTokenPerRequest(t *testing.T) {
	rec := &recorder{}
	ts := newRecordingServer(t, rec)
	defer ts.Close()

	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("token-1"), 0600))

	tr := &tokenTransport{
		credentials: &PATCredentials{User: "user", TokenFile: tokenFile},
	}
	client := &http.Client{Transport: tr}

	resp1, err := doGet(t, client, ts.URL)
	Ok(t, err)
	resp1.Body.Close()

	// Rotate the token on disk between requests; the next request must use the
	// new value without recreating the client or transport.
	Ok(t, os.WriteFile(tokenFile, []byte("token-2"), 0600))

	resp2, err := doGet(t, client, ts.URL)
	Ok(t, err)
	resp2.Body.Close()

	Equals(t, []string{"token-1", "token-2"}, rec.seen())
}

func TestTokenTransport_RoundTrip_TrimsWhitespace(t *testing.T) {
	rec := &recorder{}
	ts := newRecordingServer(t, rec)
	defer ts.Close()

	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("  spaced-token\n"), 0600))

	tr := &tokenTransport{
		credentials: &PATCredentials{User: "user", TokenFile: tokenFile},
	}
	client := &http.Client{Transport: tr}

	resp, err := doGet(t, client, ts.URL)
	Ok(t, err)
	resp.Body.Close()

	Equals(t, []string{"spaced-token"}, rec.seen())
}

func TestTokenTransport_RoundTrip_ErrorWhenTokenUnavailable(t *testing.T) {
	rec := &recorder{}
	ts := newRecordingServer(t, rec)
	defer ts.Close()

	tr := &tokenTransport{
		credentials: &PATCredentials{User: "user", TokenFile: "/does/not/exist"},
	}
	client := &http.Client{Transport: tr}

	_, err := doGet(t, client, ts.URL)
	Assert(t, err != nil, "expected an error when the token cannot be read")
	Assert(t, !rec.wasHit(), "no request should reach the server when the token is unavailable")
}

func TestTokenTransport_RoundTrip_ErrorWhenTokenEmpty(t *testing.T) {
	rec := &recorder{}
	ts := newRecordingServer(t, rec)
	defer ts.Close()

	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	// A wiped/empty (whitespace-only) token file.
	Ok(t, os.WriteFile(tokenFile, []byte("  \n"), 0600))

	tr := &tokenTransport{
		credentials: &PATCredentials{User: "user", TokenFile: tokenFile},
	}
	client := &http.Client{Transport: tr}

	_, err := doGet(t, client, ts.URL)
	Assert(t, err != nil, "expected an error when the token is empty")
	Assert(t, !rec.wasHit(), "no request should be sent with an empty token")

	// Once the token is restored the very next request succeeds (self-healing).
	Ok(t, os.WriteFile(tokenFile, []byte("restored-token"), 0600))
	resp, err := doGet(t, client, ts.URL)
	Ok(t, err)
	resp.Body.Close()
	Equals(t, []string{"restored-token"}, rec.seen())
}

// TestTokenTransport_RoundTrip_ConcurrentRequests ensures the transport is safe
// for concurrent use (Atlantis serves many pull requests at once). It must not
// share mutable state across requests; run with -race to catch regressions.
func TestTokenTransport_RoundTrip_ConcurrentRequests(t *testing.T) {
	rec := &recorder{}
	ts := newRecordingServer(t, rec)
	defer ts.Close()

	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("concurrent-token"), 0600))

	tr := &tokenTransport{
		credentials: &PATCredentials{User: "user", TokenFile: tokenFile},
	}
	client := &http.Client{Transport: tr}

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			resp, err := doGet(t, client, ts.URL)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	seen := rec.seen()
	Equals(t, n, len(seen))
	for _, tok := range seen {
		Equals(t, "concurrent-token", tok)
	}
}
