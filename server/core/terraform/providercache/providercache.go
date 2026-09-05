// Copyright 2024 contributors to runatlantis/atlantis.
// SPDX-License-Identifier: Apache-2.0
//
// Package providercache implements a caching proxy for Terraform providers.
//
// When Atlantis runs many `terraform init` commands in parallel (across
// workspaces and pull requests) each of those processes independently reaches
// out to the upstream provider registry (e.g. registry.terraform.io) and
// downloads the same provider archives from releases.hashicorp.com. That wastes
// bandwidth, is slow, and is prone to upstream rate limiting.
//
// This package runs a small HTTP server on localhost that speaks the Terraform
// Provider Registry Protocol. Terraform is pointed at it via a `host` block in
// the CLI configuration file (see the tfclient package), which redirects
// service discovery for the configured registry hostnames to this proxy. The
// proxy forwards provider metadata requests upstream and rewrites the archive
// download URLs so that Terraform fetches the archives (and their SHA256SUMS /
// signature files) back through the proxy. The proxy downloads each archive
// from upstream exactly once, caches it on disk, and serves the cached copy to
// every subsequent request — including the many concurrent requests a burst of
// parallel `terraform init` runs produces, which are de-duplicated so only a
// single upstream download happens per artifact.
//
// The archive bytes are served verbatim, so Terraform's normal checksum and
// GPG-signature verification of providers is unaffected: this proxy only
// changes where the bytes come from, never what they are.
//
// This mirrors the provider cache server that Terragrunt already ships.
package providercache

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/singleflight"

	"github.com/runatlantis/atlantis/server/logging"
)

const (
	// loopbackHost is the address the proxy binds to. It is only ever reached
	// by terraform processes running on the same host as Atlantis, so it must
	// not be exposed on a routable interface.
	loopbackHost = "127.0.0.1"

	// discoveryTimeout bounds service-discovery and metadata requests. Archive
	// downloads are deliberately not bounded by this because they can be large.
	metadataTimeout = 30 * time.Second
)

// Server is the caching provider proxy. Create it with New and start it with
// Start.
type Server struct {
	log logging.SimpleLogging

	// cacheDir is where downloaded artifacts (archives, SHA256SUMS, signatures)
	// are stored on disk, keyed by a hash of their upstream URL.
	cacheDir string

	// registries is the set of registry hostnames (e.g. "registry.terraform.io")
	// that this proxy will serve. A CLI-config host block is generated for each.
	registries []string

	// hmacKey signs the rewritten artifact URLs so that the artifact endpoint
	// only ever fetches URLs the proxy itself produced, preventing it from being
	// abused as a generic outbound HTTP proxy (SSRF).
	hmacKey []byte

	metadataClient *http.Client
	// downloadClient has no overall timeout because provider archives can be
	// hundreds of megabytes; the transport still bounds connect/idle time.
	downloadClient *http.Client

	// disco caches the resolved providers.v1 base URL per registry host.
	discoMu sync.Mutex
	disco   map[string]string

	// sf de-duplicates concurrent downloads of the same artifact so a burst of
	// parallel `terraform init` runs triggers a single upstream download.
	sf singleflight.Group

	listener   net.Listener
	httpServer *http.Server
}

// New constructs a provider cache proxy. cacheDir must already exist. registries
// is the list of registry hostnames to serve; it must contain at least one
// entry. port is the TCP port to listen on; pass 0 to let the OS choose a free
// port (recommended).
func New(log logging.SimpleLogging, cacheDir string, registries []string, port int) (*Server, error) {
	if len(registries) == 0 {
		return nil, errors.New("at least one registry host is required")
	}
	if info, err := os.Stat(cacheDir); err != nil {
		return nil, fmt.Errorf("provider cache dir %q is not usable: %w", cacheDir, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("provider cache dir %q is not a directory", cacheDir)
	}

	hmacKey := make([]byte, 32)
	if _, err := rand.Read(hmacKey); err != nil {
		return nil, fmt.Errorf("generating provider cache signing key: %w", err)
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(loopbackHost, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, fmt.Errorf("listening on %s: %w", loopbackHost, err)
	}

	s := &Server{
		log:            log,
		cacheDir:       cacheDir,
		registries:     registries,
		hmacKey:        hmacKey,
		metadataClient: &http.Client{Timeout: metadataTimeout},
		downloadClient: &http.Client{},
		disco:          make(map[string]string),
		listener:       listener,
	}

	router := mux.NewRouter()
	// Terraform Provider Registry Protocol endpoints, namespaced by the registry
	// host so a single proxy can serve multiple registries.
	router.HandleFunc("/{host}/v1/providers/{namespace}/{type}/versions", s.handleVersions).Methods(http.MethodGet)
	router.HandleFunc("/{host}/v1/providers/{namespace}/{type}/{version}/download/{os}/{arch}", s.handleDownload).Methods(http.MethodGet)
	// The rewritten archive/checksum download endpoint that actually caches.
	router.HandleFunc("/artifact", s.handleArtifact).Methods(http.MethodGet)
	router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }).Methods(http.MethodGet)

	s.httpServer = &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s, nil
}

// Start begins serving in a background goroutine. It returns immediately once
// the listener is accepting connections (the listener is already open after
// New), so the address returned by Addr is valid as soon as New returns.
func (s *Server) Start() {
	go func() {
		if err := s.httpServer.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Err("provider cache proxy stopped: %s", err)
		}
	}()
	s.log.Info("provider cache proxy listening on %s, caching to %s", s.Addr(), s.cacheDir)
}

// Stop gracefully shuts the proxy down.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Addr is the host:port the proxy is listening on.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// MirrorBaseURL is the base URL of the proxy, ending in a trailing slash. It is
// used to build the CLI-config host blocks.
func (s *Server) MirrorBaseURL() string {
	return fmt.Sprintf("http://%s/", s.Addr())
}

// Registries returns the registry hostnames this proxy serves.
func (s *Server) Registries() []string {
	return s.registries
}

// handleVersions proxies the "list available versions" registry endpoint. The
// response contains no URLs, so it is passed straight through.
func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	base, err := s.discover(r.Context(), vars["host"])
	if err != nil {
		s.proxyError(w, "service discovery", vars["host"], err)
		return
	}
	upstream := base + url.PathEscape(vars["namespace"]) + "/" + url.PathEscape(vars["type"]) + "/versions"
	s.pipe(w, r, upstream)
}

// downloadResponse is the subset of the "find a provider package" response we
// need to rewrite. Unknown fields are preserved via the raw map so that
// signing_keys, shasum, filename, protocols, etc. reach Terraform unchanged.
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	base, err := s.discover(r.Context(), vars["host"])
	if err != nil {
		s.proxyError(w, "service discovery", vars["host"], err)
		return
	}
	upstream := base +
		url.PathEscape(vars["namespace"]) + "/" +
		url.PathEscape(vars["type"]) + "/" +
		url.PathEscape(vars["version"]) + "/download/" +
		url.PathEscape(vars["os"]) + "/" +
		url.PathEscape(vars["arch"])

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream, nil)
	if err != nil {
		s.proxyError(w, "building request", upstream, err)
		return
	}
	resp, err := s.metadataClient.Do(req)
	if err != nil {
		s.proxyError(w, "download metadata", upstream, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s.relayStatus(w, resp)
		return
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		s.proxyError(w, "decoding download metadata", upstream, err)
		return
	}

	// Rewrite the three fields that point at externally-hosted files so that
	// Terraform fetches them back through the caching artifact endpoint.
	for _, field := range []string{"download_url", "shasums_url", "shasums_signature_url"} {
		orig, ok := body[field].(string)
		if !ok || orig == "" {
			continue
		}
		signed, err := s.artifactURL(orig)
		if err != nil {
			s.proxyError(w, "rewriting "+field, orig, err)
			return
		}
		body[field] = signed
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		s.log.Err("provider cache: writing download response: %s", err)
	}
}

// handleArtifact downloads (once) and caches the artifact identified by the
// signed url query parameter, then serves it from disk.
func (s *Server) handleArtifact(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	sig := r.URL.Query().Get("sig")
	if raw == "" || !s.validSignature(raw, sig) {
		http.Error(w, "invalid or unsigned artifact url", http.StatusForbidden)
		return
	}
	parsed, err := url.Parse(raw)
	if err != nil || !allowedArtifactURL(parsed) {
		http.Error(w, "artifact url must be https (or http to a loopback address)", http.StatusBadRequest)
		return
	}

	cachePath, err := s.ensureCached(r.Context(), raw)
	if err != nil {
		s.proxyError(w, "caching artifact", raw, err)
		return
	}

	// Give archives and checksum files sensible content types. http.ServeFile
	// handles Range requests, caching headers and streaming from disk.
	switch {
	case strings.HasSuffix(parsed.Path, ".zip"):
		w.Header().Set("Content-Type", "application/zip")
	case strings.HasSuffix(parsed.Path, ".sig"):
		w.Header().Set("Content-Type", "application/octet-stream")
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	http.ServeFile(w, r, cachePath)
}

// ensureCached returns the on-disk path of the artifact at rawURL, downloading
// it from upstream if it is not already cached. Concurrent callers for the same
// URL share a single download.
func (s *Server) ensureCached(ctx context.Context, rawURL string) (string, error) {
	cachePath := s.cachePath(rawURL)
	if _, err := os.Stat(cachePath); err == nil {
		s.log.Debug("provider cache hit: %s", rawURL)
		return cachePath, nil
	}

	_, err, _ := s.sf.Do(rawURL, func() (any, error) {
		// Re-check under the singleflight barrier: a sibling request may have
		// finished the download while we were waiting.
		if _, err := os.Stat(cachePath); err == nil {
			return nil, nil
		}
		s.log.Info("provider cache miss, downloading %s", rawURL)
		return nil, s.download(ctx, rawURL, cachePath)
	})
	if err != nil {
		return "", err
	}
	return cachePath, nil
}

// download streams rawURL to cachePath atomically (via a temp file + rename) so
// a partial download can never be observed as a complete cache entry.
func (s *Server) download(ctx context.Context, rawURL, cachePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := s.downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream returned %d for %s", resp.StatusCode, rawURL)
	}

	tmp, err := os.CreateTemp(s.cacheDir, ".download-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Clean up the temp file on any error path.
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, cachePath); err != nil {
		return err
	}
	success = true
	return nil
}

// discover resolves the providers.v1 base URL for a registry host, caching the
// result. It uses Terraform's remote service discovery well-known document and
// falls back to the conventional /v1/providers/ path if discovery fails.
func (s *Server) discover(ctx context.Context, host string) (string, error) {
	if host == "" {
		return "", errors.New("empty registry host")
	}
	s.discoMu.Lock()
	cached, ok := s.disco[host]
	s.discoMu.Unlock()
	if ok {
		return cached, nil
	}

	base := s.discoverUncached(ctx, host)

	s.discoMu.Lock()
	s.disco[host] = base
	s.discoMu.Unlock()
	return base, nil
}

func (s *Server) discoverUncached(ctx context.Context, host string) string {
	fallback := fmt.Sprintf("https://%s/v1/providers/", host)

	discoURL := fmt.Sprintf("https://%s/.well-known/terraform.json", host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoURL, nil)
	if err != nil {
		return fallback
	}
	resp, err := s.metadataClient.Do(req)
	if err != nil {
		s.log.Warn("provider cache: service discovery for %s failed (%s); using %s", host, err, fallback)
		return fallback
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fallback
	}
	var doc map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fallback
	}
	providersV1, ok := doc["providers.v1"].(string)
	if !ok || providersV1 == "" {
		return fallback
	}
	// providers.v1 may be relative to the discovery host.
	ref, err := url.Parse(providersV1)
	if err != nil {
		return fallback
	}
	resolved := (&url.URL{Scheme: "https", Host: host}).ResolveReference(ref)
	out := resolved.String()
	if !strings.HasSuffix(out, "/") {
		out += "/"
	}
	return out
}

// pipe forwards a GET to upstream and copies the status, content type and body
// back to the client unchanged.
func (s *Server) pipe(w http.ResponseWriter, r *http.Request, upstream string) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream, nil)
	if err != nil {
		s.proxyError(w, "building request", upstream, err)
		return
	}
	resp, err := s.metadataClient.Do(req)
	if err != nil {
		s.proxyError(w, "proxying", upstream, err)
		return
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		s.log.Debug("provider cache: copying response for %s: %s", upstream, err)
	}
}

// relayStatus forwards a non-200 upstream response (status + body) to the
// client so Terraform sees the registry's own error (e.g. 404 for an unknown
// provider).
func (s *Server) relayStatus(w http.ResponseWriter, resp *http.Response) {
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) proxyError(w http.ResponseWriter, action, target string, err error) {
	s.log.Err("provider cache: %s %s: %s", action, target, err)
	http.Error(w, fmt.Sprintf("provider cache: %s failed", action), http.StatusBadGateway)
}

// artifactURL builds the signed, proxy-local URL Terraform should use to fetch
// an upstream artifact through the cache.
func (s *Server) artifactURL(upstream string) (string, error) {
	if _, err := url.Parse(upstream); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("url", upstream)
	q.Set("sig", s.sign(upstream))
	return s.MirrorBaseURL() + "artifact?" + q.Encode(), nil
}

func (s *Server) sign(v string) string {
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write([]byte(v))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Server) validSignature(v, sig string) bool {
	expected := s.sign(v)
	return hmac.Equal([]byte(expected), []byte(sig))
}

// cachePath is the on-disk location for a cached artifact. The name is a hash of
// the upstream URL, preserving the archive extension so served files keep a
// meaningful suffix.
func (s *Server) cachePath(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	name := hex.EncodeToString(sum[:])
	if ext := path.Ext(pathOnly(rawURL)); ext != "" {
		name += ext
	}
	return filepath.Join(s.cacheDir, name)
}

// allowedArtifactURL reports whether the proxy is willing to fetch the given
// URL. Real registries always hand out https download URLs; http is permitted
// only for loopback addresses so that local test registries work. Combined with
// the HMAC signature check this bounds where the artifact endpoint can reach.
func allowedArtifactURL(u *url.URL) bool {
	switch u.Scheme {
	case "https":
		return true
	case "http":
		host := u.Hostname()
		if host == "localhost" {
			return true
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return true
		}
		return false
	default:
		return false
	}
}

// pathOnly returns just the path component of a URL, used to derive a file
// extension without query-string noise.
func pathOnly(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil {
		return u.Path
	}
	return rawURL
}
