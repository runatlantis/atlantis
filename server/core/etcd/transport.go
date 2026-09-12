// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the internal command transport (design §"Command dispatch
// and local fencing", §506). It carries an immutable command identity, validates
// a mandatory internal token in constant time over HTTPS, limits body and error
// sizes, does not follow cross-origin redirects with credentials, and forwards
// only to allowlisted destinations. VCS credentials are never forwarded.
package etcd

import (
	"bytes"
	"context"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// InternalCommandPath is the path the internal command server listens on.
const InternalCommandPath = "/internal/command"

// maxInternalBodyBytes bounds request and response bodies on the internal
// transport (design §506 "limits body and error sizes").
const maxInternalBodyBytes = 1 << 20 // 1 MiB

// Command is the credential-free envelope forwarded between replicas. It carries
// the immutable identity and pull scope plus opaque command bytes; it never
// carries VCS credentials.
type Command struct {
	Identity   CommandIdentity `json:"identity"`
	Scope      PullScope       `json:"scope"`
	Generation Generation      `json:"generation"`
	Body       json.RawMessage `json:"body,omitempty"`
}

// Result is the internal transport response. Status mirrors the HTTP status the
// owner-aware dispatch contract defines: 202 admitted, 409 stale claim, 503
// unavailable.
type Result struct {
	Status  int            `json:"status"`
	State   AdmissionState `json:"state,omitempty"`
	Message string         `json:"message,omitempty"`
}

// Allowlist matches a destination host against configured hosts and CIDRs to
// prevent SSRF and token disclosure (design §815).
type Allowlist struct {
	hosts map[string]struct{}
	nets  []*net.IPNet
}

// NewAllowlist parses host-or-CIDR entries.
func NewAllowlist(entries []string) (*Allowlist, error) {
	a := &Allowlist{hosts: map[string]struct{}{}}
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if _, ipnet, err := net.ParseCIDR(e); err == nil {
			a.nets = append(a.nets, ipnet)
			continue
		}
		a.hosts[strings.ToLower(e)] = struct{}{}
	}
	return a, nil
}

// Allows reports whether host (a hostname or IP) is permitted as a forwarding
// destination.
func (a *Allowlist) Allows(host string) bool {
	host = strings.ToLower(host)
	if _, ok := a.hosts[host]; ok {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		for _, n := range a.nets {
			if n.Contains(ip) {
				return true
			}
		}
	}
	return false
}

// InternalClient forwards commands to an owning replica's advertise URL.
type InternalClient struct {
	token     string
	allowlist *Allowlist
	http      *http.Client
}

// NewInternalClient builds a forwarding client. tlsCfg is nil only in insecure
// development. The client refuses to follow redirects so credentials are never
// re-sent cross-origin.
func NewInternalClient(token string, allowlist *Allowlist, tlsCfg *tls.Config, timeout time.Duration) *InternalClient {
	return &InternalClient{
		token:     token,
		allowlist: allowlist,
		http: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return errors.New("internal transport does not follow redirects")
			},
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		},
	}
}

// Send forwards cmd to advertiseURL and returns the parsed Result. It enforces
// the destination allowlist before dialing.
func (c *InternalClient) Send(ctx context.Context, advertiseURL string, cmd Command) (Result, error) {
	u, err := url.Parse(advertiseURL)
	if err != nil {
		return Result{}, fmt.Errorf("parsing advertise url: %w", err)
	}
	if c.allowlist != nil && !c.allowlist.Allows(u.Hostname()) {
		return Result{}, fmt.Errorf("forwarding destination %q is not allowlisted", u.Hostname())
	}

	body, err := json.Marshal(cmd)
	if err != nil {
		return Result{}, err
	}
	endpoint := strings.TrimRight(advertiseURL, "/") + InternalCommandPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("forwarding command: %w", err)
	}
	defer resp.Body.Close()

	var result Result
	dec := json.NewDecoder(io.LimitReader(resp.Body, maxInternalBodyBytes))
	if err := dec.Decode(&result); err != nil {
		// Fall back to the HTTP status if the body is unparseable.
		return Result{Status: resp.StatusCode, Message: "unparseable internal response"}, nil
	}
	if result.Status == 0 {
		result.Status = resp.StatusCode
	}
	return result, nil
}

// ForwardedHandler receives a validated forwarded command and returns the
// result. Implemented by the router (HandleForwarded).
type ForwardedHandler interface {
	HandleForwarded(ctx context.Context, cmd Command) Result
}

// InternalServer is the HTTP handler for forwarded commands. It authenticates
// the internal token in constant time and enforces the body size limit before
// dispatching.
type InternalServer struct {
	token   string
	handler ForwardedHandler
}

// NewInternalServer builds the internal command handler.
func NewInternalServer(token string, handler ForwardedHandler) *InternalServer {
	return &InternalServer{token: token, handler: handler}
}

func (s *InternalServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResult(w, Result{Status: http.StatusMethodNotAllowed, Message: "method not allowed"})
		return
	}
	if !s.authenticate(r) {
		// Do not reveal whether the token was present or malformed.
		writeResult(w, Result{Status: http.StatusUnauthorized, Message: "unauthorized"})
		return
	}

	var cmd Command
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxInternalBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cmd); err != nil {
		writeResult(w, Result{Status: http.StatusBadRequest, Message: "invalid command envelope"})
		return
	}

	result := s.handler.HandleForwarded(r.Context(), cmd)
	writeResult(w, result)
}

// authenticate compares the bearer token in constant time (design §811). The
// header is trimmed and the "Bearer" scheme matched case-insensitively with an
// optional space, so an empty token (development only) is not mangled by HTTP
// header whitespace trimming.
func (s *InternalServer) authenticate(r *http.Request) bool {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	const scheme = "bearer"
	if len(h) < len(scheme) || !strings.EqualFold(h[:len(scheme)], scheme) {
		return false
	}
	got := strings.TrimSpace(h[len(scheme):])
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) == 1
}

func writeResult(w http.ResponseWriter, result Result) {
	if result.Status == 0 {
		result.Status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Status)
	_ = json.NewEncoder(w).Encode(result)
}
