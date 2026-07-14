// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package github

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// If the hostname is github.com, should use normal BaseURL.
func TestNew_GithubCom(t *testing.T) {
	client, err := New("github.com", &UserCredentials{"user", "pass", ""}, Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	Equals(t, "https://api.github.com/", client.client.BaseURL())
}

// If the hostname is a non-github hostname should use the right BaseURL.
func TestNew_NonGithub(t *testing.T) {
	client, err := New("example.com", &UserCredentials{"user", "pass", ""}, Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	Equals(t, "https://example.com/api/v3/", client.client.BaseURL())
	// If possible in the future, test the GraphQL client's URL as well. But at the
	// moment the shurcooL library doesn't expose it.
}

func TestNewSecondaryRateLimitHTTPClientDoesNotApplyPrimaryLimits(t *testing.T) {
	var requests []string
	baseTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.Path)

		switch req.URL.Path {
		case "/api/v3/repos/example/repo":
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Status:     "403 Forbidden",
				Header: http.Header{
					"x-ratelimit-remaining": []string{"0"},
					"x-ratelimit-reset":     []string{strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)},
					"x-ratelimit-resource":  []string{"core"},
				},
				Body:    io.NopCloser(bytes.NewReader(nil)),
				Request: req,
			}, nil
		case "/api/graphql":
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(nil)),
				Request:    req,
			}, nil
		default:
			t.Fatalf("unexpected request path %q", req.URL.Path)
			return nil, nil
		}
	})

	baseClient := &http.Client{
		Transport: baseTransport,
		Timeout:   time.Minute,
	}
	client := newSecondaryRateLimitHTTPClient(baseClient)

	restReq, err := http.NewRequest(http.MethodGet, "https://example.com/api/v3/repos/example/repo", nil)
	Ok(t, err)
	restResp, err := client.Do(restReq)
	Ok(t, err)
	Equals(t, http.StatusForbidden, restResp.StatusCode)
	Ok(t, restResp.Body.Close())

	graphqlReq, err := http.NewRequest(http.MethodPost, "https://example.com/api/graphql", nil)
	Ok(t, err)
	graphqlResp, err := client.Do(graphqlReq)
	Ok(t, err)
	Equals(t, http.StatusOK, graphqlResp.StatusCode)
	Ok(t, graphqlResp.Body.Close())

	Equals(t, time.Minute, client.Timeout)
	Equals(t, []string{"/api/v3/repos/example/repo", "/api/graphql"}, requests)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
