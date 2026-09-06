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
// # Trust boundaries
//
// The proxy makes outbound requests, so it is careful about what it will reach:
//   - The registry host in every request path must be one of the operator-
//     configured registries (allowedRegistry); requests for any other host are
//     refused. So the registry it talks to is never attacker-controlled.
//   - Every other path segment (provider namespace, type, version, os, arch)
//     is validated against a strict character allowlist before it is used to
//     build an upstream URL, so it cannot alter the request target.
//   - The archive / checksum / signature files are fetched from the URLs the
//     trusted registry itself returns in its download-metadata response, never
//     from a URL supplied in the incoming request. Terraform only ever receives
//     coordinate-based artifact URLs from this proxy.
//
// This mirrors the provider cache server that Terragrunt already ships.
package providercache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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

	// metadataTimeout bounds service-discovery and metadata requests. Archive
	// downloads are deliberately not bounded by this because they can be large.
	metadataTimeout = 30 * time.Second
)

// segmentPattern is the allowlist a provider coordinate path segment (namespace,
// type, version, os, arch) must match before it is used to build an upstream
// URL. It permits only unreserved provider-address characters and cannot
// express a path-traversal or host-manipulation sequence.
var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// artifactKinds maps the artifact path suffix to the field of the registry
// download-metadata response that holds its real URL.
var artifactKinds = map[string]string{
	"archive":   "download_url",
	"shasums":   "shasums_url",
	"signature": "shasums_signature_url",
}

// Server is the caching provider proxy. Create it with New and start it with
// Start.
type Server struct {
	log logging.SimpleLogging

	// cacheDir is where downloaded artifacts (archives, SHA256SUMS, signatures)
	// are stored on disk, keyed by a hash of their upstream URL.
	cacheDir string

	// registries is the ordered set of registry hostnames (e.g.
	// "registry.terraform.io") that this proxy will serve. A CLI-config host
	// block is generated for each.
	registries []string

	metadataClient *http.Client
	// downloadClient has no overall timeout because provider archives can be
	// hundreds of megabytes; the transport still bounds connect/idle time.
	downloadClient *http.Client

	// disco caches the resolved providers.v1 base URL per registry host.
	discoMu sync.Mutex
	disco   map[string]string

	// meta caches the download-metadata response per provider coordinate, so the
	// three artifact fetches (archive, shasums, signature) for one platform share
	// a single upstream metadata request.
	metaMu sync.Mutex
	meta   map[string]map[string]any

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

	listener, err := net.Listen("tcp", net.JoinHostPort(loopbackHost, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, fmt.Errorf("listening on %s: %w", loopbackHost, err)
	}

	s := &Server{
		log:            log,
		cacheDir:       cacheDir,
		registries:     registries,
		metadataClient: &http.Client{Timeout: metadataTimeout},
		downloadClient: &http.Client{},
		disco:          make(map[string]string),
		meta:           make(map[string]map[string]any),
		listener:       listener,
	}

	router := mux.NewRouter()
	// Terraform Provider Registry Protocol endpoints, namespaced by the registry
	// host so a single proxy can serve multiple registries.
	router.HandleFunc("/{host}/v1/providers/{namespace}/{type}/versions", s.handleVersions).Methods(http.MethodGet)
	router.HandleFunc("/{host}/v1/providers/{namespace}/{type}/{version}/download/{os}/{arch}", s.handleDownload).Methods(http.MethodGet)
	// The coordinate-addressed artifact endpoint that actually caches. It carries
	// no URL: the proxy re-resolves the real download location from the trusted
	// registry's metadata response.
	router.HandleFunc("/artifact/{host}/{namespace}/{type}/{version}/{os}/{arch}/{kind}", s.handleArtifact).Methods(http.MethodGet)
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

// allowedRegistry reports whether host is one of the configured registries and,
// if so, returns the trusted, configured spelling of it. Returning the value
// from the configured list (rather than the request-derived string) bounds all
// outbound requests to registries the operator explicitly enabled.
func (s *Server) allowedRegistry(host string) (string, bool) {
	for _, r := range s.registries {
		if r == host {
			return r, true
		}
	}
	return "", false
}

// coordinates validates and returns the provider path segments common to the
// download and artifact endpoints. ok is false (and an error already written) if
// the host is not an allowed registry or any segment is malformed.
func (s *Server) coordinates(w http.ResponseWriter, vars map[string]string, keys ...string) (host string, segs map[string]string, ok bool) {
	host, ok = s.allowedRegistry(vars["host"])
	if !ok {
		http.Error(w, "unknown registry host", http.StatusNotFound)
		return "", nil, false
	}
	segs = make(map[string]string, len(keys))
	for _, k := range keys {
		v := vars[k]
		if !segmentPattern.MatchString(v) {
			http.Error(w, "invalid provider "+k, http.StatusBadRequest)
			return "", nil, false
		}
		segs[k] = v
	}
	return host, segs, true
}

// handleVersions proxies the "list available versions" registry endpoint. The
// response contains no URLs, so it is passed straight through.
func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	host, segs, ok := s.coordinates(w, mux.Vars(r), "namespace", "type")
	if !ok {
		return
	}
	base, err := s.discover(r.Context(), host)
	if err != nil {
		s.proxyError(w, "service discovery", host, err)
		return
	}
	upstream := base + segs["namespace"] + "/" + segs["type"] + "/versions"
	s.pipe(w, r, upstream)
}

// handleDownload proxies the "find a provider package" registry endpoint,
// rewriting the archive/checksum/signature URLs so Terraform fetches them back
// through the caching artifact endpoint. Every other field (signing_keys,
// shasum, filename, protocols, ...) reaches Terraform unchanged.
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, segs, ok := s.coordinates(w, vars, "namespace", "type", "version", "os", "arch")
	if !ok {
		return
	}

	body, status, err := s.downloadMetadata(r.Context(), host, segs)
	if err != nil {
		s.proxyError(w, "download metadata", host, err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, "registry returned status "+fmt.Sprint(status), status)
		return
	}

	// Rewrite the three externally-hosted file URLs to coordinate-addressed
	// artifact endpoints. Terraform never sees the real upstream URL. Copy first
	// so the cached metadata (which handleArtifact reads to find the real URLs)
	// keeps its original values.
	out := make(map[string]any, len(body))
	maps.Copy(out, body)
	for suffix, field := range artifactKinds {
		if orig, ok := out[field].(string); ok && orig != "" {
			out[field] = s.artifactURL(host, segs, suffix)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		s.log.Err("provider cache: writing download response: %s", err)
	}
}

// handleArtifact serves (and caches on first request) the archive, checksum or
// signature file for a provider coordinate. The real download location is taken
// from the trusted registry's metadata response, not from the incoming request.
func (s *Server) handleArtifact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, segs, ok := s.coordinates(w, vars, "namespace", "type", "version", "os", "arch")
	if !ok {
		return
	}
	field, ok := artifactKinds[vars["kind"]]
	if !ok {
		http.Error(w, "unknown artifact kind", http.StatusNotFound)
		return
	}

	body, status, err := s.downloadMetadata(r.Context(), host, segs)
	if err != nil {
		s.proxyError(w, "download metadata", host, err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, "registry returned status "+fmt.Sprint(status), status)
		return
	}
	// target originates from the registry's response, so it is not attacker-
	// controlled data from the incoming request.
	target, _ := body[field].(string)
	if target == "" {
		http.Error(w, "artifact not available", http.StatusNotFound)
		return
	}
	if parsed, perr := url.Parse(target); perr != nil || !allowedArtifactURL(parsed) {
		s.proxyError(w, "resolving artifact", host, fmt.Errorf("registry returned unusable url"))
		return
	}

	cachePath, err := s.ensureCached(r.Context(), target)
	if err != nil {
		s.proxyError(w, "caching artifact", host, err)
		return
	}

	switch vars["kind"] {
	case "archive":
		w.Header().Set("Content-Type", "application/zip")
	case "signature":
		w.Header().Set("Content-Type", "application/octet-stream")
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	// #nosec G703 -- cachePath is filepath.Join(cacheDir, hex-sha256(url)); the file name is a fixed-length [0-9a-f] hash with no separators, so it cannot escape cacheDir.
	http.ServeFile(w, r, cachePath)
}

// artifactURL builds the coordinate-addressed, proxy-local URL Terraform should
// use to fetch an artifact of the given kind through the cache.
func (s *Server) artifactURL(host string, segs map[string]string, kind string) string {
	return s.MirrorBaseURL() + "artifact/" +
		url.PathEscape(host) + "/" +
		url.PathEscape(segs["namespace"]) + "/" +
		url.PathEscape(segs["type"]) + "/" +
		url.PathEscape(segs["version"]) + "/" +
		url.PathEscape(segs["os"]) + "/" +
		url.PathEscape(segs["arch"]) + "/" +
		kind
}

// downloadMetadata fetches (and caches per coordinate) the registry's "find a
// provider package" response. host is a trusted registry and every segment has
// been validated, so the upstream URL target is not attacker-controlled.
func (s *Server) downloadMetadata(ctx context.Context, host string, segs map[string]string) (map[string]any, int, error) {
	key := host + "/" + segs["namespace"] + "/" + segs["type"] + "/" + segs["version"] + "/" + segs["os"] + "/" + segs["arch"]

	s.metaMu.Lock()
	cached, ok := s.meta[key]
	s.metaMu.Unlock()
	if ok {
		return cached, http.StatusOK, nil
	}

	base, err := s.discover(ctx, host)
	if err != nil {
		return nil, 0, err
	}
	upstream := base +
		segs["namespace"] + "/" +
		segs["type"] + "/" +
		segs["version"] + "/download/" +
		segs["os"] + "/" +
		segs["arch"]

	// #nosec G704 -- host is an operator-configured registry (allowedRegistry) and every path segment is regexp-validated (segmentPattern), so the request target is not attacker-controlled.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstream, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := s.metadataClient.Do(req) // #nosec G704 -- see above; upstream host is trusted and segments are validated.
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, 0, err
	}

	s.metaMu.Lock()
	s.meta[key] = body
	s.metaMu.Unlock()
	return body, http.StatusOK, nil
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
// a partial download can never be observed as a complete cache entry. rawURL is
// a location returned by a trusted registry (see handleArtifact), not a value
// from the incoming request.
func (s *Server) download(ctx context.Context, rawURL, cachePath string) error {
	// #nosec G704 -- rawURL is a location returned by a trusted registry's download-metadata response (see handleArtifact), not a value from the incoming request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := s.downloadClient.Do(req) // #nosec G704 -- see above; the URL comes from a trusted registry response and is scheme-checked.
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
	// #nosec G704 -- host is an operator-configured registry (allowedRegistry), not attacker-controlled.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoURL, nil)
	if err != nil {
		return fallback
	}
	resp, err := s.metadataClient.Do(req) // #nosec G704 -- see above; discovery host is a trusted, configured registry.
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
// back to the client unchanged. upstream is built from a trusted registry host
// and validated path segments.
func (s *Server) pipe(w http.ResponseWriter, r *http.Request, upstream string) {
	// #nosec G704 -- upstream is built from a trusted, configured registry host and regexp-validated path segments (see handleVersions), so it is not attacker-controlled.
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream, nil)
	if err != nil {
		s.proxyError(w, "building request", upstream, err)
		return
	}
	resp, err := s.metadataClient.Do(req) // #nosec G704 -- see above; upstream host is trusted and segments are validated.
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

func (s *Server) proxyError(w http.ResponseWriter, action, target string, err error) {
	s.log.Err("provider cache: %s %s: %s", action, target, err)
	http.Error(w, fmt.Sprintf("provider cache: %s failed", action), http.StatusBadGateway)
}

// cachePath is the on-disk location for a cached artifact. The file name is
// solely the hex SHA-256 of the upstream URL: a fixed-length string over
// [0-9a-f] with no separators or dots, so it cannot contain path-traversal
// sequences regardless of the URL. filepath.Base is applied as a defensive
// barrier to guarantee the name stays a single path element.
func (s *Server) cachePath(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	name := filepath.Base(hex.EncodeToString(sum[:]))
	return filepath.Join(s.cacheDir, name)
}

// allowedArtifactURL reports whether the proxy is willing to fetch the given
// URL. Real registries always hand out https download URLs; http is permitted
// only for loopback addresses so that local test registries work.
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
