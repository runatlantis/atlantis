// Copyright 2024 contributors to runatlantis/atlantis.
// SPDX-License-Identifier: Apache-2.0

package providercache

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// registryHost is a stand-in registry label. The real upstream is an httptest
// server whose base URL we pre-seed into the discovery cache, so the label never
// needs to resolve.
const registryHost = "registry.example.com"

const (
	archiveFilename    = "terraform-provider-null_3.2.1_linux_amd64.zip"
	providerBinaryName = "terraform-provider-null_v3.2.1_x5"
)

// testSigningKey lazily creates one ephemeral PGP keypair, shared by every
// test in this package, and returns its entity (for signing) and its
// ASCII-armored public key (for the fake registry's signing_keys metadata).
// Using a real keypair - rather than stubbing signature verification out -
// means these tests exercise the installer's actual verification path.
var testSigningKey = sync.OnceValues(func() (*openpgp.Entity, string) {
	entity, err := openpgp.NewEntity("Test Registry", "", "test@example.com", nil)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	w, err := armor.Encode(&buf, openpgp.PublicKeyType, nil)
	if err != nil {
		panic(err)
	}
	if err := entity.Serialize(w); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return entity, buf.String()
})

// validArchive builds a real zip archive containing a single (fake) provider
// binary. It has to be a genuine zip: the installer really unpacks it.
func validArchive(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.CreateHeader(&zip.FileHeader{Name: providerBinaryName, Method: zip.Deflate})
	Ok(t, err)
	_, err = fw.Write([]byte("#!/bin/sh\necho fake provider\n"))
	Ok(t, err)
	Ok(t, zw.Close())
	return buf.Bytes()
}

// signedShasums returns a SHA256SUMS file listing archive's real digest under
// archiveFilename, and a binary (non-armored) detached signature of it from
// testSigningKey - matching what the archive/shasums/shasums_signature
// endpoints of a real registry return.
func signedShasums(t *testing.T, archive []byte) (shasums, signature []byte) {
	t.Helper()
	sum := sha256.Sum256(archive)
	shasums = []byte(hex.EncodeToString(sum[:]) + "  " + archiveFilename + "\n")

	entity, _ := testSigningKey()
	var sigBuf bytes.Buffer
	Ok(t, openpgp.DetachSign(&sigBuf, entity, bytes.NewReader(shasums), nil))
	return shasums, sigBuf.Bytes()
}

// newUpstream returns a fake registry server implementing the subset of the
// provider registry protocol the proxy uses, plus an artifact endpoint,
// serving a real, correctly-signed provider archive. The returned counter
// tracks how many times the archive itself was downloaded from upstream.
func newUpstream(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var archiveHits int32
	archive := validArchive(t)
	shasums, signature := signedShasums(t, archive)
	sum := sha256.Sum256(archive)
	_, publicKeyArmor := testSigningKey()

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
			"filename":              archiveFilename,
			"download_url":          self.URL + "/archives/" + archiveFilename,
			"shasums_url":           self.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS",
			"shasums_signature_url": self.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS.sig",
			"shasum":                hex.EncodeToString(sum[:]),
			"signing_keys": map[string]any{
				"gpg_public_keys": []any{
					map[string]any{"ascii_armor": publicKeyArmor},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/archives/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".zip"):
			atomic.AddInt32(&archiveHits, 1)
			_, _ = w.Write(archive)
		case strings.HasSuffix(r.URL.Path, ".sig"):
			_, _ = w.Write(signature)
		default:
			_, _ = w.Write(shasums)
		}
	})

	self = httptest.NewServer(mux)
	t.Cleanup(self.Close)
	return self, &archiveHits
}

// newUpstreamWithoutSigningKeys is like newUpstream, except the registry
// offers no signing keys for the provider - as a misconfigured or malicious
// registry might - so any signature the installer receives can never be
// verified against anything.
func newUpstreamWithoutSigningKeys(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var archiveHits int32
	archive := validArchive(t)
	shasums, signature := signedShasums(t, archive)
	sum := sha256.Sum256(archive)

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
			"filename":              archiveFilename,
			"download_url":          self.URL + "/archives/" + archiveFilename,
			"shasums_url":           self.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS",
			"shasums_signature_url": self.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS.sig",
			"shasum":                hex.EncodeToString(sum[:]),
			"signing_keys":          map[string]any{"gpg_public_keys": []any{}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/archives/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".zip"):
			atomic.AddInt32(&archiveHits, 1)
			_, _ = w.Write(archive)
		case strings.HasSuffix(r.URL.Path, ".sig"):
			_, _ = w.Write(signature)
		default:
			_, _ = w.Write(shasums)
		}
	})

	self = httptest.NewServer(mux)
	t.Cleanup(self.Close)
	return self, &archiveHits
}

// waitFor polls cond until it returns true or timeout elapses, failing the
// test with msg (formatted with args) if it never does. Used for asserting
// on the installer's background work, which has no synchronous completion
// signal from the caller's perspective by design (see handleArtifact).
func waitFor(t *testing.T, timeout time.Duration, cond func() bool, msg string, args ...any) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf(msg, args...)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// newProxy starts a proxy pointed (via a pre-seeded discovery entry) at upstream.
func newProxy(t *testing.T, upstream *httptest.Server) *Server {
	t.Helper()
	s, err := New(logging.NewNoopLogger(t), t.TempDir(), []string{registryHost}, 0, 0, 0)
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

	artifactBase := s.MirrorBaseURL() + "artifact/" + registryHost + "/hashicorp/null/3.2.1/linux/amd64/"
	expected := map[string]string{
		"download_url":          artifactBase + "archive",
		"shasums_url":           artifactBase + "shasums",
		"shasums_signature_url": artifactBase + "signature",
	}
	for field, want := range expected {
		v, _ := meta[field].(string)
		Equals(t, want, v)
		Assert(t, !strings.Contains(v, upstream.URL), "%s should not leak the upstream URL, got %q", field, v)
	}
	// Fields that must survive untouched.
	Equals(t, archiveFilename, meta["filename"])
	wantSum := sha256.Sum256(validArchive(t))
	Equals(t, hex.EncodeToString(wantSum[:]), meta["shasum"])
}

// Test that an archive request never serves bytes to the caller (that would
// let Terraform install it itself, which is exactly the race this proxy
// exists to avoid - see the package doc comment): it always answers 423 and,
// in the background, verifies and installs the provider into the mirror
// directory in Terraform's own provider_installation.filesystem_mirror
// layout. A burst of concurrent requests, as parallel `terraform init` runs
// would produce, collapses to a single upstream download.
func TestProxy_ArtifactInstallsIntoMirrorAndDedupes(t *testing.T) {
	upstream, archiveHits := newUpstream(t)
	s := newProxy(t, upstream)

	_, body := mustGet(t, s.MirrorBaseURL()+registryHost+"/v1/providers/hashicorp/null/3.2.1/download/linux/amd64")
	var meta map[string]any
	Ok(t, json.Unmarshal([]byte(body), &meta))
	artifactURL := meta["download_url"].(string)

	const n = 15
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			resp, _ := mustGet(t, artifactURL)
			Equals(t, http.StatusLocked, resp.StatusCode)
		})
	}
	wg.Wait()

	mirrorPath := filepath.Join(s.MirrorDir(), registryHost, "hashicorp", "null", "3.2.1", "linux_amd64")
	waitFor(t, 2*time.Second, func() bool {
		info, err := os.Stat(mirrorPath)
		return err == nil && info.IsDir()
	}, "provider was not installed into the mirror at %s", mirrorPath)

	installed, err := os.ReadFile(filepath.Join(mirrorPath, providerBinaryName))
	Ok(t, err)
	Equals(t, "#!/bin/sh\necho fake provider\n", string(installed))

	Equals(t, int32(1), atomic.LoadInt32(archiveHits))
}

// Test that a provider whose signature the proxy cannot verify (here: the
// registry offers no signing keys at all) is never published to the mirror -
// the installer fails closed rather than installing something it can't
// verify.
func TestProxy_ArtifactRefusesUnverifiableProvider(t *testing.T) {
	upstream, _ := newUpstreamWithoutSigningKeys(t)
	s := newProxy(t, upstream)

	_, body := mustGet(t, s.MirrorBaseURL()+registryHost+"/v1/providers/hashicorp/null/3.2.1/download/linux/amd64")
	var meta map[string]any
	Ok(t, json.Unmarshal([]byte(body), &meta))
	artifactURL := meta["download_url"].(string)

	resp, _ := mustGet(t, artifactURL)
	Equals(t, http.StatusLocked, resp.StatusCode)

	mirrorPath := filepath.Join(s.MirrorDir(), registryHost, "hashicorp", "null", "3.2.1", "linux_amd64")
	// Give the (expected to fail) background install every chance to
	// (wrongly) publish before asserting it never did.
	time.Sleep(200 * time.Millisecond)
	_, err := os.Stat(mirrorPath)
	Assert(t, os.IsNotExist(err), "an unverifiable provider must never be published to the mirror")

	// And the failure is queryable via /status, not just logged server-side -
	// this is what lets InitStepRunner surface the real cause once it gives
	// up waiting on the mirror.
	resp, body = mustGet(t, s.MirrorBaseURL()+"status/"+registryHost)
	Equals(t, http.StatusOK, resp.StatusCode)
	var status struct {
		Errors []string `json:"errors"`
	}
	Ok(t, json.Unmarshal([]byte(body), &status))
	Assert(t, len(status.Errors) == 1, "expected one recorded error, got %v", status.Errors)
	Assert(t, strings.Contains(status.Errors[0], "refusing to install an unverifiable package"), "unexpected status error: %s", status.Errors[0])
}

func TestProxy_StatusRejectsUnconfiguredRegistryHost(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	resp, _ := mustGet(t, s.MirrorBaseURL()+"status/evil.example.com")
	Equals(t, http.StatusNotFound, resp.StatusCode)
}

func TestProxy_ArtifactServesShasumsAndSignature(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	base := s.MirrorBaseURL() + "artifact/" + registryHost + "/hashicorp/null/3.2.1/linux/amd64/"

	resp, body := mustGet(t, base+"shasums")
	Equals(t, http.StatusOK, resp.StatusCode)
	Equals(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	wantSum := sha256.Sum256(validArchive(t))
	Equals(t, hex.EncodeToString(wantSum[:])+"  "+archiveFilename+"\n", body)

	resp, body = mustGet(t, base+"signature")
	Equals(t, http.StatusOK, resp.StatusCode)
	Equals(t, "application/octet-stream", resp.Header.Get("Content-Type"))
	Assert(t, len(body) > 0, "signature artifact should not be empty")
}

func TestProxy_ArtifactRejectsUnknownKind(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	resp, _ := mustGet(t, s.MirrorBaseURL()+"artifact/"+registryHost+"/hashicorp/null/3.2.1/linux/amd64/bogus")
	Equals(t, http.StatusNotFound, resp.StatusCode)
}

func TestProxy_ArtifactRejectsUnconfiguredRegistryHost(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	// The artifact endpoint is also bound to the configured registry allowlist.
	resp, _ := mustGet(t, s.MirrorBaseURL()+"artifact/evil.example.com/hashicorp/null/3.2.1/linux/amd64/archive")
	Equals(t, http.StatusNotFound, resp.StatusCode)
}

func TestSegmentPattern(t *testing.T) {
	for _, v := range []string{"hashicorp", "null", "3.2.1", "1.0.0-rc1", "linux", "amd64", "aws_v2"} {
		Assert(t, segmentPattern.MatchString(v), "%q should be a valid segment", v)
	}
	for _, v := range []string{"", "..", "a/b", "a b", "-x", ".x", "a?b", strings.Repeat("a", 200)} {
		Assert(t, !segmentPattern.MatchString(v), "%q should be an invalid segment", v)
	}
}

func TestProxy_RejectsUnconfiguredRegistryHost(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newProxy(t, upstream)

	// A host that is not in the configured registry allowlist must be refused
	// so the proxy cannot be used to reach arbitrary hosts (SSRF).
	resp, _ := mustGet(t, s.MirrorBaseURL()+"evil.example.com/v1/providers/hashicorp/null/versions")
	Equals(t, http.StatusNotFound, resp.StatusCode)
}

// newUnstartedProxy is like newProxy, but never calls Start: no HTTP server
// and, crucially, no background janitor goroutine. Used by tests that call
// ensureCached/sweepExpiredCache directly and want to backdate a file's mtime
// deterministically, without racing an hourly-ticker janitor that Start would
// otherwise launch concurrently (harmless in practice - the ticker won't fire
// within a test's lifetime - but still an unsynchronized access to s.maxAge
// under the race detector if a test mutates it after Start).
func newUnstartedProxy(t *testing.T, upstream *httptest.Server, maxAge time.Duration) *Server {
	t.Helper()
	s, err := New(logging.NewNoopLogger(t), t.TempDir(), []string{registryHost}, 0, 0, maxAge)
	Ok(t, err)
	t.Cleanup(func() { _ = s.listener.Close() })
	s.disco[registryHost] = upstream.URL + "/v1/providers/"
	return s
}

// Test that a cache hit in ensureCached bumps the blob's mtime - the signal
// sweepExpiredCache's age-based removal relies on to leave an actively-reused
// blob alone (see touch's doc comment).
func TestServer_EnsureCachedTouchesOnCacheHit(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newUnstartedProxy(t, upstream, 0)

	rawURL := upstream.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS"
	cachePath, err := s.ensureCached(context.Background(), rawURL)
	Ok(t, err)

	old := time.Now().Add(-2 * time.Hour)
	Ok(t, os.Chtimes(cachePath, old, old))

	// A second call for the same URL is a cache hit (the file already exists),
	// not a re-download.
	_, err = s.ensureCached(context.Background(), rawURL)
	Ok(t, err)

	info, err := os.Stat(cachePath)
	Ok(t, err)
	Assert(t, info.ModTime().After(old), "expected a cache hit to bump the blob's mtime past %s, got %s", old, info.ModTime())
}

// Test that sweepExpiredCache removes a raw artifact blob whose mtime is
// older than maxAge, while leaving a fresher one (or one untouched since
// download, i.e. never expired) in place.
func TestServer_SweepExpiredCacheRemovesExpiredKeepsFresh(t *testing.T) {
	upstream, _ := newUpstream(t)
	s := newUnstartedProxy(t, upstream, time.Hour)

	expiredURL := upstream.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS"
	freshURL := upstream.URL + "/archives/terraform-provider-null_3.2.1_SHA256SUMS.sig"

	expiredPath, err := s.ensureCached(context.Background(), expiredURL)
	Ok(t, err)
	freshPath, err := s.ensureCached(context.Background(), freshURL)
	Ok(t, err)

	old := time.Now().Add(-2 * time.Hour)
	Ok(t, os.Chtimes(expiredPath, old, old))

	Ok(t, s.sweepExpiredCache(time.Now()))

	_, err = os.Stat(expiredPath)
	Assert(t, os.IsNotExist(err), "expected the expired cache blob to be removed")
	_, err = os.Stat(freshPath)
	Ok(t, err)
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
