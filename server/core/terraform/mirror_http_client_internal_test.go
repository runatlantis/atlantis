// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	. "github.com/runatlantis/atlantis/testing"
)

type mirrorTestTransport func(*http.Request) (*http.Response, error)

func (transport mirrorTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

func mirrorTestClient(t *testing.T, origin string, auth APIAuth, transport mirrorTestTransport) *http.Client {
	t.Helper()
	client := newMirrorHTTPClient(origin, auth)
	retryClient := client.Transport.(*retryablehttp.RoundTripper).Client
	retryClient.RetryWaitMin = 0
	retryClient.RetryWaitMax = 0
	retryClient.RetryMax = 1
	switch authTransport := retryClient.HTTPClient.Transport.(type) {
	case *bearerTokenRoundTripper:
		authTransport.inner = transport
	case *basicAuthRoundTripper:
		authTransport.inner = transport
	default:
		t.Fatal("mirror auth must wrap the inner retry client's transport")
	}
	return client
}

func mirrorTestResponse(req *http.Request, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("response")),
		Request:    req,
	}
}

func TestMirrorHTTPClient_RedirectOrigins(t *testing.T) {
	for _, auth := range []struct {
		name   string
		config APIAuth
		header string
	}{
		{"bearer", APIAuth{BearerToken: "mirror-token"}, "Bearer mirror-token"},
		{"basic", APIAuth{Username: "user", Password: "pass"}, "Basic dXNlcjpwYXNz"},
		{"token_precedence", APIAuth{BearerToken: "mirror-token", Username: "user", Password: "pass"}, "Bearer mirror-token"},
	} {
		for _, c := range []struct {
			name       string
			origin     string
			target     string
			authorized bool
		}{
			{"same_origin", "https://mirror.example:8443", "https://mirror.example:8443", true},
			{"scheme", "https://mirror.example:8443", "http://mirror.example:8443", false},
			{"hostname", "https://mirror.example:8443", "https://other.example:8443", false},
			{"subdomain", "https://mirror.example:8443", "https://cdn.mirror.example:8443", false},
			{"port", "https://mirror.example:8443", "https://mirror.example:9443", false},
			{"default_https_port", "https://mirror.example", "https://mirror.example:443", true},
			{"default_http_port", "http://mirror.example", "http://mirror.example:80", true},
			{"hostname_case", "https://MIRROR.example:443", "https://mirror.example", true},
		} {
			t.Run(auth.name+"/"+c.name, func(t *testing.T) {
				var headers []string
				client := mirrorTestClient(t, c.origin, auth.config, func(req *http.Request) (*http.Response, error) {
					headers = append(headers, req.Header.Get("Authorization"))
					if req.URL.Path == "/start" {
						response := mirrorTestResponse(req, http.StatusFound)
						response.Header.Set("Location", c.target+"/finish")
						return response, nil
					}
					return mirrorTestResponse(req, http.StatusOK), nil
				})
				req, err := http.NewRequest(http.MethodGet, c.origin+"/start", nil)
				Ok(t, err)
				response, err := client.Do(req)
				Ok(t, err)
				Ok(t, response.Body.Close())
				expected := ""
				if c.authorized {
					expected = auth.header
				}
				Equals(t, []string{auth.header, expected}, headers)
				Equals(t, "", req.Header.Get("Authorization"))
			})
		}
	}
}

func TestMirrorHTTPClient_RetriesKeepAuth(t *testing.T) {
	for _, auth := range []APIAuth{{BearerToken: "mirror-token"}, {Username: "user", Password: "pass"}} {
		t.Run(auth.BearerToken+auth.Username, func(t *testing.T) {
			var headers []string
			client := mirrorTestClient(t, "https://mirror.example", auth, func(req *http.Request) (*http.Response, error) {
				headers = append(headers, req.Header.Get("Authorization"))
				status := http.StatusOK
				if len(headers) == 1 {
					status = http.StatusServiceUnavailable
				}
				return mirrorTestResponse(req, status), nil
			})
			response, err := client.Get("https://mirror.example/index.json")
			Ok(t, err)
			Ok(t, response.Body.Close())
			Equals(t, 2, len(headers))
			Assert(t, headers[0] != "", "retry requests must be authenticated")
			Equals(t, headers[0], headers[1])
		})
	}
}

func TestMirrorHTTPClient_NoCustomClientWithoutCredentials(t *testing.T) {
	Equals(t, (*http.Client)(nil), newMirrorHTTPClient("https://mirror.example", APIAuth{}))
	Equals(t, (*http.Client)(nil), newMirrorHTTPClient("https://%", APIAuth{BearerToken: "token"}))
}

func TestSelectDistribution_ReusesConfiguredInstances(t *testing.T) {
	for _, defaultName := range []string{"terraform", "opentofu"} {
		t.Run(defaultName, func(t *testing.T) {
			configured := NewDistribution(defaultName, "https://mirror.example", APIAuth{BearerToken: "token"})
			Equals(t, configured, SelectDistribution(configured, nil))
			for _, override := range []string{"terraform", "opentofu"} {
				first := SelectDistribution(configured, &override)
				Assert(t, first == SelectDistribution(configured, &override), "project selection must reuse the configured instance")
			}
		})
	}
}
