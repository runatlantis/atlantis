// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/github/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestAppCredentials_GetUserJWT(t *testing.T) {
	for _, host := range []string{"github.com", "github.example.com", "tenant.ghe.com"} {
		for _, slug := range []string{"", "configured-slug"} {
			// A configured slug on GitHub.com/GHES retains the existing
			// installation-token lookup, covered by TestClient_GetUser_AppSlug.
			if slug != "" && !isGHECloud(host) {
				continue
			}
			for _, response := range []struct {
				name    string
				status  int
				body    string
				wantErr bool
			}{
				{"success", http.StatusOK, `{"slug":"actual-app"}`, false},
				{"unauthorized", http.StatusUnauthorized, `{"message":"bad credentials"}`, true},
				{"missing slug", http.StatusOK, `{}`, true},
			} {
				t.Run(host+"/"+slug+"/"+response.name, func(t *testing.T) {
					previous := http.DefaultTransport
					t.Cleanup(func() { http.DefaultTransport = previous })
					http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
						path := "/app"
						if host == "github.example.com" {
							path = "/api/v3/app"
						}
						Equals(t, path, req.URL.Path)
						apiHost := host
						if host == "github.com" || isGHECloud(host) {
							apiHost = "api." + host
						}
						Equals(t, apiHost, req.URL.Host)
						Assert(t, strings.HasPrefix(req.Header.Get("Authorization"), "Bearer "), "app lookup must use JWT authentication")
						return &http.Response{StatusCode: response.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response.body)), Request: req}, nil
					})
					creds := &AppCredentials{AppID: 1, InstallationID: 1, Key: []byte(testdata.PrivateKey), Hostname: host, AppSlug: slug}
					user, err := creds.GetUser()
					if response.wantErr {
						Assert(t, err != nil, "expected app identity lookup error")
						Equals(t, "", user)
						if slug == "" {
							client, err := New(host, creds, Config{}, 0, logging.NewNoopLogger(t))
							Ok(t, err)
							self, err := client.IsAtlantisUser(models.Repo{}, models.User{Username: "actual-app[bot]"})
							Ok(t, err)
							Equals(t, false, self)
						}
					} else {
						Ok(t, err)
						Equals(t, "actual-app[bot]", user)
					}
				})
			}
		}
	}
}
