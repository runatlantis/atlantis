// Copyright 2024 contributors to runatlantis/atlantis.
// SPDX-License-Identifier: Apache-2.0

package providercache

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// registryHost is a stand-in registry label. The real upstream is an httptest
// server whose base URL we pre-seed into the discovery cache, so the label never
// needs to resolve.
const registryHost = "registry.example.com"

// zipBody is the fake provider archive returned by the upstream server.
const zipBody = "PK\x03\x04 fake provider archive"

// newUpstream returns a fake registry server implementing the subset of the
// provider registry protocol the proxy uses, plus an artifact endpoint. The
// returned counter tracks how many times the archive itself was downloaded.
func newUpstream(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var archiveHits int32
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/providers/hashicorp/null/versions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"versions":[{"version":"3.2.1","protocols":["5.0"],"platforms":[{"os":"linux","arch":"amd64"}]}]}`)
	})

	var self *httptest.Server
	mux.HandleFunc("/v1/providers/hashicorp/null/3.2.1/download/linux/amd64", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"protocols":             []string{"5.0"},
			"os":                    "linux",
			"arch":                  "amd64",
			"filename":              "terraform-provider-null_3.2.1_linux_amd64.zip",
			"download_url":          self.URL + "/archives/terraform-provider-null_3.2.1_linux_amd64.zip",
			"shasums_url":           self.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS",
			"shasums_signature_url": self.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS.sig",
			"shasum":                "abc123",
			"signing_keys":          map[string]any{"gpg_public_keys": []any{}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/archives/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".zip") {
			atomic.AddInt32(&archiveHits, 1)
		}
		_, _ = io.WriteString(w, zipBody)
	})

	self = httptest.NewServer(mux)
	t.Cleanup(self.Close)
	return self, &archiveHits
}

// newProxy starts a proxy pointed (via a pre-seeded discovery entry) at upstream.
func newProxy(t *testing.T, upstream *httptest.Server) *Server {
	t.Helper()
	s, err := New(logging.NewNoopLogger(t), t.TempDir(), []string{registryHost}, 0)
	Ok(t, err)
	// Bypass real (https) service discovery by seeding the resolved base URL.
	s.disco[registryHost] = upstream.URL + "/v1/providers/"
	s.Start()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Stop(ctx)
	})
	return s
}

func mustGet(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Get(url) //nolint:gosec // test-controlled URL
	Ok(t, err)
	body, err := io.ReadAll(resp.Body)
	Ok(t, err)
	resp.Body.Close()
	return resp, string(body)
}

func TestProxy_VersionsPassThrough(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	resp, body := mustGet(t, s.MirrorBaseURL()+registryHost+"/v1/providers/hashicorp/null/versions")
	Equals(t, http.StatusOK, resp.StatusCode)
	Assert(t, strings.Contains(body, `"version":"3.2.1"`), "versions body should be passed through, got %q", body)
}

func TestProxy_DownloadRewritesURLsThroughArtifactEndpoint(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	resp, body := mustGet(t, s.MirrorBaseURL()+registryHost+"/v1/providers/hashicorp/null/3.2.1/download/linux/amd64")
	Equals(t, http.StatusOK, resp.StatusCode)

	var meta map[string]any
	Ok(t, json.Unmarshal([]byte(body), &meta))

	for _, field := range []string{"download_url", "shasums_url", "shasums_signature_url"} {
		v, _ := meta[field].(string)
		Assert(t, strings.HasPrefix(v, s.MirrorBaseURL()+"artifact?"),
			"%s should be rewritten through the artifact endpoint, got %q", field, v)
		Assert(t, !strings.Contains(v, upstream.URL), "%s should not leak the upstream URL, got %q", field, v)
	}
	// Fields that must survive untouched.
	Equals(t, "terraform-provider-null_3.2.1_linux_amd64.zip", meta["filename"])
	Equals(t, "abc123", meta["shasum"])
}

func TestProxy_ArtifactCachesAndDedupes(t *testing.T) {
	upstream, archiveHits := newUpstream(t)
	s := newProxy(t, upstream)

	// Get the rewritten (signed) download URL from the metadata endpoint.
	_, body := mustGet(t, s.MirrorBaseURL()+registryHost+"/v1/providers/hashicorp/null/3.2.1/download/linux/amd64")
	var meta map[string]any
	Ok(t, json.Unmarshal([]byte(body), &meta))
	artifactURL := meta["download_url"].(string)

	// Fire many concurrent requests, as a burst of parallel `terraform init`
	// runs would. Only a single upstream download should happen.
	const n = 15
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, got := mustGet(t, artifactURL)
			Equals(t, http.StatusOK, resp.StatusCode)
			Equals(t, zipBody, got)
			Equals(t, "application/zip", resp.Header.Get("Content-Type"))
		}()
	}
	wg.Wait()

	Equals(t, int32(1), atomic.LoadInt32(archiveHits))
}

func TestProxy_ArtifactRejectsUnsignedURL(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	// A caller cannot supply an arbitrary URL without a valid signature.
	resp, _ := mustGet(t, s.MirrorBaseURL()+"artifact?url="+upstream.URL+"/archives/x.zip")
	Equals(t, http.StatusForbidden, resp.StatusCode)

	// Even a correctly signed but tampered URL is rejected.
	signed, err := s.artifactURL(upstream.URL + "/archives/terraform-provider-null_3.2.1_linux_amd64.zip")
	Ok(t, err)
	tampered := strings.Replace(signed, "null_3.2.1", "null_9.9.9", 1)
	resp2, _ := mustGet(t, tampered)
	Equals(t, http.StatusForbidden, resp2.StatusCode)
}

func TestAllowedArtifactURL(t *testing.T) {
	cases := map[string]bool{
		"https://releases.hashicorp.com/x.zip": true,
		"http://127.0.0.1:8080/x.zip":          true,
		"http://localhost:8080/x.zip":          true,
		"http://example.com/x.zip":             false,
		"ftp://example.com/x.zip":              false,
	}
	for raw, want := range cases {
		u, err := url.Parse(raw)
		Ok(t, err)
		Equals(t, want, allowedArtifactURL(u))
	}
}
